module github.com/grafana/pyroscope-go/x/k6

go 1.26.0

replace github.com/grafana/pyroscope-go => ../../

require (
	github.com/grafana/pyroscope-go v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/otel v1.47.0
	google.golang.org/grpc v1.83.2
)

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.12 // indirect
	github.com/klauspost/compress v1.18.7 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
