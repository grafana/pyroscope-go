package compat

import (
	"bytes"
	"runtime"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/godeltaprof"
)

// deepRecurse recurses depth-1 times before invoking fn, producing a stack
// chain of `depth` frames inside this function plus the test/runtime frames.
//
//go:noinline
func deepRecurse(depth int, fn func()) {
	if depth <= 1 {
		fn()

		return
	}
	deepRecurse(depth-1, fn)
}

// TestDeepStackHeap verifies that godeltaprof preserves stack frames beyond
// the historical 32-frame Stack0 limit. On Go 1.23+ the runtime captures up
// to GODEBUG=profstackdepth frames (default 128); godeltaprof should now
// surface all of them in the heap profile.
func TestDeepStackHeap(t *testing.T) {
	const recursionDepth = 64

	prevRate := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() { runtime.MemProfileRate = prevRate }()

	hp := godeltaprof.NewHeapProfiler()

	var buf bytes.Buffer
	require.NoError(t, hp.Profile(&buf)) // prime the delta baseline
	buf.Reset()

	var sink [][]byte
	deepRecurse(recursionDepth, func() {
		sink = append(sink, make([]byte, 64*1024))
	})
	runtime.GC()

	require.NoError(t, hp.Profile(&buf))

	profile, err := gprofile.ParseData(buf.Bytes())
	require.NoError(t, err)

	maxFrames := 0
	for _, s := range profile.Sample {
		frames := 0
		for _, loc := range s.Location {
			if len(loc.Line) > 0 {
				frames += len(loc.Line)
			} else {
				frames++
			}
		}
		if frames > maxFrames {
			maxFrames = frames
		}
	}

	if maxFrames <= 32 {
		t.Fatalf("expected at least one sample with >32 frames, got max=%d", maxFrames)
	}
	_ = sink
}
