//go:build go1.23
// +build go1.23

package pprof

import _ "unsafe"

// MemProfileRecord mirrors internal/profilerecord.MemProfileRecord layout.
// The runtime writes into these via //go:linkname to pprof_memProfileInternal,
// so the field layout MUST match the runtime's definition exactly.
type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64
	AllocObjects, FreeObjects int64
	Stack                     []uintptr
}

func (r *MemProfileRecord) InUseObjects() int64 { return r.AllocObjects - r.FreeObjects }

type BlockProfileRecord struct {
	Count  int64
	Cycles int64
	Stack  []uintptr
}

func memRecordStack(r *MemProfileRecord) []uintptr     { return r.Stack }
func blockRecordStack(r *BlockProfileRecord) []uintptr { return r.Stack }

//go:linkname pprof_memProfileInternal runtime.pprof_memProfileInternal
func pprof_memProfileInternal(p []MemProfileRecord, inuseZero bool) (n int, ok bool)

//go:linkname pprof_blockProfileInternal runtime.pprof_blockProfileInternal
func pprof_blockProfileInternal(p []BlockProfileRecord) (n int, ok bool)

//go:linkname pprof_mutexProfileInternal runtime.pprof_mutexProfileInternal
func pprof_mutexProfileInternal(p []BlockProfileRecord) (n int, ok bool)

func MemProfile(inuseZero bool) []MemProfileRecord {
	var p []MemProfileRecord
	n, _ := pprof_memProfileInternal(nil, inuseZero)
	for {
		p = make([]MemProfileRecord, n+50)
		var ok bool
		n, ok = pprof_memProfileInternal(p, inuseZero)
		if ok {
			return p[:n]
		}
	}
}

func BlockProfile() []BlockProfileRecord { return fetchBlockLike(pprof_blockProfileInternal) }
func MutexProfile() []BlockProfileRecord { return fetchBlockLike(pprof_mutexProfileInternal) }

func fetchBlockLike(f func([]BlockProfileRecord) (int, bool)) []BlockProfileRecord {
	var p []BlockProfileRecord
	n, _ := f(nil)
	for {
		p = make([]BlockProfileRecord, n+50)
		var ok bool
		n, ok = f(p)
		if ok {
			return p[:n]
		}
	}
}
