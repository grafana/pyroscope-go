package compat

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var exportFileCache sync.Map

func goListExport(pkg string) (string, error) {
	if v, ok := exportFileCache.Load(pkg); ok {
		return v.(string), nil
	}
	out, err := exec.Command("go", "list", "-export", "-f", "{{.Export}}", pkg).Output()
	if err != nil {
		return "", fmt.Errorf("go list -export %s: %w", pkg, err)
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", fmt.Errorf("no export file for %s", pkg)
	}
	exportFileCache.Store(pkg, path)
	return path, nil
}

func checkSignature(t *testing.T, pkg string, name string, expectedSignature string) {
	t.Helper()

	// Use go list -json to find source files for the package.
	// We need source (not just export data) because the functions
	// we're checking are unexported.
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	require.NoError(t, err, "go list -json failed for %s", pkg)

	var pkgInfo struct {
		Dir     string
		GoFiles []string
	}
	require.NoError(t, json.Unmarshal(out, &pkgInfo))

	fset := token.NewFileSet()
	var files []*ast.File
	for _, fname := range pkgInfo.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkgInfo.Dir, fname), nil, 0)
		require.NoError(t, err)
		files = append(files, f)
	}

	// Type-check the package from source.
	// Dependencies are loaded from export data via go list -export.
	conf := types.Config{
		Importer: importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
			exportFile, err := goListExport(path)
			if err != nil {
				return nil, err
			}
			return os.Open(exportFile)
		}),
	}
	p, err := conf.Check(pkg, fset, files, nil)
	require.NoError(t, err, "type-check failed for %s", pkg)

	f := p.Scope().Lookup(name)
	require.NotNilf(t, f, "function %s not found in %s", name, pkg)
	ff, ok := f.(*types.Func)
	require.True(t, ok)
	assert.Equal(t, expectedSignature, ff.String())
}
