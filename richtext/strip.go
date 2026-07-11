package richtext

import (
	"strings"
	"unicode/utf8"

	"github.com/bitwizeshift/go-cli/richtext/internal/token"
)

// Strip returns s with every valid tag removed, leaving only the visible text.
// An unknown namespace is not a valid tag, so it and its closing tag are kept
// verbatim, matching how [Writer] treats them.
func Strip(s string) string {
	var scanner token.Scanner
	var b strings.Builder
	for _, tok := range scanner.Scan([]byte(s)) {
		writeVisible(&b, tok)
	}
	if tok, ok := scanner.Flush(); ok {
		b.WriteString(tok.Raw)
	}
	return b.String()
}

// Len returns the number of runes in the visible text of s, ignoring valid
// tags. It is intended for alignment and width calculations.
func Len(s string) int {
	return utf8.RuneCountInString(Strip(s))
}

func writeVisible(b *strings.Builder, tok token.Token) {
	if tok.Kind == token.Text || !isKnownNamespace(tok.Namespace) {
		b.WriteString(tok.Raw)
	}
}
