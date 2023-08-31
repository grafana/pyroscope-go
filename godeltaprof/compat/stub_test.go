package compat

import (
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestSignatureExpandFinalInlineFrame(t *testing.T) {
	checkSignature(t, "github.com/grafana/pyroscope-go/godeltaprof/internal/pprof",
		"runtime_expandFinalInlineFrame",
		"func github.com/grafana/pyroscope-go/godeltaprof/internal/pprof.runtime_expandFinalInlineFrame(stk []uintptr) []uintptr")
}

func TestSignatureCyclesPerSecond(t *testing.T) {
	checkSignature(t, "github.com/grafana/pyroscope-go/godeltaprof/internal/pprof",
		"runtime_cyclesPerSecond",
		"func github.com/grafana/pyroscope-go/godeltaprof/internal/pprof.runtime_cyclesPerSecond() int64")
}

func TestSignatureCyclesPerSecondRuntime(t *testing.T) {
	checkSignature(t, "runtime/pprof",
		"runtime_cyclesPerSecond",
		"func runtime/pprof.runtime_cyclesPerSecond() int64")
}

func TestSignatureExpandFinalInlineFrameRuntime(t *testing.T) {
	checkSignature(t, "runtime/pprof",
		"runtime_expandFinalInlineFrame",
		"func runtime/pprof.runtime_expandFinalInlineFrame(stk []uintptr) []uintptr")
}

func checkSignature(t *testing.T, pkg string, name string, expectedSignature string) {
	cfg := &packages.Config{
		Mode:  packages.NeedImports | packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, pkg)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range pkgs {
		if strings.Contains(p.ID, ".test") {
			continue
		}
		f := p.Types.Scope().Lookup(name)
		if f != nil {
			found = true
			ff := f.(*types.Func)
			if ff.String() != expectedSignature {
				t.Fatalf("expected %s, got %s", expectedSignature, ff.String())
			}
		}
	}
	if !found {
		t.Fatalf("function %s %s not found", pkg, name)
	}
}
