module github.com/grafana/pyroscope-go/godeltaprof/otlp

go 1.19

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6
	go.opentelemetry.io/proto/otlp v1.2.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/grafana/pyroscope-go/godeltaprof => ../

//todo https://github.com/open-telemetry/opentelemetry-proto-go/pull/170
replace go.opentelemetry.io/proto/otlp => github.com/florianl/opentelemetry-proto-go/otlp v0.0.0-20240515144740-5317dc5b90ad
