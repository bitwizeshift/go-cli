//go:build ignore

// Command golden_gen regenerates the checked-in panic report under testdata. It
// renders the sample panic context (using the fake stack trace in
// testdata/stack.txt) and writes the result to disk. Run it with "go generate".
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
)

// Sample panic inputs. These must match the constants in golden_test.go.
const (
	sampleMessage = "runtime error: index out of range [3] with length 3"
	sampleURL     = "https://github.com/bitwizeshift/go-cli/issues/new"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "golden_gen:", err)
		os.Exit(1)
	}
}

func run() error {
	stack, err := os.ReadFile(filepath.Join("testdata", "stack.txt"))
	if err != nil {
		return err
	}
	ctx := panichandler.PanicContext{
		Err:      errors.New(sampleMessage),
		Stack:    stack,
		IssueURL: sampleURL,
	}
	var buf bytes.Buffer
	if err := (panichandler.Renderer{}).Render(&buf, ctx); err != nil {
		return err
	}
	rendered, err := plain.Render(buf.String())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("testdata", "panic.golden.txt"), []byte(rendered), 0o644)
}
