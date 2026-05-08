module github.com/grafana/pyroscope-go/x/k6

go 1.25.0

replace github.com/grafana/pyroscope-go => ../../

require (
	github.com/grafana/pyroscope-go v1.2.8
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel v1.41.0
	google.golang.org/grpc v1.80.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.10 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
