package pyroscope

import (
	"bytes"
	"time"

	"github.com/grafana/pyroscope-go/upstream"
	"github.com/grafana/pyroscope-io/fgprof"
)

type wallClockProfileCollector struct {
	name string
	dur  time.Duration

	upstream upstream.Upstream
	logger   Logger

	buf         *bytes.Buffer
	timeStarted time.Time
	stop        chan struct{}
	done        chan struct{}
}

func newWallClockProfileCollector(
	name string,
	upstream upstream.Upstream,
	logger Logger,
	period time.Duration,
) *wallClockProfileCollector {
	buf := bytes.NewBuffer(make([]byte, 0, 1<<10))
	return &wallClockProfileCollector{
		name:     name,
		dur:      period,
		upstream: upstream,
		logger:   logger,
		buf:      buf,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (w *wallClockProfileCollector) Start() {
	t := time.NewTicker(w.dur)
	defer func() {
		t.Stop()
		close(w.done)
	}()
	for {
		w.timeStarted = time.Now()
		stop := fgprof.Start(w.buf)
		select {
		case <-t.C:
			_ = stop()
			w.upload()
		case <-w.stop:
			_ = stop()
			w.upload()
			return
		}
	}
}

func (w *wallClockProfileCollector) Stop() {
	close(w.stop)
	<-w.done
}

func (w *wallClockProfileCollector) Flush() error {
	return nil // TODO
}

func (w *wallClockProfileCollector) upload() {
	if w.timeStarted.IsZero() {
		return
	}
	buf := w.buf.Bytes()
	if len(buf) == 0 {
		return
	}
	w.upstream.Upload(&upstream.UploadJob{
		Name:            w.name,
		StartTime:       w.timeStarted,
		EndTime:         time.Now(),
		SpyName:         "gospy",
		SampleRate:      99,
		Units:           "nanoseconds",
		AggregationType: "sum",
		Format:          upstream.FormatPprof,
		Profile:         copyBuf(buf),
	})
	w.buf.Reset()
}
