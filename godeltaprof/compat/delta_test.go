package compat

import (
	"runtime"
	"testing"
)

var (
	stack0       = [32]uintptr{}
	stack0Marker string
	stack1       = [32]uintptr{}
	stack1Marker string
)

func init() {
	fs := getFunctionPointers()
	stack0 = [32]uintptr{fs[0], fs[1]}
	stack1 = [32]uintptr{fs[2], fs[3]}
	stack0Marker = runtime.FuncForPC(fs[1]).Name() + ";" + runtime.FuncForPC(fs[0]).Name()
	stack1Marker = runtime.FuncForPC(fs[3]).Name() + ";" + runtime.FuncForPC(fs[2]).Name()
}

func TestDeltaHeap(t *testing.T) {
	// scale 0 0 0
	// scale 1 2 705084
	// scale 2 4 1410169
	// scale 3 6 2115253
	// scale 4 8 2820338
	// scale 5 10 3525422
	// scale 6 12 4230507
	// scale 7 15 4935592
	// scale 8 17 5640676
	// scale 9 19 6345761

	const testMemProfileRate = 524288
	const testObjectSize = 327680

	h := newHeapTestHelper()

	p1 := h.dump(
		h.r(0, 0, 0, 0, stack0),
		h.r(0, 0, 0, 0, stack1),
	)
	expectEmptyProfile(t, p1)

	p2 := h.dump(
		h.r(5, 5*testObjectSize, 0, 0, stack0),
		h.r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	expectStackFrames(t, p2, stack0Marker, 10, 3525422, 10, 3525422)
	expectStackFrames(t, p2, stack1Marker, 6, 2115253, 0, 0)

	for i := 0; i < 3; i++ {
		// if we write same data, stack0 is in use, stack1 should not be present
		p3 := h.dump(
			h.r(5, 5*testObjectSize, 0, 0, stack0),
			h.r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
		)
		expectStackFrames(t, p3, stack0Marker, 0, 0, 10, 3525422)
		expectNoStackFrames(t, p3, stack1Marker)
	}

	p4 := h.dump(
		h.r(5, 5*testObjectSize, 5, 5*testObjectSize, stack0),
		h.r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	expectEmptyProfile(t, p4)

	p5 := h.dump(
		h.r(8, 8*testObjectSize, 5, 5*testObjectSize, stack0),
		h.r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	// note, this value depends on scale order, it currently tests the current implementation, but we may change it
	// to alloc objects to be scale(8) - scale(5) = 17-10 = 7
	expectStackFrames(t, p5, stack0Marker, 6, 2115253, 6, 2115253)
	expectNoStackFrames(t, p5, stack1Marker)
}

func TestDeltaBlockProfile(t *testing.T) {
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

			p1 := h.dump(
				h.r(0, 0, stack0),
				h.r(0, 0, stack1),
			)
			expectEmptyProfile(t, p1)

			const cycles = 42
			p2 := h.dump(
				h.r(239, 239*cycles, stack0),
				h.r(0, 0, stack1),
			)
			count0, nanos0 := h.scale(239, 239*cycles)
			expectStackFrames(t, p2, stack0Marker, count0, nanos0)
			expectNoStackFrames(t, p2, stack1Marker)

			for j := 0; j < 2; j++ {
				p3 := h.dump(
					h.r(239, 239*cycles, stack0),
					h.r(0, 0, stack1),
				)
				expectEmptyProfile(t, p3)
			}

			count1, nanos1 := h.scale(240, 240*cycles)
			p4 := h.dump(
				h.r(240, 240*cycles, stack0),
			)
			expectStackFrames(t, p4, stack0Marker, count1-count0, nanos1-nanos0)
			expectNoStackFrames(t, p4, stack1Marker)
		})
	}
}

func BenchmarkHeapDelta(b *testing.B) {
	h := newHeapTestHelper()
	fs := h.generateMemProfileRecords(512, 32)
	builder := &noopBuilder{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.dp.WriteHeapProto(builder, fs, int64(runtime.MemProfileRate))
		h.mutate(fs)
	}
}

func BenchmarkMutexDelta(b *testing.B) {
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
			builder := &noopBuilder{}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = h.dp.PrintCountCycleProfile(builder, scaler, fs)
				h.mutate(fs)
			}
		})

	}
}
