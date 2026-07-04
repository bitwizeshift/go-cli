package versiontest

import (
	"bytes"
	"runtime/debug"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/internal/template/version"
	"github.com/spf13/cobra"
)

// BuildInfo returns the fixed build information used to render deterministic
// version output.
func BuildInfo() *debug.BuildInfo {
	return &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Main:      debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "GOOS", Value: "darwin"},
			{Key: "GOARCH", Value: "arm64"},
			{Key: "-tags", Value: "netgo"},
			{Key: "vcs", Value: "git"},
			{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
			{Key: "vcs.time", Value: "2024-08-05T20:18:23Z"},
		},
	}
}

// Command returns the fixture command whose name appears in the version output.
func Command() *cobra.Command {
	return &cobra.Command{Use: "widget"}
}

// Render renders the version template for cmd with plain (uncoloured) styling.
// It reads build information from [tmplfuncs.DefaultBuild], which callers set to
// [BuildInfo] for deterministic output.
func Render(cmd *cobra.Command) (string, error) {
	tmpl, err := template.New("version").
		Funcs(tmplfuncs.NewFunc(palette.NoColour{})).
		Parse(version.Template())
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cmd); err != nil {
		return "", err
	}
	return buf.String(), nil
}
