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
		entry := d.m.Lookup(r.Stack(), 0)
		entry.acc.v1 += r.Count // accumulate unscaled
		entry.acc.v2 += r.Cycles
	}
	for _, r := range records {
		entry := d.m.Lookup(r.Stack(), 0)
		accCount := entry.acc.v1
		accCycles := entry.acc.v2
		if accCount == 0 && accCycles == 0 { //todo check if this is correct
			continue
		}
		entry.acc = count{}
		count, nanosec := ScaleMutexProfile(scaler, accCount, float64(accCycles)/cpuGHz)
		inanosec := int64(nanosec)

		// do the delta
		values[0] = count - entry.prev.v1
		values[1] = inanosec - entry.prev.v2
		entry.prev.v1 = count
		entry.prev.v2 = inanosec

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
