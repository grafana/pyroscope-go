module github.com/grafana/pyroscope-go

go 1.17

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.24.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
