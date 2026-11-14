package pyroscope

import (
	"testing"
	"time"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"

	"github.com/grafana/pyroscope-go/internal/testutil"
)

func TestHeapProfileTypes(t *testing.T) {
	tests := []struct {
		name         string
		profileTypes []ProfileType
		want         []string
	}{
		{
			name:         "inuse objects",
			profileTypes: []ProfileType{ProfileInuseObjects},
			want:         []string{"inuse_objects"},
		},
		{
			name:         "allocated objects",
			profileTypes: []ProfileType{ProfileAllocObjects},
			want:         []string{"alloc_objects"},
		},
		{
			name:         "bytes only",
			profileTypes: []ProfileType{ProfileInuseSpace, ProfileAllocSpace},
			want:         []string{"alloc_space", "inuse_space"},
		},
		{
			name:         "repeated selection",
			profileTypes: []ProfileType{ProfileInuseObjects, ProfileInuseObjects},
			want:         []string{"inuse_objects"},
		},
		{
			name:         "default profile types",
			profileTypes: DefaultProfileTypes,
			want:         []string{"alloc_objects", "alloc_space", "inuse_objects", "inuse_space"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := new(mockUpstream)
			session, err := NewSession(SessionConfig{
				Upstream:       uploader,
				Logger:         testutil.NewTestLogger(),
				AppName:        "heap-types",
				ProfilingTypes: tt.profileTypes,
			})
			require.NoError(t, err)

			session.dumpHeapProfile(time.Unix(1, 0), time.Unix(2, 0))
			require.Len(t, uploader.uploaded, 1)
			job := uploader.uploaded[0]
			profile, err := gprofile.ParseData(job.Profile)
			require.NoError(t, err)
			require.NoError(t, profile.CheckValid())
			var sampleTypes []string
			for _, sampleType := range profile.SampleType {
				sampleTypes = append(sampleTypes, sampleType.Type)
			}
			require.Equal(t, tt.want, sampleTypes)
			require.Len(t, job.SampleTypeConfig, len(tt.want))
			for _, name := range tt.want {
				require.Equal(t, sampleTypeConfigHeap[name], job.SampleTypeConfig[name])
			}
		})
	}
}
