package compat

import (
	"go/types"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
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

func generateMemProfileRecords(n, depth, seed int) []runtime.MemProfileRecord {
	var records []runtime.MemProfileRecord
	rng := rand.NewSource(int64(seed))
	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(rng.Int63())) % 1000000
		r := runtime.MemProfileRecord{
			AllocObjects: int64(nobj),
			AllocBytes:   int64(nobj * 1024),
			FreeObjects:  int64(nobj), // pretend inuse is zero
			FreeBytes:    int64(nobj * 1024),
		}
		for j := 0; j < depth; j++ {
			r.Stack0[j] = fs[int(uint64(rng.Int63()))%len(fs)]
		}
		records = append(records, r)
	}
	return records
}

func generateBlockProfileRecords(n, depth, seed int) []runtime.BlockProfileRecord {
	var records []runtime.BlockProfileRecord
	rng := rand.NewSource(int64(seed))
	fs := getFunctionPointers()
	for i := 0; i < n; i++ {
		nobj := int(uint64(rng.Int63())) % 1000000
		r := runtime.BlockProfileRecord{
			Count:  int64(nobj),
			Cycles: int64(nobj * 10),
		}
		for j := 0; j < depth; j++ {
			r.Stack0[j] = fs[int(uint64(rng.Int63()))%len(fs)]
		}
		records = append(records, r)
	}
	return records
}

func getFunctions(t testing.TB, pkg string) []*types.Func {
	var res []*types.Func
	cfg := &packages.Config{
		Mode:  packages.NeedImports | packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, pkg)
	require.NoError(t, err)
	for _, p := range pkgs {
		if strings.Contains(p.ID, ".test") {
			continue
		}
		for _, name := range p.Types.Scope().Names() {
			f := p.Types.Scope().Lookup(name)

			if f != nil {
				ff, ok := f.(*types.Func)
				if ok {
					res = append(res, ff)
				}
			}
		}
	}
	return res
}
