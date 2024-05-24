package compat

import (
	"fmt"
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var (
	stack0       = [32]uintptr{}
	stack0Marker string
	stack1       = [32]uintptr{}
	stack1Marker string
	stack2       = [32]uintptr{}
	stack2Marker string
	stack3       = [32]uintptr{}
	stack4       = [32]uintptr{}
)

func init() {
	fs := getFunctionPointers()

	stack0 = [32]uintptr{fs[0], fs[1]}
	stack1 = [32]uintptr{fs[2], fs[3]}
	stack2 = [32]uintptr{
		reflect.ValueOf(runtime.GC).Pointer(),
		reflect.ValueOf(runtime.FuncForPC).Pointer(),
		reflect.ValueOf(TestDeltaBlockProfile).Pointer(),
		reflect.ValueOf(TestDeltaHeap).Pointer(),
	}
	stack3 = [32]uintptr{ // equal , but difference in runtime
		reflect.ValueOf(runtime.GC).Pointer() + 1,
		reflect.ValueOf(runtime.FuncForPC).Pointer(),
		reflect.ValueOf(TestDeltaBlockProfile).Pointer(),
		reflect.ValueOf(TestDeltaHeap).Pointer(),
	}

	stack4 = [32]uintptr{ // equal , but difference in non runtime frame
		reflect.ValueOf(runtime.GC).Pointer(),
		reflect.ValueOf(runtime.FuncForPC).Pointer(),
		reflect.ValueOf(TestDeltaBlockProfile).Pointer() + 1,
		reflect.ValueOf(TestDeltaHeap).Pointer(),
	}
	marker := func(stk []uintptr) string {
		res := []string{}
		for i := range stk {
			f := stk[len(stk)-1-i]
			res = append(res, runtime.FuncForPC(f).Name())
		}
		return strings.Join(res, ";")
	}
	stack0Marker = marker(stack0[:2])
	stack1Marker = marker(stack1[:2])
	stack2Marker = marker(stack2[:2])
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
	h.rate = testMemProfileRate

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

func TestMutexDuplicates(t *testing.T) {
	h := newMutexTestHelper()
	const cycles = 42
	p := h.dump(
		h.r(239, 239*cycles, stack0),
		h.r(42, 42*cycles, stack1),
		h.r(7, 7*cycles, stack0),
	)

	expectStackFrames(t, p, stack0Marker, h.scale2(239+7, (239+7)*cycles)...)
	expectStackFrames(t, p, stack1Marker, h.scale2(42, (42)*cycles)...)

	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack0Marker), 1, h.scale2(239+7, (239+7)*cycles)...)
	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack1Marker), 1, h.scale2(42, 42*cycles)...)

	p = h.dump(
		h.r(239, 239*cycles, stack0),
		h.r(42, 42*cycles, stack1),
		h.r(7, 7*cycles, stack0),
	)
	expectEmptyProfile(t, p)
}

func TestHeapDuplicates(t *testing.T) {
	const testMemProfileRate = 524288
	h := newHeapTestHelper()
	h.rate = testMemProfileRate
	const blockSize = 1024
	const blockSize2 = 1024
	p := h.dump(
		h.r(239, 239*blockSize, 239, 239*blockSize, stack0),
		h.r(3, 3*blockSize2, 3, 3*blockSize2, stack0),
		h.r(42, 42*blockSize, 42, 42*blockSize, stack1),
		h.r(7, 7*blockSize, 7, 7*blockSize, stack0),
		h.r(3, 3*blockSize, 3, 3*blockSize, stack2),
		h.r(5, 5*blockSize, 5, 5*blockSize, stack3),
		h.r(11, 11*blockSize, 11, 11*blockSize, stack4),
	)
	scale := func(c, b int) []int64 {
		c1, b1 := pprof.ScaleHeapSample(int64(c), int64(b), testMemProfileRate)
		return []int64{c1, b1, 0, 0}
	}
	expectStackFrames(t, p, stack0Marker, scale(239+7, (239+7)*blockSize)...)
	expectStackFrames(t, p, stack1Marker, scale(42, 42*blockSize)...)

	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack0Marker), 1, scale(239+7, (239+7)*blockSize)...)
	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack1Marker), 1, scale(42, 42*blockSize)...)
	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack2Marker), 1, scale(3, 3*blockSize)...)
	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack2Marker), 1, scale(5, 5*blockSize)...)
	expectPPROFLocations(t, p, fmt.Sprintf("^%s$", stack2Marker), 1, scale(11, 11*blockSize)...)

	p = h.dump(
		h.r(239, 239*blockSize, 239, 239*blockSize, stack0),
		h.r(3, 3*blockSize2, 3, 3*blockSize2, stack0),
		h.r(42, 42*blockSize, 42, 42*blockSize, stack1),
		h.r(7, 7*blockSize, 7, 7*blockSize, stack0),
		h.r(3, 3*blockSize, 3, 3*blockSize, stack2),
		h.r(5, 5*blockSize, 5, 5*blockSize, stack3),
		h.r(11, 11*blockSize, 11, 11*blockSize, stack4),
	)
	expectEmptyProfile(t, p)
}
