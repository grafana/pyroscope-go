# Pyroscope Go SDK k6 extension

This library provides extension functions to provide support for Grafana's
integration between Pyroscope and k6. Namely, it provides HTTP and gRPC
middleware to dynamically label the profiling context with k6 test metadata.

> [!CAUTION]
> Maintainers: Be aware this project has its own `go.work` file and is built
> independently of the rest of the Go SDK. This is because this project has a
> dependency on `google.golang.org/grpc` which is only supported for the last
> 2 major versions of Go. Since it's not acceptable to extend this restriction
> to the remainder of the SDK, this module remains independent from the rest of
> the code base.
