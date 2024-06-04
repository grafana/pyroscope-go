package otlp

import (
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	otlpprofile "go.opentelemetry.io/proto/otlp/profiles/v1experimental"

	"runtime"
	"sync"
)

type HeapProfiler struct {
	mutex sync.Mutex
	impl  pprof.DeltaHeapProfiler
	mem   []pprof.MemMap
	opt   pprof.ProfileBuilderOptions
}

func NewHeapProfilerWithOptions(options godeltaprof.ProfileOptions) *HeapProfiler {
	return &HeapProfiler{
		opt: pprof.ProfileBuilderOptions{
			GenericsFrames: options.GenericsFrames,
			LazyMapping:    options.LazyMappings,
		},
		impl: pprof.DeltaHeapProfiler{},
	}
}

func (hp *HeapProfiler) Profile() (*otlpprofile.Profile, error) {

	// Find out how many records there are (MemProfile(nil, true)),
	// allocate that many records, and get the data.
	// There's a race—more records might be added between
	// the two calls—so allocate a few extra records for safety
	// and also try again if we're very unlucky.
	// The loop should only execute one iteration in the common case.
	var p []runtime.MemProfileRecord
	n, ok := runtime.MemProfile(nil, true)
	for {
		// Allocate room for a slightly bigger profile,
		// in case a few more entries have been added
		// since the call to MemProfile.
		p = make([]runtime.MemProfileRecord, n+50)
		n, ok = runtime.MemProfile(p, true)
		if ok {
			p = p[0:n]
			break
		}
		// Profile grew; try again.
	}

	return hp.ProfileFromRecords(p)
}

func (hp *HeapProfiler) ProfileFromRecords(p []runtime.MemProfileRecord) (*otlpprofile.Profile, error) {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	rate := int64(runtime.MemProfileRate)
	stc := pprof.HeapProfileConfig(rate)
	b := newOTLPProtoBuilder(stc, &hp.opt)

	err := hp.impl.WriteHeapProto(b, p, rate)
	if err != nil {
		return nil, err
	}

	return b.Proto(), nil
}
