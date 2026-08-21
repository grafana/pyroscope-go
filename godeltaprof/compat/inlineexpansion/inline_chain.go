package inlineexpansion

import (
	"runtime"
	"sync"
	"time"
)

// The helpers below build a call chain that the compiler collapses into fewer
// physical frames than there are logical ones, currently:
//
//	reproEntry
//	  -> reproL5 (reproL4..reproL1 inlined into it)
//	    -> reproLeaf (sync.(*Mutex).Unlock inlined into it)
//
// Where exactly the inliner draws the line does not matter, only that one
// physical frame covers several logical ones. reproLeaf contends on reproMu, so
// the runtime records that frame in mutex and block profiles, either as a single
// physical PC or as one PC per logical frame, depending on how the stack was
// captured. See mutex_inline_expansion_test.go.

var reproMu sync.Mutex //nolint:gochecknoglobals

//go:noinline
func reproSpin(d time.Duration) {
	start := time.Now()
	for time.Since(start) < d {
	}
}

// reproLeaf holds reproMu for d, and optionally records its own call stack.
// The recorded stack is a "logical" one: runtime.Callers fully expands inlined
// frames, same as the runtime traceback used for contention events when frame
// pointer unwinding is unavailable.
//
//go:noinline
func reproLeaf(d time.Duration, pcs []uintptr) int {
	reproMu.Lock()
	reproSpin(d)
	reproMu.Unlock()

	if pcs == nil {
		return 0
	}

	return runtime.Callers(1, pcs)
}

func reproL1(d time.Duration, pcs []uintptr) int { return reproLeaf(d, pcs) }
func reproL2(d time.Duration, pcs []uintptr) int { return reproL1(d, pcs) }
func reproL3(d time.Duration, pcs []uintptr) int { return reproL2(d, pcs) }
func reproL4(d time.Duration, pcs []uintptr) int { return reproL3(d, pcs) }
func reproL5(d time.Duration, pcs []uintptr) int { return reproL4(d, pcs) }

//go:noinline
func reproEntry(d time.Duration, pcs []uintptr) int { return reproL5(d, pcs) }

const reproLeafFrame = "github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproLeaf"

// inlinedFrames are the frames of the reproEntry call chain, leaf first. A
// correct mutex profile contains all of them.
var inlinedFrames = []string{ //nolint:gochecknoglobals
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproL1",
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproL2",
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproL3",
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproL4",
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproL5",
	"github.com/grafana/pyroscope-go/godeltaprof/compat/inlineexpansion.reproEntry",
}

// contendReproMu creates contention on reproMu through reproEntry, so that the
// runtime records mutex profile events with reproEntry on the stack.
func contendReproMu(workers int, hold, total time.Duration) {
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				reproEntry(hold, nil)
			}
		}()
	}
	time.Sleep(total)
	close(stop)
	wg.Wait()
}
