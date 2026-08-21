package inlineexpansion

import (
	"bytes"
	"io"
	"runtime"
	"slices"
	"testing"
	"time"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/grafana/pyroscope-go/godeltaprof/compat"
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

// Since Go 1.23 the runtime records block/mutex contention stacks in two
// different shapes (runtime/mprof.go saveblockevent):
//
//   - frame pointer unwinding, the common case, stores *physical* return
//     addresses: one PC per physical frame, inlined callers are not expanded
//     (runtime.fpTracebackPartialExpand);
//   - a full traceback, used when GODEBUG=tracefpunwindoff=1 or when the M has
//     cgo on the stack (runtime.m.hasCgoOnStack), stores *logical* PCs: one PC
//     per logical frame, including the virtual PCs the compiler emits for
//     inlined calls.
//
// runtime/pprof normalizes every record with expandInlinedFrames before handing
// it to the profile builder (printCountCycleProfile). godeltaprof does not: it
// passes the raw record stack to profileBuilder.LocsForStack, which assumes
// logical PCs. That has two consequences, one per test below.

// dumpMutexProfile calls profiler.Profile and reports a panic as a test failure
// instead of taking the whole test binary down. pyroscope-go recovers the same
// panic in Session.dumpMutexProfile, which is why the bug is quiet in
// production: the dump is simply discarded.
func dumpMutexProfile(t *testing.T, profiler *godeltaprof.BlockProfiler, w io.Writer) (err error) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("godeltaprof panicked building the mutex profile: %v", r)
		}
	}()

	return profiler.Profile(w)
}

// captureLogicalStack returns the fully expanded stack of reproLeaf, i.e. the
// shape the runtime records for a contention event when it cannot use frame
// pointer unwinding. It is captured on a fresh goroutine to keep the stack as
// shallow as a real worker goroutine stack.
func captureLogicalStack(t *testing.T) []uintptr {
	t.Helper()

	pcs := make([]uintptr, 64)
	var n int
	done := make(chan struct{})
	go func() {
		n = reproEntry(0, pcs)
		close(done)
	}()
	<-done
	require.NotZero(t, n)

	return pcs[:n]
}

// collapseInlineGroups converts a logical stack into the physical stack frame
// pointer unwinding records for the same goroutine: every group of logical PCs
// belonging to one physical frame collapses into its first PC, which is the
// real return address.
func collapseInlineGroups(stk []uintptr) []uintptr {
	physical := make([]uintptr, 0, len(stk))
	var prevEntry uintptr
	for _, pc := range stk {
		var entry uintptr
		if f := runtime.FuncForPC(pc); f != nil {
			entry = f.Entry()
		}
		if len(physical) > 0 && entry == prevEntry {
			continue // a virtual PC of the physical frame already recorded
		}
		physical = append(physical, pc)
		prevEntry = entry
	}

	return physical
}

// expandInlinedFrames mirrors runtime/pprof.expandInlinedFrames, the
// normalization step godeltaprof is missing.
func expandInlinedFrames(stk []uintptr) []uintptr {
	expanded := make([]uintptr, 0, len(stk))
	frames := runtime.CallersFrames(stk)
	for {
		f, more := frames.Next()
		// f.PC is a "call PC", consumers expect "return PCs"
		expanded = append(expanded, f.PC+1)
		if !more {
			break
		}
	}

	return expanded
}

// sampleFuncs returns the function names of the first profile sample that
// contains want, leaf last.
func sampleFuncs(t *testing.T, profile *gprofile.Profile, want string) []string {
	t.Helper()

	var found []string
	for _, s := range profile.Sample {
		var funcs []string
		for i := len(s.Location) - 1; i >= 0; i-- {
			for j := len(s.Location[i].Line) - 1; j >= 0; j-- {
				funcs = append(funcs, s.Location[i].Line[j].Function.Name)
			}
		}
		if found == nil && slices.Contains(funcs, want) {
			found = funcs
		}
	}
	require.NotNilf(t, found, "no sample contains %s", want)

	return found
}

