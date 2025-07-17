package compat

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"runtime"

	"github.com/stretchr/testify/assert"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
)

func getFunctionPointers() []uintptr {
	return []uintptr{
		reflect.ValueOf(assert.Truef).Pointer(),
		reflect.ValueOf(assert.CallerInfo).Pointer(),
		reflect.ValueOf(assert.Condition).Pointer(),
		reflect.ValueOf(assert.Conditionf).Pointer(),
		reflect.ValueOf(assert.Contains).Pointer(),
		reflect.ValueOf(assert.Containsf).Pointer(),
		reflect.ValueOf(assert.DirExists).Pointer(),
		reflect.ValueOf(assert.DirExistsf).Pointer(),
		reflect.ValueOf(assert.ElementsMatch).Pointer(),
		reflect.ValueOf(assert.ElementsMatchf).Pointer(),
		reflect.ValueOf(assert.Empty).Pointer(),
		reflect.ValueOf(assert.Emptyf).Pointer(),
		reflect.ValueOf(assert.Equal).Pointer(),
		reflect.ValueOf(assert.EqualError).Pointer(),
		reflect.ValueOf(assert.EqualErrorf).Pointer(),
		reflect.ValueOf(assert.EqualValues).Pointer(),
		reflect.ValueOf(assert.EqualValuesf).Pointer(),
		reflect.ValueOf(assert.Equalf).Pointer(),
		reflect.ValueOf(assert.Error).Pointer(),
		reflect.ValueOf(assert.ErrorAs).Pointer(),
		reflect.ValueOf(assert.ErrorAsf).Pointer(),
		reflect.ValueOf(assert.ErrorIs).Pointer(),
		reflect.ValueOf(assert.ErrorIsf).Pointer(),
		reflect.ValueOf(assert.Errorf).Pointer(),
		reflect.ValueOf(assert.Eventually).Pointer(),
		reflect.ValueOf(assert.Eventuallyf).Pointer(),
		reflect.ValueOf(assert.Exactly).Pointer(),
		reflect.ValueOf(assert.Exactlyf).Pointer(),
		reflect.ValueOf(assert.Fail).Pointer(),
		reflect.ValueOf(assert.FailNow).Pointer(),
		reflect.ValueOf(assert.FailNowf).Pointer(),
		reflect.ValueOf(assert.Failf).Pointer(),
		reflect.ValueOf(assert.False).Pointer(),
		reflect.ValueOf(assert.Falsef).Pointer(),
		reflect.ValueOf(assert.FileExists).Pointer(),
		reflect.ValueOf(assert.FileExistsf).Pointer(),
		reflect.ValueOf(assert.Greater).Pointer(),
		reflect.ValueOf(assert.GreaterOrEqual).Pointer(),
		reflect.ValueOf(assert.GreaterOrEqualf).Pointer(),
		reflect.ValueOf(assert.Greaterf).Pointer(),
		reflect.ValueOf(assert.HTTPBody).Pointer(),
		reflect.ValueOf(assert.HTTPBodyContains).Pointer(),
		reflect.ValueOf(assert.HTTPBodyContainsf).Pointer(),
		reflect.ValueOf(assert.HTTPBodyNotContains).Pointer(),
		reflect.ValueOf(assert.HTTPBodyNotContainsf).Pointer(),
		reflect.ValueOf(assert.HTTPError).Pointer(),
		reflect.ValueOf(assert.HTTPErrorf).Pointer(),
		reflect.ValueOf(assert.HTTPRedirect).Pointer(),
		reflect.ValueOf(assert.HTTPRedirectf).Pointer(),
		reflect.ValueOf(assert.HTTPStatusCode).Pointer(),
		reflect.ValueOf(assert.HTTPStatusCodef).Pointer(),
		reflect.ValueOf(assert.HTTPSuccess).Pointer(),
		reflect.ValueOf(assert.HTTPSuccessf).Pointer(),
		reflect.ValueOf(assert.Implements).Pointer(),
		reflect.ValueOf(assert.Implementsf).Pointer(),
		reflect.ValueOf(assert.InDelta).Pointer(),
		reflect.ValueOf(assert.InDeltaMapValues).Pointer(),
		reflect.ValueOf(assert.InDeltaMapValuesf).Pointer(),
		reflect.ValueOf(assert.InDeltaSlice).Pointer(),
		reflect.ValueOf(assert.InDeltaSlicef).Pointer(),
		reflect.ValueOf(assert.InDeltaf).Pointer(),
		reflect.ValueOf(assert.InEpsilon).Pointer(),
		reflect.ValueOf(assert.InEpsilonSlice).Pointer(),
		reflect.ValueOf(assert.InEpsilonSlicef).Pointer(),
		reflect.ValueOf(assert.InEpsilonf).Pointer(),
		reflect.ValueOf(assert.IsDecreasing).Pointer(),
		reflect.ValueOf(assert.IsDecreasingf).Pointer(),
		reflect.ValueOf(assert.IsIncreasing).Pointer(),
		reflect.ValueOf(assert.IsIncreasingf).Pointer(),
		reflect.ValueOf(assert.IsNonDecreasing).Pointer(),
		reflect.ValueOf(assert.IsNonDecreasingf).Pointer(),
		reflect.ValueOf(assert.IsNonIncreasing).Pointer(),
		reflect.ValueOf(assert.IsNonIncreasingf).Pointer(),
		reflect.ValueOf(assert.IsType).Pointer(),
		reflect.ValueOf(assert.IsTypef).Pointer(),
		reflect.ValueOf(assert.JSONEq).Pointer(),
		reflect.ValueOf(assert.JSONEqf).Pointer(),
		reflect.ValueOf(assert.Len).Pointer(),
		reflect.ValueOf(assert.Lenf).Pointer(),
		reflect.ValueOf(assert.Less).Pointer(),
		reflect.ValueOf(assert.LessOrEqual).Pointer(),
		reflect.ValueOf(assert.LessOrEqualf).Pointer(),
		reflect.ValueOf(assert.Lessf).Pointer(),
		reflect.ValueOf(assert.Negative).Pointer(),
		reflect.ValueOf(assert.Negativef).Pointer(),
		reflect.ValueOf(assert.Never).Pointer(),
		reflect.ValueOf(assert.Neverf).Pointer(),
		reflect.ValueOf(assert.New).Pointer(),
		reflect.ValueOf(assert.Nil).Pointer(),
		reflect.ValueOf(assert.Nilf).Pointer(),
		reflect.ValueOf(assert.NoDirExists).Pointer(),
		reflect.ValueOf(assert.NoDirExistsf).Pointer(),
		reflect.ValueOf(assert.NoError).Pointer(),
		reflect.ValueOf(assert.NoErrorf).Pointer(),
		reflect.ValueOf(assert.NoFileExists).Pointer(),
		reflect.ValueOf(assert.NoFileExistsf).Pointer(),
		reflect.ValueOf(assert.NotContains).Pointer(),
		reflect.ValueOf(assert.NotContainsf).Pointer(),
		reflect.ValueOf(assert.NotEmpty).Pointer(),
		reflect.ValueOf(assert.NotEmptyf).Pointer(),
		reflect.ValueOf(assert.NotEqual).Pointer(),
		reflect.ValueOf(assert.NotEqualValues).Pointer(),
		reflect.ValueOf(assert.NotEqualValuesf).Pointer(),
		reflect.ValueOf(assert.NotEqualf).Pointer(),
		reflect.ValueOf(assert.NotErrorIs).Pointer(),
		reflect.ValueOf(assert.NotErrorIsf).Pointer(),
		reflect.ValueOf(assert.NotNil).Pointer(),
		reflect.ValueOf(assert.NotNilf).Pointer(),
		reflect.ValueOf(assert.NotPanics).Pointer(),
		reflect.ValueOf(assert.NotPanicsf).Pointer(),
		reflect.ValueOf(assert.NotRegexp).Pointer(),
		reflect.ValueOf(assert.NotRegexpf).Pointer(),
		reflect.ValueOf(assert.NotSame).Pointer(),
		reflect.ValueOf(assert.NotSamef).Pointer(),
		reflect.ValueOf(assert.NotSubset).Pointer(),
		reflect.ValueOf(assert.NotSubsetf).Pointer(),
		reflect.ValueOf(assert.NotZero).Pointer(),
		reflect.ValueOf(assert.NotZerof).Pointer(),
		reflect.ValueOf(assert.ObjectsAreEqual).Pointer(),
		reflect.ValueOf(assert.ObjectsAreEqualValues).Pointer(),
		reflect.ValueOf(assert.Panics).Pointer(),
		reflect.ValueOf(assert.PanicsWithError).Pointer(),
		reflect.ValueOf(assert.PanicsWithErrorf).Pointer(),
		reflect.ValueOf(assert.PanicsWithValue).Pointer(),
		reflect.ValueOf(assert.PanicsWithValuef).Pointer(),
		reflect.ValueOf(assert.Panicsf).Pointer(),
		reflect.ValueOf(assert.Positive).Pointer(),
		reflect.ValueOf(assert.Positivef).Pointer(),
		reflect.ValueOf(assert.Regexp).Pointer(),
		reflect.ValueOf(assert.Regexpf).Pointer(),
		reflect.ValueOf(assert.Same).Pointer(),
		reflect.ValueOf(assert.Samef).Pointer(),
		reflect.ValueOf(assert.Subset).Pointer(),
		reflect.ValueOf(assert.Subsetf).Pointer(),
		reflect.ValueOf(assert.True).Pointer(),
		reflect.ValueOf(assert.Truef).Pointer(),
		reflect.ValueOf(assert.WithinDuration).Pointer(),
		reflect.ValueOf(assert.WithinDurationf).Pointer(),
		reflect.ValueOf(assert.YAMLEq).Pointer(),
		reflect.ValueOf(assert.YAMLEqf).Pointer(),
		reflect.ValueOf(assert.Zero).Pointer(),
		reflect.ValueOf(assert.Zerof).Pointer(),
	}
}

