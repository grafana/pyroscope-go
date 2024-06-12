module github.com/grafana/pyroscope-go/godeltaprof/otlp

go 1.19

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6
	go.opentelemetry.io/proto/otlp v1.3.1
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/grafana/pyroscope-go/godeltaprof => ../
