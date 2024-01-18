package svcinfo

import (
	"os"
	"runtime"
	"runtime/debug"
	"sync"
)

const (
	LabelNameServiceRepository = "service_repository"
	LabelNameServiceGitRef     = "service_git_ref"

	EnvServiceInfo          = "PYROSCOPE_SERVICE_INFO"
	EnvServiceVCSRepository = "PYROSCOPE_SERVICE_VCS_REPOSITORY"
	EnvServiceVCSRevision   = "PYROSCOPE_SERVICE_VCS_REVISION"
)

// https://github.com/grafana/pyroscope/commit/fbd156187335b8500e7001227e32280c4fd4f8ba#diff-7bd23599b57d18abcd8339f39053d0bc233bd6ad26af63527d36b078fb413206R453
type ServiceVersion struct {
	Repository string `json:"repository,omitempty"`
	GitRef     string `json:"git_ref,omitempty"`
	// Note: this is the process ELF mapping buildID, not service build ID
	BuildID string `json:"build_id,omitempty"`

	GoVersion           string `json:"go_version,omitempty"`
	PyroscopeSDKVersion string `json:"pyroscope_sdk_version,omitempty"`
	GodeltaprofVersion  string `json:"godeltaprof_version,omitempty"`
}

var info ServiceVersion
var once sync.Once

func GetServiceVersion() ServiceVersion {
	once.Do(func() {
		if os.Getenv(EnvServiceInfo) == "false" {
			return
		}
		it, ok := debug.ReadBuildInfo()
		if it != nil && ok {
			info.Repository = it.Path
			info.GoVersion = it.GoVersion
			for _, setting := range it.Settings {
				if setting.Key == "vcs.revision" {
					info.GitRef = setting.Value
				}
			}
			for _, dep := range it.Deps {
				if dep.Path == "github.com/grafana/pyroscope-go" {
					info.PyroscopeSDKVersion = dep.Version
				}
				if dep.Path == "github.com/grafana/pyroscope-go/godeltaprof" {
					info.GodeltaprofVersion = dep.Version
				}
			}
			if it.Main.Path == "github.com/grafana/pyroscope-go" && info.PyroscopeSDKVersion == "" {
				info.PyroscopeSDKVersion = it.Main.Version
			}
			if it.Main.Path == "github.com/grafana/pyroscope-go/godeltaprof" && info.GodeltaprofVersion == "" {
				info.GodeltaprofVersion = it.Main.Version
			}
		}
		if repo := os.Getenv(EnvServiceVCSRepository); repo != "" {
			info.Repository = repo
		}
		if rev := os.Getenv(EnvServiceVCSRevision); rev != "" {
			info.GitRef = rev
		}
		if info.GoVersion == "" {
			info.GoVersion = runtime.Version()
		}
	})
	return info
}
