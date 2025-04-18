package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/grafana/pyroscope-go/upstream/remote"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/grafana/pyroscope-go"
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

func fastFunction(c context.Context, wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	pyroscope.TagWrapper(c, pyroscope.Labels("function", "fast"), func(c context.Context) {
		work(200000000)
	})
	wg.Done()
}

func slowFunction(c context.Context, wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	// standard pprof.Do wrappers work as well
	pprof.Do(c, pprof.Labels("function", "slow"), func(c context.Context) {
		work(800000000)
	})
	wg.Done()
}

func main() {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	config := pyroscope.Config{
		ApplicationName:   "simple.golang.app-new",
		ServerAddress:     "http://localhost:4040",
		Logger:            pyroscope.StandardLogger,
		AuthToken:         os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		TenantID:          os.Getenv("PYROSCOPE_TENANT_ID"),
		BasicAuthUser:     os.Getenv("PYROSCOPE_BASIC_AUTH_USER"),
		BasicAuthPassword: os.Getenv("PYROSCOPE_BASIC_AUTH_PASSWORD"),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
		HTTPHeaders: map[string]string{"X-Extra-Header": "extra-header-value"},
	}
	httpClient, err := createHttpClient()
	if err != nil {
		log.Fatal(err)
	}
	err = start(config, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	pyroscope.TagWrapper(context.Background(), pyroscope.Labels("foo", "bar"), func(c context.Context) {
		for {
			wg := sync.WaitGroup{}
			wg.Add(2)
			go fastFunction(c, &wg)
			go slowFunction(c, &wg)
			wg.Wait()
		}
	})
}

func start(cfg pyroscope.Config, httpClient remote.HTTPClient) error {
	rc := remote.Config{
		TenantID:          cfg.TenantID,
		BasicAuthUser:     cfg.BasicAuthUser,
		BasicAuthPassword: cfg.BasicAuthPassword,
		HTTPHeaders:       cfg.HTTPHeaders,
		Address:           cfg.ServerAddress,
		Threads:           5, // per each profile type upload
		Timeout:           30 * time.Second,
		Logger:            cfg.Logger,
		HTTPClient:        httpClient,
	}
	uploader, err := remote.NewRemote(rc)
	if err != nil {
		return err
	}

	sc := pyroscope.SessionConfig{
		Upstream:               uploader,
		Logger:                 cfg.Logger,
		AppName:                cfg.ApplicationName,
		Tags:                   cfg.Tags,
		ProfilingTypes:         cfg.ProfileTypes,
		DisableGCRuns:          cfg.DisableGCRuns,
		DisableAutomaticResets: cfg.DisableAutomaticResets,
		UploadRate:             cfg.UploadRate,
	}

	s, err := pyroscope.NewSession(sc)
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	uploader.Start()
	if err = s.Start(); err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	return nil
}

func createHttpClient() (remote.HTTPClient, error) {
	cert, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	clientCert := fmt.Sprintf("./certs/client.crt")
	clientKey := fmt.Sprintf("./certs/client.key")
	log.Println("Load key pairs - ", clientCert, clientKey)
	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("could not load certificate: %v", err)
	}

	client := &http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	return client, nil
}
