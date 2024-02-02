//go:build go1.16 && !go1.18
// +build go1.16,!go1.18

package svcinfo

import "runtime/debug"

func gitRef(_ *debug.BuildInfo) string {
	return ""
}
