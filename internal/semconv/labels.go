package semconv

import (
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/grafana/pyroscope-go/internal/labelset"
)

const (
	labelScopeName             = "otel.scope.name"
	labelScopeVersion          = "otel.scope.version"
	labelProcessRuntimeName    = "process.runtime.name"
	labelProcessRuntimeVersion = "process.runtime.version"
	labelPyroscopeSessionID    = "__session_id__"
	scopeSDK                   = "com.grafana.pyroscope/go"
	scopeGodeltaprof           = "com.grafana.pyroscope/godeltaprof"
)

var (
	scopeVersions = versions{}
	versionOnce   = sync.Once{}
)

type versions struct {
	sdk         string
	godeltaprof string
}

type AppNames struct {
	SDK         string
	Godeltaprof string
}

func getScopeVersions() versions {
	versionOnce.Do(func() {
		var sdk = ""
		var godeltaprof = ""
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}
		for _, dep := range info.Deps {
			switch dep.Path {
			case "github.com/grafana/pyroscope-go/godeltaprof":
				godeltaprof = dep.Version
			case "github.com/grafana/pyroscope-go":
				sdk = dep.Version
			}
		}
		scopeVersions = versions{
			sdk:         sdk,
			godeltaprof: godeltaprof,
		}
	})
	return scopeVersions
}

func getRuntimeName() string {
	if runtime.Compiler == "gc" {
		return "go"
	}
	return runtime.Compiler
}

func getRuntimeVersion() string {
	return runtime.Version()
}

// MergeTagsWithAppName validates user input and merges explicitly specified
// tags with tags from app name.
//
// App name may be in the full form including tags (app.name{foo=bar,baz=qux}).
// Returned application name is always short, any tags that were included are
// moved to tags map. When merged with explicitly provided tags (config/CLI),
// last take precedence.
//
// App name may be an empty string. Tags must not contain reserved keys,
// the map is modified in place.
func MergeTagsWithAppName(appName string, sid string, tags map[string]string) (AppNames, error) {
	k, err := labelset.Parse(appName)
	if err != nil {
		return AppNames{}, err
	}
	for tagKey, tagValue := range tags {
		if labelset.IsLabelNameReserved(tagKey) {
			continue
		}
		err = labelset.ValidateLabelName(tagKey)
		if err != nil {
			return AppNames{}, err
		}
		k.Add(tagKey, tagValue)
	}

	k.Add(labelPyroscopeSessionID, sid)
	k.Add(labelProcessRuntimeName, getRuntimeName())
	k.Add(labelProcessRuntimeVersion, getRuntimeVersion())
	vs := getScopeVersions()
	return AppNames{
		SDK:         buildAppName(k, scopeSDK, vs.sdk),
		Godeltaprof: buildAppName(k, scopeGodeltaprof, vs.godeltaprof),
	}, nil
}

func buildAppName(builder *labelset.LabelSet, scope, version string) string {
	builder.Add(labelScopeName, scope)
	if version != "" {
		builder.Add(labelScopeVersion, version)
	} else {
		builder.Add(labelScopeVersion, "") // delete
	}
	return builder.Normalized()
}
