package spec_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestPanicError_Error_ReportsRecoveredValue(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := spec.PanicError{Err: "kaboom"}

	// Act
	message := sut.Error()

	// Assert
	if got, want := message, "kaboom"; got != want {
		t.Errorf("PanicError.Error() = %q, want %q", got, want)
	}
}

func TestPanicError_Unwrap_YieldsErrPanic(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := spec.PanicError{Err: "kaboom"}

	// Act
	err := sut.Unwrap()

	// Assert
	if got, want := err, spec.ErrPanic; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("PanicError.Unwrap() = %v, want %v", got, want)
	}
}
