package heap

import (
	"bytes"
	"github.com/pyroscope-io/client/heap/internal/pprof"
	"runtime"
)

type DeltaHeapProfiler struct {
}

func (d *DeltaHeapProfiler) Profile() error {

	// Find out how many records there are (MemProfile(nil, true)),
	// allocate that many records, and get the data.
	// There's a race—more records might be added between
	// the two calls—so allocate a few extra records for safety
	// and also try again if we're very unlucky.
	// The loop should only execute one iteration in the common case.
	var p []runtime.MemProfileRecord
	n, ok := runtime.MemProfile(nil, true)
	for {
		// Allocate room for a slightly bigger profile,
		// in case a few more entries have been added
		// since the call to MemProfile.
		p = make([]runtime.MemProfileRecord, n+50)
		n, ok = runtime.MemProfile(p, true)
		if ok {
			p = p[0:n]
			break
		}
		// Profile grew; try again.
	}

	w := bytes.NewBuffer(nil)

	err := pprof.WriteHeapProto(w, p, int64(runtime.MemProfileRate), "")
	return err
}
