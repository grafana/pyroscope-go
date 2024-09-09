package remote

import (
	"bytes"
	"io"
	"net/http"
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
			expectedAuthHeader: "",
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
	return args.Get(0).(*http.Response), args.Error(1)
}
