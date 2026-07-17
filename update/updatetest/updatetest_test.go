package updatetest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bitwizeshift/go-cli/update/updatetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := updatetest.Provider("v1.2.3")
	ctx := context.Background()

	// Act
	version, err := sut.LatestVersion(ctx)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Provider.LatestVersion() error = %v, want %v", got, want)
	}
	if got, want := version, "v1.2.3"; !cmp.Equal(got, want) {
		t.Errorf("Provider.LatestVersion() = %q, want %q", got, want)
	}
}

func TestErrProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	// Arrange
	errLookup := errors.New("release lookup")
	sut := updatetest.ErrProvider(errLookup)
	ctx := context.Background()

	// Act
	version, err := sut.LatestVersion(ctx)

	// Assert
	if got, want := err, errLookup; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrProvider.LatestVersion() error = %v, want %v", got, want)
	}
	if got, want := version, ""; !cmp.Equal(got, want) {
		t.Errorf("ErrProvider.LatestVersion() = %q, want %q", got, want)
	}
}
