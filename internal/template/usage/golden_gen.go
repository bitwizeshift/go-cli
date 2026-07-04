//go:build ignore

// Command golden_gen regenerates the checked-in golden usage output under
// testdata. It renders the shared fixture commands and writes the result to
// disk. Run it with "go generate".
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitwizeshift/go-cli/internal/template/usage"
	"github.com/bitwizeshift/go-cli/internal/template/usage/usagetest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "golden_gen:", err)
		os.Exit(1)
	}
}

func run() error {
	for _, c := range usagetest.Cases() {
		var buf bytes.Buffer
		if err := (usage.Renderer{}).Render(&buf, c.Command); err != nil {
			return err
		}
		path := filepath.Join("testdata", c.Name)
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}
