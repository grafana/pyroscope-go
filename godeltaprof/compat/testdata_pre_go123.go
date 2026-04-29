//go:build !go1.23
// +build !go1.23

package compat

import (
	"runtime"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

//nolint:unparam
func (h *heapTestHelper) generateMemProfileRecords(n, depth int) []pprof.MemProfileRecord {
	var records []pprof.MemProfileRecord

	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		r := runtime.MemProfileRecord{
			AllocObjects: int64(nobj),
			AllocBytes:   int64(nobj * 1024),
			FreeObjects:  int64(nobj), // pretend inuse is zero
			FreeBytes:    int64(nobj * 1024),
		}
		for j := 0; j < depth; j++ {
			r.Stack0[j] = fs[int(uint64(h.rng.Int63()))%len(fs)] //nolint:gosec
		}
		records = append(records, r)
	}

	return records
}

//nolint:unparam
func (h *mutexTestHelper) generateBlockProfileRecords(n, depth int) []pprof.BlockProfileRecord {
	var records []pprof.BlockProfileRecord
	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		r := runtime.BlockProfileRecord{
			Count:  int64(nobj),
			Cycles: int64(nobj * 10),
		}
		for j := 0; j < depth; j++ {
			r.Stack0[j] = fs[int(uint64(h.rng.Int63()))%len(fs)] //nolint:gosec
		}
		records = append(records, r)
	}

	return records
}

func (h *mutexTestHelper) r(count, cycles int64, s [32]uintptr) pprof.BlockProfileRecord {
	return runtime.BlockProfileRecord{
		Count:  count,
		Cycles: cycles,
		StackRecord: runtime.StackRecord{
			Stack0: s,
		},
	}
}

func (h *heapTestHelper) r(allocObjects, allocBytes, freeObjects, freeBytes int64,
	s [32]uintptr) pprof.MemProfileRecord {
	return runtime.MemProfileRecord{
		AllocObjects: allocObjects,
		AllocBytes:   allocBytes,
		FreeBytes:    freeBytes,
		FreeObjects:  freeObjects,
		Stack0:       s,
	}
}
