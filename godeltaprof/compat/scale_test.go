package compat

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	gprofile "github.com/google/pprof/profile"
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var m sync.Mutex

func TestScaleMutex(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(-1)
	defer runtime.SetMutexProfileFraction(prev)

	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	profiler := godeltaprof.NewMutexProfiler()
	err := profiler.Profile(io.Discard)
	require.NoError(t, err)

	const fraction = 5
	const iters = 5000
	const workers = 2
	const expectedCount = workers * iters
	const expectedTime = expectedCount * 1000000

	runtime.SetMutexProfileFraction(fraction)

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for j := 0; j < workers; j++ {
		go func() {
			for i := 0; i < iters; i++ {
				m.Lock()
				time.Sleep(time.Millisecond)
				m.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()

	err = profiler.Profile(buffer)
	require.NoError(t, err)

	profile, err := gprofile.Parse(buffer)
	require.NoError(t, err)

	res := stackCollapseProfile(profile)

	my := findStack(res, "github.com/grafana/pyroscope-go/godeltaprof/compat.TestScaleMutex")
	require.NotNil(t, my)

	fmt.Println(my.value[0], my.value[1])
	fmt.Println(expectedCount, expectedTime)

	assert.Less(t, math.Abs(float64(my.value[0])-float64(expectedCount)), 0.4*float64(expectedCount))
	assert.Less(t, math.Abs(float64(my.value[1])-float64(expectedTime)), 0.4*float64(expectedTime))
}

func TestScaleBlock(t *testing.T) {
	defer runtime.SetBlockProfileRate(0)

	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	profiler := godeltaprof.NewBlockProfiler()
	err := profiler.Profile(io.Discard)
	require.NoError(t, err)

	const fraction = 5
	const iters = 5000
	const workers = 2
	const expectedCount = workers * iters
	const expectedTime = expectedCount * 1000000

	runtime.SetBlockProfileRate(fraction)

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for j := 0; j < workers; j++ {
		go func() {
			for i := 0; i < iters; i++ {
				m.Lock()
				time.Sleep(time.Millisecond)
				m.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()

	err = profiler.Profile(buffer)
	require.NoError(t, err)

	profile, err := gprofile.Parse(buffer)
	require.NoError(t, err)

	res := stackCollapseProfile(profile)

	my := findStack(res, "github.com/grafana/pyroscope-go/godeltaprof/compat.TestScaleBlock")
	require.NotNil(t, my)

	fmt.Println(my.value[0], my.value[1])
	fmt.Println(expectedCount, expectedTime)

	assert.Less(t, math.Abs(float64(my.value[0])-float64(expectedCount)), 0.4*float64(expectedCount))
	assert.Less(t, math.Abs(float64(my.value[1])-float64(expectedTime)), 0.4*float64(expectedTime))
}

var bufs [][]byte

//go:noinline
func appendBuf(sz int) {
	elems := make([]byte, 0, sz)
	bufs = append(bufs, elems)
}

func TestScaleHeap(t *testing.T) {
	prev := runtime.MemProfileRate
	runtime.MemProfileRate = 0

	const size = 64 * 1024
	const iters = 1024

	const expectedCount = iters
	const expectedTime = 1000000

	bufs = make([][]byte, 0, iters)
	defer func() {
		bufs = nil
		runtime.MemProfileRate = prev
	}()

	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	profiler := godeltaprof.NewHeapProfiler()
	err := profiler.Profile(io.Discard)
	require.NoError(t, err)

	runtime.MemProfileRate = 1
	for i := 0; i < iters; i++ {
		appendBuf(size)
	}

	time.Sleep(time.Second)
	runtime.GC()
	time.Sleep(time.Second)

	expected := []int64{iters, iters * size, iters, iters * size}
	err = profiler.Profile(buffer)
	require.NoError(t, err)

	profile, err := gprofile.Parse(buffer)
	require.NoError(t, err)

	res := stackCollapseProfile(profile)

	my := findStack(res, "github.com/grafana/pyroscope-go/godeltaprof/compat.TestScaleHeap;github.com/grafana/pyroscope-go/godeltaprof/compat.appendBuf")
	require.NotNil(t, my)

	fmt.Println(my.value)
	fmt.Println(expected)
	for i := range my.value {
		assert.Less(t, math.Abs(float64(my.value[i])-float64(expected[i])), 0.1*float64(expected[i]))
	}
}
