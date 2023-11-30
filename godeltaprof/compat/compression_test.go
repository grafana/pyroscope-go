package compat

import (
	"io"
	"math/rand"
	"runtime"
	"testing"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func BenchmarkHeapCompression(b *testing.B) {
	dh := pprof.DeltaHeapProfiler{}
	fs := generateMemProfileRecords(512, 32, 239)
	rng := rand.NewSource(239)
	objSize := fs[0].AllocBytes / fs[0].AllocObjects
	nMutations := int(uint(rng.Int63())) % len(fs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dh.WriteHeapProto(io.Discard, fs, int64(runtime.MemProfileRate), "")
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

			dh := pprof.DeltaMutexProfiler{}
			fs := generateBlockProfileRecords(512, 32, 239)
			rng := rand.NewSource(239)
			nMutations := int(uint(rng.Int63())) % len(fs)
			oneBlockCycles := fs[0].Cycles / fs[0].Count
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = dh.PrintCountCycleProfile(io.Discard, "contentions", "delay", scaler, fs)
				for j := 0; j < nMutations; j++ {
					idx := int(uint(rng.Int63())) % len(fs)
					fs[idx].Count += 1
					fs[idx].Cycles += oneBlockCycles
				}
			}
		})

	}
}