//nolint:unparam
func (h *heapTestHelper) generateMemProfileRecords(n, depth int) []runtime.MemProfileRecord {
	var records []runtime.MemProfileRecord

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
func (h *mutexTestHelper) generateBlockProfileRecords(n, depth int) []runtime.BlockProfileRecord {
	var records []runtime.BlockProfileRecord
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

type mutexTestHelper struct {
	dp     *pprof.DeltaMutexProfiler
	opt    *pprof.ProfileBuilderOptions
	scaler pprof.MutexProfileScaler
	rng    rand.Source
}

func newMutexTestHelper() *mutexTestHelper {
	res := &mutexTestHelper{
		dp: &pprof.DeltaMutexProfiler{},
		opt: &pprof.ProfileBuilderOptions{
			GenericsFrames: true,
			LazyMapping:    true,
		},
		scaler: pprof.ScalerMutexProfile,
		rng:    rand.NewSource(239),
	}

	return res
}

func (h *mutexTestHelper) scale(rcount, rcycles int64) (int64, int64) {
	cpuGHz := float64(pprof.Runtime_cyclesPerSecond()) / 1e9
	count, nanosec := pprof.ScaleMutexProfile(h.scaler, rcount, float64(rcycles)/cpuGHz)
	inanosec := int64(nanosec)

	return count, inanosec
}

func (h *mutexTestHelper) scale2(rcount, rcycles int64) []int64 {
	c, n := h.scale(rcount, rcycles)

	return []int64{c, n}
}

func (h *mutexTestHelper) dump(r ...runtime.BlockProfileRecord) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	err := PrintCountCycleProfile(h.dp, h.opt, buf, h.scaler, r)
	if err != nil { // never happens
		panic(err)
	}

	return buf
}

func (h *mutexTestHelper) r(count, cycles int64, s [32]uintptr) runtime.BlockProfileRecord {
	return runtime.BlockProfileRecord{
		Count:  count,
		Cycles: cycles,
		StackRecord: runtime.StackRecord{
			Stack0: s,
		},
	}
}

func (h *mutexTestHelper) mutate(nmutations int, fs []runtime.BlockProfileRecord) {
	oneBlockCycles := fs[0].Cycles / fs[0].Count
	for j := 0; j < nmutations; j++ {
		idx := int(uint(h.rng.Int63())) % len(fs) //nolint:gosec
		fs[idx].Count += 1
		fs[idx].Cycles += oneBlockCycles
	}
}

type heapTestHelper struct {
	dp   *pprof.DeltaHeapProfiler
	opt  *pprof.ProfileBuilderOptions
	rate int64
	rng  rand.Source
}

func newHeapTestHelper() *heapTestHelper {
	res := &heapTestHelper{
		dp: &pprof.DeltaHeapProfiler{},
		opt: &pprof.ProfileBuilderOptions{
			GenericsFrames: true,
			LazyMapping:    true,
		},
		rng:  rand.NewSource(239),
		rate: int64(runtime.MemProfileRate),
	}

	return res
}

func (h *heapTestHelper) dump(r ...runtime.MemProfileRecord) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	err := WriteHeapProto(h.dp, h.opt, buf, r, h.rate)
	if err != nil { // never happens
		panic(err)
	}

	return buf
}

