module github.com/grafana/pyroscope-go/x/k6

go 1.22

replace github.com/grafana/pyroscope-go => ../../

require (
	github.com/grafana/pyroscope-go v1.1.1
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.24.0
	google.golang.org/grpc v1.67.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.8 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
