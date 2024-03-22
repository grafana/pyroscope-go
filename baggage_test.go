package pyroscope

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
)

func Test_getBaggageLabels(t *testing.T) {
	t.Run("empty values are skipped", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"blank": "",
		})

		filters := []FilterFunc{}
		transforms := []TransformFunc{}

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("no filters or transforms", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		filters := []FilterFunc{}
		transforms := []TransformFunc{}

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("with filters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		filters := []FilterFunc{
			func(key string) bool {
				return strings.HasPrefix(key, "k6.")
			},
		}
		transforms := []TransformFunc{}

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{
			"k6.test_run_id": "123",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("with transforms", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		filters := []FilterFunc{}
		transforms := []TransformFunc{
			func(key string) string {
				return strings.ReplaceAll(key, ".", "_")
			},
		}

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{
			"k6_test_run_id":        "123",
			"not_k6_some_other_key": "value",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("with filters and transforms", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		filters := []FilterFunc{
			func(key string) bool {
				return strings.HasPrefix(key, "k6.")
			},
		}
		transforms := []TransformFunc{
			func(key string) string {
				return strings.ReplaceAll(key, ".", "_")
			},
		}

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{
			"k6_test_run_id": "123",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("with K6Options", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		cfg := &BaggageConfig{}
		for _, opt := range K6Options() {
			opt(cfg)
		}
		filters := cfg.filters
		transforms := cfg.transforms

		labelSet := getBaggageLabels(req, filters, transforms)
		gotLabels := testPprofLabelsToMap(t, labelSet)

		expectedLabels := map[string]string{
			"k6_test_run_id": "123",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})
}

func testRequestWithBaggage(t *testing.T, req *http.Request, bag map[string]string) *http.Request {
	t.Helper()

	members := []baggage.Member{}
	for k, v := range bag {
		member, err := baggage.NewMember(k, v)
		require.NoError(t, err)

		members = append(members, member)
	}

	b, err := baggage.New(members...)
	require.NoError(t, err)

	req.Header.Add("Baggage", b.String())
	return req
}

func testPprofLabelsToMap(t *testing.T, labelSet pprof.LabelSet) map[string]string {
	t.Helper()

	gotLabels := map[string]string{}
	ctx := pprof.WithLabels(context.Background(), labelSet)
	pprof.ForLabels(ctx, func(key, value string) bool {
		gotLabels[key] = value
		return true
	})

	return gotLabels
}
