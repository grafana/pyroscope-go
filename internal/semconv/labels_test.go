package semconv

import (
	"testing"

	"github.com/grafana/pyroscope-go/internal/labelset"
	"github.com/stretchr/testify/require"
)

func TestMergeTagsWithAppName(t *testing.T) {
	tags := map[string]string{"foo": "bar"}
	names, err := MergeTagsWithAppName("testAppapp", "239", tags)
	require.NoError(t, err)

	sdkLabels := parseAppName(t, names.SDK)
	require.Equal(t, "testAppapp", sdkLabels[labelset.ReservedLabelNameName])
	require.Equal(t, "239", sdkLabels[labelPyroscopeSessionID])
	require.Equal(t, "bar", sdkLabels["foo"])
	require.Equal(t, scopeSDK, sdkLabels[labelScopeName])
	require.Equal(t, getRuntimeName(), sdkLabels[labelProcessRuntimeName])
	require.Equal(t, getRuntimeVersion(), sdkLabels[labelProcessRuntimeVersion])

	godeltaprofLabels := parseAppName(t, names.Godeltaprof)
	require.Equal(t, "testAppapp", godeltaprofLabels[labelset.ReservedLabelNameName])
	require.Equal(t, "239", godeltaprofLabels[labelPyroscopeSessionID])
	require.Equal(t, "bar", godeltaprofLabels["foo"])
	require.Equal(t, scopeGodeltaprof, godeltaprofLabels[labelScopeName])
	require.Equal(t, getRuntimeName(), godeltaprofLabels[labelProcessRuntimeName])
	require.Equal(t, getRuntimeVersion(), godeltaprofLabels[labelProcessRuntimeVersion])
}

func TestMergeTagsWithAppNamePreservesUserProvidedSemconvTags(t *testing.T) {
	tags := map[string]string{
		labelScopeName:             "custom-scope",
		labelScopeVersion:          "custom-scope-version",
		labelProcessRuntimeName:    "custom-runtime",
		labelProcessRuntimeVersion: "custom-runtime-version",
	}

	names, err := MergeTagsWithAppName("testApp", "239", tags)
	require.NoError(t, err)

	for _, name := range []string{names.SDK, names.Godeltaprof} {
		labels := parseAppName(t, name)
		require.Equal(t, "custom-scope", labels[labelScopeName])
		require.Equal(t, "custom-scope-version", labels[labelScopeVersion])
		require.Equal(t, "custom-runtime", labels[labelProcessRuntimeName])
		require.Equal(t, "custom-runtime-version", labels[labelProcessRuntimeVersion])
	}
}

func TestMergeTagsWithAppNamePreservesEmbeddedSemconvTags(t *testing.T) {
	names, err := MergeTagsWithAppName(
		"testApp{otel.scope.name=embedded-scope,otel.scope.version=embedded-scope-version,process.runtime.name=embedded-runtime,process.runtime.version=embedded-runtime-version}", //nolint:lll
		"239",
		nil,
	)
	require.NoError(t, err)

	for _, name := range []string{names.SDK, names.Godeltaprof} {
		labels := parseAppName(t, name)
		require.Equal(t, "embedded-scope", labels[labelScopeName])
		require.Equal(t, "embedded-scope-version", labels[labelScopeVersion])
		require.Equal(t, "embedded-runtime", labels[labelProcessRuntimeName])
		require.Equal(t, "embedded-runtime-version", labels[labelProcessRuntimeVersion])
	}
}

func TestMergeTagsWithAppNameTagsOverrideEmbeddedTags(t *testing.T) {
	tags := map[string]string{
		"foo":          "from-tags",
		labelScopeName: "tag-scope",
	}

	names, err := MergeTagsWithAppName("testApp{foo=from-app,otel.scope.name=app-scope}", "239", tags)
	require.NoError(t, err)

	for _, name := range []string{names.SDK, names.Godeltaprof} {
		labels := parseAppName(t, name)
		require.Equal(t, "from-tags", labels["foo"])
		require.Equal(t, "tag-scope", labels[labelScopeName])
	}
}

func parseAppName(t *testing.T, name string) map[string]string {
	t.Helper()

	labels, err := labelset.Parse(name)
	require.NoError(t, err)

	return labels.Labels()
}