func (h *heapTestHelper) r(allocObjects, allocBytes, freeObjects, freeBytes int64,
	s [32]uintptr) runtime.MemProfileRecord {
	return runtime.MemProfileRecord{
		AllocObjects: allocObjects,
		AllocBytes:   allocBytes,
		FreeBytes:    freeBytes,
		FreeObjects:  freeObjects,
		Stack0:       s,
	}
}

func (h *heapTestHelper) mutate(nmutations int, fs []runtime.MemProfileRecord) {
	objSize := fs[0].AllocBytes / fs[0].AllocObjects
	for j := 0; j < nmutations; j++ {
		idx := int(uint(h.rng.Int63())) % len(fs) //nolint:gosec
		fs[idx].AllocObjects += 1
		fs[idx].AllocBytes += objSize
		fs[idx].FreeObjects += 1
		fs[idx].FreeBytes += objSize
	}
}

func WriteHeapProto(dp *pprof.DeltaHeapProfiler, opt *pprof.ProfileBuilderOptions, w io.Writer,
	p []runtime.MemProfileRecord, rate int64) error {
	stc := pprof.HeapProfileConfig(rate)
	b := pprof.NewProfileBuilder(w, opt, stc)

	return dp.WriteHeapProto(b, p, rate)
}

func PrintCountCycleProfile(d *pprof.DeltaMutexProfiler, opt *pprof.ProfileBuilderOptions, w io.Writer,
	scaler pprof.MutexProfileScaler, records []runtime.BlockProfileRecord) error {
	stc := pprof.MutexProfileConfig()
	b := pprof.NewProfileBuilder(w, opt, stc)

	return d.PrintCountCycleProfile(b, scaler, records)
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
