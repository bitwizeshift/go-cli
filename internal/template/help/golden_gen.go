//go:build ignore

// Command golden_gen regenerates the checked-in golden help output under
// testdata. It renders the shared fixture command hierarchy for each golden
// case and writes the result to disk. Run it with "go generate".
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/internal/template/help/helptest"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "golden_gen:", err)
		os.Exit(1)
	}
}

func run() error {
	for _, c := range helptest.Cases() {
		var buf bytes.Buffer
		renderer := help.Renderer{Columns: c.Columns}
		if err := renderer.Render(&buf, c.Command, c.CL); err != nil {
			return err
		}
		rendered, err := plain.Render(buf.String())
		if err != nil {
			return err
		}
		path := filepath.Join("testdata", c.Name)
		if err := os.WriteFile(path, []byte(rendered), 0o644); err != nil {
			return err
		}
	}
	return nil
}
