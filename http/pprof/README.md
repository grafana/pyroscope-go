# Pyroscope CPU Profiler HTTP Handler

This package facilitates the collection of CPU profiles via HTTP without disrupting
the background operation of the Pyroscope profiler. It enables you to seamlessly gather
CPU profiles through HTTP while continuously sending them to Pyroscope.

The standard Go pprof HTTP endpoint `/debug/pprof/profile` returns an error if profiling
is already started:

> Could not enable CPU profiling: CPU profiling already in use

The Pyroscope CPU Profiler HTTP handler serve this gracefully by communicating with
the Pyroscope profiler, which collects profiles in the background.

## Usage

The package does not register the handler automatically. It is highly recommended to
avoid using the standard path `/debug/pprof/profile` and the default mux because
attempting to register the handler on the same path will cause a panic. In many cases,
the `net/http/pprof` package is imported by dependencies, and therefore there is no
reliable way to avoid the conflict.

```go
import (
    "net/http"

    "github.com/grafana/pyroscope-go/http/pprof"
)

func main() {
	http.Handle("/debug/pprof/cpu", pprof.Profile())
}
```

With each invocation of the handler, it suspends the Pyroscope profiler, gathers a CPU
profile, dispatches the collected profile to both the caller and the Pyroscope profiler,
and subsequently resumes the profiler.
