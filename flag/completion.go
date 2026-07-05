package flag

import (
	"strings"

	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// completerFunc computes shell completion candidates for the partial word
// toComplete, returning the candidates and a directive describing how the shell
// should treat them.
type completerFunc func(toComplete string) ([]string, annotation.CompletionDirective)

// CompleterFunc completes the flag with the candidates returned by fn for the
// word being completed. File completion is suppressed so only fn's candidates
// are offered.
func CompleterFunc(fn func(toComplete string) []string) Option {
	return completionOption(func(toComplete string) ([]string, annotation.CompletionDirective) {
		return fn(toComplete), annotation.CompletionNoFileComp
	})
}

// CompleteFrom completes the flag with the members of options that are prefixed
// by the word being completed. File completion is suppressed so only the given
// options are offered.
func CompleteFrom(options ...string) Option {
	return completionOption(func(toComplete string) ([]string, annotation.CompletionDirective) {
		var matches []string
		for _, option := range options {
			if strings.HasPrefix(option, toComplete) {
				matches = append(matches, option)
			}
		}
		return matches, annotation.CompletionNoFileComp
	})
}

// CompleteFiles completes the flag with file names, deferring to the shell's
// default file completion.
func CompleteFiles() Option {
	return completionOption(func(string) ([]string, annotation.CompletionDirective) {
		return nil, annotation.CompletionDefault
	})
}

// CompleteFilesMatching completes the flag with file names whose extension is
// one of exts. A leading dot on an extension is optional, so ".json" and "json"
// are equivalent.
func CompleteFilesMatching(exts ...string) Option {
	normalized := make([]string, len(exts))
	for i, ext := range exts {
		normalized[i] = strings.TrimPrefix(ext, ".")
	}
	return completionOption(func(string) ([]string, annotation.CompletionDirective) {
		return normalized, annotation.CompletionFilterFileExt
	})
}

// CompleteDirs completes the flag with directory names only.
func CompleteDirs() Option {
	return completionOption(func(string) ([]string, annotation.CompletionDirective) {
		return nil, annotation.CompletionFilterDirs
	})
}

// completionOption builds an [Option] that assigns complete as the flag's
// completer, panicking if a completer was already assigned.
func completionOption(complete completerFunc) Option {
	return option(func(c *config) {
		if c.completer != nil {
			panic("flag: multiple completion options specified for one flag")
		}
		c.completer = complete
	})
}
