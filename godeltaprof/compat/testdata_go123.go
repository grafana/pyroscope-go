//go:build !go1.27
// +build !go1.27

package compat

import (
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func (h *heapTestHelper) generateMemProfileRecords(n, depth int) []pprof.MemProfileRecord {
	records := make([]pprof.MemProfileRecord, 0, n)

	fs := getFunctionPointers()
	for range n {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		stack := make([]uintptr, depth)
		for j := range depth {
			stack[j] = fs[int(uint64(h.rng.Int63()))%len(fs)] //nolint:gosec
		}
		records = append(records, pprof.MemProfileRecord{
			AllocObjects: int64(nobj),
			AllocBytes:   int64(nobj * 1024),
			FreeObjects:  int64(nobj), // pretend inuse is zero
			FreeBytes:    int64(nobj * 1024),
			Stack:        stack,
		})
	}

	return records
}

func (h *mutexTestHelper) generateBlockProfileRecords(n, depth int) []pprof.BlockProfileRecord {
	records := make([]pprof.BlockProfileRecord, 0, n)
	fs := getFunctionPointers()
	for range n {
		nobj := int(uint64(h.rng.Int63())) % 1000000 //nolint:gosec
		stack := make([]uintptr, depth)
		for j := range depth {
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

// stackFromArray copies the prefix of s up to (but not including) the first
// zero entry into a fresh slice — matching the convention runtime.MemProfileRecord.Stack()
// uses on Stack0.
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
	return pprof.MemProfileRecord{
		AllocObjects: allocObjects,
		AllocBytes:   allocBytes,
		FreeBytes:    freeBytes,
		FreeObjects:  freeObjects,
		Stack:        stackFromArray(s),
	}
}

func (h *heapTestHelper) mutate(nmutations int, fs []pprof.MemProfileRecord) {
	objSize := fs[0].AllocBytes / fs[0].AllocObjects
	for range nmutations {
		idx := int(uint(h.rng.Int63())) % len(fs) //nolint:gosec
		fs[idx].AllocObjects += 1
		fs[idx].AllocBytes += objSize
		fs[idx].FreeObjects += 1
		fs[idx].FreeBytes += objSize
	}
}
