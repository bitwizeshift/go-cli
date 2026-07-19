package arg

import (
	"strings"

	"github.com/bitwizeshift/go-cli/internal/completion"
)

// completerFunc computes shell completion candidates for the partial word
// toComplete, returning the candidates and a directive describing how the shell
// should treat them.
type completerFunc func(toComplete string) ([]string, completion.Directive)

// CompleterFunc completes the argument with the candidates returned by fn for the
// word being completed. File completion is suppressed so only fn's candidates
// are offered.
func CompleterFunc(fn func(toComplete string) []string) Option {
	return completionOption(func(toComplete string) ([]string, completion.Directive) {
		return fn(toComplete), completion.NoFileComp
	})
}

// CompleteFrom completes the argument with the members of options that are prefixed
// by the word being completed. File completion is suppressed so only the given
// options are offered.
func CompleteFrom(options ...string) Option {
	return completionOption(func(toComplete string) ([]string, completion.Directive) {
		var matches []string
		for _, option := range options {
			if strings.HasPrefix(option, toComplete) {
				matches = append(matches, option)
			}
		}
		return matches, completion.NoFileComp
	})
}

// CompleteFiles completes the argument with file names, deferring to the shell's
// default file completion.
func CompleteFiles() Option {
	return completionOption(func(string) ([]string, completion.Directive) {
		return nil, completion.Default
	})
}

// CompleteFilesMatching completes the argument with file names whose extension is
// one of exts. A leading dot on an extension is optional, so ".json" and "json"
// are equivalent.
func CompleteFilesMatching(exts ...string) Option {
	normalized := make([]string, len(exts))
	for i, ext := range exts {
		normalized[i] = strings.TrimPrefix(ext, ".")
	}
	return completionOption(func(string) ([]string, completion.Directive) {
		return normalized, completion.FilterFileExt
	})
}

// CompleteDirs completes the argument with directory names only.
func CompleteDirs() Option {
	return completionOption(func(string) ([]string, completion.Directive) {
		return nil, completion.FilterDirs
	})
}

// completionOption builds an [Option] that assigns complete as the argument's
// completer, panicking if a completer was already assigned.
func completionOption(complete completerFunc) Option {
	return option(func(c *config) {
		if c.completer != nil {
			panic("arg: multiple completion options specified for one argument")
		}
		c.completer = complete
	})
}
