package otlp

import (
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	otlpprofile "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
	"runtime"
	"sort"
	"sync"
)

type BlockProfiler struct {
	impl           pprof.DeltaMutexProfiler
	mutex          sync.Mutex
	runtimeProfile func([]runtime.BlockProfileRecord) (int, bool)
	scaleProfile   pprof.MutexProfileScaler
	options        pprof.ProfileBuilderOptions
}

func NewMutexProfilerWithOptions(options godeltaprof.ProfileOptions) *BlockProfiler {
	return &BlockProfiler{
		runtimeProfile: runtime.MutexProfile,
		scaleProfile:   pprof.ScalerMutexProfile,
		impl:           pprof.DeltaMutexProfiler{},
		options: pprof.ProfileBuilderOptions{
			GenericsFrames: options.GenericsFrames,
			LazyMapping:    options.LazyMappings,
		},
	}
}

func NewBlockProfilerWithOptions(options godeltaprof.ProfileOptions) *BlockProfiler {
	return &BlockProfiler{
		runtimeProfile: runtime.BlockProfile,
		scaleProfile:   pprof.ScalerBlockProfile,
		impl:           pprof.DeltaMutexProfiler{},
		options: pprof.ProfileBuilderOptions{
			GenericsFrames: options.GenericsFrames,
			LazyMapping:    options.LazyMappings,
		},
	}
}

func (d *BlockProfiler) Profile() (*otlpprofile.Profile, error) {
	var p []runtime.BlockProfileRecord
	n, ok := d.runtimeProfile(nil)
	for {
		p = make([]runtime.BlockProfileRecord, n+50)
		n, ok = d.runtimeProfile(p)
		if ok {
			p = p[:n]
			break
		}
	}

	sort.Slice(p, func(i, j int) bool { return p[i].Cycles > p[j].Cycles })

	return d.ProfileFromRecords(p)
}

func (d *BlockProfiler) ProfileFromRecords(p []runtime.BlockProfileRecord) (*otlpprofile.Profile, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	stc := pprof.MutexProfileConfig()
	b := newOTLPProtoBuilder(stc, &d.options)
	err := d.impl.PrintCountCycleProfile(b, d.scaleProfile, p)
	if err != nil {
		return nil, err
	}
	return b.Proto(), nil
}
