//go:build go1.21

package compat

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	n := 10
	fib[int](&n)

	runtime.GC()

	const expectedRealShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsShape;github.com/grafana/pyroscope-go/godeltaprof/compat.fib\\[go.shape.int\\]$"
	const expectedOmmitedShape = "github.com/grafana/pyroscope-go/godeltaprof/compat.TestGenericsShape;github.com/grafana/pyroscope-go/godeltaprof/compat.fib\\[\\.\\.\\.\\]$"

	t.Run("go runtime", func(t *testing.T) {
		buffer = bytes.NewBuffer(nil)
		err = pprof.WriteHeapProfile(buffer)
		require.NoError(t, err)
		profile, err := gprofile.Parse(buffer)
		require.NoError(t, err)
		line := findStack(stackCollapseProfile(profile), expectedRealShape)
		assert.NotNil(t, line)
		if line != nil {
			assert.Equal(t, int64(2), line.value[0])
			assert.Equal(t, int64(32), line.value[1])
		}
	})

	t.Run("godeltaprof disabled", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfiler()
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		profile, err := gprofile.Parse(buffer)
		require.NoError(t, err)
		line := findStack(stackCollapseProfile(profile), expectedOmmitedShape)
		assert.NotNil(t, line)
		if line != nil {
			assert.Equal(t, int64(2), line.value[0])
			assert.Equal(t, int64(32), line.value[1])
		}
	})

	t.Run("godeltaprof disabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: false,
		})
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		profile, err := gprofile.Parse(buffer)
		require.NoError(t, err)
		line := findStack(stackCollapseProfile(profile), expectedOmmitedShape)
		assert.NotNil(t, line)
		if line != nil {
			assert.Equal(t, int64(2), line.value[0])
			assert.Equal(t, int64(32), line.value[1])
		}
	})

	t.Run("godeltaprof enabled explicitly", func(t *testing.T) {
		profiler := godeltaprof.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
			GenericsFrames: true,
		})
		buffer = bytes.NewBuffer(nil)
		err = profiler.Profile(buffer)
		require.NoError(t, err)
		profile, err := gprofile.Parse(buffer)
		require.NoError(t, err)
		line := findStack(stackCollapseProfile(profile), expectedRealShape)
		assert.NotNil(t, line)
		if line != nil {
			assert.Equal(t, int64(2), line.value[0])
			assert.Equal(t, int64(32), line.value[1])
		}
	})
}

func TestBlock(t *testing.T) {
	t.Fail() //todo
}
func TestMutex(t *testing.T) {
	t.Fail() //todo
}
