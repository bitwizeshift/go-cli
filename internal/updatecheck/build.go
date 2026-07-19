package updatecheck

// BuildInfo describes how the running binary was built and distributed.
type BuildInfo struct {
	// Version is the semantic version of the running binary, or a non-version
	// marker such as "snapshot" or "(devel)" when it was not built from a release.
	Version string

	// Source names the distribution channel the binary was installed from, such
	// as "github" or "brew". It selects which registered [Provider] a [Checker]
	// queries.
	Source string
}
