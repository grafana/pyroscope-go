//go:build go1.18
// +build go1.18

package svcinfo

import "runtime/debug"

func gitRef(it *debug.BuildInfo) string {
	for _, setting := range it.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
}
