//go:build !go1.23
// +build !go1.23

package pprof

import "runtime"

type MemProfileRecord = runtime.MemProfileRecord
type BlockProfileRecord = runtime.BlockProfileRecord

func memRecordStack(r *MemProfileRecord) []uintptr     { return r.Stack() }
func blockRecordStack(r *BlockProfileRecord) []uintptr { return r.Stack() }

func MemProfile(inuseZero bool) []MemProfileRecord {
	var p []MemProfileRecord
	n, _ := runtime.MemProfile(nil, inuseZero)
	for {
		p = make([]MemProfileRecord, n+50)
		var ok bool
		n, ok = runtime.MemProfile(p, inuseZero)
		if ok {
			return p[:n]
		}
	}
}

func BlockProfile() []BlockProfileRecord { return fetchBlockLike(runtime.BlockProfile) }
func MutexProfile() []BlockProfileRecord { return fetchBlockLike(runtime.MutexProfile) }

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
