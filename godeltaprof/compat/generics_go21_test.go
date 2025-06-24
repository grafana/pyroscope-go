//go:build go1.21
// +build go1.21

package compat

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/pprof/profile"
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

// TestGenerics tests that post go1.21 we emmit [...] as generics by default and [go.shape.int] if enabled
func TestGenericsShape(t *testing.T) {
	var buffer *bytes.Buffer
	var err error

	prev := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	runtime.GC()

	defer func() {
		runtime.MemProfileRate = prev
	}()

	it := genericAllocFunc[int](239)
	escape(it)

	runtime.GC()

	const expectedRealShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsShape;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc\\[go\\.shape\\.int\\]$"
	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsShape;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc\\[\\.\\.\\.\\]$"

	t.Run("go runtime", func(t *testing.T) {
		buffer = bytes.NewBuffer(nil)
		err = pprof.WriteHeapProfile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 1, 2048)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfiler()
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 1, 2048)
	})

	t.Run("godeltaprof generics disabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: false,
		})
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedOmmitedShape, 1, 2048)
	})

	t.Run("godeltaprof generics enabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: true,
		})
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 1, 2048)
	})
}

func TestBlock(t *testing.T) {
	defer runtime.SetBlockProfileRate(0)
	runtime.SetBlockProfileRate(1) // every block

	triggerGenericBlock()

	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[\\.\\.\\.\\];sync\\.\\(\\*Mutex\\)\\.Lock"

	const expectedRealShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[go\\.shape\\.int];sync\\.\\(\\*Mutex\\)\\.Lock"

	t.Run("go runtime", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		err := pprof.Lookup("block").WriteTo(buffer, 0)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 19)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewBlockProfiler()
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 19)
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
		expectStackFrames(t, buffer, expectedRealShape, 19)
	})
}

func TestMutex(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(-1)
	defer runtime.SetMutexProfileFraction(prev)
	runtime.SetMutexProfileFraction(1)

	triggerGenericBlock()

	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[\\.\\.\\.\\];sync\\.\\(\\*Mutex\\)\\.Unlock"

	const expectedRealShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.triggerGenericBlock.func1;github.com/grafana/pyroscope-go/godeltaprof/" +
		"compat\\.genericBlock\\[go\\.shape\\.int];sync\\.\\(\\*Mutex\\)\\.Unlock"

	t.Run("go runtime", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		err := pprof.Lookup("mutex").WriteTo(buffer, 0)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 19)
	})

	t.Run("godeltaprof generics enabled by default", func(t *testing.T) {
		profiler := godeltaprof.NewMutexProfiler()
		buffer := bytes.NewBuffer(nil)
		err := profiler.Profile(buffer)
		require.NoError(t, err)
		expectStackFrames(t, buffer, expectedRealShape, 19)
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
		expectStackFrames(t, buffer, expectedRealShape, 19)
	})
}

func profileToStrings(p *profile.Profile) []string {
	var res []string
	for _, s := range p.Sample {
		res = append(res, sampleToString(s))
	}
	return res
}

func sampleToString(s *profile.Sample) string {
	var funcs []string
	for i := len(s.Location) - 1; i >= 0; i-- {
		loc := s.Location[i]
		funcs = locationToStrings(loc, funcs)
	}
	return fmt.Sprintf("%s %v", strings.Join(funcs, ";"), s.Value)
}

func locationToStrings(loc *profile.Location, funcs []string) []string {
	for j := range loc.Line {
		line := loc.Line[len(loc.Line)-1-j]
		funcs = append(funcs, line.Function.Name)
	}
	return funcs
}

// This is a regression test for https://go.dev/issue/64528 .
func TestGenericsHashKeyInPprofBuilder(t *testing.T) {
	previousRate := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() {
		runtime.MemProfileRate = previousRate
	}()
	for _, sz := range []int{128, 256} {
		it := genericAllocFunc[uint32](sz / 4)
		escape(it)
	}
	for _, sz := range []int{32, 64} {
		it := genericAllocFunc[uint64](sz / 8)
		escape(it)
	}

	runtime.GC()
	buf := bytes.NewBuffer(nil)
	if err := WriteHeapProfile(buf); err != nil {
		t.Fatalf("writing profile: %v", err)
	}
	p, err := profile.Parse(buf)
	if err != nil {
		t.Fatalf("profile.Parse: %v", err)
	}

	actual := profileToStrings(p)
	expected := []string{
		"testing.tRunner;github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsHashKeyInPprofBuilder;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc[go.shape.uint32] [1 128 0 0]",
		"testing.tRunner;github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsHashKeyInPprofBuilder;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc[go.shape.uint32] [1 256 0 0]",
		"testing.tRunner;github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsHashKeyInPprofBuilder;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc[go.shape.uint64] [1 32 0 0]",
		"testing.tRunner;github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsHashKeyInPprofBuilder;github.com/grafana/pyroscope-go/godeltaprof/compat.genericAllocFunc[go.shape.uint64] [1 64 0 0]",
	}

	for _, l := range expected {
		if !slices.Contains(actual, l) {
			t.Errorf("profile = %v\nwant = %v", strings.Join(actual, "\n"), l)
		}
	}
}

