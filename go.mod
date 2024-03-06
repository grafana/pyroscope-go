module github.com/grafana/pyroscope-go

go 1.17

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require (
	github.com/google/pprof v0.0.0-20231127191134-f3a68a39ae15
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6
)

require github.com/klauspost/compress v1.17.3 // indirect
