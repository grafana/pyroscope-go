//go:build go1.16 && !go1.18
// +build go1.16,!go1.18

package svcinfo

func gitRef(_ *debug.BuildInfo) string {
	return ""
}
