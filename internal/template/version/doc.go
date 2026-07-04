// Package version renders the CLI's --version output.
//
// It exists because cobra offers only a version template string and no hook for
// a render function. This package supplies that template via [Template]; callers
// install it with cobra.Command.SetVersionTemplate and register the shared
// template functions with cobra.AddTemplateFuncs so the template can style and
// align the build details.
package version

//go:generate go run golden_gen.go
