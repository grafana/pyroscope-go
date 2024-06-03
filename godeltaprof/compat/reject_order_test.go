package compat

import (
	"bytes"
	"io"
	"runtime"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeapReject(t *testing.T) {
	dh := new(pprof.DeltaHeapProfiler)
	opt := new(pprof.ProfileBuilderOptions)
	fs := generateMemProfileRecords(512, 32, 239)
	p1 := bytes.NewBuffer(nil)
	err := WriteHeapProto(dh, opt, p1, fs, int64(runtime.MemProfileRate))
	assert.NoError(t, err)
	p1Size := p1.Len()
	profile, err := gprofile.Parse(p1)
	require.NoError(t, err)
	ls := stackCollapseProfile(t, profile)
	assert.Len(t, ls, 512)
	assert.Len(t, profile.Location, 141)
	t.Log("p1 size", p1Size)

	p2 := bytes.NewBuffer(nil)
	err = WriteHeapProto(dh, opt, p2, fs, int64(runtime.MemProfileRate))
	assert.NoError(t, err)
	p2Size := p2.Len()
	assert.Less(t, p2Size, 1000)
	profile, err = gprofile.Parse(p2)
	require.NoError(t, err)
	ls = stackCollapseProfile(t, profile)
	assert.Len(t, ls, 0)
	assert.Len(t, profile.Location, 0)
	t.Log("p2 size", p2Size)
}

func BenchmarkHeapRejectOrder(b *testing.B) {
	opt := &pprof.ProfileBuilderOptions{
		GenericsFrames: false,
		LazyMapping:    true,
	}
	dh := &pprof.DeltaHeapProfiler{}
	fs := generateMemProfileRecords(512, 32, 239)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteHeapProto(dh, opt, io.Discard, fs, int64(runtime.MemProfileRate))
	}
}

var mutexProfileScalers = []pprof.MutexProfileScaler{
	pprof.ScalerMutexProfile,
	pprof.ScalerBlockProfile,
}

func TestMutexReject(t *testing.T) {
	for i, scaler := range mutexProfileScalers {
		name := "ScalerMutexProfile"
		if i == 1 {
			name = "ScalerBlockProfile"
		}
		t.Run(name, func(t *testing.T) {
			prevMutexProfileFraction := runtime.SetMutexProfileFraction(-1)
			runtime.SetMutexProfileFraction(5)
			defer runtime.SetMutexProfileFraction(prevMutexProfileFraction)

			dh := new(pprof.DeltaMutexProfiler)
			opt := new(pprof.ProfileBuilderOptions)
			fs := generateBlockProfileRecords(512, 32, 239)
			p1 := bytes.NewBuffer(nil)
			err := PrintCountCycleProfile(dh, opt, p1, scaler, fs)
			assert.NoError(t, err)
			p1Size := p1.Len()
			profile, err := gprofile.Parse(p1)
			require.NoError(t, err)
			ls := stackCollapseProfile(t, profile)
			assert.Len(t, ls, 512)
			assert.Len(t, profile.Location, 141)
			t.Log("p1 size", p1Size)

			p2 := bytes.NewBuffer(nil)
			err = PrintCountCycleProfile(dh, opt, p2, scaler, fs)
			assert.NoError(t, err)
			p2Size := p2.Len()
			assert.Less(t, p2Size, 1000)
			profile, err = gprofile.Parse(p2)
			require.NoError(t, err)
			ls = stackCollapseProfile(t, profile)
			assert.Len(t, ls, 0)
			assert.Len(t, profile.Location, 0)
			t.Log("p2 size", p2Size)
		})
	}
}

func BenchmarkMutexRejectOrder(b *testing.B) {
	for i, scaler := range mutexProfileScalers {
		name := "ScalerMutexProfile"
		if i == 1 {
			name = "ScalerBlockProfile"
		}
		b.Run(name, func(b *testing.B) {
			prevMutexProfileFraction := runtime.SetMutexProfileFraction(-1)
			runtime.SetMutexProfileFraction(5)
			defer runtime.SetMutexProfileFraction(prevMutexProfileFraction)
			opt := &pprof.ProfileBuilderOptions{
				GenericsFrames: false,
				LazyMapping:    true,
			}
			dh := &pprof.DeltaMutexProfiler{}
			fs := generateBlockProfileRecords(512, 32, 239)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				PrintCountCycleProfile(dh, opt, io.Discard, scaler, fs)
			}
		})

	}
}
