package session

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/pyroscope-io/client/internal/flameql"
	"github.com/pyroscope-io/client/internal/types"
)

const errorThrottlerPeriod = 10 * time.Second

type Session struct {
	// configuration, doesn't change
	upstream      Upstream
	sampleRate    uint32
	profileTypes  []types.ProfileType
	uploadRate    time.Duration
	disableGCRuns bool
	pid           int

	logger    types.Logger
	stopOnce  sync.Once
	stopCh    chan struct{}
	trieMutex sync.Mutex

	// these things do change:
	cpuBuf       *bytes.Buffer
	memBuf       *bytes.Buffer
	memPrevBytes []byte

	lastGCGeneration uint32
	appName          string
	startTime        time.Time
}

type SessionConfig struct {
	Upstream
	types.Logger
	AppName        string
	Tags           map[string]string
	ProfilingTypes []types.ProfileType
	DisableGCRuns  bool
	SampleRate     uint32
	UploadRate     time.Duration
}

// implements upstream.Payload
type Single []byte

func (s Single) Bytes() []byte {
	return s
}

// TODO: I don't think append is very efficient
func (s Single) BytesWithLength() []byte {
	lenBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBytes, uint64(len(s)))
	return append(lenBytes, s...)
}

// implements upstream.Payload
type Double struct {
	prev    Single
	current Single
}

// TODO: I don't think append is very efficient
func (d Double) Bytes() []byte {
	return append(d.prev.BytesWithLength(), d.current.BytesWithLength()...)
}

// TODO: horrible API, change this
type SetAppNamer interface {
	SetAppName(string)
}

func NewSession(c SessionConfig) (*Session, error) {
	appName, err := mergeTagsWithAppName(c.AppName, c.Tags)
	if err != nil {
		return nil, err
	}

	ps := &Session{
		upstream:      c.Upstream,
		appName:       appName,
		profileTypes:  c.ProfilingTypes,
		disableGCRuns: c.DisableGCRuns,
		sampleRate:    c.SampleRate,
		uploadRate:    c.UploadRate,
		stopCh:        make(chan struct{}),
		logger:        c.Logger,
		cpuBuf:        &bytes.Buffer{},
		memBuf:        &bytes.Buffer{},
	}

	return ps, nil
}

func addSuffix(name string, ptype types.ProfileType) (string, error) {
	k, err := flameql.ParseKey(name)
	if err != nil {
		return "", err
	}
	k.Add("__name__", k.AppName()+"."+string(ptype))
	return k.Normalized(), nil
}

// mergeTagsWithAppName validates user input and merges explicitly specified
// tags with tags from app name.
//
// App name may be in the full form including tags (app.name{foo=bar,baz=qux}).
// Returned application name is always short, any tags that were included are
// moved to tags map. When merged with explicitly provided tags (config/CLI),
// last take precedence.
//
// App name may be an empty string. Tags must not contain reserved keys,
// the map is modified in place.
func mergeTagsWithAppName(appName string, tags map[string]string) (string, error) {
	k, err := flameql.ParseKey(appName)
	if err != nil {
		return "", err
	}
	for tagKey, tagValue := range tags {
		if flameql.IsTagKeyReserved(tagKey) {
			continue
		}
		if err = flameql.ValidateTagKey(tagKey); err != nil {
			return "", err
		}
		k.Add(tagKey, tagValue)
	}
	return k.Normalized(), nil
}

// revive:disable-next-line:cognitive-complexity complexity is fine
func (ps *Session) takeSnapshots() {
	ticker := time.NewTicker(time.Second / time.Duration(ps.sampleRate))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if ps.isDueForReset() {
				ps.reset()
			}
		case <-ps.stopCh:
			return
		}
	}
}

func copyBuf(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)
	return r
}

func (ps *Session) Start() error {
	ps.reset()

	go ps.takeSnapshots()
	return nil
}

func (ps *Session) isDueForReset() bool {
	// TODO: duration should be either taken from config or ideally passed from server
	now := time.Now().Truncate(ps.uploadRate)
	start := ps.startTime.Truncate(ps.uploadRate)

	return !start.Equal(now)
}

func (ps *Session) isCPUEnabled() bool {
	for _, t := range ps.profileTypes {
		if t == types.ProfileCPU {
			return true
		}
	}
	return false
}

func (ps *Session) isMemEnabled() bool {
	for _, t := range ps.profileTypes {
		if t == types.ProfileInuseObjects || t == types.ProfileAllocObjects || t == types.ProfileInuseSpace || t == types.ProfileAllocSpace {
			return true
		}
	}
	return false
}

func (ps *Session) reset() {
	now := time.Now()
	endTime := now.Truncate(ps.uploadRate)
	startTime := endTime.Add(-ps.uploadRate)
	ps.logger.Debugf("profiling session reset %s", startTime.String())

	ps.startTime = endTime

	// first reset should not result in an upload
	if !ps.startTime.IsZero() {
		if ps.isCPUEnabled() {
			pprof.StopCPUProfile()
			defer func() {
				pprof.StartCPUProfile(ps.cpuBuf)
			}()
			ps.upstream.Upload(&UploadJob{
				Name:            ps.appName,
				StartTime:       startTime,
				EndTime:         endTime,
				SpyName:         "gospy",
				SampleRate:      100,
				Units:           "samples",
				AggregationType: "sum",
				Format:          Pprof,
				Profile:         copyBuf(ps.cpuBuf.Bytes()),
			})
			ps.cpuBuf.Reset()
		}

		if ps.isMemEnabled() {
			currentGCGeneration := numGC()
			// sometimes GC doesn't run within 10 seconds
			//   in such cases we force a GC run
			//   users can disable it with disableGCRuns option
			if currentGCGeneration == ps.lastGCGeneration && !ps.disableGCRuns {
				runtime.GC()
				currentGCGeneration = numGC()
			}
			if currentGCGeneration != ps.lastGCGeneration {
				pprof.WriteHeapProfile(ps.memBuf)
				curMemBytes := copyBuf(ps.memBuf.Bytes())
				ps.memBuf.Reset()
				if ps.memPrevBytes != nil {
					ps.upstream.Upload(&UploadJob{
						Name:        ps.appName,
						StartTime:   startTime,
						EndTime:     endTime,
						SpyName:     "gospy",
						SampleRate:  100,
						Format:      Pprof,
						Profile:     curMemBytes,
						PrevProfile: ps.memPrevBytes,
					})
				}
				ps.memPrevBytes = curMemBytes
				ps.lastGCGeneration = currentGCGeneration
			}
		}
	}
}

func (ps *Session) Stop() {
	ps.trieMutex.Lock()
	defer ps.trieMutex.Unlock()

	ps.stopOnce.Do(func() {
		// TODO: wait for stopCh consumer to finish!
		close(ps.stopCh)
		// before stopping, upload the tries
		ps.uploadData(time.Now())
	})
}

func (ps *Session) uploadData(now time.Time) {

}

func numGC() uint32 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.NumGC
}