type opAlloc struct {
	buf [128]byte
}

type opCall struct {
}

var sink []byte

func storeAlloc() {
	sink = make([]byte, 16)
}

func nonRecursiveGenericAllocFunction[CurrentOp any, OtherOp any](alloc bool) {
	if alloc {
		storeAlloc()
	} else {
		nonRecursiveGenericAllocFunction[OtherOp, CurrentOp](true)
	}
}

func TestGenericsInlineLocations(t *testing.T) {
	if OptimizationOff() {
		t.Skip("skipping test with optimizations disabled")
	}

	previousRate := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() {
		runtime.MemProfileRate = previousRate
		sink = nil
	}()

	nonRecursiveGenericAllocFunction[opAlloc, opCall](true)
	nonRecursiveGenericAllocFunction[opCall, opAlloc](false)

	runtime.GC()

	buf := bytes.NewBuffer(nil)
	if err := WriteHeapProfile(buf); err != nil {
		t.Fatalf("writing profile: %v", err)
	}
	p, err := profile.Parse(buf)
	if err != nil {
		t.Fatalf("profile.Parse: %v", err)
	}

	const expectedSample = "testing.tRunner;github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsInlineLocations;github.com/grafana/pyroscope-go/godeltaprof/compat.nonRecursiveGenericAllocFunction[go.shape.struct {},go.shape.struct { github.com/grafana/pyroscope-go/godeltaprof/compat.buf [128]uint8 }];github.com/grafana/pyroscope-go/godeltaprof/compat.nonRecursiveGenericAllocFunction[go.shape.struct { github.com/grafana/pyroscope-go/godeltaprof/compat.buf [128]uint8 },go.shape.struct {}];github.com/grafana/pyroscope-go/godeltaprof/compat.storeAlloc [1 16 1 16]"
	const expectedLocation = "github.com/grafana/pyroscope-go/godeltaprof/compat.nonRecursiveGenericAllocFunction[go.shape.struct {},go.shape.struct { github.com/grafana/pyroscope-go/godeltaprof/compat.buf [128]uint8 }];github.com/grafana/pyroscope-go/godeltaprof/compat.nonRecursiveGenericAllocFunction[go.shape.struct { github.com/grafana/pyroscope-go/godeltaprof/compat.buf [128]uint8 },go.shape.struct {}];github.com/grafana/pyroscope-go/godeltaprof/compat.storeAlloc"
	const expectedLocationNewInliner = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsInlineLocations;" + expectedLocation
	var s *profile.Sample
	for _, sample := range p.Sample {
		if sampleToString(sample) == expectedSample {
			s = sample
			break
		}
	}
	if s == nil {
		t.Fatalf("expected \n%s\ngot\n%s", expectedSample, strings.Join(profileToStrings(p), "\n"))
	}
	loc := s.Location[0]
	actual := strings.Join(locationToStrings(loc, nil), ";")
	if expectedLocation != actual && expectedLocationNewInliner != actual {
		t.Errorf("expected a location with at least 3 functions\n%s\ngot\n%s\n", expectedLocation, actual)
	}
}

func OptimizationOff() bool {
	optimizationMarker := func() uintptr {
		pc, _, _, _ := runtime.Caller(0)
		return pc
	}
	pc := optimizationMarker()
	f := runtime.FuncForPC(runtime.FuncForPC(pc).Entry())
	return f.Name() == "github.com/grafana/pyroscope-go/godeltaprof/compat.OptimizationOff.func1"
}

func WriteHeapProfile(w io.Writer) error {
	runtime.GC()
	dh := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})
	return dh.Profile(w)
}

var blackhole []any

// make sure a is on the heap
// https://go-review.googlesource.com/c/go/+/649035
// https://go-review.googlesource.com/c/go/+/653856
func escape(a any) {
	blackhole = append(blackhole, a)
	blackhole[0] = nil
	blackhole = blackhole[:0]
}
