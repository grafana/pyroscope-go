package compat

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	gprofile "github.com/google/pprof/profile"
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/stretchr/testify/require"
)

var m sync.Mutex

func TestScaleMutex(t *testing.T) {
	prev := runtime.SetMutexProfileFraction(-1)
	defer runtime.SetMutexProfileFraction(prev)

	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	profiler := godeltaprof.NewMutexProfiler()
	err := profiler.Profile(io.Discard)
	require.NoError(t, err)

	const fraction = 2
	const iters = 1000
	const workers = 2
	runtime.SetMutexProfileFraction(fraction)

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for j := 0; j < workers; j++ {
		go func() {
			for i := 0; i < iters; i++ {
				m.Lock()
				time.Sleep(time.Millisecond)
				m.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()

	err = profiler.Profile(buffer)
	require.NoError(t, err)

	profile, err := gprofile.Parse(buffer)
	require.NoError(t, err)

	res := StackCollapseProfile(profile, 0, 1)
	for _, re := range res {
		fmt.Println(re)
	}
}

func StackCollapseProfile(p *gprofile.Profile, valueIDX int, scale float64) []string {
	type stack struct {
		funcs string
		value int64
	}

	var ret []stack
	for _, s := range p.Sample {
		var funcs []string
		for i := range s.Location {

			loc := s.Location[i]
			for _, line := range loc.Line {
				f := line.Function
				funcs = append(funcs, f.Name)
			}
		}
		for i := 0; i < len(funcs)/2; i++ {
			j := len(funcs) - i - 1
			funcs[i], funcs[j] = funcs[j], funcs[i]
		}

		v := s.Value[valueIDX]
		if scale != 1 {
			v = int64(float64(v) * scale)
		}
		ret = append(ret, stack{
			funcs: strings.Join(funcs, ";"),
			value: v,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return strings.Compare(ret[i].funcs, ret[j].funcs) < 0
	})
	var unique []stack
	for _, s := range ret {
		if len(unique) == 0 {
			unique = append(unique, s)
			continue
		}
		if unique[len(unique)-1].funcs == s.funcs {
			unique[len(unique)-1].value += s.value
			continue
		}
		unique = append(unique, s)

	}

	res := []string{}
	for _, s := range unique {
		res = append(res, fmt.Sprintf("%s %d", s.funcs, s.value))
	}
	return res
}
