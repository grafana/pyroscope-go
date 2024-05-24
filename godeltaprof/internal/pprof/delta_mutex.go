package pprof

import (
	"runtime"
)

type DeltaMutexProfiler struct {
	m profMap
}

// PrintCountCycleProfile outputs block profile records (for block or mutex profiles)
// as the pprof-proto format output. Translations from cycle count to time duration
// are done because The proto expects count and time (nanoseconds) instead of count
// and the number of cycles for block, contention profiles.
// Possible 'scaler' functions are scaleBlockProfile and scaleMutexProfile.
func (d *DeltaMutexProfiler) PrintCountCycleProfile(b ProfileBuilder, scaler MutexProfileScaler, records []runtime.BlockProfileRecord) error {

	cpuGHz := float64(runtime_cyclesPerSecond()) / 1e9

	values := []int64{0, 0}
	var locs []uint64
	for _, r := range records {
		count, nanosec := ScaleMutexProfile(scaler, r.Count, float64(r.Cycles)/cpuGHz)
		inanosec := int64(nanosec)

		// do the delta
		entry := d.m.Lookup(r.Stack(), 0)
		values[0] = count - entry.count.v1
		values[1] = inanosec - entry.count.v2
		entry.count.v1 = count
		entry.count.v2 = inanosec

		if values[0] < 0 || values[1] < 0 {
			continue
		}
		if values[0] == 0 && values[1] == 0 {
			continue
		}

		// For count profiles, all stack addresses are
		// return PCs, which is what appendLocsForStack expects.
		locs = b.LocsForStack(r.Stack())
		b.Sample(values, locs, 0)
	}
	b.Build()
	return nil
}

func MutexProfileConfig() ProfileConfig {
	return ProfileConfig{
		PeriodType: ValueType{"contentions", "count"},
		Period:     1,
		SampleType: []ValueType{
			{"contentions", "count"},
			{"delay", "nanoseconds"},
		},
	}
}
