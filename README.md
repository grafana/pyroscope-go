# Pyroscope Golang Client

This is a golang integration for Pyroscope — open source continuous profiling platform.

For more information, please visit our [golang integration documentation](https://pyroscope.io/docs/golang).

### Profiling Go applications
To start profiling a Go application, you need to include our go module in your app:
```
# make sure you also upgrade pyroscope server to version 0.3.1 or higher
go get github.com/pyroscope-io/client/pyroscope
```

Then add the following code to your application:

```go
package main

import "github.com/pyroscope-io/client/pyroscope"

func main() {
  pyroscope.Start(pyroscope.Config{
    ApplicationName: "simple.golang.app",

    // replace this with the address of pyroscope server
    ServerAddress:   "http://pyroscope-server:4040",

    // you can disable logging by setting this to nil
    Logger:          pyroscope.StandardLogger,

    // optionally, if authentication is enabled, specify the API key:
    // AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

    // by default all profilers are enabled,
    // but you can select the ones you want to use:
    ProfileTypes: []pyroscope.ProfileType{
      pyroscope.ProfileCPU,
      pyroscope.ProfileAllocObjects,
      pyroscope.ProfileAllocSpace,
      pyroscope.ProfileInuseObjects,
      pyroscope.ProfileInuseSpace,
    },
  })

  // your code goes here
}
```

### Tags

It is possible to add tags (labels) to the profiling data. These tags can be used to filter the data in the UI.

```go
// these two ways of adding tags are equivalent:
pyroscope.TagWrapper(context.Background(), pyroscope.Labels("controller", "slow_controller"), func(c context.Context) {
  slowCode()
})

pprof.Do(context.Background(), pprof.Labels("controller", "slow_controller"), func(c context.Context) {
  slowCode()
})
```

### Pull Mode

Go integration supports pull mode, which means that you can profile applications without adding any extra code. For that to work you will need to make sure you have profiling routes (`/debug/pprof`) enabled in your http server. Generally, that means that you need to add `net/http/pprof` package:
```go
import _ "net/http/pprof"
```

### Examples

Check out the [examples](https://github.com/pyroscope-io/pyroscope/tree/main/examples/golang-push) directory in our repository to learn more 🔥
