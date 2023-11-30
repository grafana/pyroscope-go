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
	dh := pprof.DeltaHeapProfiler{}
	fs := generateMemProfileRecords(512, 32, 239)
	p1 := bytes.NewBuffer(nil)
	err := dh.WriteHeapProto(p1, fs, int64(runtime.MemProfileRate), "")
	assert.NoError(t, err)
	p1Size := p1.Len()
	profile, err := gprofile.Parse(p1)
	require.NoError(t, err)
	ls := stackCollapseProfile(t, profile)
	assert.Len(t, ls, 512)
	assert.Len(t, profile.Location, 141)
	t.Log("p1 size", p1Size)

	p2 := bytes.NewBuffer(nil)
	err = dh.WriteHeapProto(p2, fs, int64(runtime.MemProfileRate), "")
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
	dh := pprof.DeltaHeapProfiler{
		Options: pprof.ProfileBuilderOptions{
			GenericsFrames: false,
			LazyMapping:    true,
		},
	}
	fs := generateMemProfileRecords(512, 32, 239)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dh.WriteHeapProto(io.Discard, fs, int64(runtime.MemProfileRate), "")
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

			dh := pprof.DeltaMutexProfiler{}
			fs := generateBlockProfileRecords(512, 32, 239)
			p1 := bytes.NewBuffer(nil)
			err := dh.PrintCountCycleProfile(p1, "contentions", "delay", scaler, fs)
			assert.NoError(t, err)
			p1Size := p1.Len()
			profile, err := gprofile.Parse(p1)
			require.NoError(t, err)
			ls := stackCollapseProfile(t, profile)
			assert.Len(t, ls, 512)
			assert.Len(t, profile.Location, 141)
			t.Log("p1 size", p1Size)

			p2 := bytes.NewBuffer(nil)
			err = dh.PrintCountCycleProfile(p2, "contentions", "delay", scaler, fs)
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

			dh := pprof.DeltaMutexProfiler{
				Options: pprof.ProfileBuilderOptions{
					GenericsFrames: false,
					LazyMapping:    true,
				},
			}
			fs := generateBlockProfileRecords(512, 32, 239)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				dh.PrintCountCycleProfile(io.Discard, "contentions", "delay", scaler, fs)
			}
		})

	}
}
