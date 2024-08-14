//go:build go1.22

package pprof

import "os"

func routePrefix() string {
	// As of go 1.23
	// https://github.com/golang/go/blob/9fcffc53593c5cd103630d0d24ef8bd91e17246d/src/net/http/pprof/pprof.go#L98-L97
	// https://github.com/golang/go/commit/9fcffc53593c5cd103630d0d24ef8bd91e17246d
	prefix := ""
	//if godebug.New("httpmuxgo121").Value() != "1" { // todo, how to check it?
	if os.Getenv("PYROSCOPE_GODELTAPROF_HTTPMUXGO121") != "1" {
		prefix = "GET "
	}
	return prefix

}
