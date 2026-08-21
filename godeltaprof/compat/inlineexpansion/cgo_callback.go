//go:build cgo

package inlineexpansion

/*
#include "cgo_bridge.h"
*/
import "C"

import "time"

// cgoHoldDuration is how long the cgo callback holds reproMu.
var cgoHoldDuration time.Duration //nolint:gochecknoglobals

//export godeltaprofCgoCallback
func godeltaprofCgoCallback() {
	reproEntry(cgoHoldDuration, nil)
}

// callGoViaC enters C, which calls back into Go. While the callback runs the M
// has cgo on its stack (runtime.m.hasCgoOnStack), so the runtime records mutex
// contention with a full traceback instead of frame pointer unwinding.
func callGoViaC() {
	C.godeltaprofCallGoFromC()
}
