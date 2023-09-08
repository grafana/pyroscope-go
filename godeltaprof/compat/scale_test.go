package compat

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"runtime"
	"sort"
	"strings"
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

type stack struct {
	funcs []string
	line  string
	value []int64
}

func findStack(res []stack, line string) *stack {
	for i, re := range res {
		//fmt.Println(re.line, re.value)
		if strings.Contains(re.line, line) {
			return &res[i]
		}
	}

	return nil
}

func stackCollapseProfile(p *gprofile.Profile) []stack {

	var ret []stack
	for _, s := range p.Sample {
		var funcs []string
		for i := range s.Location {

			loc := s.Location[i]
			for _, line := range loc.Line {
				f := line.Function
				//funcs = append(funcs, fmt.Sprintf("%s:%d", f.Name, line.Line))
				funcs = append(funcs, f.Name)
			}
		}
		for i := 0; i < len(funcs)/2; i++ {
			j := len(funcs) - i - 1
			funcs[i], funcs[j] = funcs[j], funcs[i]
		}

		ret = append(ret, stack{
			line:  strings.Join(funcs, ";"),
			funcs: funcs,
			value: s.Value,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return strings.Compare(ret[i].line, ret[j].line) < 0
	})
	var unique []stack
	for _, s := range ret {
		if len(unique) == 0 {
			unique = append(unique, s)
			continue
		}
		if unique[len(unique)-1].line == s.line {
			for i := 0; i < len(s.value); i++ {
				unique[len(unique)-1].value[i] += s.value[i]
			}
			continue
		}
		unique = append(unique, s)

	}

	return unique
}
