package pprof

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"strings"
)

type DeltaHeapProfiler struct {
	m profMap
}

// WriteHeapProto writes the current heap profile in protobuf format to w.
func (d *DeltaHeapProfiler) WriteHeapProto(w io.Writer, p []runtime.MemProfileRecord, rate int64, defaultSampleType string) error {
	b := newProfileBuilder(w)
	b.pbValueType(tagProfile_PeriodType, "space", "bytes")
	b.pb.int64Opt(tagProfile_Period, rate)
	b.pbValueType(tagProfile_SampleType, "alloc_objects", "count")
	b.pbValueType(tagProfile_SampleType, "alloc_space", "bytes")
	b.pbValueType(tagProfile_SampleType, "inuse_objects", "count")
	b.pbValueType(tagProfile_SampleType, "inuse_space", "bytes")
	if defaultSampleType != "" {
		b.pb.int64Opt(tagProfile_DefaultSampleType, b.stringIndex(defaultSampleType))
	}

	values := []int64{0, 0, 0, 0}
	var locs []uint64
	for _, r := range p {
		hideRuntime := true
		for tries := 0; tries < 2; tries++ {
			stk := r.Stack()
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
			locs = b.appendLocsForStack(locs[:0], stk)
			if len(locs) > 0 {
				break
			}
			hideRuntime = false // try again, and show all frames next time.
		}

		// do the delta
		pp := func(prefix string, stk []uintptr) {
			frames1 := runtime.CallersFrames(stk)
			for {
				f, more := frames1.Next()
				if !more {
					break
				}
				fmt.Printf("[dhp] %s %s\n", prefix, f.Func.Name())
			}
		}
		if r.AllocBytes == 0 && r.AllocObjects == 0 && r.FreeObjects == 0 && r.FreeBytes == 0 {
			// it is a fresh bucket and it will be published after next 1-2 gc cycles
			continue
		}
		size := int64(0)
		if r.AllocObjects != 0 {
			if r.AllocBytes%r.AllocObjects != 0 {
				fmt.Printf("[dhp] warning: %d %d\n", r.AllocObjects, r.AllocBytes)
				pp("mod", r.Stack())
			}
			size = r.AllocBytes / r.AllocObjects
		} else {
			fmt.Printf("[dhp] zero\n")
			pp("zero", r.Stack())
		}
		entry := d.m.Lookup(r.Stack(), uintptr(size))

		if (r.AllocObjects - entry.count.v1) < 0 {
			pp("neg 1", r.Stack())
			pp("neg 2", entry.stk[:])
			continue
		}
		AllocObjects := r.AllocObjects - entry.count.v1
		AllocBytes := r.AllocBytes - entry.count.v2
		entry.count.v1 = r.AllocObjects
		entry.count.v2 = r.AllocBytes

		values[0], values[1] = scaleHeapSample(AllocObjects, AllocBytes, rate)
		values[2], values[3] = scaleHeapSample(r.InUseObjects(), r.InUseBytes(), rate)

		if values[0] == 0 && values[1] == 0 && values[2] == 0 && values[3] == 0 {
			continue
		}
		//var blockSize int64
		//if r.AllocObjects > 0 {
		//	blockSize = r.AllocBytes / r.AllocObjects
		//}
		// todo should we skip if values are 0?
		b.pbSample(values, locs, func() {
			//if blockSize != 0 {
			//	b.pbLabel(tagSample_Label, "bytes", "", blockSize)
			//}
		})
	}
	b.build()
	return nil
}

// scaleHeapSample adjusts the data from a heap Sample to
// account for its probability of appearing in the collected
// data. heap profiles are a sampling of the memory allocations
// requests in a program. We estimate the unsampled value by dividing
// each collected sample by its probability of appearing in the
// profile. heap profiles rely on a poisson process to determine
// which samples to collect, based on the desired average collection
// rate R. The probability of a sample of size S to appear in that
// profile is 1-exp(-S/R).
func scaleHeapSample(count, size, rate int64) (int64, int64) {
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
