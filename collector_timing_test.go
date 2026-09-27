package pyroscope

import (
	"io"
	"testing"
	"testing/synctest"
	"time"

	"github.com/grafana/pyroscope-go/internal/testutil"
)

func Test_CPUProfile_early_tick_waits_remaining_interval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t.Helper()
		period := time.Minute
		collector := &delayedCollector{starts: make(chan time.Time, 4), delay: time.Nanosecond}
		c := newCPUProfileCollector("test", new(mockUpstream), testutil.NewTestLogger(), period)
		c.collector = collector
		go c.Start()
		first := <-collector.starts
		defer c.Stop()

		time.Sleep(period)
		synctest.Wait()
		select {
		case next := <-collector.starts:
			if next.Sub(first) != period {
				t.Fatalf("profile duration = %v, want %v", next.Sub(first), period)
			}
		default:
			t.Fatal("early ticker delayed profile beyond the configured interval")
		}
	})
}

func Test_CPUProfile_flush_resets_remaining_interval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t.Helper()
		period := time.Minute
		collector := &delayedCollector{starts: make(chan time.Time, 4)}
		c := newCPUProfileCollector("test", new(mockUpstream), testutil.NewTestLogger(), period)
		c.collector = collector
		go c.Start()
		<-collector.starts
		defer c.Stop()

		time.Sleep(10 * time.Second)
		if err := c.Flush(); err != nil {
			t.Fatal(err)
		}
		resumed := <-collector.starts
		time.Sleep(period)
		synctest.Wait()
		select {
		case next := <-collector.starts:
			if next.Sub(resumed) != period {
				t.Fatalf("profile duration after flush = %v, want %v", next.Sub(resumed), period)
			}
		default:
			t.Fatal("flush delayed profile beyond the configured interval")
		}
	})
}

type delayedCollector struct {
	starts chan time.Time
	delay  time.Duration
}

func (c *delayedCollector) StartCPUProfile(io.Writer) error {
	c.starts <- time.Now()

	return nil
}

func (c *delayedCollector) StopCPUProfile() {
	if c.delay != 0 {
		time.Sleep(c.delay)
		c.delay = 0
	}
}
