package compat

import (
	"io"
	"runtime"
	"testing"
)

const (
	scalerMutexProfileName = "ScalerMutexProfile"
	scalerBlockProfileName = "ScalerBlockProfile"
)

func BenchmarkHeapCompression(b *testing.B) {
	h := newHeapTestHelper()
	fs := h.generateMemProfileRecords(512, 32)
	h.rng.Seed(239)
	nmutations := int(h.rng.Int63() % int64(len(fs)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i == 1000 {
			v := h.rng.Int63()
			if v != 7817861117094116717 {
				b.Errorf("unexpected random value: %d. "+
					"The bench should be deterministic for better comparison.", v)
			}
		}
		_ = WriteHeapProto(h.dp, h.opt, io.Discard, fs, int64(runtime.MemProfileRate))
		h.mutate(nmutations, fs)
	}
}

func BenchmarkMutexCompression(b *testing.B) {
	for i, scaler := range mutexProfileScalers {
		name := scalerMutexProfileName
		if i == 1 {
			name = scalerBlockProfileName
		}
		b.Run(name, func(b *testing.B) {
			prevMutexProfileFraction := runtime.SetMutexProfileFraction(-1)
			runtime.SetMutexProfileFraction(5)
			defer runtime.SetMutexProfileFraction(prevMutexProfileFraction)

			h := newMutexTestHelper()
			h.scaler = scaler
			fs := h.generateBlockProfileRecords(512, 32)
			h.rng.Seed(239)
			nmutations := int(h.rng.Int63() % int64(len(fs)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if i == 1000 {
					v := h.rng.Int63()
					if v != 7817861117094116717 {
						b.Errorf("unexpected random value: %d. "+
							"The bench should be deterministic for better comparison.", v)
					}
				}
				_ = PrintCountCycleProfile(h.dp, h.opt, io.Discard, scaler, fs)
				h.mutate(nmutations, fs)
			}
		})
	}
}
