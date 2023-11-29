//go:build go1.18 && !go1.21
// +build go1.18,!go1.21

package compat

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/stretchr/testify/require"
)

func genericAllocFunc[T any](n int) []T {
	return make([]T, n)
}

func genericBlock[T any](n int) {
	for i := 0; i < n; i++ {
		m.Lock()
		time.Sleep(100 * time.Millisecond)
		m.Unlock()
	}
}

func triggerGenericBlock() {
	const iters = 2
	const workers = 10

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for j := 0; j < workers; j++ {
		go func() {
			genericBlock[int](iters)
			wg.Done()
		}()
	}
	wg.Wait()
}

// TestGenerics tests that pre go1.21 we emmit [...] as generics
func TestGenericsShape(t *testing.T) {

	prev := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	runtime.GC()

	defer func() {
		runtime.MemProfileRate = prev
	}()

	_ = genericAllocFunc[int](239)

	runtime.GC()

	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsShape;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat.genericAllocFunc\\[\\.\\.\\.\\]$"

	t.Run("go runtime", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		err := pprof.WriteHeapProfile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 1, 2048)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfiler()
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 1, 2048)
	})

	t.Run("godeltaprof generics disabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: false,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 1, 2048)
	})

	t.Run("godeltaprof generics enabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: true,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 1, 2048)
	})
}

func TestBlock(t *testing.T) {
	defer runtime.SetBlockProfileRate(0)

	runtime.SetBlockProfileRate(1) // every block

	triggerGenericBlock()

	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[\\.\\.\\.\\];sync\\.\\(\\*Mutex\\)\\.Lock"

	t.Run("go runtime", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		err := pprof.Lookup("block").WriteTo(buffer, 0)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewBlockProfiler()
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics disabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewBlockProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: false,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics enabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewBlockProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: true,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})
}

func TestMutex(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(-1)
	defer runtime.SetMutexProfileFraction(prev)
	runtime.SetMutexProfileFraction(1)

	triggerGenericBlock()

	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[\\.\\.\\.\\];sync\\.\\(\\*Mutex\\)\\.Unlock"

	t.Run("go runtime", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		err := pprof.Lookup("mutex").WriteTo(buffer, 0)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewMutexProfiler()
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics disabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewMutexProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: false,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})

	t.Run("godeltaprof generics enabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewMutexProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: true,
		})
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 19)
	})
}
