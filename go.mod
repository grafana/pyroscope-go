module github.com/grafana/pyroscope-go

go 1.25.0

toolchain go1.25.11

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.11
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
