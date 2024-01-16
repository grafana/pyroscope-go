package pprof

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	internal "github.com/grafana/pyroscope-go/internal/pprof"
)

// Profile responds with the pprof-formatted cpu profile.
// Profiling lasts for duration specified in seconds GET parameter, or for 30 seconds if not specified.
// The package initialization registers it as /debug/pprof/profile.
func Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec <= 0 || err != nil {
		sec = 30
	}

	if durationExceedsWriteTimeout(r, float64(sec)) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}

	// Set Content Type assuming StartCPUProfile will work,
	// because if it does it start writing.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sec)*time.Second)
	defer cancel()
	if err = collectCPUProfile(ctx, w); err != nil {
		serveError(w, http.StatusInternalServerError, fmt.Sprintf("Could not enable CPU profiling: %s", err))
	}
}

func durationExceedsWriteTimeout(r *http.Request, seconds float64) bool {
	srv, ok := r.Context().Value(http.ServerContextKey).(*http.Server)
	return ok && srv.WriteTimeout != 0 && seconds >= srv.WriteTimeout.Seconds()
}

func serveError(w http.ResponseWriter, status int, txt string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Go-Pprof", "1")
	w.Header().Del("Content-Disposition")
	w.WriteHeader(status)
	_, _ = fmt.Fprintln(w, txt)
}

func collectCPUProfile(ctx context.Context, w io.Writer) error {
	if err := internal.StartCPUProfile(w); err != nil {
		return err
	}
	defer internal.StopCPUProfile()
	<-ctx.Done()
	return nil
}
