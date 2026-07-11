//go:build ignore

// Command golden_gen regenerates the checked-in golden version output under
// testdata. It renders the fixture command against fixed build information and
// writes the result to disk. Run it with "go generate".
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/bitwizeshift/go-cli/internal/template/plain"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/internal/template/version/versiontest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "golden_gen:", err)
		os.Exit(1)
	}
}

func run() error {
	tmplfuncs.DefaultBuild.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
		return versiontest.BuildInfo(), true
	}
	markup, err := versiontest.Render(versiontest.Command())
	if err != nil {
		return err
	}
	rendered, err := plain.Render(markup)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("testdata", "version.golden.txt"), []byte(rendered), 0o644)
}
