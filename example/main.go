package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/grafana/pyroscope-golang/profiler"
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

	profiler.TagWrapper(c, profiler.Labels("function", "fast"), func(c context.Context) {
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
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	profiler.Start(profiler.Config{
		ApplicationName:   "simple.golang.app-new",
		ServerAddress:     "http://localhost:4040",
		Logger:            profiler.StandardLogger,
		AuthToken:         os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		TenantID:          os.Getenv("PYROSCOPE_TENANT_ID"),
		BasicAuthUser:     os.Getenv("PYROSCOPE_BASIC_AUTH_USER"),
		BasicAuthPassword: os.Getenv("PYROSCOPE_BASIC_AUTH_PASSWORD"),
		ProfileTypes: []profiler.ProfileType{
			profiler.ProfileCPU,
			profiler.ProfileInuseObjects,
			profiler.ProfileAllocObjects,
			profiler.ProfileInuseSpace,
			profiler.ProfileAllocSpace,
			profiler.ProfileGoroutines,
			profiler.ProfileMutexCount,
			profiler.ProfileMutexDuration,
			profiler.ProfileBlockCount,
			profiler.ProfileBlockDuration,
		},
		HTTPHeaders: map[string]string{"X-Extra-Header": "extra-header-value"},
	})

	profiler.TagWrapper(context.Background(), profiler.Labels("foo", "bar"), func(c context.Context) {
		for {
			wg := sync.WaitGroup{}
			wg.Add(2)
			go fastFunction(c, &wg)
			go slowFunction(c, &wg)
			wg.Wait()
		}
	})
}
