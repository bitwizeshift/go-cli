package tag

// Themed wraps s in a theme tag for role, which a richtext writer resolves to
// the role's style. An empty role produces no styling.
func Themed(role, s string) string {
	return "[theme:" + role + "]" + s + "[/theme]"
}

// Raw wraps s in a passthrough region so its contents are written verbatim
// instead of being parsed as tags. Use it for arbitrary text that may itself
// contain bracketed sequences, such as error messages or stack traces.
func Raw(s string) string {
	return "[richtext:off]" + s + "[/richtext]"
}
