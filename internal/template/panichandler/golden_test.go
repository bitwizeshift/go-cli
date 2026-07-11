package panichandler_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
)

// Sample panic inputs. These must match the constants in golden_gen.go.
const (
	sampleMessage = "runtime error: index out of range [3] with length 3"
	sampleURL     = "https://github.com/bitwizeshift/go-cli/issues/new"
)

func TestRenderer_Render_Golden(t *testing.T) {
	t.Parallel()

	// Arrange
	stack := readTestdata(t, "stack.txt")
	want := string(readTestdata(t, "panic.golden.txt"))
	ctx := panichandler.PanicContext{
		Err:      errors.New(sampleMessage),
		Stack:    stack,
		IssueURL: sampleURL,
	}
	sut := panichandler.Renderer{}
	var buf bytes.Buffer

	// Act
	renderErr := sut.Render(&buf, ctx)
	rendered, stripErr := plain.Render(buf.String())

	// Assert
	if renderErr != nil {
		t.Fatalf("Render(...) = %v, want nil", renderErr)
	}
	if stripErr != nil {
		t.Fatalf("plain.Render(...) = %v, want nil", stripErr)
	}
	if got, want := rendered, want; !cmp.Equal(got, want) {
		t.Errorf("Render(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

// readTestdata reads the file named name from the testdata directory.
func readTestdata(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("readTestdata(%q) = %v, want nil", name, err)
	}
	return data
}
