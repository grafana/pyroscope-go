package godeltaprof

import (
	"bytes"
	"testing"
)

func BenchmarkHeap(b *testing.B) {
	p := NewHeapProfiler()
	buf := bytes.NewBuffer(nil)
	for range b.N {
		err := p.Profile(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
