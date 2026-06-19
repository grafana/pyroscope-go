package semconv

import (
	"strings"
	"testing"
)

func TestMergeTagsWithAppName(t *testing.T) {
	tags := map[string]string{"foo": "bar"}
	names, err := MergeTagsWithAppName("testAppapp", "239", tags)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(names.SDK, "testAppapp{__session_id__=239,foo=bar,otel.scope.name=com.grafana.pyroscope/go,process.runtime.name=go,process.runtime.version=") {
		t.Fatalf("unexpected sdk name %s", names.SDK)
	}
	if !strings.HasPrefix(names.Godeltaprof, "testAppapp{__session_id__=239,foo=bar,otel.scope.name=com.grafana.pyroscope/godeltaprof,process.runtime.name=go,process.runtime.version=") {
		t.Fatalf("unexpected godeltaprof name %s", names.Godeltaprof)
	}
}
