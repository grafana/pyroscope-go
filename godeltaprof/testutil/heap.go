package testutil

import (
	"io"
	"runtime"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

type HeapProfiler interface {
	WriteHeapProto(w io.Writer, p []runtime.MemProfileRecord, rate int64, defaultSampleType string) error
}

type heapProfiler struct {
	impl pprof.DeltaHeapProfiler
}

func (h *heapProfiler) WriteHeapProto(w io.Writer, p []runtime.MemProfileRecord, rate int64, defaultSampleType string) error {
	return h.impl.WriteHeapProto(w, p, rate, defaultSampleType)
}

func NewHeapProfiler(generics bool, lazyMapping bool) HeapProfiler {
	return &heapProfiler{
		impl: pprof.DeltaHeapProfiler{
			Options: pprof.ProfileBuilderOptions{
				GenericsFrames: generics,
				LazyMapping:    lazyMapping,
			},
		},
	}
}
