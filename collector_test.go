package pyroscope

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/grafana/pyroscope-go/upstream"
)

func Test_StartCPUProfile_interrupts_background_profiling(t *testing.T) {
	logger := new(testLogger)
	collector := new(mockCollector)
	c := newCPUProfileCollector(
		"test",
		new(mockUpstream),
		logger,
		100*time.Millisecond,
	)
	c.collector = collector

	go c.Start()
	<-collector.waitStartCPUProfile()

	// Background profile is being collected.
	// Try to interrupt it with StartCPUProfile.
	start := collector.waitStartCPUProfile()
	stop := collector.waitStopCPUProfile()
	if err := c.StartCPUProfile(io.Discard); err != nil {
		t.Fatal("failed to start CPU profiling")
	}
	<-stop
	<-start

	// Foreground profile is being collected.
	// Resume background profiling with StopCPUProfile.
	start = collector.waitStartCPUProfile()
	stop = collector.waitStopCPUProfile()
	c.StopCPUProfile()
	<-stop
	<-start

	c.Stop()

	if !reflect.DeepEqual(logger.lines, []string{
		"starting cpu profile collector",
		"cpu profile collector interrupted with StartCPUProfile",
		"cpu profile collector restored",
		"stopping cpu profile collector",
		"stopping cpu profile collector stopped",
	}) {
		for _, line := range logger.lines {
			t.Log(line)
		}
		t.Fatal("^ unexpected even sequence")
	}
}

func Test_StartCPUProfile_blocks_Stop(t *testing.T) {
	logger := new(testLogger)
	collector := new(mockCollector)
	c := newCPUProfileCollector(
		"test",
		new(mockUpstream),
		logger,
		100*time.Millisecond,
	)
	c.collector = collector

	go c.Start()
	<-collector.waitStartCPUProfile()

	// Background profile is being collected.
	// Try to interrupt it with StartCPUProfile.
	start := collector.waitStartCPUProfile()
	stop := collector.waitStopCPUProfile()
	if err := c.StartCPUProfile(io.Discard); err != nil {
		t.Fatal("failed to start CPU profiling")
	}
	<-stop
	<-start

	go c.StopCPUProfile()
	c.Stop()

	if !reflect.DeepEqual(logger.lines, []string{
		"starting cpu profile collector",
		"cpu profile collector interrupted with StartCPUProfile",
		"stopping cpu profile collector",
		"cpu profile collector restored",
		"stopping cpu profile collector stopped",
	}) {
		for _, line := range logger.lines {
			t.Log(line)
		}
		t.Fatal("^ unexpected even sequence")
	}
}

type mockCollector struct {
	sync.Mutex
	start chan struct{}
	stop  chan struct{}
}

func (m *mockCollector) waitStartCPUProfile() <-chan struct{} {
	m.Lock()
	c := make(chan struct{})
	m.start = c
	m.Unlock()
	return c
}

func (m *mockCollector) waitStopCPUProfile() <-chan struct{} {
	m.Lock()
	c := make(chan struct{})
	m.stop = c
	m.Unlock()
	return c
}

func (m *mockCollector) StartCPUProfile(_ io.Writer) error {
	m.Lock()
	if m.start != nil {
		close(m.start)
		m.start = nil
	}
	m.Unlock()
	return nil
}

func (m *mockCollector) StopCPUProfile() {
	m.Lock()
	if m.stop != nil {
		close(m.stop)
		m.stop = nil
	}
	m.Unlock()
}

type mockUpstream struct{ uploaded []*upstream.UploadJob }

func (m *mockUpstream) Upload(j *upstream.UploadJob) { m.uploaded = append(m.uploaded, j) }

func (*mockUpstream) Flush() {}

type testLogger struct {
	sync.Mutex
	lines []string
}

func (t *testLogger) Debugf(format string, args ...interface{}) { t.put(format, args...) }
func (t *testLogger) Infof(format string, args ...interface{})  { t.put(format, args...) }
func (t *testLogger) Errorf(format string, args ...interface{}) { t.put(format, args...) }

func (t *testLogger) put(format string, args ...interface{}) {
	t.Lock()
	t.lines = append(t.lines, fmt.Sprintf(format, args...))
	t.Unlock()
}
