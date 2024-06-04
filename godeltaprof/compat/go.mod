module github.com/grafana/pyroscope-go/godeltaprof/compat

go 1.18

require (
	github.com/google/pprof v0.0.0-20231127191134-f3a68a39ae15
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6
	github.com/grafana/pyroscope-go/godeltaprof/otlp v0.1.5
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/proto/otlp v1.2.0
	golang.org/x/tools v0.16.0
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.34.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//todo https://github.com/open-telemetry/opentelemetry-proto-go/pull/170
replace go.opentelemetry.io/proto/otlp => github.com/florianl/opentelemetry-proto-go/otlp v0.0.0-20240515144740-5317dc5b90ad

replace github.com/grafana/pyroscope-go/godeltaprof/otlp => ../otlp
