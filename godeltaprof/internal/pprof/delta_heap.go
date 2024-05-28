package pprof

import (
	"math"
	"runtime"
	"strings"
)

type heapPrevValue struct {
	allocObjects int64
}

type heapAccValue struct {
	allocObjects int64
	inuseObjects int64
}

type DeltaHeapProfiler struct {
	m profMap[heapPrevValue, heapAccValue]
	//todo consider adding an option to remove block size label and merge allocations of different size
}

// WriteHeapProto writes the current heap profile in protobuf format to w.
func (d *DeltaHeapProfiler) WriteHeapProto(b ProfileBuilder, p []runtime.MemProfileRecord, rate int64) error {
	values := []int64{0, 0, 0, 0}
	var locs []uint64
	for i := range p {
		r := &p[i]
		if r.AllocBytes == 0 && r.AllocObjects == 0 && r.FreeObjects == 0 && r.FreeBytes == 0 {
			// it is a fresh bucket and it will be published after next 1-2 gc cycles
			continue
		}
		var blockSize int64
		if r.AllocObjects > 0 {
			blockSize = r.AllocBytes / r.AllocObjects
		}
		entry := d.m.Lookup(stack(r.Stack0[:]), uintptr(blockSize))
		entry.acc.allocObjects += r.AllocObjects
		entry.acc.inuseObjects += r.InUseObjects()
	}
	for i := range p {
		r := &p[i]
		// do the delta
		if r.AllocBytes == 0 && r.AllocObjects == 0 && r.FreeObjects == 0 && r.FreeBytes == 0 {
			// it is a fresh bucket and it will be published after next 1-2 gc cycles
			continue
		}
		var blockSize int64
		if r.AllocObjects > 0 {
			blockSize = r.AllocBytes / r.AllocObjects
		}
		entry := d.m.Lookup(stack(r.Stack0[:]), uintptr(blockSize))
		if entry.acc == (heapAccValue{}) {
			continue
		}

		AllocObjects := entry.acc.allocObjects - entry.prev.allocObjects
		if AllocObjects < 0 {
			continue
		}
		AllocBytes := AllocObjects * blockSize
		entry.prev.allocObjects = entry.acc.allocObjects

		values[0], values[1] = ScaleHeapSample(AllocObjects, AllocBytes, rate)
		values[2], values[3] = ScaleHeapSample(entry.acc.inuseObjects, entry.acc.inuseObjects*blockSize, rate)

		entry.acc = heapAccValue{}

		if values[0] == 0 && values[1] == 0 && values[2] == 0 && values[3] == 0 {
			continue
		}

		hideRuntime := true
		for tries := 0; tries < 2; tries++ {
			stk := stack(r.Stack0[:])
			// For heap profiles, all stack
			// addresses are return PCs, which is
			// what appendLocsForStack expects.
			if hideRuntime {
				for i, addr := range stk {
					if f := runtime.FuncForPC(addr); f != nil && strings.HasPrefix(f.Name(), "runtime.") {
						continue
					}
					// Found non-runtime. Show any runtime uses above it.
					stk = stk[i:]
					break
				}
			}
			locs = b.LocsForStack(stk)
			if len(locs) > 0 {
				break
			}
			hideRuntime = false // try again, and show all frames next time.
		}

		b.Sample(values, locs, blockSize)
	}
	b.Build()
	return nil
}

// ScaleHeapSample adjusts the data from a heap Sample to
// account for its probability of appearing in the collected
// data. heap profiles are a sampling of the memory allocations
// requests in a program. We estimate the unsampled value by dividing
// each collected sample by its probability of appearing in the
// profile. heap profiles rely on a poisson process to determine
// which samples to collect, based on the desired average collection
// rate R. The probability of a sample of size S to appear in that
// profile is 1-exp(-S/R).
func ScaleHeapSample(count, size, rate int64) (int64, int64) {
	if count == 0 || size == 0 {
		return 0, 0
	}

	if rate <= 1 {
		// if rate==1 all samples were collected so no adjustment is needed.
		// if rate<1 treat as unknown and skip scaling.
		return count, size
	}

	avgSize := float64(size) / float64(count)
	scale := 1 / (1 - math.Exp(-avgSize/float64(rate)))

	return int64(float64(count) * scale), int64(float64(size) * scale)
}

func HeapProfileConfig(rate int64) ProfileConfig {
	return ProfileConfig{
		PeriodType: ValueType{Typ: "space", Unit: "bytes"},
		Period:     rate,
		SampleType: []ValueType{
			{"alloc_objects", "count"},
			{"alloc_space", "bytes"},
			{"inuse_objects", "count"},
			{"inuse_space", "bytes"},
		},
		DefaultSampleType: "",
	}
}

func stack(stk []uintptr) []uintptr {
	for i, v := range stk {
		if v == 0 {
			return stk[0:i]
		}
	}
	return stk
}
