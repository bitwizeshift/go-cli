package update_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
	"github.com/bitwizeshift/go-cli/update"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBrewProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr error
	}{
		{
			name:   "StableVersion",
			status: http.StatusOK,
			body:   `{"versions":{"stable":"3.1.4"}}`,
			want:   "v3.1.4",
		}, {
			name:    "NotFound",
			status:  http.StatusNotFound,
			body:    "",
			wantErr: update.ErrUnexpectedStatus,
		}, {
			name:    "MalformedBody",
			status:  http.StatusOK,
			body:    "{",
			wantErr: update.ErrDecodeResponse,
		}, {
			name:    "InvalidStable",
			status:  http.StatusOK,
			body:    `{"versions":{"stable":"HEAD"}}`,
			wantErr: updatecheck.ErrInvalidVersion,
		}, {
			name:    "EmptyStable",
			status:  http.StatusOK,
			body:    `{"versions":{"stable":""}}`,
			wantErr: updatecheck.ErrInvalidVersion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			t.Cleanup(server.Close)
			sut := &update.BrewProvider{
				Name:    "go-cli",
				BaseURL: server.URL,
			}
			ctx := context.Background()

			// Act
			version, err := sut.LatestVersion(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("BrewProvider.LatestVersion() error = %v, want %v", got, want)
			}
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("BrewProvider.LatestVersion() = %q, want %q", got, want)
			}
		})
	}
}
