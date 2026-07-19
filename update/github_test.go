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

func TestGitHubProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr error
	}{
		{
			name:   "LatestRelease",
			status: http.StatusOK,
			body:   `{"tag_name":"v1.2.3"}`,
			want:   "v1.2.3",
		}, {
			name:   "TagWithoutPrefix",
			status: http.StatusOK,
			body:   `{"tag_name":"1.2.3"}`,
			want:   "v1.2.3",
		}, {
			name:   "NonCanonicalTag",
			status: http.StatusOK,
			body:   `{"tag_name":"v1.2"}`,
			want:   "v1.2.0",
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
			name:    "InvalidTag",
			status:  http.StatusOK,
			body:    `{"tag_name":"nightly"}`,
			wantErr: updatecheck.ErrInvalidVersion,
		}, {
			name:    "EmptyTag",
			status:  http.StatusOK,
			body:    `{"tag_name":""}`,
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
			sut := &update.GitHubProvider{
				Owner:   "bitwizeshift",
				Repo:    "go-cli",
				BaseURL: server.URL,
			}
			ctx := context.Background()

			// Act
			version, err := sut.LatestVersion(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("GitHubProvider.LatestVersion() error = %v, want %v", got, want)
			}
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GitHubProvider.LatestVersion() = %q, want %q", got, want)
			}
		})
	}
}

func TestGitHubProvider_LatestVersion_WithInvalidBaseURL_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &update.GitHubProvider{
		Owner:   "bitwizeshift",
		Repo:    "go-cli",
		BaseURL: "://bad",
	}
	ctx := context.Background()

	// Act
	version, err := sut.LatestVersion(ctx)

	// Assert
	if got, want := err, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("GitHubProvider.LatestVersion() error = %v, want %v", got, want)
	}
	if got, want := version, ""; !cmp.Equal(got, want) {
		t.Errorf("GitHubProvider.LatestVersion() = %q, want %q", got, want)
	}
}

func TestGitHubProvider_LatestVersion_WithCanceledContext_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &update.GitHubProvider{
		Owner: "bitwizeshift",
		Repo:  "go-cli",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	version, err := sut.LatestVersion(ctx)

	// Assert
	if got, want := err, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("GitHubProvider.LatestVersion() error = %v, want %v", got, want)
	}
	if got, want := version, ""; !cmp.Equal(got, want) {
		t.Errorf("GitHubProvider.LatestVersion() = %q, want %q", got, want)
	}
}
