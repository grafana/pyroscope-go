module github.com/grafana/pyroscope-go

go 1.21

toolchain go1.23.1

replace github.com/grafana/pyroscope-go/godeltaprof => ./godeltaprof

require github.com/grafana/pyroscope-go/godeltaprof v0.1.6

require (
	github.com/klauspost/compress v1.17.8 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
