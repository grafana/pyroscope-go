package pyroscope

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfilerStartStop(t *testing.T) {
	profiler, err := Start(Config{
		ApplicationName: "test",
	})
	require.NoError(t, err)
	err = profiler.Stop()
	require.NoError(t, err)
}
