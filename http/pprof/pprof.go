package pprof

import (
	"fmt"
	"github.com/pyroscope-io/client/delta"
	"io"
	"net/http"
	"runtime"
	"strconv"
)

var deltaHeapProfiler = delta.NewHeapProfiler()
var deltaBlockProfiler = delta.NewBlockProfiler()
var deltaMutexProfiler = delta.NewMutexProfiler()

type deltaProfiler interface {
	Profile(w io.Writer) error
}

func init() {
	http.HandleFunc("/debug/pprof/delta_heap", Heap)
	http.HandleFunc("/debug/pprof/delta_block", Block)
	http.HandleFunc("/debug/pprof/delta_mutex", Mutex)
}

func Heap(w http.ResponseWriter, r *http.Request) {
	WriteDeltaProfile(&deltaHeapProfiler, "heap", w, r)
}

func Block(w http.ResponseWriter, r *http.Request) {
	WriteDeltaProfile(&deltaBlockProfiler, "block", w, r)
}

func Mutex(w http.ResponseWriter, r *http.Request) {
	WriteDeltaProfile(&deltaMutexProfiler, "mutex", w, r)
}

func WriteDeltaProfile(p deltaProfiler, name string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	gc, _ := strconv.Atoi(r.FormValue("gc"))
	if gc > 0 {
		runtime.GC()
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pprof.gz"`, name))
	_ = p.Profile(w)
}
