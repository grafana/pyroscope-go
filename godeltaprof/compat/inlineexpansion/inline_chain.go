package inlineexpansion

import (
	"sync"
	"time"
)

// The helpers below build a call chain that the compiler collapses into fewer
// physical frames than there are logical ones, currently:
//
//	reproEntry (reproL5..reproL1 inlined into it)
//	  -> reproLeaf (sync.(*Mutex).Unlock inlined into it)
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

//go:noinline
func reproLeaf(d time.Duration) {
	reproMu.Lock()
	reproSpin(d)
	reproMu.Unlock()
}

func reproL1(d time.Duration) { reproLeaf(d) }
func reproL2(d time.Duration) { reproL1(d) }
func reproL3(d time.Duration) { reproL2(d) }
func reproL4(d time.Duration) { reproL3(d) }
func reproL5(d time.Duration) { reproL4(d) }

//go:noinline
func reproEntry(d time.Duration) { reproL5(d) }

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
				reproEntry(hold)
			}
		}()
	}
	time.Sleep(total)
	close(stop)
	wg.Wait()
}
