package main

import (
	"net/http"
	_ "net/http/pprof" // register pprof HTTP handlers.

	pyroscope "github.com/grafana/pyroscope-go/http/pprof"
)

func init() {
	// Optionally, you can also register Pyroscope handler, that enables
	// you to seamlessly gather CPU profiles through HTTP while continuously
	// sending them to Pyroscope.
	go func() {
		http.HandleFunc("/debug/pprof/cpu", pyroscope.Profile)
		_ = http.ListenAndServe(":8080", nil)
	}()
}
