package main

import (
	"context"
	"fmt"
	"github.com/grafana/pyroscope-go/godeltaprof"
	"github.com/grafana/pyroscope-go/godeltaprof/otlp"
	otlpcollector "go.opentelemetry.io/proto/otlp/collector/profiles/v1experimental"
	otlpcommon "go.opentelemetry.io/proto/otlp/common/v1"
	otlpprofile "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
	"go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

//go:noinline
func work(n int) {
	// revive:disable:empty-block this is fine because this is a example app, not real production code
	for i := 0; i < n; i++ {
	}
	fmt.Printf("work\n")
	// revive:enable:empty-block
}

var m sync.Mutex

func fastFunction(wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	work(200000000)
	wg.Done()
}

func slowFunction(wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	// standard pprof.Do wrappers work as well
	work(800000000)
	wg.Done()
}

func exportLoop() {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	con, err := grpc.DialContext(context.Background(), "localhost:9095", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	svc := otlpcollector.NewProfilesServiceClient(con)

	mutex := otlp.NewMutexProfilerWithOptions(godeltaprof.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})

	block := otlp.NewBlockProfilerWithOptions(godeltaprof.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})

	heap := otlp.NewHeapProfilerWithOptions(godeltaprof.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})

	profiles := []func() (*otlpprofile.Profile, error){
		mutex.Profile,
		block.Profile,
		heap.Profile,
	}

	_ = heap
	_ = block
	_ = mutex

	for {
		startTime := time.Now()
		time.Sleep(5 * time.Second)
		endTime := time.Now()
		req := &otlpcollector.ExportProfilesServiceRequest{
			ResourceProfiles: []*otlpprofile.ResourceProfiles{
				{
					Resource: &v1.Resource{
						Attributes: []*otlpcommon.KeyValue{
							{
								Key: "service_name",
								Value: &otlpcommon.AnyValue{
									Value: &otlpcommon.AnyValue_StringValue{StringValue: "otlp_example"},
								},
							},
						},
					},
					ScopeProfiles: []*otlpprofile.ScopeProfiles{{}},
				},
			},
		}
		for _, profile := range profiles {
			data, err := profile()
			if err != nil {
				fmt.Printf("profile dump error: %v\n", err)
			} else {
				id := make([]byte, 16)
				rand.Read(id)
				req.ResourceProfiles[0].ScopeProfiles[0].Profiles = append(req.ResourceProfiles[0].ScopeProfiles[0].Profiles, &otlpprofile.ProfileContainer{
					StartTimeUnixNano: uint64(startTime.UnixNano()),
					EndTimeUnixNano:   uint64(endTime.UnixNano()),
					ProfileId:         id,
					Profile:           data,
				})
			}
		}
		res, err := svc.Export(context.Background(), req)
		if err != nil {
			fmt.Printf("export error: %v\n", err)
		} else {
			fmt.Printf("export response: %v\n", res)
		}
	}
}

func main() {
	go exportLoop()

	for {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go fastFunction(&wg)
		go slowFunction(&wg)
		wg.Wait()
	}
}
