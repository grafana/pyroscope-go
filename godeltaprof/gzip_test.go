package godeltaprof

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func TestGz(t *testing.T) {
	blobs := [][]byte{
		[]byte("Hello, World! This is the first test blob with some data to compress."),
		[]byte("This is a second blob with different content for compression testing."),
	}

	var g gz
	var bufs []bytes.Buffer
	
	for i, blob := range blobs {
		var buf bytes.Buffer
		gzw := g.get(&buf)
		if _, err := gzw.Write(blob); err != nil {
			t.Fatalf("Failed to write blob %d: %v", i, err)
		}
		if err := gzw.Close(); err != nil {
			t.Fatalf("Failed to close gzip writer %d: %v", i, err)
		}
		bufs = append(bufs, buf)
	}
	
	for i, blob := range blobs {
		gzr, err := gzip.NewReader(&bufs[i])
		if err != nil {
			t.Fatalf("Failed to create gzip reader for blob %d: %v", i, err)
		}
		
		decompressed, err := io.ReadAll(gzr)
		if err != nil {
			gzr.Close()
			t.Fatalf("Failed to decompress blob %d: %v", i, err)
		}
		gzr.Close()
		
		if !bytes.Equal(blob, decompressed) {
			t.Errorf("Blob %d mismatch:\nOriginal:     %q\nDecompressed: %q", i, blob, decompressed)
		}
		
		if bytes.Equal(blob, bufs[i].Bytes()) {
			t.Errorf("Buffer %d should contain compressed data, not original", i)
		}
	}
}