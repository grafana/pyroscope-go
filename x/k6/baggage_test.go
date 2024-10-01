package k6

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
)

func TestLabelsFromBaggageHandler(t *testing.T) {
	t.Run("adds_baggage_to_context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testAddBaggageToRequest(t, req,
			"k6.test_run_id", "123",
			"not_k6.some_other_key", "value",
		)

		handler := LabelsFromBaggageHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := baggage.FromContext(r.Context())
			require.NotNil(t, b)

			testAssertEqualMembers(t, b.Members(),
				"k6.test_run_id", "123",
				"not_k6.some_other_key", "value",
			)
		}))

		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("passthrough_requests_with_no_baggage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)

		handler := LabelsFromBaggageHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := baggage.FromContext(r.Context())
			require.Equal(t, 0, b.Len())
		}))

		handler.ServeHTTP(httptest.NewRecorder(), req)
	})
}

func Test_setBaggageContextFromHeader(t *testing.T) {
	t.Run("sets_baggage_context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testAddBaggageToRequest(t, req,
			"k6.test_run_id", "123",
			"not_k6.some_other_key", "value",
		)

		req, found := setBaggageContextFromHeader(req)
		require.True(t, found)

		b := baggage.FromContext(req.Context())
		testAssertEqualMembers(t, b.Members(),
			"k6.test_run_id", "123",

			// Also passthrough non-k6 baggage. We should avoid clobbering
			// other baggage members that the system might also be using.
			"not_k6.some_other_key", "value",
		)
	})

	t.Run("does_not_set_baggage_context_with_no_baggage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)

		req, found := setBaggageContextFromHeader(req)
		require.False(t, found)

		b := baggage.FromContext(req.Context())
		require.Equal(t, 0, b.Len())
	})

	t.Run("does_not_set_baggage_context_invalid_baggage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Add("Baggage", "invalid")

		req, found := setBaggageContextFromHeader(req)
		require.False(t, found)

		b := baggage.FromContext(req.Context())
		require.Equal(t, 0, b.Len())
	})
}

func Test_getBaggageLabels(t *testing.T) {
	t.Run("with_k6_baggage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testAddBaggageToRequest(t, req,
			"k6.test_run_id", "123",
			"not_k6.some_other_key", "value",
		)

		labelSet := getBaggageLabelsFromContext(req.Context())
		require.NotNil(t, labelSet)

		gotLabels := testPprofLabelsToMap(t, *labelSet)
		expectedLabels := map[string]string{
			"k6_test_run_id": "123",
		}

		require.Equal(t, expectedLabels, gotLabels)
	})

	t.Run("empty_values_are_skipped", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req = testAddBaggageToRequest(t, req)

		labelSet := getBaggageLabelsFromContext(req.Context())
		require.Nil(t, labelSet)
	})

	t.Run("skips_missing_baggage_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)

		labelSet := getBaggageLabelsFromContext(req.Context())
		require.Nil(t, labelSet)
	})

	t.Run("skips_invalid_baggage_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Add("Baggage", "invalid")

		labelSet := getBaggageLabelsFromContext(req.Context())
		require.Nil(t, labelSet)
	})
}

func testAddBaggageToRequest(t *testing.T, req *http.Request, kvPairs ...string) *http.Request {
	t.Helper()

	require.Equal(t, 0, len(kvPairs)%2, "kvPairs must be a multiple of 2")

	members := make([]baggage.Member, 0, len(kvPairs)/2)
	for i := 0; i < len(kvPairs); i += 2 {
		key := kvPairs[i]
		value := kvPairs[i+1]

		member, err := baggage.NewMember(key, value)
		require.NoError(t, err)

		members = append(members, member)
	}

	b, err := baggage.New(members...)
	require.NoError(t, err)

	ctx := baggage.ContextWithBaggage(req.Context(), b)
	req = req.WithContext(ctx)
	req.Header.Add("Baggage", b.String())

	return req
}

func testMustNewMember(t *testing.T, key string, value string) baggage.Member {
	t.Helper()

	member, err := baggage.NewMember(key, value)
	require.NoError(t, err)

	return member
}

// testAssertEqualMembers verifies the two slices of Members are equal by
// sorting them and comparing them as pairs of key-value strings.
//
// This is necessary because Baggage.Members() returns an unordered slice of
// Members.
func testAssertEqualMembers(t *testing.T, got []baggage.Member, wants ...string) {
	t.Helper()

	type Pair struct {
		Key   string
		Value string
	}

	require.Equal(t, 0, len(wants)%2, "wants must be a multiple of 2")
	require.Equal(t, len(wants)/2, len(got))

	wantPairs := make([]Pair, 0, len(wants)/2)
	for i := 0; i < len(wants); i += 2 {
		key := wants[i]
		value := wants[i+1]

		wantPairs = append(wantPairs, Pair{
			Key:   key,
			Value: value,
		})
	}

	gotPairs := make([]Pair, 0, len(got))
	for _, m := range got {
		gotPairs = append(gotPairs, Pair{m.Key(), m.Value()})
	}

	sort.Slice(wantPairs, func(i, j int) bool {
		return wantPairs[i].Key < wantPairs[j].Key
	})
	sort.Slice(gotPairs, func(i, j int) bool {
		return gotPairs[i].Key < gotPairs[j].Key
	})

	require.Equal(t, wantPairs, gotPairs)
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
