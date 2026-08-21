//go:build cgo

#include "cgo_bridge.h"

extern void godeltaprofCgoCallback(void);

// godeltaprofCallGoFromC calls back into Go from C, so that the Go code runs
// with cgo on the M's stack.
void godeltaprofCallGoFromC(void) { godeltaprofCgoCallback(); }
