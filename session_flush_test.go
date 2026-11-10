package pyroscope

import (
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	internal "github.com/grafana/pyroscope-go/internal/pprof"
	"github.com/grafana/pyroscope-go/upstream"
)

func TestSessionFlushWithoutCPU(t *testing.T) {
	for _, profileType := range []ProfileType{ProfileAllocSpace, ProfileGoroutines} {
		for _, wait := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/wait=%t", profileType, wait), func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					t.Helper()
					u := &sessionFlushUpstream{}
					s, err := NewSession(SessionConfig{
						AppName:        "test.session.flush",
						ProfilingTypes: []ProfileType{profileType},
						UploadRate:     time.Hour,
						Upstream:       u,
						Logger:         noopLogger,
					})
					require.NoError(t, err)
					require.NoError(t, s.Start())

					// Exercise the real session loop and profiles, without starting
					// CPU profiling or uploading over the network. Repeated flushes
					// and Stop must also return after an asynchronous flush.
					const flushes = 3
					for range flushes {
						s.flush(wait)
					}
					s.Stop()

					require.Equal(t, flushes, u.flushes)
					require.Len(t, u.uploaded, flushes)
					for _, job := range u.uploaded {
						require.NotEmpty(t, job.Profile)
					}
				})
			})
		}
	}
}

func TestSessionFlushWithCPU(t *testing.T) {
	for _, wait := range []bool{true, false} {
		t.Run(fmt.Sprintf("wait=%t", wait), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				t.Helper()
				defer internal.ResetCollector()
				u := &sessionFlushUpstream{}
				s, err := NewSession(SessionConfig{
					AppName:        "test.session.flush",
					ProfilingTypes: DefaultProfileTypes,
					UploadRate:     time.Hour,
					Upstream:       u,
					Logger:         noopLogger,
				})
				require.NoError(t, err)
				collector := new(mockCollector)
				s.cpu.collector = collector
				started := collector.waitStartCPUProfile()
				require.NoError(t, s.Start())
				defer s.Stop()
				<-started

				const flushes = 3
				for range flushes {
					restarted := collector.waitStartCPUProfile()
					s.flush(wait)
					synctest.Wait()
					select {
					case <-restarted:
					default:
						t.Fatal("flush did not restart CPU profiling")
					}
				}
				s.Stop()

				require.Equal(t, flushes, u.flushes)
			})
		})
	}
}

type sessionFlushUpstream struct {
	uploaded []*upstream.UploadJob
	flushes  int
}

func (u *sessionFlushUpstream) Upload(job *upstream.UploadJob) {
	u.uploaded = append(u.uploaded, job)
}

func (u *sessionFlushUpstream) Flush() {
	u.flushes++
}
