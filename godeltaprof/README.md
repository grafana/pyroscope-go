# godeltaprof

godeltaprof is an efficient delta profiler for memory, mutex, block.

# Why

In golang, allocation, mutex and block profiles are cumulative - they only grow over time and show allocations happened since the beginning of the running program.
Not only values grow, but the size of the profile itself grows as well. It could grow up to megabytes in size for long-running processes. These megabytes profiles are called huge profiles in this document.

In many cases it is much more useful to see the difference between two points in time.

There is a way to do it with the original runtime/pprof package, they call it delta profile, it requires passing seconds argument to pprof endpoint query.

```
go tool pprof http://localhost:6060/debug/pprof/heap?seconds=30
```

What this does:
1. dump profile p0
2. sleep
3. dump profile p1
4. decompress and parse protobuf p0
5. decompress and parse protobuf p1
6. subtract p0 from p1
7. serialize protobuf and compress the result

The resulting profile is *usually* much smaller(p0 may be megabytes, while result is usually tens of kilobytes).

There are number of issues with this approach.
1. Heap profile contains both allocation values and inuse values. Inuse values are not cumulative. So inuse values are corrupted by the subtraction.
Note: it can be fixed if runtime/pprof package would do the following `p0.ScaleN([]float64{-1,-1,0,0})` instead of `p0.Scale(-1)` - that would substract allocation values and zero out inuse values in p0.
2. It requires dumping two profiles.
3. It produces a lot of allocations putting pressure on GC.


## DataDog's fastdelta

Another approach found is DataDog's [fastdelta profiler](https://github.com/DataDog/dd-trace-go/blob/30e1406c2cb62af749df03d559853e1d1de0e3bf/profiler/internal/fastdelta/fd.go#L75)

It improves the runtime/pprof approach by keeping a copy of the previous profile and subtracting the current profile from it.
It also does it in a more efficient way by using a custom protobuf pprof parser that doesn't allocate as much memory.
So it is much faster and produces less garbage. Also does not require dumping two profiles.
However, it still parses huge profiles up to megabytes, just to discard most of it.

## godeltaprof

godeltaprof does a similar job but slightly differently.

The main difference is that delta computation happens before serializing any pprof files using `runtime.MemprofileRecord` and `BlockProfileRecord`
This way we don't need to parse huge profiles, we compute the delta on raw records, reject all zeros and serialize and compress the result.
We dont parse huge profiles, we serialize and compress only small delta profiles.

The source code for godeltaprof is based(forked) on the original [runtime/pprof package](https://github.com/golang/go/tree/master/src/runtime/pprof)  .
It is modified to include delta computation before serialization and to expose the new endpoints.
There are other small improvements/benefits:
- using `github.com/klauspost/compress/gzip` instead of `compress/gzip`
- optional lazy mappings reading (they don't change over time for most applications)
- it is a separate package from runtime, so we can update it independently of 

# benchmarks

I used memory profiles from [pyroscope](https://github.com/grafana/pyroscope) server.

BenchmarkOG - dumps memory profile with runtime/pprof package
BenchmarkFastDelta - dumps memory profile with runtime/pprof package and computes delta using fastdelta
BenchmarkGodeltaprof - does not dump profile with runtime/pprof, computes delta, outputs it results

Each benchmark also outputs produced profile sizes.
```
BenchmarkOG
       3	 703322458 ns/op
profile sizes: [211872 212311 212039 212915 213467 214227 214152 214235 216195 217801 217875 218251 218094 218079 218157]

BenchmarkFastDelta
       2	 603072104 ns/op
profile sizes: [173458 47415 43212 48645 49923 42830 14556 15618 21177 24979 20851 18199 13250 13198 13577]

BenchmarkGodeltaprof
      15	  92276847 ns/op
profile sizes: [218803 56600 52205 58711 60330 51712 15659 16926 24197 28538 23579 20280 14199 14367 14627]
```

Notice how BenchmarkOG profiles are ~200k and BenchmarkGodeltaprof and BenchmarkFastDelta are ~14k - that is because a lof of samples
with zero values are discarded after delta computation.

Source code of benchmarks could be found [here](https://github.com/grafana/pyroscope/compare/godeltaprofbench?expand=1) 

CPU profiles: [BenchmarkOG](https://flamegraph.com/share/665822d1-9819-11ee-a502-466f68d203a5), [BenchmarkFastDelta](https://flamegraph.com/share/b06774de-9819-11ee-9a0d-f2c25703e557),  [BenchmarkGodeltaprof]( https://flamegraph.com/share/192c77c5-9819-11ee-a502-466f68d203a5)



# upstreaming

TODO(korniltsev): create golang issue and ask if godeltaprof is something that could be considered merging to upstream golang repo
in some way(maybe not as is, maybe with different APIs)



