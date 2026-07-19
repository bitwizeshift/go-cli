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

func TestGitLabProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr error
	}{
		{
			name:   "MostRecentRelease",
			status: http.StatusOK,
			body:   `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`,
			want:   "v2.0.0",
		}, {
			name:   "NoReleases",
			status: http.StatusOK,
			body:   `[]`,
			want:   "",
		}, {
			name:    "NotFound",
			status:  http.StatusNotFound,
			body:    "",
			wantErr: update.ErrUnexpectedStatus,
		}, {
			name:    "MalformedBody",
			status:  http.StatusOK,
			body:    "[",
			wantErr: update.ErrDecodeResponse,
		}, {
			name:    "InvalidTag",
			status:  http.StatusOK,
			body:    `[{"tag_name":"latest"}]`,
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
			sut := &update.GitLabProvider{
				Project: "42",
				BaseURL: server.URL,
			}
			ctx := context.Background()

			// Act
			version, err := sut.LatestVersion(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("GitLabProvider.LatestVersion() error = %v, want %v", got, want)
			}
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GitLabProvider.LatestVersion() = %q, want %q", got, want)
			}
		})
	}
}
