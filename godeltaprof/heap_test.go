package godeltaprof

import (
	"bytes"
	"testing"
)

func BenchmarkHeap(b *testing.B) {
	p := NewHeapProfiler()
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		err := p.Profile(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
