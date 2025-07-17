package remote

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/grafana/pyroscope-go/internal/testutil"
	"github.com/grafana/pyroscope-go/upstream"
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
				StatusCode: http.StatusOK,
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
			require.NoError(t, err)

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
	switch typed := a0.(type) {
	case *http.Response:
		return typed, err
	case func() *http.Response:
		return typed(), err
	default:
		return nil, fmt.Errorf("unknown mock arg type arg %+v %w", a0, err)
	}
}

func TestConcurrentUploadFlushRace(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockClient.On("Do", mock.Anything).Return(func() *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
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

func newJob(name string) *upstream.UploadJob {
	return &upstream.UploadJob{
		Name: name,
	}
}
