package strcase

import "unicode"

// class categorizes a rune for the purpose of word-boundary detection.
type class int

const (
	other class = iota // neither a letter nor a digit: a separator
	lower              // a lowercase or caseless letter
	upper              // an uppercase letter
	digit              // a decimal digit
)

// classify reports the [class] of r. Letters without case (such as CJK
// ideographs) are treated as [lower] so that they count as word content and do
// not introduce spurious casing boundaries.
func classify(r rune) class {
	switch {
	case unicode.IsUpper(r):
		return upper
	case unicode.IsLower(r):
		return lower
	case unicode.IsDigit(r):
		return digit
	case unicode.IsLetter(r):
		return lower
	default:
		return other
	}
}

// split tokenizes s into words using casing and separator transitions rather
// than a fixed delimiter. Runes that are neither letters nor digits act as
// separators and are dropped, and consecutive separators collapse. Digits stay
// attached to the word on their left, so "Base64Binary" yields
// ["Base64", "Binary"] and "AES512" yields ["AES512"]. Input without any letter
// or digit yields no words.
func split(s string) []string {
	runes := []rune(s)
	var words []string
	var current []rune

	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0]
		}
	}

	for i, r := range runes {
		if classify(r) == other {
			flush()
			continue
		}
		if i > 0 && boundaryBefore(runes, i) {
			flush()
		}
		current = append(current, r)
	}
	flush()

	return words
}

// boundaryBefore reports whether a new word begins at runes[i] based on the
// transition from the preceding rune. It assumes runes[i] is a letter or digit.
func boundaryBefore(runes []rune, i int) bool {
	prev, cur := classify(runes[i-1]), classify(runes[i])
	switch {
	case prev == lower && cur == upper:
		return true
	case prev == digit && cur == upper:
		return true
	case prev == upper && cur == upper && i+1 < len(runes) && classify(runes[i+1]) == lower:
		return true
	default:
		return false
	}
}
