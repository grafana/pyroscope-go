package compat

import (
	"io"
	"math/bits"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/godeltaprof"
)

//nolint:gochecknoglobals
var (
	// profiledAllocation keeps the allocation of allocateAt reachable.
	profiledAllocation []byte
	// callSites makes the four calls in allocateAt return to different code.
	callSites [4]int
)

// allocateAt allocates from a call stack that is unique to path: every level
// returns to one of four call sites, picked by two bits of path.
//
//go:noinline
func allocateAt(depth int, path uint64) {
	if depth == 0 {
		profiledAllocation = make([]byte, 64)

		return
	}
	next := bits.RotateLeft64(path, 2)
	switch path & 3 {
	case 0:
		allocateAt(depth-1, next)
		callSites[0]++
	case 1:
		allocateAt(depth-1, next)
		callSites[1]++
	case 2:
		allocateAt(depth-1, next)
		callSites[2]++
	default:
		allocateAt(depth-1, next)
		callSites[3]++
	}
}

func liveHeapBytes() int64 {
	runtime.GC()
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return int64(stats.HeapAlloc) //nolint:gosec
}

// TestHeapProfilerKeepsNoStackCopy verifies that the delta heap profiler does
// not hold its own copy of every profiled stack. The runtime keeps each stack
// in a profile bucket for the life of the process, and the profiler has one
// map entry per bucket, so a copy doubles the memory that profiling costs.
func TestHeapProfilerKeepsNoStackCopy(t *testing.T) {
	const (
		stacks = 20_000
		// With the frames of the test below it, a stack has about 95 frames.
		depth = 90
	)

	prevRate := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() { runtime.MemProfileRate = prevRate }()

	for i := range uint64(stacks) {
		// The odd multiplier gives each index a path of its own.
		allocateAt(depth, (i+1)*0x9E3779B97F4A7C15)
	}
	profiledAllocation = nil
	// The runtime publishes a bucket two collections after its allocation.
	runtime.GC()
	runtime.GC()
	// The profiler's own allocations must not add buckets.
	runtime.MemProfileRate = prevRate

	before := liveHeapBytes()
	hp := godeltaprof.NewHeapProfiler()
	require.NoError(t, hp.Profile(io.Discard))
	retained := liveHeapBytes() - before
	runtime.KeepAlive(hp)

	// A copy of a 95-frame stack is 760 bytes. A map entry that points at the
	// runtime's bucket is about 100.
	require.Less(t, retained/stacks, int64(300), "heap bytes kept per profiled stack")
}
