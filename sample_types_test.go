package pyroscope

import (
	"bytes"
	"runtime"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"
)

func TestFilterHeapSampleTypes(t *testing.T) {
	tests := []struct {
		name         string
		profileTypes []ProfileType
		wantTypes    []string
		wantUnits    []string
		wantValues   [][]int64
		wantDefault  string
	}{
		{
			name:         "one column",
			profileTypes: []ProfileType{ProfileInuseObjects},
			wantTypes:    []string{"inuse_objects"},
			wantUnits:    []string{"count"},
			wantValues:   [][]int64{{7}, {13}},
		},
		{
			name:         "nonadjacent columns",
			profileTypes: []ProfileType{ProfileAllocSpace, ProfileInuseSpace},
			wantTypes:    []string{"alloc_space", "inuse_space"},
			wantUnits:    []string{"bytes", "bytes"},
			wantValues:   [][]int64{{110, 70}, {230, 130}},
			wantDefault:  "alloc_space",
		},
		{
			name:         "all columns",
			profileTypes: DefaultProfileTypes,
			wantTypes:    []string{"alloc_objects", "alloc_space", "inuse_objects", "inuse_space"},
			wantUnits:    []string{"count", "bytes", "count", "bytes"},
			wantValues:   [][]int64{{11, 110, 7, 70}, {23, 230, 13, 130}},
			wantDefault:  "alloc_space",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := heapProfileFixture()
			var encoded bytes.Buffer
			require.NoError(t, input.Write(&encoded))
			data := encoded.Bytes()
			filtered, err := filterHeapSampleTypes(data, selectHeapSampleTypes(tt.profileTypes))
			require.NoError(t, err)
			output, err := gprofile.ParseData(filtered)
			require.NoError(t, err)
			require.NoError(t, output.CheckValid())

			var types []string
			var units []string
			for _, sampleType := range output.SampleType {
				types = append(types, sampleType.Type)
				units = append(units, sampleType.Unit)
			}
			require.Equal(t, tt.wantTypes, types)
			require.Equal(t, tt.wantUnits, units)
			require.Equal(t, tt.wantDefault, output.DefaultSampleType)
			require.Len(t, output.Sample, len(tt.wantValues))
			for i, sample := range output.Sample {
				require.Equal(t, tt.wantValues[i], sample.Value)
				require.Equal(t, input.Sample[i].Label, sample.Label)
				require.Equal(t, input.Sample[i].NumLabel, sample.NumLabel)
				require.Equal(t, input.Sample[i].Location[0].Line[0].Function.Name,
					sample.Location[0].Line[0].Function.Name)
			}
			require.Equal(t, input.TimeNanos, output.TimeNanos)
			require.Equal(t, input.DurationNanos, output.DurationNanos)
			require.Equal(t, input.PeriodType.Type, output.PeriodType.Type)
			require.Equal(t, input.PeriodType.Unit, output.PeriodType.Unit)
			require.Equal(t, input.Period, output.Period)
			if len(tt.wantTypes) == 4 {
				require.Equal(t, data, filtered)
			}
		})
	}
}

func heapProfileFixture() *gprofile.Profile {
	function := &gprofile.Function{ID: 1, Name: "allocate", Filename: "main.go"}
	location := &gprofile.Location{ID: 1, Line: []gprofile.Line{{Function: function, Line: 42}}}

	return &gprofile.Profile{
		SampleType: []*gprofile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		},
		Sample: []*gprofile.Sample{
			{Location: []*gprofile.Location{location}, Value: []int64{11, 110, 7, 70},
				Label: map[string][]string{"region": {"west"}}, NumLabel: map[string][]int64{"bytes": {10}}},
			{Location: []*gprofile.Location{location}, Value: []int64{23, 230, 13, 130},
				Label: map[string][]string{"region": {"east"}}, NumLabel: map[string][]int64{"bytes": {10}}},
		},
		Location:          []*gprofile.Location{location},
		Function:          []*gprofile.Function{function},
		TimeNanos:         123456789,
		DurationNanos:     1000000,
		PeriodType:        &gprofile.ValueType{Type: "space", Unit: "bytes"},
		Period:            524288,
		DefaultSampleType: "alloc_space",
	}
}

func BenchmarkHeapProfileSelection(b *testing.B) {
	profile := heapProfileFixture()
	for i := range 1000 {
		profile.Sample = append(profile.Sample, &gprofile.Sample{
			Location: profile.Location,
			Value:    []int64{int64(i), int64(i * 100), int64(i / 2), int64(i * 50)},
		})
	}
	var encoded bytes.Buffer
	require.NoError(b, profile.Write(&encoded))
	payload := encoded.Bytes()
	b.Run("legacy copy", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(payload)))
		var result []byte
		for b.Loop() {
			result = copyBuf(payload)
		}
		runtime.KeepAlive(result)
	})
	for _, tt := range []struct {
		name  string
		types []ProfileType
	}{
		{"all types", DefaultProfileTypes},
		{"one type", []ProfileType{ProfileInuseObjects}},
	} {
		b.Run(tt.name, func(b *testing.B) {
			selected := selectHeapSampleTypes(tt.types)
			b.ReportAllocs()
			b.SetBytes(int64(len(payload)))
			var result []byte
			for b.Loop() {
				var err error
				result, err = filterHeapSampleTypes(copyBuf(payload), selected)
				if err != nil {
					b.Fatal(err)
				}
			}
			runtime.KeepAlive(result)
		})
	}
}
