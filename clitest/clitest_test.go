package clitest_test

import (
	"context"
	"testing"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/clitest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestWithStorage_RoundTrips(t *testing.T) {
	t.Parallel()

	// Arrange
	_, app := clitest.WithStorage(context.Background())

	// Act
	writeErr := app.Data.WriteFile("state.bin", []byte("value"))
	data, readErr := app.Data.ReadFile("state.bin")

	// Assert
	if got, want := writeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("WriteFile(...) = %v, want %v", got, want)
	}
	if got, want := readErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadFile(...) = %v, want %v", got, want)
	}
	if got, want := string(data), "value"; got != want {
		t.Errorf("ReadFile(...) = %q, want %q", got, want)
	}
}

func TestWithStorage_PopulatesContext(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, _ := clitest.WithStorage(context.Background())

	// Act
	app := cli.StorageFrom(ctx)

	// Assert
	if got, want := app != nil, true; got != want {
		t.Errorf("StorageFrom(ctx) present = %t, want %t", got, want)
	}
}
