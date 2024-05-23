package compat

import (
	"io"
	"math/rand"
	"runtime"
	"testing"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func BenchmarkHeapCompression(b *testing.B) {
	opt := &pprof.ProfileBuilderOptions{
		GenericsFrames: true,
		LazyMapping:    true,
	}
	dh := new(pprof.DeltaHeapProfiler)
	fs := generateMemProfileRecords(512, 32, 239)
	rng := rand.NewSource(239)
	objSize := fs[0].AllocBytes / fs[0].AllocObjects
	nMutations := int(uint(rng.Int63())) % len(fs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WriteHeapProto(dh, opt, io.Discard, fs, int64(runtime.MemProfileRate))
		for j := 0; j < nMutations; j++ {
			idx := int(uint(rng.Int63())) % len(fs)
			fs[idx].AllocObjects += 1
			fs[idx].AllocBytes += objSize
			fs[idx].FreeObjects += 1
			fs[idx].FreeBytes += objSize
		}
	}
}

func BenchmarkMutexCompression(b *testing.B) {
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
				GenericsFrames: true,
				LazyMapping:    true,
			}
			dh := new(pprof.DeltaMutexProfiler)
			fs := generateBlockProfileRecords(512, 32, 239)
			rng := rand.NewSource(239)
			nMutations := int(uint(rng.Int63())) % len(fs)
			oneBlockCycles := fs[0].Cycles / fs[0].Count
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = PrintCountCycleProfile(dh, opt, io.Discard, scaler, fs)
				for j := 0; j < nMutations; j++ {
					idx := int(uint(rng.Int63())) % len(fs)
					fs[idx].Count += 1
					fs[idx].Cycles += oneBlockCycles
				}
			}
		})

	}
}

func WriteHeapProto(dp *pprof.DeltaHeapProfiler, opt *pprof.ProfileBuilderOptions, w io.Writer, p []runtime.MemProfileRecord, rate int64) error {
	stc := pprof.HeapProfileConfig(rate)
	b := pprof.NewProfileBuilder(w, opt, stc)
	return dp.WriteHeapProto(b, p, rate)
}

func PrintCountCycleProfile(d *pprof.DeltaMutexProfiler, opt *pprof.ProfileBuilderOptions, w io.Writer, scaler pprof.MutexProfileScaler, records []runtime.BlockProfileRecord) error {
	stc := pprof.MutexProfileConfig()
	b := pprof.NewProfileBuilder(w, opt, stc)
	return d.PrintCountCycleProfile(b, scaler, records)
}
