package pprof

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"
)

// TestRuntimeProfileRecordLayout verifies that the locally declared
// MemProfileRecord and BlockProfileRecord exactly match the layout of
// internal/profilerecord.{Mem,Block}ProfileRecord in the active Go toolchain.
// The runtime writes into our types via //go:linkname pprof_*ProfileInternal,
// so any drift in field count, name, type, order, offset or struct size
// silently corrupts the captured profile.
func TestRuntimeProfileRecordLayout(t *testing.T) {
	src := findRuntimeProfileRecordSource(t)
	pkg := loadProfileRecordPackage(t, src)

	cases := []struct {
		name string
		ours reflect.Type
	}{
		{"MemProfileRecord", reflect.TypeOf(MemProfileRecord{})},
		{"BlockProfileRecord", reflect.TypeOf(BlockProfileRecord{})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			theirs := lookupStruct(t, pkg, c.name)
			compareLayout(t, c.name, theirs, c.ours)
		})
	}
}

func findRuntimeProfileRecordSource(t *testing.T) string {
	t.Helper()
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		var stdout bytes.Buffer
		cmd := exec.Command("go", "env", "GOROOT")
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			t.Fatalf("go env GOROOT: %v", err)
		}
		goroot = string(bytes.TrimSpace(stdout.Bytes()))
	}
	if goroot == "" {
		t.Fatal("GOROOT not set; cannot locate runtime profilerecord source")
	}
	candidates := []string{
		filepath.Join(goroot, "src", "internal", "profilerecord", "profilerecord.go"),
		filepath.Join(goroot, "src", "runtime", "internal", "profilerecord", "profilerecord.go"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Fatalf("runtime profilerecord source not found in any of: %v", candidates)

	return ""
}

func loadProfileRecordPackage(t *testing.T, src string) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, src, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", src, err)
	}
	conf := &types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("profilerecord", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("type-check %s: %v", src, err)
	}

	return pkg
}

func lookupStruct(t *testing.T, pkg *types.Package, name string) *types.Struct {
	t.Helper()
	obj := pkg.Scope().Lookup(name)
	if obj == nil {
		t.Fatalf("type %s not found in %s", name, pkg.Path())
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		t.Fatalf("%s is not a type name (got %T)", name, obj)
	}
	st, ok := tn.Type().Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("%s underlying is %T, want *types.Struct", name, tn.Type().Underlying())
	}

	return st
}

func compareLayout(t *testing.T, name string, theirs *types.Struct, ours reflect.Type) {
	t.Helper()
	if ours.Kind() != reflect.Struct {
		t.Fatalf("%s: ours is %s, want struct", name, ours.Kind())
	}

	if theirs.NumFields() != ours.NumField() {
		t.Fatalf("%s: field count mismatch: runtime=%d ours=%d",
			name, theirs.NumFields(), ours.NumField())
	}

	ptrSize := int64(unsafe.Sizeof(uintptr(0)))
	sizes := &types.StdSizes{WordSize: ptrSize, MaxAlign: ptrSize}

	theirFields := make([]*types.Var, theirs.NumFields())
	for i := range theirs.NumFields() {
		theirFields[i] = theirs.Field(i)
	}
	theirOffsets := sizes.Offsetsof(theirFields)

	for i := range theirs.NumFields() {
		their := theirs.Field(i)
		our := ours.Field(i)

		if their.Name() != our.Name {
			t.Errorf("%s field %d: name mismatch: runtime=%q ours=%q",
				name, i, their.Name(), our.Name)
		}

		theirType := their.Type().String()
		ourType := our.Type.String()
		if theirType != ourType {
			t.Errorf("%s field %q: type mismatch: runtime=%s ours=%s",
				name, their.Name(), theirType, ourType)
		}

		theirSize := sizes.Sizeof(their.Type())
		ourSize := int64(our.Type.Size())
		if theirSize != ourSize {
			t.Errorf("%s field %q: size mismatch: runtime=%d ours=%d",
				name, their.Name(), theirSize, ourSize)
		}

		theirOffset := theirOffsets[i]
		ourOffset := int64(our.Offset)
		if theirOffset != ourOffset {
			t.Errorf("%s field %q: offset mismatch: runtime=%d ours=%d",
				name, their.Name(), theirOffset, ourOffset)
		}
	}

	theirSize := sizes.Sizeof(theirs)
	ourSize := int64(ours.Size())
	if theirSize != ourSize {
		t.Errorf("%s: struct size mismatch: runtime=%d ours=%d", name, theirSize, ourSize)
	}
}
