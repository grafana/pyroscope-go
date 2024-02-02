package pyroscope

import (
	"testing"
)

func TestProfilerStartStop(t *testing.T) {
	profiler, err := Start(Config{
		ApplicationName: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	profiler.Stop()
}
