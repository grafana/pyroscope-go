//go:build go1.20
// +build go1.20

package pprof

type MutexProfileScaler struct {
}

func scaleMutexProfile(_ MutexProfileScaler, cnt int64, ns float64) (int64, float64) {
	return cnt, ns
}

var ScalerMutexProfile = MutexProfileScaler{}

var ScalerBlockProfile = MutexProfileScaler{}
