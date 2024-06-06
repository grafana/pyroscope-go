package k6

import (
	"context"
	"net/http"
	"strings"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/baggage"
)

// FilterFunc returns true if this key should be used.
type FilterFunc func(key string) bool

// TransformFunc transforms the key.
type TransformFunc func(baggage string) string

// BaggageConfig contains configuration options for filtering and transforming
// baggage members.
type BaggageConfig struct {
	filters    []FilterFunc
	transforms []TransformFunc
}

// BaggageOption sets a option on a BaggageConfig.
type BaggageOption func(config *BaggageConfig)

// WithFilters sets filtering functions to apply to the baggage.
func WithFilters(filters ...FilterFunc) BaggageOption {
	return func(config *BaggageConfig) {
		config.filters = append(config.filters, filters...)
	}
}

// WithTransforms sets transformation functions to apply to the baggage.
func WithTransforms(transforms ...TransformFunc) BaggageOption {
	return func(config *BaggageConfig) {
		config.transforms = append(config.transforms, transforms...)
	}
}

// K6Options provides default options to select k6 members from the baggage.
func K6Options() []BaggageOption {
	return []BaggageOption{
		WithFilters(func(key string) bool {
			return strings.HasPrefix(key, "k6.")
		}),
		WithTransforms(func(key string) string {
			return strings.ReplaceAll(key, ".", "_")
		}),
	}
}

// LabelsFromBaggageHandler is a middleware that will extract key-value pairs
// from the request baggage and make them profiling labels. Filtering options
// will be applied first, followed by transformation options.
func LabelsFromBaggageHandler(handler http.Handler, opts ...BaggageOption) http.Handler {
	lh := &labelHandler{
		innerHandler: handler,
		cfg: BaggageConfig{
			filters:    []FilterFunc{},
			transforms: []TransformFunc{},
		},
	}

	for _, opt := range opts {
		opt(&lh.cfg)
	}

	return lh
}

type labelHandler struct {
	innerHandler http.Handler
	cfg          BaggageConfig
}

func (lh *labelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	labels := getBaggageLabels(r, lh.cfg.filters, lh.cfg.transforms)
	if labels == nil {
		lh.innerHandler.ServeHTTP(w, r)
		return
	}

	pyroscope.TagWrapper(r.Context(), *labels, func(ctx context.Context) {
		lh.innerHandler.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getBaggageLabels applies filters and transformations to request baggage and
// returns the resulting LabelSet.
func getBaggageLabels(r *http.Request, filters []FilterFunc, transforms []TransformFunc) *pyroscope.LabelSet {
	b, err := baggage.Parse(r.Header.Get("Baggage"))
	if err != nil {
		return nil
	}

	labels := make([]string, 0, b.Len()*2)
Outer:
	for _, m := range b.Members() {
		if len(m.Value()) == 0 {
			// Skip keys with no value
			continue
		}

		for _, filter := range filters {
			if !filter(m.Key()) {
				continue Outer
			}
		}

		key := m.Key()
		for _, transform := range transforms {
			key = transform(key)
		}
		labels = append(labels, key, m.Value())
	}

	lbls := pyroscope.Labels(labels...)
	return &lbls
}
