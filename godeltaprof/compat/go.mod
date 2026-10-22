module github.com/grafana/pyroscope-go/godeltaprof/compat

go 1.25.0

toolchain go1.26.10

replace github.com/grafana/pyroscope-go/godeltaprof => ../

require (
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936
	github.com/grafana/pyroscope-go/godeltaprof v0.1.12
	github.com/klauspost/compress v1.18.7
	github.com/stretchr/testify v1.12.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect
