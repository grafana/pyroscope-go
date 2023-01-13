package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pyroscope-io/client/heap"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/pyroscope-io/client/pyroscope"
)

//go:noinline
func work(n int) {
	// revive:disable:empty-block this is fine because this is a example app, not real production code
	for i := 0; i < n; i++ {
	}
	fmt.Printf("work\n")
	// revive:enable:empty-block
}

var m sync.Mutex

func fastFunction(c context.Context, wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	pyroscope.TagWrapper(c, pyroscope.Labels("function", "fast"), func(c context.Context) {
		work(200000000)
	})
	wg.Done()
}

func slowFunction(c context.Context, wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	// standard pprof.Do wrappers work as well
	pprof.Do(c, pprof.Labels("function", "slow"), func(c context.Context) {
		work(800000000)
	})
	wg.Done()
}

func main() {
	go func() {
		hp := heap.DeltaHeapProfiler{}
		for {
			buf := bytes.NewBuffer(nil)
			err := hp.Profile(buf)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	pyroscope.Start(pyroscope.Config{
		ApplicationName: "simple.golang.app-new",
		ServerAddress:   "http://localhost:4040", // this will run inside docker-compose, hence `pyroscope` for hostname
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	pyroscope.TagWrapper(context.Background(), pyroscope.Labels("foo", "bar"), func(c context.Context) {
		for {
			wg := sync.WaitGroup{}
			wg.Add(2)
			go fastFunction(c, &wg)
			go slowFunction(c, &wg)
			wg.Wait()
		}
	})

}
