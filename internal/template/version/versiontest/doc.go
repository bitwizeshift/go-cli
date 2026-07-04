// Package versiontest provides the shared fixtures used by the version
// renderer's golden test and its generator.
//
// It exists so that the checked-in golden output and the code that regenerates
// it use the same fixed build information and render it identically. Use
// [BuildInfo] for the deterministic build data, [Command] for the fixture
// command, and [Render] to render the version template.
package versiontest
