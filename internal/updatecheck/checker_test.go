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

func TestChecker_Check(t *testing.T) {
	t.Parallel()

	errLookup := errors.New("release lookup")

	testCases := []struct {
		name     string
		build    updatecheck.BuildInfo
		registry updatecheck.ProviderRegistry
		want     updatecheck.Result
		wantErr  error
	}{
		{
			name:     "UpdateAvailable",
			build:    updatecheck.BuildInfo{Version: "v1.0.0", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: true,
				Current:   "v1.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "NonCanonicalCurrent",
			build:    updatecheck.BuildInfo{Version: "1.0", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: true,
				Current:   "v1.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "UpToDate",
			build:    updatecheck.BuildInfo{Version: "v2.0.0", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: false,
				Current:   "v2.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "CurrentNewer",
			build:    updatecheck.BuildInfo{Version: "v3.0.0", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: false,
				Current:   "v3.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "SnapshotCurrent",
			build:    updatecheck.BuildInfo{Version: "snapshot", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: false,
				Current:   "snapshot",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "NoProviderForSource",
			build:    updatecheck.BuildInfo{Version: "v1.0.0", Source: "brew"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: updatecheck.Result{
				Available: false,
				Current:   "v1.0.0",
				Latest:    "",
			},
		}, {
			name:     "ProviderError",
			build:    updatecheck.BuildInfo{Version: "v1.0.0", Source: "github"},
			registry: updatecheck.ProviderRegistry{"github": updatetest.ErrProvider(errLookup)},
			want:     updatecheck.Result{},
			wantErr:  errLookup,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := updatecheck.NewChecker(tc.build, &tc.registry)
			ctx := context.Background()

			// Act
			result, err := sut.Check(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Checker.Check() error = %v, want %v", got, want)
			}
			if got, want := result, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Checker.Check() = %+v, want %+v", got, want)
			}
		})
	}
}
