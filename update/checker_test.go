package update_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bitwizeshift/go-cli/update"
	"github.com/bitwizeshift/go-cli/update/updatetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestChecker_Check(t *testing.T) {
	t.Parallel()

	errLookup := errors.New("release lookup")

	testCases := []struct {
		name     string
		build    update.BuildInfo
		registry update.ProviderRegistry
		want     update.Result
		wantErr  error
	}{
		{
			name:     "UpdateAvailable",
			build:    update.BuildInfo{Version: "v1.0.0", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: true,
				Current:   "v1.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "NonCanonicalCurrent",
			build:    update.BuildInfo{Version: "1.0", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: true,
				Current:   "v1.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "UpToDate",
			build:    update.BuildInfo{Version: "v2.0.0", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: false,
				Current:   "v2.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "CurrentNewer",
			build:    update.BuildInfo{Version: "v3.0.0", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: false,
				Current:   "v3.0.0",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "SnapshotCurrent",
			build:    update.BuildInfo{Version: "snapshot", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: false,
				Current:   "snapshot",
				Latest:    "v2.0.0",
			},
		}, {
			name:     "NoProviderForSource",
			build:    update.BuildInfo{Version: "v1.0.0", Source: "brew"},
			registry: update.ProviderRegistry{"github": updatetest.Provider("v2.0.0")},
			want: update.Result{
				Available: false,
				Current:   "v1.0.0",
				Latest:    "",
			},
		}, {
			name:     "ProviderError",
			build:    update.BuildInfo{Version: "v1.0.0", Source: "github"},
			registry: update.ProviderRegistry{"github": updatetest.ErrProvider(errLookup)},
			want:     update.Result{},
			wantErr:  errLookup,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := update.NewChecker(tc.build, &tc.registry)
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
