package strcase

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ToPascal converts s to PascalCase by title-casing each detected word and
// concatenating them. It returns "" when s contains no letters or digits.
// Acronyms are normalized, so "HTTPServer" becomes "HttpServer".
func ToPascal(s string) string {
	words := split(s)
	for i, w := range words {
		words[i] = title(w)
	}
	return strings.Join(words, "")
}

// ToCamel converts s to camelCase by lowercasing the first detected word,
// title-casing the rest, and concatenating them. It returns "" when s contains
// no letters or digits. Acronyms are normalized, so "HTTPServer" becomes
// "httpServer".
func ToCamel(s string) string {
	words := split(s)
	for i, w := range words {
		if i == 0 {
			words[i] = strings.ToLower(w)
			continue
		}
		words[i] = title(w)
	}
	return strings.Join(words, "")
}

// ToKebab converts s to kebab-case: each detected word lowercased and joined
// with "-". It returns "" when s contains no letters or digits.
func ToKebab(s string) string {
	return join(split(s), "-", strings.ToLower)
}

// ToSnake converts s to snake_case: each detected word lowercased and joined
// with "_". It returns "" when s contains no letters or digits.
func ToSnake(s string) string {
	return join(split(s), "_", strings.ToLower)
}

// ToScreamingSnake converts s to SCREAMING_SNAKE_CASE: each detected word
// uppercased and joined with "_". It returns "" when s contains no letters or
// digits.
func ToScreamingSnake(s string) string {
	return join(split(s), "_", strings.ToUpper)
}

// join applies transform to every word and concatenates the results with sep.
func join(words []string, sep string, transform func(string) string) string {
	for i, w := range words {
		words[i] = transform(w)
	}
	return strings.Join(words, sep)
}

// title uppercases the first rune of w and lowercases the remainder. It is only
// called with the non-empty words produced by [split].
func title(w string) string {
	r, size := utf8.DecodeRuneInString(w)
	return string(unicode.ToUpper(r)) + strings.ToLower(w[size:])
}
