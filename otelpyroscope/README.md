# OpenTelemetry tracing integration

The package provides means to integrate tracing with profiling. More specifically, a `TracerProvider` implementation,
that annotates profiling data with span IDs: when a new trace span emerges, the tracer adds a `span_id` [pprof tag](https://github.com/google/pprof/blob/master/doc/README.md#tag-filtering)
that points to the span. This makes it possible to filter out a profile of a particular trace span in [Pyroscope](https://pyroscope.io).

Note that the module does not control `pprof` profiler itself – it still needs to be started for profiles to be
collected. This can be done either via `runtime/pprof` package, or using the [Pyroscope client](https://github.com/grafana/pyroscope-go).

By default, only the root span gets labeled (the first span created locally): such spans are marked with the
`pyroscope.profile.id` attribute set to the span ID. Please note that presence of the attribute does not necessarily
indicate that the span has a profile: stack trace samples might not be collected, if the utilized CPU time is
less than the sample interval (10ms).

Limitations:
- Only CPU profiling is fully supported at the moment.

## Example

You can find a complete example setup in the [Pyroscope repository](https://github.com/grafana/pyroscope/tree/main/examples/tracing/tempo).
