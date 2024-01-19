package pprof

import (
	"io"
	"runtime"
	"testing"
)

func Test_SetCollector(t *testing.T) {
	for i := 0; i < 20; i++ {
		_ = StartCPUProfile(io.Discard)
		// SetCollector blocks until StopCPUProfile is called.
		done := make(chan struct{})
		go func() {
			ResetCollector()
			close(done)
		}()
		runtime.Gosched()
		StopCPUProfile()
		<-done
		if c.Collector != nil {
			t.Fatal("collector was not reset")
		}
	}
}

func Test_DefaultCollector(t *testing.T) {
	if err := StartCPUProfile(io.Discard); err != nil {
		t.Fatalf("Default collector StartCPUProfile: %v", err)
	}
	if err := StartCPUProfile(io.Discard); err == nil {
		t.Fatalf("Default collector must fail on consecuitive StartCPUProfile")
	}
	StopCPUProfile()
}
