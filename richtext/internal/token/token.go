package token

import "strings"

// Kind classifies a [Token].
type Kind uint8

const (
	// Text is a run of literal text.
	Text Kind = iota
	// Open is an opening tag, such as "[fg:red]".
	Open
	// Close is a closing tag, such as "[/fg]".
	Close
)

const (
	// RawNamespace opens a passthrough region whose contents are scanned as
	// literal text until the region is closed.
	RawNamespace = "richtext"

	// RawField is the only field RawNamespace accepts to begin a passthrough
	// region.
	RawField = "off"
)

// isRawOpen reports whether t opens a passthrough region.
func isRawOpen(t Token) bool {
	return t.Kind == Open && t.Namespace == RawNamespace && t.Field == RawField
}

// Token is a single unit produced by scanning tag markup. Raw always holds the
// exact source text of the token; Namespace and Field carry the parsed parts of
// a tag.
type Token struct {
	Kind      Kind
	Raw       string
	Namespace string // set for Open and Close
	Field     string // set for Open
}

func textToken(b []byte) Token {
	return Token{Kind: Text, Raw: string(b)}
}

// parseTag interprets a complete bracketed span (including the surrounding
// brackets) as a tag. It reports false when the span is not a well-formed tag,
// in which case the caller keeps the span as literal text.
func parseTag(raw string) (Token, bool) {
	inner := raw[1 : len(raw)-1]
	if rest, ok := strings.CutPrefix(inner, "/"); ok {
		if !isNamespace(rest) {
			return Token{}, false
		}
		return Token{Kind: Close, Raw: raw, Namespace: rest}, true
	}
	ns, field, ok := strings.Cut(inner, ":")
	if !ok || !isNamespace(ns) || field == "" {
		return Token{}, false
	}
	return Token{Kind: Open, Raw: raw, Namespace: ns, Field: field}, true
}

// isNamespace reports whether s is a syntactically valid namespace: a non-empty
// run of ASCII letters. Whether the namespace is meaningful is decided by the
// consumer, not here.
func isNamespace(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}
