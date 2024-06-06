package k6

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
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

		labelSet := getBaggageLabels(req)
		gotLabels := testPprofLabelsToMap(t, *labelSet)

		expectedLabels := map[string]string{}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("with K6Options", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testRequestWithBaggage(t, req, map[string]string{
			"k6.test_run_id":        "123",
			"not_k6.some_other_key": "value",
		})

		labelSet := getBaggageLabels(req)
		gotLabels := testPprofLabelsToMap(t, *labelSet)

		expectedLabels := map[string]string{
			"k6_test_run_id": "123",
		}
		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("does not allocate with failure to parse baggage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Add("Baggage", "invalid")

		labelSet := getBaggageLabels(req)
		require.Nil(t, labelSet)
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
