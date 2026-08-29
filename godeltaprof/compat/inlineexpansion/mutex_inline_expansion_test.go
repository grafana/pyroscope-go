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

// TestMutexProfileDropsInlinedFrames covers the silent half of the bug: no cgo,
// no panic, but the frames inlined into reproEntry never make it into the
// profile. A physical PC is expanded with CallersFrames one PC at a time
// (profileBuilder.allFrames), and runtime.Frames only inserts the virtual PCs of
// an inlined call when it can peek at the next PC of the stack.
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
