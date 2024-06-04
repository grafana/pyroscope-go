module github.com/grafana/pyroscope-go

go 1.17

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require github.com/grafana/pyroscope-go/godeltaprof v0.1.6

require github.com/klauspost/compress v1.17.8 // indirect
