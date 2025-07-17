package compat

import (
	"bytes"
	"io"
	"runtime"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func TestHeapReject(t *testing.T) {
	h := newHeapTestHelper()
	fs := h.generateMemProfileRecords(512, 32)
	p1 := bytes.NewBuffer(nil)
	err := WriteHeapProto(h.dp, h.opt, p1, fs, int64(runtime.MemProfileRate))
	assert.NoError(t, err)
	p1Size := p1.Len()
	profile, err := gprofile.Parse(p1)
	require.NoError(t, err)
	ls := stackCollapseProfile(profile)
	assert.Len(t, ls, 512)
	assert.Len(t, profile.Location, 141)
	t.Log("p1 size", p1Size)

	p2 := bytes.NewBuffer(nil)
	err = WriteHeapProto(h.dp, h.opt, p2, fs, int64(runtime.MemProfileRate))
	assert.NoError(t, err)
	p2Size := p2.Len()
	assert.Less(t, p2Size, 1000)
	profile, err = gprofile.Parse(p2)
	require.NoError(t, err)
	ls = stackCollapseProfile(profile)
	assert.Empty(t, ls)
	assert.Empty(t, profile.Location)
	t.Log("p2 size", p2Size)
}

func BenchmarkHeapRejectOrder(b *testing.B) {
	h := newHeapTestHelper()
	fs := h.generateMemProfileRecords(512, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteHeapProto(h.dp, h.opt, io.Discard, fs, int64(runtime.MemProfileRate))
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

			h := newMutexTestHelper()
			h.scaler = scaler
			fs := h.generateBlockProfileRecords(512, 32)
			p1 := bytes.NewBuffer(nil)
			err := PrintCountCycleProfile(h.dp, h.opt, p1, scaler, fs)
			assert.NoError(t, err)
			profile, err := gprofile.Parse(p1)
			require.NoError(t, err)
			ls := stackCollapseProfile(profile)
			assert.Len(t, ls, 512)
			assert.Len(t, profile.Location, 141)

			p2 := bytes.NewBuffer(nil)
			err = PrintCountCycleProfile(h.dp, h.opt, p2, scaler, fs)
			assert.NoError(t, err)
			p2Size := p2.Len()
			assert.Less(t, p2Size, 1000)
			profile, err = gprofile.Parse(p2)
			require.NoError(t, err)
			ls = stackCollapseProfile(profile)
			assert.Empty(t, ls)
			assert.Empty(t, profile.Location)
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
			h := newMutexTestHelper()
			h.scaler = scaler
			fs := h.generateBlockProfileRecords(512, 32)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				PrintCountCycleProfile(h.dp, h.opt, io.Discard, scaler, fs)
			}
		})
	}
}
