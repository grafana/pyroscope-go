//go:build go1.21
// +build go1.21

package compat

import (
	"testing"
)

func TestRuntimeFrameSymbolName(t *testing.T) {
	checkSignature(t, "runtime/pprof",
		"runtime_FrameSymbolName",
		"func runtime/pprof.runtime_FrameSymbolName(f *runtime.Frame) string")
	checkSignature(t, "github.com/grafana/pyroscope-go/godeltaprof/internal/pprof",
		"runtime_FrameSymbolName",
		"func github.com/grafana/pyroscope-go/godeltaprof/internal/pprof.runtime_FrameSymbolName(f *runtime.Frame) string")
}

func TestRuntimeFrameStartLine(t *testing.T) {
	checkSignature(t, "runtime/pprof",
		"runtime_FrameStartLine",
		"func runtime/pprof.runtime_FrameStartLine(f *runtime.Frame) int")
	checkSignature(t, "github.com/grafana/pyroscope-go/godeltaprof/internal/pprof",
		"runtime_FrameStartLine",
		"func github.com/grafana/pyroscope-go/godeltaprof/internal/pprof.runtime_FrameStartLine(f *runtime.Frame) int")
}