// TestMutexProfilePhysicalStackDropsInlinedFrames shows the silent half of the
// bug on a single record: a physical stack, the shape frame pointer unwinding
// produces, loses every frame inlined into reproL5. A physical PC is expanded
// with CallersFrames one PC at a time (profileBuilder.allFrames), and
// runtime.Frames only inserts the virtual PCs of an inlined call when it can
// peek at the next PC of the stack.
func TestMutexProfilePhysicalStackDropsInlinedFrames(t *testing.T) {
	physical := collapseInlineGroups(captureLogicalStack(t))

	buf := dumpRecords(t, pprof.BlockProfileRecord{Count: 1, Cycles: 1000, Stack: physical})

	profile, err := gprofile.ParseData(buf.Bytes())
	require.NoError(t, err)
	funcs := sampleFuncs(t, profile, reproLeafFrame)
	for _, f := range inlinedFrames {
		assert.Containsf(t, funcs, f, "frame %s missing from %v", f, funcs)
	}
}

// TestMutexProfileMixedStackShapes reproduces the LocsForStack panic reported in
// https://github.com/grafana/pyroscope-go/issues/245.
//
// A single dump may contain records of both shapes: on a cgo heavy service some
// contention events are sampled while cgo is on the stack, and those get a full
// traceback. The logical record teaches the location cache how many PCs the
// reproL5 frame is made of; the physical record spells that very same frame as a
// single PC, so LocsForStack skips more PCs than the stack has left.
func TestMutexProfileMixedStackShapes(t *testing.T) {
	logical := captureLogicalStack(t)
	physical := collapseInlineGroups(logical)

	// Self check: the synthetic physical stack is exactly the one the runtime
	// would have recorded, i.e. expanding it back yields the logical stack.
	require.Less(t, len(physical), len(logical))
	require.Equal(t, logical, expandInlinedFrames(physical))

	var buf *bytes.Buffer
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("godeltaprof panicked building a mutex profile "+
					"out of a logical and a physical stack: %v", r)
			}
		}()
		// The logical record comes first: BlockProfiler.Profile sorts records by
		// cycles, and it is the cgo path that holds locks the longest.
		buf = dumpRecords(t,
			pprof.BlockProfileRecord{Count: 10, Cycles: 10000, Stack: logical},
			pprof.BlockProfileRecord{Count: 1, Cycles: 1000, Stack: physical},
		)
	}()
	if buf == nil {
		return // already reported
	}

	profile, err := gprofile.ParseData(buf.Bytes())
	require.NoError(t, err)
	funcs := sampleFuncs(t, profile, reproLeafFrame)
	for _, f := range inlinedFrames {
		assert.Containsf(t, funcs, f, "frame %s missing from %v", f, funcs)
	}
}

// dumpRecords writes the given block profile records the way
// godeltaprof.BlockProfiler.Profile does.
func dumpRecords(t *testing.T, records ...pprof.BlockProfileRecord) *bytes.Buffer {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	err := compat.PrintCountCycleProfile(
		&pprof.DeltaMutexProfiler{},
		&pprof.ProfileBuilderOptions{GenericsFrames: true, LazyMapping: true},
		buf,
		pprof.ScalerMutexProfile,
		records,
	)
	require.NoError(t, err)

	return buf
}

// TestMutexProfileDropsInlinedFrames is the end to end version of
// TestMutexProfilePhysicalStackDropsInlinedFrames: real contention, real
// records, no cgo and no panic, just missing frames.
func TestMutexProfileDropsInlinedFrames(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(prev)

	profiler := godeltaprof.NewMutexProfiler()
	require.NoError(t, dumpMutexProfile(t, profiler, io.Discard))

	contendReproMu(4, 50*time.Microsecond, 200*time.Millisecond)

	buf := bytes.NewBuffer(nil)
	require.NoError(t, dumpMutexProfile(t, profiler, buf))
	if t.Failed() {
		return // already reported
	}

	profile, err := gprofile.ParseData(buf.Bytes())
	require.NoError(t, err)
	funcs := sampleFuncs(t, profile, reproLeafFrame)
	for _, f := range inlinedFrames {
		assert.Containsf(t, funcs, f, "frame %s missing from %v", f, funcs)
	}
}
