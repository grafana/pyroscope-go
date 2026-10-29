package pprof_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof" //nolint:gosec // Exercise registration together; no network listener is started.
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	deltapprof "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"
)

func TestDefaultMuxRegistration(t *testing.T) {
	// Both packages register handlers during initialization. The mux mode is
	// also selected at startup, so changing GODEBUG in this process is too late.
	if mode := os.Getenv("GODELTAPROF_HTTP_TEST_MODE"); mode != "" {
		recorder := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("stdlib pprof index returned %d: %s", recorder.Code, recorder.Body.String())
		}

		if mode == "0" {
			for _, name := range []string{"heap", "block", "mutex"} {
				path := "/debug/pprof/delta_" + name
				recorder = httptest.NewRecorder()
				http.DefaultServeMux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
				checkDeltaProfile(t, recorder, name)

				recorder = httptest.NewRecorder()
				http.DefaultServeMux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
				if recorder.Code != http.StatusMethodNotAllowed {
					t.Fatalf("POST %s returned %d, want %d", path, recorder.Code, http.StatusMethodNotAllowed)
				}
			}
		}

		return
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	for _, mode := range []string{"0", "1"} {
		t.Run("httpmuxgo121="+mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			//nolint:gosec // Run this test executable, not external input.
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestDefaultMuxRegistration$", "-test.v")
			cmd.Env = append(os.Environ(),
				"GODEBUG="+os.Getenv("GODEBUG")+",httpmuxgo121="+mode,
				"GODELTAPROF_HTTP_TEST_MODE="+mode,
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("pprof registration with httpmuxgo121=%s failed: %v\n%s", mode, err, output)
			}
		})
	}
}

func TestHandlers(t *testing.T) {
	for _, handler := range []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "heap", fn: deltapprof.Heap},
		{name: "block", fn: deltapprof.Block},
		{name: "mutex", fn: deltapprof.Mutex},
	} {
		t.Run(handler.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.fn(recorder, httptest.NewRequest(http.MethodGet, "/debug/pprof/delta_"+handler.name, nil))
			checkDeltaProfile(t, recorder, handler.name)
		})
	}
}

func TestHeapGC(t *testing.T) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	recorder := httptest.NewRecorder()
	deltapprof.Heap(recorder, httptest.NewRequest(http.MethodGet, "/debug/pprof/delta_heap?gc=1", nil))
	runtime.ReadMemStats(&after)

	if after.NumGC <= before.NumGC {
		t.Fatal("heap handler did not run the requested garbage collection")
	}
	checkDeltaProfile(t, recorder, "heap")
}

func checkDeltaProfile(t *testing.T, recorder *httptest.ResponseRecorder, name string) {
	t.Helper()

	if recorder.Code != http.StatusOK {
		t.Fatalf("%s profile returned %d: %s", name, recorder.Code, recorder.Body.String())
	}
	for header, want := range map[string]string{
		"Content-Type":           "application/octet-stream",
		"Content-Disposition":    `attachment; filename="` + name + `.pprof.gz"`,
		"X-Content-Type-Options": "nosniff",
	} {
		if got := recorder.Header().Get(header); got != want {
			t.Errorf("%s: %s = %q, want %q", name, header, got, want)
		}
	}

	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("%s profile is not gzip: %v", name, err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	})
	profile, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading %s profile: %v", name, err)
	}
	if len(profile) == 0 {
		t.Errorf("%s profile has an empty payload", name)
	}
	if !strings.Contains(string(profile), "inuse_space") && name == "heap" {
		t.Error("heap profile is missing its inuse_space sample type")
	}
}
