package k6

import (
	"net/http"
	"runtime/pprof"
	"strings"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/baggage"
)

// LabelsFromBaggageHandler is a middleware that will extract key-value pairs
// from the request baggage and make them profiling labels.
func LabelsFromBaggageHandler(handler http.Handler) http.Handler {
	lh := &labelHandler{
		innerHandler: handler,
	}

	return lh
}

type labelHandler struct {
	innerHandler http.Handler
}

func (lh *labelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var found bool
	r, found = setBaggageContextFromHeader(r)
	if !found {
		lh.innerHandler.ServeHTTP(w, r)
		return
	}

	labels := getBaggageLabels(r)
	if labels == nil {
		lh.innerHandler.ServeHTTP(w, r)
		return
	}

	// Inlined version of pyroscope.TagWrapper and pprof.Do to reduce noise in
	// the stack trace.
	ctx := r.Context()
	defer pprof.SetGoroutineLabels(ctx)
	ctx = pprof.WithLabels(ctx, *labels)
	pprof.SetGoroutineLabels(ctx)

	lh.innerHandler.ServeHTTP(w, r.WithContext(ctx))
}

func setBaggageContextFromHeader(r *http.Request) (*http.Request, bool) {
	baggageHeader := r.Header.Get("Baggage")
	if baggageHeader == "" {
		return r, false
	}

	b, err := baggage.Parse(baggageHeader)
	if err != nil {
		return r, false
	}

	ctx := baggage.ContextWithBaggage(r.Context(), b)
	return r.WithContext(ctx), true
}

// getBaggageLabels applies filters and transformations to request baggage and
// returns the resulting LabelSet.
func getBaggageLabels(r *http.Request) *pyroscope.LabelSet {
	b := baggage.FromContext(r.Context())
	if b.Len() == 0 {
		return nil
	}

	labels := baggageToLabels(b)
	return &labels
}

// baggageToLabels converts request baggage to a LabelSet.
func baggageToLabels(b baggage.Baggage) pyroscope.LabelSet {
	labelPairs := make([]string, 0, len(b.Members())*2)
	for _, m := range b.Members() {
		if !strings.HasPrefix(m.Key(), "k6.") {
			continue
		}

		if m.Value() == "" {
			continue
		}

		key := strings.ReplaceAll(m.Key(), ".", "_")
		labelPairs = append(labelPairs, key, m.Value())
	}

	return pyroscope.Labels(labelPairs...)
}
