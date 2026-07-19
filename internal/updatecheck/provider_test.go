package updatecheck_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
	"github.com/bitwizeshift/go-cli/update/updatetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestProviderRegistry_LatestVersion(t *testing.T) {
	t.Parallel()

	errLookup := errors.New("release lookup")

	testCases := []struct {
		name     string
		registry updatecheck.ProviderRegistry
		source   string
		want     string
		wantErr  error
	}{
		{
			name:     "RegisteredSource",
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v1.0.0")},
			source:   "github",
			want:     "v1.0.0",
		}, {
			name:     "UnregisteredSource",
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v1.0.0")},
			source:   "brew",
			want:     "",
		}, {
			name:     "ProviderError",
			registry: updatecheck.ProviderRegistry{"github": updatetest.ErrProvider(errLookup)},
			source:   "github",
			want:     "",
			wantErr:  errLookup,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.registry
			ctx := context.Background()

			// Act
			version, err := sut.LatestVersion(ctx, tc.source)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ProviderRegistry.LatestVersion() error = %v, want %v", got, want)
			}
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("ProviderRegistry.LatestVersion() = %q, want %q", got, want)
			}
		})
	}
}
