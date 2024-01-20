package main

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/grafana/pyroscope-go"
	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
)

func main() {
	// Starting pyroscope profiler
	pyroscope.Start(pyroscope.Config{
		ApplicationName: "example-app",
		ServerAddress:   "http://pyroscope:4040",
	})

	// Setting up HTTP server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Standard pprof routes (copied from /net/http/pprof)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// This route is special: note that we're using Pyroscope handler here
	mux.HandleFunc("/debug/pprof/profile", pyroscope_pprof.Profile)

	go doMeaninglessWork()

	server.ListenAndServe()
}

// doMeaninglessWork does meaningless work to show CPU usage
func doMeaninglessWork() {
	for {
		for t := time.Now(); time.Now().Sub(t).Seconds() < 1; {
		}
	}
}
