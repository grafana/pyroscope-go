# Pyroscope CPU Profiler HTTP Handler

This is an example of how you can use `github.com/grafana/pyroscope-go/http/pprof` package to use standard Go `/debug/pprof/profile` profiler along with Pyroscope continuous profiler.

### Problem

The standard Go pprof HTTP endpoint `/debug/pprof/profile` returns an error if profiling
is already started:

> Could not enable CPU profiling: CPU profiling already in use

This prevents you from using the standard Go pprof HTTP endpoint `/debug/pprof/profile` along with Pyroscope continuous profiler.

### Solution

This package facilitates the collection of CPU profiles via HTTP without disrupting the background operation of the Pyroscope profiler. It enables you to seamlessly gather CPU profiles through HTTP while continuously sending them to Pyroscope.

### How Does It Work

With each invocation of the handler, it suspends the Pyroscope profiler, gathers a CPU profile, dispatches the collected profile to both the caller and the Pyroscope profiler, and subsequently resumes the Pyroscope profiler.

### Recommendations

Standard `net/http/pprof` package registers its handlers automatically for the default HTTP server, but this package does not to avoid runtime errors due to the standard Go pprof HTTP endpoint `/debug/pprof/profile` being already registered.

It is highly recommended that you create a separate HTTP server and register pprof handlers on it. Then you can use Pyroscope's `github.com/grafana/pyroscope-go/http/pprof` package to register a handler that will collect CPU profiles and send them to Pyroscope while allowing you to still use the standard Go pprof HTTP endpoint `/debug/pprof/profile`.

### Example

```go
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
```

### Docker Compose Example

You can find a complete example of how to use this package in the [docker-compose.yml](./docker-compose.yml) file.

To run it:
```bash
docker-compose up --build
```

You can see Pyroscope data at [`http://localhost:4040/`](http://localhost:4040/?query=process_cpu%3Acpu%3Ananoseconds%3Acpu%3Ananoseconds%7Bservice_name%3D%22example-app%22%7D&from=now-5m), and if you want to see data from the standard Go pprof HTTP endpoint you can go to `http://localhost:8080/debug/pprof/profile`. Note that when you do that it does not disrupt the Pyroscope profiler:

```bash
curl http://localhost:8080/debug/pprof/profile > profile.pprof
pprof -http=:8081 profile.pprof
```

