package pyroscope

import (
	"bytes"
	"fmt"
	"io"
	"runtime/pprof"
	"time"

	internal "github.com/grafana/pyroscope-go/internal/pprof"
	"github.com/grafana/pyroscope-go/upstream"
)

type cpuProfileCollector struct {
	name        string
	dur         time.Duration
	buf         *bytes.Buffer
	upstream    upstream.Upstream
	timeStarted time.Time

	// started indicates whether the collector
	// is interrupted with StartCPUProfile.
	started bool
	events  chan event

	halt chan struct{}
	done chan struct{}
}

type event struct {
	typ  eventType
	done chan error
	w    io.Writer
}

type eventType int

const (
	startEvent eventType = iota
	stopEvent
	flushEvent
)

func newEvent(typ eventType) event {
	return event{typ: typ, done: make(chan error)}
}

func (e event) send(c chan<- event) error {
	c <- e
	return <-e.done
}

func newStartEvent(w io.Writer) event {
	e := newEvent(startEvent)
	e.w = w
	return e
}

func newCPUProfileCollector(
	name string,
	upstream upstream.Upstream,
	period time.Duration,
) *cpuProfileCollector {
	buf := bytes.NewBuffer(make([]byte, 0, 1<<10))
	return &cpuProfileCollector{
		name:     name,
		dur:      period,
		buf:      buf,
		upstream: upstream,
		events:   make(chan event),
		halt:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (c *cpuProfileCollector) Start() {
	// From now on, internal pprof.StartCPUProfile
	// is handled by this collector.
	internal.SetCollector(c)
	t := time.NewTicker(c.dur)

	// Force pprof.StartCPUProfile: if CPU profiling is already
	// in progress (pprof.StartCPUProfile called outside the
	// package), profiling will start once it finishes.
	_ = c.reset(nil)
	for {
		select {
		case n := <-t.C:
			// Skip and adjust the timer, if the actual
			// profile duration is less than the desired,
			// which may happen if the collector has been
			// interrupted and then resumed, or flushed.
			if d := n.Sub(c.timeStarted); d < c.dur {
				t.Reset(d)
				continue
			}
			t.Reset(c.dur)
			if !c.started {
				// Collector can't start collecting profiles
				// in background while profiling started with
				// StartCPUProfile (foreground).
				_ = c.reset(nil)
			}

		case <-c.halt:
			t.Stop()
			if c.started {
				// Collector can't be stopped in-between
				// StartCPUProfile and StopCPUProfile calls.
				continue
			}
			pprof.StopCPUProfile()
			c.upload()
			close(c.done)
			return

		case e := <-c.events:
			c.handleEvent(e)
		}
	}
}

func (c *cpuProfileCollector) handleEvent(e event) {
	var err error
	defer func() {
		e.done <- err
		close(e.done)
	}()

	switch e.typ {
	case startEvent:
		if c.started { // Misuse.
			// Just to avoid interruption of the background
			// profiling that will fail immediately.
			err = fmt.Errorf("cpu profiling already started")
		} else {
			err = c.reset(e.w)
		}
		c.started = err == nil

	case stopEvent:
		if c.started {
			err = c.reset(nil)
			c.started = false
		}

	case flushEvent:
		if c.started {
			// Flush can't be done if StartCPUProfile is called,
			// as we'd need stopping the foreground collector first.
			err = fmt.Errorf("flush rejected: cpu profiling is in progress")
		} else {
			err = c.reset(nil)
		}
	}
}

func (c *cpuProfileCollector) Stop() {
	// Switches back to the standard pprof collector.
	// If internal pprof.StartCPUProfile is called,
	// the function blocks until StopCPUProfile.
	internal.ResetCollector()
	// Note that "halt" is not an event, but rather state
	// of the collector: the channel can be read multiple
	// times before the collector stops.
	close(c.halt)
	<-c.done
}

func (c *cpuProfileCollector) StartCPUProfile(w io.Writer) error {
	return newStartEvent(w).send(c.events)
}

func (c *cpuProfileCollector) StopCPUProfile() {
	_ = newEvent(stopEvent).send(c.events)
}

func (c *cpuProfileCollector) Flush() error {
	return newEvent(flushEvent).send(c.events)
}

func (c *cpuProfileCollector) reset(w io.Writer) error {
	pprof.StopCPUProfile()
	c.upload()
	var d io.Writer = c.buf
	if w != nil {
		// pprof.StopCPUProfile dumps gzipped
		// profile ignoring any writer failure.
		d = io.MultiWriter(d, w)
	}
	c.timeStarted = time.Now()
	if err := pprof.StartCPUProfile(d); err != nil {
		c.timeStarted = time.Time{}
		c.buf.Reset()
		return err
	}
	return nil
}

func (c *cpuProfileCollector) upload() {
	if c.timeStarted.IsZero() {
		return
	}
	c.upstream.Upload(&upstream.UploadJob{
		Name:            c.name,
		StartTime:       c.timeStarted,
		EndTime:         time.Now(),
		SpyName:         "gospy",
		SampleRate:      100,
		Units:           "samples",
		AggregationType: "sum",
		Format:          upstream.FormatPprof,
		Profile:         copyBuf(c.buf.Bytes()),
	})
	c.buf.Reset()
}
