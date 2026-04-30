//go:build go1.27
// +build go1.27

package compat

import (
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func (h *heapTestHelper) generateMemProfileRecords(n, depth int) []pprof.MemProfileRecord {
	var records []pprof.MemProfileRecord

	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		stack := make([]uintptr, depth)
		for j := 0; j < depth; j++ {
			stack[j] = fs[int(uint64(h.rng.Int63()))%len(fs)] //nolint:gosec
		}
		records = append(records, pprof.MemProfileRecord{
			ObjectSize:   1024,
			AllocObjects: int64(nobj),
			FreeObjects:  int64(nobj), // pretend inuse is zero
			Stack:        stack,
		})
	}

	return records
}

func (h *mutexTestHelper) generateBlockProfileRecords(n, depth int) []pprof.BlockProfileRecord {
	var records []pprof.BlockProfileRecord
	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		stack := make([]uintptr, depth)
		for j := 0; j < depth; j++ {
			stack[j] = fs[int(uint64(h.rng.Int63()))%len(fs)] //nolint:gosec
		}
		records = append(records, pprof.BlockProfileRecord{
			Count:  int64(nobj),
			Cycles: int64(nobj * 10),
			Stack:  stack,
		})
	}

	return records
}

// stackFromArray copies the prefix of s up to the first zero entry into a
// fresh slice — matching runtime.MemProfileRecord.Stack0 -> Stack() behavior.
func stackFromArray(s [32]uintptr) []uintptr {
	for i, v := range s {
		if v == 0 {
			out := make([]uintptr, i)
			copy(out, s[:i])

			return out
		}
	}
	out := make([]uintptr, len(s))
	copy(out, s[:])

	return out
}

func (h *mutexTestHelper) r(count, cycles int64, s [32]uintptr) pprof.BlockProfileRecord {
	return pprof.BlockProfileRecord{
		Count:  count,
		Cycles: cycles,
		Stack:  stackFromArray(s),
	}
}

func (h *heapTestHelper) r(allocObjects, allocBytes, freeObjects, freeBytes int64,
	s [32]uintptr) pprof.MemProfileRecord {
	var size int64
	switch {
	case allocObjects > 0:
		size = allocBytes / allocObjects
	case freeObjects > 0:
		size = freeBytes / freeObjects
	}
	return pprof.MemProfileRecord{
		ObjectSize:   size,
		AllocObjects: allocObjects,
		FreeObjects:  freeObjects,
		Stack:        stackFromArray(s),
	}
}

func (h *heapTestHelper) mutate(nmutations int, fs []pprof.MemProfileRecord) {
	for j := 0; j < nmutations; j++ {
		idx := int(uint(h.rng.Int63())) % len(fs) //nolint:gosec
		fs[idx].AllocObjects += 1
		fs[idx].FreeObjects += 1
	}
}
