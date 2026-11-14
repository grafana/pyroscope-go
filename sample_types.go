package pyroscope

import (
	"bytes"

	gprofile "github.com/google/pprof/profile"

	"github.com/grafana/pyroscope-go/upstream"
)

var (
	sampleTypeConfigHeap = map[string]*upstream.SampleType{ //nolint:gochecknoglobals
		"alloc_objects": {
			Units:      "objects",
			Cumulative: false,
		},
		"alloc_space": {
			Units:      "bytes",
			Cumulative: false,
		},
		"inuse_space": {
			Units:       "bytes",
			Aggregation: "average",
			Cumulative:  false,
		},
		"inuse_objects": {
			Units:       "objects",
			Aggregation: "average",
			Cumulative:  false,
		},
	}
	sampleTypeConfigMutex = map[string]*upstream.SampleType{ //nolint:gochecknoglobals
		"contentions": {
			DisplayName: "mutex_count",
			Units:       "lock_samples",
			Cumulative:  false,
		},
		"delay": {
			DisplayName: "mutex_duration",
			Units:       "lock_nanoseconds",
			Cumulative:  false,
		},
	}
	sampleTypeConfigBlock = map[string]*upstream.SampleType{ //nolint:gochecknoglobals
		"contentions": {
			DisplayName: "block_count",
			Units:       "lock_samples",
			Cumulative:  false,
		},
		"delay": {
			DisplayName: "block_duration",
			Units:       "lock_nanoseconds",
			Cumulative:  false,
		},
	}
	sampleTypeConfigGoroutineLeak = map[string]*upstream.SampleType{ //nolint:gochecknoglobals
		"goroutineleak": {
			DisplayName: "goroutine_leak",
			Units:       "goroutines",
			Aggregation: "average",
		},
	}
)

func selectHeapSampleTypes(profileTypes []ProfileType) map[string]*upstream.SampleType {
	selected := make(map[string]*upstream.SampleType)
	for _, profileType := range profileTypes {
		name := string(profileType)
		if sampleType, ok := sampleTypeConfigHeap[name]; ok {
			selected[name] = sampleType
		}
	}

	return selected
}

func filterHeapSampleTypes(data []byte, selected map[string]*upstream.SampleType) ([]byte, error) {
	// The default selection needs no decoding or recompression.
	if len(selected) == len(sampleTypeConfigHeap) {
		return data, nil
	}

	profile, err := gprofile.ParseData(data)
	if err != nil {
		return nil, err
	}

	indices := make([]int, 0, len(selected))
	sampleTypes := make([]*gprofile.ValueType, 0, len(selected))
	for i, sampleType := range profile.SampleType {
		if _, ok := selected[sampleType.Type]; ok {
			indices = append(indices, i)
			sampleTypes = append(sampleTypes, sampleType)
		}
	}
	profile.SampleType = sampleTypes
	for _, sample := range profile.Sample {
		for i, index := range indices {
			sample.Value[i] = sample.Value[index]
		}
		sample.Value = sample.Value[:len(indices)]
	}
	if _, ok := selected[profile.DefaultSampleType]; !ok {
		profile.DefaultSampleType = ""
	}

	var filtered bytes.Buffer
	if err := profile.Write(&filtered); err != nil {
		return nil, err
	}

	return filtered.Bytes(), nil
}
