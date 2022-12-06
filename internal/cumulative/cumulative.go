package cumulative

import (
	"bytes"
	"fmt"
	pprofile "github.com/google/pprof/profile"
	"github.com/pyroscope-io/client/upstream"
	"time"
)

type ProfileMerger struct {
	SampleTypes      []string
	MergeRatios      []float64
	SampleTypeConfig map[string]*upstream.SampleType

	prev *pprofile.Profile
	name string
}

type Mergers struct {
	Heap  *ProfileMerger
	Block *ProfileMerger
	Mutex *ProfileMerger
}

func NewMergers() *Mergers {
	return &Mergers{
		Block: &ProfileMerger{
			SampleTypes: []string{"contentions", "delay"},
			MergeRatios: []float64{-1, -1},
			SampleTypeConfig: map[string]*upstream.SampleType{
				"contentions": {
					DisplayName: "block_count",
					Units:       "lock_samples",
				},
				"delay": {
					DisplayName: "block_duration",
					Units:       "lock_nanoseconds",
				},
			},
			name: "block",
		},
		Mutex: &ProfileMerger{
			SampleTypes: []string{"contentions", "delay"},
			MergeRatios: []float64{-1, -1},
			SampleTypeConfig: map[string]*upstream.SampleType{
				"contentions": {
					DisplayName: "mutex_count",
					Units:       "lock_samples",
				},
				"delay": {
					DisplayName: "mutex_duration",
					Units:       "lock_nanoseconds",
				},
			},
			name: "mutex",
		},
		Heap: &ProfileMerger{
			SampleTypes: []string{"alloc_objects", "alloc_space", "inuse_objects", "inuse_space"},
			MergeRatios: []float64{-1, -1, 0, 0},
			SampleTypeConfig: map[string]*upstream.SampleType{
				"alloc_objects": {
					Units: "objects",
				},
				"alloc_space": {
					Units: "bytes",
				},
				"inuse_space": {
					Units:       "bytes",
					Aggregation: "average",
				},
				"inuse_objects": {
					Units:       "objects",
					Aggregation: "average",
				},
			},
			name: "heap",
		},
	}
}

// todo should we filter by enabled ps.profileTypes to reduce profile size ? maybe add a separate option ?
func (m *ProfileMerger) Merge(j *upstream.UploadJob) error {
	t1 := time.Now()
	defer func() { //todo remove before merge
		t2 := time.Now()
		fmt.Printf("Profile %v merged in %v\n", m.SampleTypes, t2.Sub(t1))
	}()
	p2, err := m.parseProfile(j.Profile)
	if err != nil {
		return err
	}
	p1 := m.prev
	if p1 == nil {
		p1, err = m.parseProfile(j.PrevProfile)
		if err != nil {
			return err
		}
	}

	err = p1.ScaleN(m.MergeRatios)
	if err != nil {
		return err
	}

	p, err := pprofile.Merge([]*pprofile.Profile{p1, p2})
	if err != nil {
		return err
	}

	negative := 0
	for _, sample := range p.Sample {
		if sample.Value[0] < 0 {
			for i := range sample.Value {
				sample.Value[i] = 0
				negative += 1
			}
		}
	}
	p = p.Compact()

	var prof bytes.Buffer
	err = p.Write(&prof)
	if err != nil {
		return err
	}

	m.prev = p2
	j.Profile = prof.Bytes()
	j.PrevProfile = nil
	j.SampleTypeConfig = m.SampleTypeConfig

	cb := upstream.DebugStatsCallback
	if cb != nil {
		cb(m.name, len(p1.Sample), len(p2.Sample), len(p.Sample), negative)
	}
	return nil
}

func (m *ProfileMerger) parseProfile(bs []byte) (*pprofile.Profile, error) {
	var prof = bytes.NewBuffer(bs)
	p, err := pprofile.Parse(prof)
	if err != nil {
		return nil, err
	}
	if got := len(p.SampleType); got != len(m.SampleTypes) {
		return nil, fmt.Errorf("invalid  profile: got %d sample types, want %d", got, len(m.SampleTypes))
	}
	for i, want := range m.SampleTypes {
		if got := p.SampleType[i].Type; got != want {
			return nil, fmt.Errorf("invalid profile: got %q sample type at index %d, want %q", got, i, want)
		}
	}
	return p, nil
}
