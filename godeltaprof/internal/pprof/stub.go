package pprof

import "runtime"

func Runtime_cyclesPerSecond() int64 {
	return runtime_cyclesPerSecond()
}

func Runtime_expandFinalInlineFrame(stk []uintptr) []uintptr {
	return runtime_expandFinalInlineFrame(stk)
}

func Runtime_FrameSymbolName(f *runtime.Frame) string {
	return runtime_FrameSymbolName(f)
}

func Runtime_FrameStartLine(f *runtime.Frame) int {
	return runtime_FrameStartLine(f)
}
