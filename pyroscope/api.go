package pyroscope

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/pyroscope-io/client/internal/session"
	"github.com/pyroscope-io/client/internal/types"
)

type Config struct {
	ApplicationName string // e.g backend.purchases
	Tags            map[string]string
	ServerAddress   string // e.g http://pyroscope.services.internal:4040
	AuthToken       string // specify this token when using pyroscope cloud
	SampleRate      uint32
	Logger          types.Logger
	ProfileTypes    []types.ProfileType
	DisableGCRuns   bool // this will disable automatic runtime.GC runs between getting the heap profiles
}

type Profiler struct {
	session *session.Session
}

// Start starts continuously profiling go code
func Start(cfg Config) (*Profiler, error) {
	if len(cfg.ProfileTypes) == 0 {
		cfg.ProfileTypes = types.DefaultProfileTypes
	}
	if cfg.SampleRate == 0 {
		cfg.SampleRate = types.DefaultSampleRate
	}
	if cfg.Logger == nil {
		cfg.Logger = noopLogger
	}

	rc := session.RemoteConfig{
		AuthToken:              cfg.AuthToken,
		UpstreamAddress:        cfg.ServerAddress,
		UpstreamThreads:        4,
		UpstreamRequestTimeout: 30 * time.Second,
	}
	uploader, err := session.NewRemote(rc, cfg.Logger)
	if err != nil {
		return nil, err
	}

	sc := session.SessionConfig{
		Upstream:       uploader,
		Logger:         cfg.Logger,
		AppName:        cfg.ApplicationName,
		Tags:           cfg.Tags,
		ProfilingTypes: cfg.ProfileTypes,
		DisableGCRuns:  cfg.DisableGCRuns,
		SampleRate:     cfg.SampleRate,
		UploadRate:     10 * time.Second,
	}
	cfg.Logger.Infof("starting profiling session:")
	cfg.Logger.Infof("  AppName:        %+v", sc.AppName)
	cfg.Logger.Infof("  Tags:           %+v", sc.Tags)
	cfg.Logger.Infof("  ProfilingTypes: %+v", sc.ProfilingTypes)
	cfg.Logger.Infof("  DisableGCRuns:  %+v", sc.DisableGCRuns)
	cfg.Logger.Infof("  SampleRate:     %+v", sc.SampleRate)
	cfg.Logger.Infof("  UploadRate:     %+v", sc.UploadRate)
	session, err := session.NewSession(sc)
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}
	if err = session.Start(); err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	return &Profiler{session: session}, nil
}

// Stop stops continious profiling session
func (p *Profiler) Stop() error {
	p.session.Stop()
	return nil
}

type LabelSet = pprof.LabelSet

var Labels = pprof.Labels

func TagWrapper(ctx context.Context, labels LabelSet, cb func(context.Context)) {
	pprof.Do(ctx, labels, func(c context.Context) { cb(c) })
}
