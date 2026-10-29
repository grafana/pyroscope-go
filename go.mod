module github.com/grafana/pyroscope-go

go 1.25.0

toolchain go1.26.10

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.12
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/klauspost/compress v1.18.7 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)
