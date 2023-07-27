package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"

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

func fastFunction(c context.Context) {
	pyroscope.TagWrapper(c, pyroscope.Labels("function", "fast"), func(c context.Context) {
		work(200000000)
	})
}

func slowFunction(c context.Context) {
	// standard pprof.Do wrappers work as well
	pprof.Do(c, pprof.Labels("function", "slow"), func(c context.Context) {
		work(800000000)
	})
}

func main() {
	sa := os.Getenv("SERVER_ADDRESS")
	if sa == "" {
		sa = "https://localhost:4317"
	}
	pyroscope.Start(pyroscope.Config{
		ApplicationName: "go-test",
		ServerAddress:   sa,
		Logger:          pyroscope.StandardLogger,
		UseOTLP:         true,
	})

	pyroscope.TagWrapper(context.Background(), pyroscope.Labels("foo", "bar"), func(c context.Context) {
		for {
			fastFunction(c)
			slowFunction(c)
		}
	})
}
