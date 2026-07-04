package version

import _ "embed"

//go:embed templates/version.tmpl
var templateText string

// Template returns the text of the --version template. Install it with
// cobra.Command.SetVersionTemplate; the template calls the shared functions from
// tmplfuncs, which must be registered with cobra.AddTemplateFuncs.
func Template() string {
	return templateText
}
