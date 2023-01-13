package pprof

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/pprof/profile"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

func init() {
	http.HandleFunc("/pyroscope/pprof/heap", Heap)
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
	fmt.Fprintln(w, txt)
}

const name = "heap"

func Heap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	p := pprof.Lookup(name)
	if p == nil {
		serveError(w, http.StatusNotFound, "Unknown profile")
		return
	}
	if sec := r.FormValue("seconds"); sec != "" {
		serveDeltaHeapProfile(w, r, p, sec)
		return
	}
	gc, _ := strconv.Atoi(r.FormValue("gc"))
	if gc > 0 {
		runtime.GC()
	}
	debug, _ := strconv.Atoi(r.FormValue("debug"))
	if debug != 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	}
	p.WriteTo(w, debug)
}

func serveDeltaHeapProfile(w http.ResponseWriter, r *http.Request, p *pprof.Profile, secStr string) {
	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil || sec <= 0 {
		serveError(w, http.StatusBadRequest, `invalid value for "seconds" - must be a positive integer`)
		return
	}
	if durationExceedsWriteTimeout(r, float64(sec)) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}
	debug, _ := strconv.Atoi(r.FormValue("debug"))
	if debug != 0 {
		serveError(w, http.StatusBadRequest, "seconds and debug params are incompatible")
		return
	}
	p0, err := collectHeapProfile(p)
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to collect profile")
		return
	}

	t := time.NewTimer(time.Duration(sec) * time.Second)
	defer t.Stop()

	select {
	case <-r.Context().Done():
		err := r.Context().Err()
		if err == context.DeadlineExceeded {
			serveError(w, http.StatusRequestTimeout, err.Error())
		} else { // TODO: what's a good status code for canceled requests? 400?
			serveError(w, http.StatusInternalServerError, err.Error())
		}
		return
	case <-t.C:
	}

	p1, err := collectHeapProfile(p)
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to collect profile")
		return
	}
	ts := p1.TimeNanos
	dur := p1.TimeNanos - p0.TimeNanos

	err = p0.ScaleN([]float64{-1, -1, 0, 0}) // subtract alloc* remove inuse*
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to collect profile")
		return
	}

	p1, err = profile.Merge([]*profile.Profile{p0, p1})
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to compute delta")
		return
	}

	p1.TimeNanos = ts // set since we don't know what profile.Merge set for TimeNanos.
	p1.DurationNanos = dur

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-delta"`, name))
	p1.Write(w)
}

func collectHeapProfile(p *pprof.Profile) (*profile.Profile, error) {
	var buf bytes.Buffer
	if err := p.WriteTo(&buf, 0); err != nil {
		return nil, err
	}
	ts := time.Now().UnixNano()
	p0, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}
	p0.TimeNanos = ts

	if got := len(p0.SampleType); got != 4 {
		return nil, fmt.Errorf("invalid heap profile: got %d sample types, want 4", got)
	}
	for i, want := range []string{"alloc_objects", "alloc_space", "inuse_objects", "inuse_space"} {
		if got := p0.SampleType[i].Type; got != want {
			return nil, fmt.Errorf("invalid heap profile: got %q sample type at index %d, want %q", got, i, want)
		}
	}
	return p0, nil
}
