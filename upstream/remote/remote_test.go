package remote

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/grafana/pyroscope-go/internal/testutil"
	"github.com/grafana/pyroscope-go/upstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadProfile(t *testing.T) {
	tests := []struct {
		name               string
		cfg                Config
		serverAddress      string
		expectedAuthHeader string
		expectWarning      bool
	}{
		{
			name: "OG Pyroscope Cloud with AuthToken",
			cfg: Config{
				AuthToken: "test-token",
				Address:   "https://example.pyroscope.cloud",
			},
			expectedAuthHeader: "Bearer test-token",
			expectWarning:      false,
		},
		{
			name: "Non-OG Server with BasicAuth",
			cfg: Config{
				BasicAuthUser:     "user",
				BasicAuthPassword: "pass",
				Address:           "https://example.com",
			},
			expectedAuthHeader: "Basic dXNlcjpwYXNz", // Base64 encoded "user:pass"
			expectWarning:      false,
		},
		{
			name: "Non-OG Server with AuthToken (Deprecated)",
			cfg: Config{
				AuthToken: "deprecated-token",
				Address:   "https://example.com",
			},
			expectedAuthHeader: "Bearer deprecated-token",
			expectWarning:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := testutil.NewTestLogger()
			mockClient := new(MockHTTPClient)

			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}, nil)

			r := &Remote{
				cfg:    tt.cfg,
				client: mockClient,
				logger: logger,
			}

			err := r.uploadProfile(&upstream.UploadJob{
				Name:       "test-profile",
				StartTime:  time.Now(),
				EndTime:    time.Now().Add(time.Minute),
				SpyName:    "test-spy",
				SampleRate: 100,
				Units:      "samples",
			})
			assert.NoError(t, err)

			if tt.expectWarning {
				assert.Contains(t, logger.Lines(), authTokenDeprecationWarning)
			} else {
				assert.NotContains(t, logger.Lines(), authTokenDeprecationWarning)
			}

			mockClient.AssertCalled(t, "Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Header.Get("Authorization") == tt.expectedAuthHeader
			}))

			mockClient.AssertExpectations(t)
		})
	}
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	err := args.Error(1)
	a0 := args.Get(0)
	if resp, ok := a0.(*http.Response); ok {
		return resp, err
	}
	if resp, ok := a0.(func() *http.Response); ok {
		return resp(), err
	}
	return nil, fmt.Errorf("unknown arg %+v %w", a0, err)
}

func TestConcurrentUploadFlushRace(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockClient.On("Do", mock.Anything).Return(func() *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
		}
	}, nil)
	r, err := NewRemote(Config{
		Threads:    2,
		Logger:     testutil.NewTestLogger(),
		HTTPClient: mockClient,
	})
	require.NoError(t, err)
	r.Start()
	defer r.Stop()

	var wg sync.WaitGroup
	wg.Add(2)
	loop := func(f func()) {
		timeout := time.After(10 * time.Millisecond)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-timeout:
					return
				default:
					f()
				}
			}
		}()
	}
	loop(func() {
		r.Upload(newJob("job1"))
	})
	loop(func() {
		r.Flush()
	})
	wg.Wait()
}

func TestStartTwice(t *testing.T) {
	r, err := NewRemote(Config{
		Threads:    2,
		Logger:     testutil.NewTestLogger(),
		HTTPClient: new(MockHTTPClient),
	})
	require.NoError(t, err)
	r.Start()
	r.Start()
	r.Stop()
}

func TestUploadNotStarted(t *testing.T) {
	c := new(MockHTTPClient)
	r, err := NewRemote(Config{
		Threads:    2,
		Logger:     testutil.NewTestLogger(),
		HTTPClient: c,
	})
	require.NoError(t, err)
	r.Upload(newJob("j1"))
	require.Len(t, r.jobs, 0)
	c.AssertExpectations(t)
}

func TestDrainJobs(t *testing.T) {
	const requestDuration = 100 * time.Millisecond
	c := new(MockHTTPClient)
	c.On("Do", mock.Anything).Return(func() *http.Response {
		time.Sleep(requestDuration)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
		}
	}, nil).Once()
	logger := testutil.NewTestLogger()
	r, err := NewRemote(Config{
		Threads:    1,
		Logger:     logger,
		HTTPClient: c,
	})
	require.NoError(t, err)
	r.Start()
	r.Upload(newJob("job1"))
	r.Upload(newJob("job2"))
	r.Upload(newJob("job3"))
	time.Sleep(time.Millisecond)
	r.Stop()
	assert.Len(t, r.jobs, 0)
	r.Flush()
	require.EqualValues(t, 2, r.droppedJobs.Load())
}

func TestStartStopMultipleTimes(t *testing.T) {
	c := new(MockHTTPClient)

	c.On("Do", mock.Anything).Return(func() *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
		}
	}, nil).Once()
	logger := testutil.NewTestLogger()
	r, err := NewRemote(Config{
		Threads:    1,
		Logger:     logger,
		HTTPClient: c,
	})
	require.NoError(t, err)
	r.Start()
	r.Stop()
	r.Start()
	defer r.Stop()
	r.Upload(newJob("j1"))
	time.Sleep(time.Millisecond)
	c.AssertExpectations(t)
}

func newJob(name string) *upstream.UploadJob {
	return &upstream.UploadJob{
		Name: name,
	}
}
