package k6

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestLabelsFromMultipleBaggageHeaders(t *testing.T) {
	countMembers := []string{"k6.test_run_id=123"}
	for i := range 64 {
		countMembers = append(countMembers, "other"+strconv.Itoa(i)+"=value")
	}

	tests := []struct {
		name        string
		values      []string
		members     map[string]string
		properties  bool
		validFields bool
	}{
		{
			name:   "k6_members_in_later_fields",
			values: []string{"other=value", "k6.test_run_id=123", "k6.scenario=checkout"},
			members: map[string]string{
				"other": "value", "k6.test_run_id": "123", "k6.scenario": "checkout",
			},
		},
		{
			name:    "duplicate_key_last_field_wins",
			values:  []string{"k6.test_run_id=old", "k6.test_run_id=new"},
			members: map[string]string{"k6.test_run_id": "new"},
		},
		{
			name:   "encoded_value_and_member_properties",
			values: []string{"other=value", "k6.test_run_id=run%2Cone;source=load"},
			members: map[string]string{
				"other": "value", "k6.test_run_id": "run,one",
			},
			properties: true,
		},
		{
			name:    "empty_fields_around_members",
			values:  []string{"", " \t", "k6.test_run_id=123", "", "k6.scenario=checkout", ""},
			members: map[string]string{"k6.test_run_id": "123", "k6.scenario": "checkout"},
		},
		{
			name:   "all_empty_fields_preserve_context",
			values: []string{"", " \t", ""},
		},
		{
			name:   "malformed_later_field_preserves_context",
			values: []string{"k6.test_run_id=123", "invalid"},
		},
		{
			name: "combined_byte_limit_preserves_context",
			values: []string{
				"k6.test_run_id=" + strings.Repeat("a", 2800),
				"other=" + strings.Repeat("b", 2800),
				"third=" + strings.Repeat("c", 2800),
			},
			validFields: true,
		},
		{
			name: "combined_member_limit_preserves_context",
			values: []string{
				strings.Join(countMembers[:32], ","),
				strings.Join(countMembers[32:], ","),
			},
			validFields: true,
		},
		{
			name:   "non_ows_field_preserves_context",
			values: []string{"k6.test_run_id=123", "\u00a0"},
		},
	}

	transports := []struct {
		name string
		run  func(*testing.T, context.Context, []string, func(context.Context))
	}{
		{
			name: "http",
			run: func(t *testing.T, ctx context.Context, values []string, check func(context.Context)) {
				t.Helper()
				req := httptest.NewRequest(http.MethodGet, "http://example.com", nil).WithContext(ctx)
				for _, value := range values {
					req.Header.Add("bAgGaGe", value)
				}
				LabelsFromBaggageHandler(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
					check(req.Context())
				})).ServeHTTP(httptest.NewRecorder(), req)
			},
		},
		{
			name: "grpc_unary",
			run: func(t *testing.T, ctx context.Context, values []string, check func(context.Context)) {
				t.Helper()
				ctx = testRepeatedBaggageMetadata(ctx, values)
				_, err := LabelsFromBaggageUnaryInterceptor(ctx, nil, nil, func(ctx context.Context, _ any) (any, error) {
					check(ctx)

					return "response", nil
				})
				require.NoError(t, err)
			},
		},
		{
			name: "grpc_stream",
			run: func(t *testing.T, ctx context.Context, values []string, check func(context.Context)) {
				t.Helper()
				stream := &grpcmiddleware.WrappedServerStream{WrappedContext: testRepeatedBaggageMetadata(ctx, values)}
				//nolint:contextcheck // gRPC passes the derived context through ServerStream.
				err := LabelsFromBaggageStreamInterceptor(nil, stream, nil, func(_ any, stream grpc.ServerStream) error {
					check(stream.Context())

					return nil
				})
				require.NoError(t, err)
			},
		},
	}

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					if tt.validFields {
						for _, value := range tt.values {
							_, err := baggage.Parse(value)
							require.NoError(t, err, "each field must be valid in isolation")
						}
					}
					parent := baggage.ContextWithBaggage(context.Background(), testMustNewBaggage(t, "parent", "keep"))
					parent = pprof.WithLabels(parent, pprof.Labels("existing", "keep", "k6_test_run_id", "parent"))
					called := false
					transport.run(t, parent, tt.values, func(ctx context.Context) {
						called = true
						testAssertBaggageHeaderContext(t, ctx, tt.members, tt.properties)
					})
					require.True(t, called)
					require.Equal(t, "keep", baggage.FromContext(parent).Member("parent").Value())
					value, ok := pprof.Label(parent, "k6_test_run_id")
					require.True(t, ok)
					require.Equal(t, "parent", value)
				})
			}
		})
	}
}

func testRepeatedBaggageMetadata(ctx context.Context, values []string) context.Context {
	md := metadata.MD{}
	for _, value := range values {
		md.Append("bAgGaGe", value)
	}

	return metadata.NewIncomingContext(ctx, md)
}

func testAssertBaggageHeaderContext(t *testing.T, ctx context.Context, members map[string]string, withProperties bool) {
	t.Helper()

	wantMembers := members
	if wantMembers == nil {
		wantMembers = map[string]string{"parent": "keep"}
	}
	gotMembers := map[string]string{}
	for _, member := range baggage.FromContext(ctx).Members() {
		gotMembers[member.Key()] = member.Value()
	}
	require.Equal(t, wantMembers, gotMembers)
	wantLabels := map[string]string{"existing": "keep", "k6_test_run_id": "parent"}
	for key, value := range members {
		if strings.HasPrefix(key, "k6.") {
			wantLabels[strings.ReplaceAll(key, ".", "_")] = value
		}
	}
	gotLabels := map[string]string{}
	pprof.ForLabels(ctx, func(key, value string) bool {
		gotLabels[key] = value

		return true
	})
	require.Equal(t, wantLabels, gotLabels)
	if withProperties {
		properties := baggage.FromContext(ctx).Member("k6.test_run_id").Properties()
		require.Len(t, properties, 1)
		require.Equal(t, "source", properties[0].Key())
		value, ok := properties[0].Value()
		require.True(t, ok)
		require.Equal(t, "load", value)
	}
}
