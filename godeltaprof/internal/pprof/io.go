package pprof

import "io"

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
