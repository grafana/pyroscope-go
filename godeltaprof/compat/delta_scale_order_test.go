package compat

import (
	"bytes"
	"regexp"
	"runtime"
	"testing"

	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	"github.com/stretchr/testify/require"
)

const rate = int64(524288)

var markerPC uintptr
var markerPCName string

func init() {
	cs := make([]uintptr, 1)
	_ = runtime.Callers(1, cs)
	markerPC = cs[0]
	markerPCName = runtime.FuncForPC(markerPC).Name()
	markerPCName = regexp.QuoteMeta(markerPCName)
}

func TestScaleBeforeDelta(t *testing.T) {
	var v = []struct {
		count, size, scaledCount, scaledSize int64
	}{
		{5, 5 * 327680, 10, 3525422},
		{8, 8 * 327680, 17, 5640676},
		{9, 9 * 327680, 19, 6345761},
	}

	dh := pprof.DeltaHeapProfiler{}

	p := func(v1, v2, v3, v4 int64) *bytes.Buffer {
		r := []runtime.MemProfileRecord{
			{AllocObjects: v1, AllocBytes: v2, FreeObjects: v3, FreeBytes: v4, Stack0: [32]uintptr{markerPC}},
		}
		buf := bytes.NewBuffer(nil)
		err := dh.WriteHeapProto(buf, r, rate, "")
		require.NoError(t, err)
		return buf
	}

	p1 := p(0, 0, 0, 0)
	expectNoFrames(t, p1)

	p2 := p(v[0].count, v[0].size, 0, 0)
	expectStackFrames(t, p2, markerPCName,
		v[0].scaledCount, v[0].scaledSize,
		v[0].scaledCount, v[0].scaledSize,
	)

	p3 := p(v[0].count, v[0].size, v[0].count, v[0].size)
	expectNoFrames(t, p3)

	p4 := p(v[1].count, v[1].size, v[0].count, v[0].size)
	expectStackFrames(t, p4, markerPCName,
		v[1].scaledCount-v[0].scaledCount, v[1].scaledSize-v[0].scaledSize,
		v[1].scaledCount-v[0].scaledCount, v[1].scaledSize-v[0].scaledSize,
	)

	p5 := p(v[2].count, v[2].size, v[0].count, v[0].size)
	expectStackFrames(t, p5, markerPCName,
		v[2].scaledCount-v[1].scaledCount, v[2].scaledSize-v[1].scaledSize,
		v[2].scaledCount-v[0].scaledCount, v[2].scaledSize-v[0].scaledSize,
	)

	p6 := p(v[2].count, v[2].size, v[2].count, v[2].size)
	expectNoFrames(t, p6)
}

func TestScaleMutexOrder(t *testing.T) {
	t.Errorf("todo")
}

func TestScaleBlockOrder(t *testing.T) {
	t.Errorf("todo")
}
