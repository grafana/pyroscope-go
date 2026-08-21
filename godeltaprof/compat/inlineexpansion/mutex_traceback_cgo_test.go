//go:build cgo

package inlineexpansion

import (
	"bytes"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/godeltaprof"
)

// TestMutexProfileCgoMixedTraceback is the end to end version of
// TestMutexProfileMixedStackShapes: no synthetic records, the mix of physical
// and logical stacks within one dump comes from the runtime itself.
//
// reproMu is contended from two kinds of goroutines. The cgo one runs the
// contending code in a Go callback invoked from C, so the M has cgo on its
// stack and the runtime records a full, inline expanded traceback for it. The
// plain Go ones get frame pointer unwinding, i.e. physical PCs. This is what a
// CGO_ENABLED=1 service looks like, and it panics the profile builder exactly
// as reported in https://github.com/grafana/pyroscope-go/issues/245.
func TestMutexProfileCgoMixedTraceback(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(prev)

	profiler := godeltaprof.NewMutexProfiler()
	require.NoError(t, dumpMutexProfile(t, profiler, io.Discard))

	// the cgo goroutine holds the mutex longer, so that its record sorts first
	// and populates the location cache before the plain Go records are written.
	cgoHoldDuration = 2 * time.Millisecond

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			callGoViaC()
		}
	}()
	contendReproMu(4, 50*time.Microsecond, 300*time.Millisecond)
	close(stop)
	wg.Wait()

	buf := bytes.NewBuffer(nil)
	require.NoError(t, dumpMutexProfile(t, profiler, buf))
	if t.Failed() {
		return // already reported
	}

	profile, err := gprofile.ParseData(buf.Bytes())
	require.NoError(t, err)
	// both the cgo and the plain Go paths must be there, with all their frames
	for _, root := range []string{
		"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.godeltaprofCgoCallback",
		"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.contendReproMu.func1",
	} {
		funcs := sampleFuncs(t, profile, root)
		for _, f := range inlinedFrames {
			assert.Containsf(t, funcs, f, "frame %s missing from %v", f, funcs)
		}
	}
}
