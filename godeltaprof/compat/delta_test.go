package compat

import (
	"bytes"
	"math/rand"
	"runtime"
	"testing"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	"github.com/stretchr/testify/assert"
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

	dh := new(pprof.DeltaHeapProfiler)
	opt := new(pprof.ProfileBuilderOptions)
	dump := func(r ...runtime.MemProfileRecord) *bytes.Buffer {
		buf := bytes.NewBuffer(nil)
		err := WriteHeapProto(dh, opt, buf, r, testMemProfileRate)
		assert.NoError(t, err)
		return buf
	}
	r := func(AllocObjects, AllocBytes, FreeObjects, FreeBytes int64, s [32]uintptr) runtime.MemProfileRecord {
		return runtime.MemProfileRecord{
			AllocObjects: AllocObjects,
			AllocBytes:   AllocBytes,
			FreeBytes:    FreeBytes,
			FreeObjects:  FreeObjects,
			Stack0:       s,
		}
	}

	p1 := dump(
		r(0, 0, 0, 0, stack0),
		r(0, 0, 0, 0, stack1),
	)
	expectEmptyProfile(t, p1)

	p2 := dump(
		r(5, 5*testObjectSize, 0, 0, stack0),
		r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	expectStackFrames(t, p2, stack0Marker, 10, 3525422, 10, 3525422)
	expectStackFrames(t, p2, stack1Marker, 6, 2115253, 0, 0)

	for i := 0; i < 3; i++ {
		// if we write same data, stack0 is in use, stack1 should not be present
		p3 := dump(
			r(5, 5*testObjectSize, 0, 0, stack0),
			r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
		)
		expectStackFrames(t, p3, stack0Marker, 0, 0, 10, 3525422)
		expectNoStackFrames(t, p3, stack1Marker)
	}

	p4 := dump(
		r(5, 5*testObjectSize, 5, 5*testObjectSize, stack0),
		r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	expectEmptyProfile(t, p4)

	p5 := dump(
		r(8, 8*testObjectSize, 5, 5*testObjectSize, stack0),
		r(3, 3*testObjectSize, 3, 3*testObjectSize, stack1),
	)
	// note, this value depends on scale order, it currently tests the current implementation, but we may change it
	// to alloc objects to be scale(8) - scale(5) = 17-10 = 7
	expectStackFrames(t, p5, stack0Marker, 6, 2115253, 6, 2115253)
	expectNoStackFrames(t, p5, stack1Marker)
}

func TestDeltaBlockProfile(t *testing.T) {
	cpuGHz := float64(pprof.Runtime_cyclesPerSecond()) / 1e9

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

			scale := func(rcount, rcycles int64) (int64, int64) {
				count, nanosec := pprof.ScaleMutexProfile(scaler, rcount, float64(rcycles)/cpuGHz)
				inanosec := int64(nanosec)
				return count, inanosec
			}
			dump := func(r ...runtime.BlockProfileRecord) *bytes.Buffer {
				buf := bytes.NewBuffer(nil)
				err := PrintCountCycleProfile(dh, opt, buf, scaler, r)
				assert.NoError(t, err)
				return buf
			}
			r := func(count, cycles int64, s [32]uintptr) runtime.BlockProfileRecord {
				return runtime.BlockProfileRecord{
					Count:  count,
					Cycles: cycles,
					StackRecord: runtime.StackRecord{
						Stack0: s,
					},
				}
			}

			p1 := dump(
				r(0, 0, stack0),
				r(0, 0, stack1),
			)
			expectEmptyProfile(t, p1)

			const cycles = 42
			p2 := dump(
				r(239, 239*cycles, stack0),
				r(0, 0, stack1),
			)
			count0, nanos0 := scale(239, 239*cycles)
			expectStackFrames(t, p2, stack0Marker, count0, nanos0)
			expectNoStackFrames(t, p2, stack1Marker)

			for j := 0; j < 2; j++ {
				p3 := dump(
					r(239, 239*cycles, stack0),
					r(0, 0, stack1),
				)
				expectEmptyProfile(t, p3)
			}

			count1, nanos1 := scale(240, 240*cycles)
			p4 := dump(
				r(240, 240*cycles, stack0),
			)
			expectStackFrames(t, p4, stack0Marker, count1-count0, nanos1-nanos0)
			expectNoStackFrames(t, p4, stack1Marker)
		})
	}
}

func BenchmarkHeapDelta(b *testing.B) {
	dh := new(pprof.DeltaHeapProfiler)
	fs := generateMemProfileRecords(512, 32, 239)
	rng := rand.NewSource(239)
	objSize := fs[0].AllocBytes / fs[0].AllocObjects
	nMutations := int(uint(rng.Int63())) % len(fs)
	builder := &noopBuilder{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dh.WriteHeapProto(builder, fs, int64(runtime.MemProfileRate))
		for j := 0; j < nMutations; j++ {
			idx := int(uint(rng.Int63())) % len(fs)
			fs[idx].AllocObjects += 1
			fs[idx].AllocBytes += objSize
			fs[idx].FreeObjects += 1
			fs[idx].FreeBytes += objSize
		}
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

			dh := new(pprof.DeltaMutexProfiler)
			fs := generateBlockProfileRecords(512, 32, 239)
			rng := rand.NewSource(239)
			nMutations := int(uint(rng.Int63())) % len(fs)
			oneBlockCycles := fs[0].Cycles / fs[0].Count
			builder := &noopBuilder{}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = dh.PrintCountCycleProfile(builder, scaler, fs)
				for j := 0; j < nMutations; j++ {
					idx := int(uint(rng.Int63())) % len(fs)
					fs[idx].Count += 1
					fs[idx].Cycles += oneBlockCycles
				}
			}
		})

	}
}

type noopBuilder struct {
}

func (b *noopBuilder) LocsForStack(_ []uintptr) []uint64 {
	return nil
}
func (b *noopBuilder) Sample(_ []int64, _ []uint64, _ int64) {

}

func (b *noopBuilder) Build() {

}
