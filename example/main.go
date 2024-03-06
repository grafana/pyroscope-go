package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/grafana/pyroscope-go"
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
		go func() {
			defer wg.Done()
			work(200000000)
		}()
	})
}

func slowFunction(c context.Context, wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()
	// standard pprof.Do wrappers work as well
	pprof.Do(c, pprof.Labels("function", "slow"), func(c context.Context) {
		go func() {
			wg.Done()
			work(800000000)
		}()
	})
}

func main() {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	pyroscope.Start(pyroscope.Config{
		ApplicationName:   "simple.golang.app-new",
		ServerAddress:     "http://localhost:4040",
		Logger:            pyroscope.StandardLogger,
		AuthToken:         os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		TenantID:          os.Getenv("PYROSCOPE_TENANT_ID"),
		BasicAuthUser:     os.Getenv("PYROSCOPE_BASIC_AUTH_USER"),
		BasicAuthPassword: os.Getenv("PYROSCOPE_BASIC_AUTH_PASSWORD"),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileWallClockTime,
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
		HTTPHeaders: map[string]string{"X-Extra-Header": "extra-header-value"},
	})

	pyroscope.TagWrapper(context.Background(), pyroscope.Labels("foo", "bar"), func(c context.Context) {
		for {
			wg := sync.WaitGroup{}
			wg.Add(2)
			fastFunction(c, &wg)
			slowFunction(c, &wg)
			wg.Wait()
		}
	})
}
