package token

import "bytes"

// rawClose is the exact sequence that ends a passthrough region opened by a
// "[richtext:off]" tag.
const rawClose = "[/" + RawNamespace + "]"

// Scanner turns successive byte chunks into tokens. Its zero value is ready to
// use. It retains only an incomplete trailing fragment between calls, never the
// whole input.
type Scanner struct {
	pending []byte // an incomplete trailing fragment carried to the next Scan
	raw     bool   // inside a "[richtext:off]...[/richtext]" passthrough region
}

// Scan consumes p and returns the tokens recognised so far. Any trailing bytes
// that form an incomplete tag, or that may begin a passthrough region's closing
// sequence, are retained and reconsidered on the next call.
func (s *Scanner) Scan(p []byte) []Token {
	data := p
	if len(s.pending) > 0 {
		data = append(s.pending, p...)
		s.pending = nil
	}

	var tokens []Token
	textStart, i := 0, 0
	for i < len(data) {
		if s.raw {
			raw, next, done := s.scanRaw(data, i)
			tokens = append(tokens, raw...)
			if !done {
				return tokens
			}
			i, textStart = next, next
			continue
		}
		if data[i] != '[' {
			i++
			continue
		}
		rel := bytes.IndexByte(data[i+1:], ']')
		if rel < 0 {
			if i > textStart {
				tokens = append(tokens, textToken(data[textStart:i]))
			}
			s.pending = append(s.pending, data[i:]...)
			return tokens
		}
		end := i + 1 + rel
		tag, ok := parseTag(string(data[i : end+1]))
		if !ok {
			i++
			continue
		}
		if i > textStart {
			tokens = append(tokens, textToken(data[textStart:i]))
		}
		tokens = append(tokens, tag)
		i, textStart = end+1, end+1
		if isRawOpen(tag) {
			s.raw = true
		}
	}
	if !s.raw && textStart < len(data) {
		tokens = append(tokens, textToken(data[textStart:]))
	}
	return tokens
}

// scanRaw consumes the passthrough region beginning at from. It returns the
// tokens recognised, the index to resume from, and whether the region's close
// was found. When the close is absent it emits the literal text seen so far,
// buffers any trailing bytes that may begin the closing sequence, and reports
// done as false.
func (s *Scanner) scanRaw(data []byte, from int) (tokens []Token, next int, done bool) {
	if rel := bytes.Index(data[from:], []byte(rawClose)); rel >= 0 {
		end := from + rel
		if end > from {
			tokens = append(tokens, textToken(data[from:end]))
		}
		tokens = append(tokens, Token{Kind: Close, Raw: rawClose, Namespace: RawNamespace})
		s.raw = false
		return tokens, end + len(rawClose), true
	}

	keep := rawClosePrefixLen(data[from:])
	cut := len(data) - keep
	if cut > from {
		tokens = append(tokens, textToken(data[from:cut]))
	}
	s.pending = append(s.pending, data[cut:]...)
	return tokens, len(data), false
}

// rawClosePrefixLen returns the length of the longest suffix of data that is a
// proper prefix of [rawClose], so those bytes can be reconsidered once more
// input arrives.
func rawClosePrefixLen(data []byte) int {
	limit := min(len(data), len(rawClose)-1)
	for keep := limit; keep > 0; keep-- {
		if bytes.Equal(data[len(data)-keep:], []byte(rawClose[:keep])) {
			return keep
		}
	}
	return 0
}

// Flush returns any buffered incomplete fragment as a literal-text token,
// reporting false when nothing is pending. Call it once no further input will
// arrive.
func (s *Scanner) Flush() (Token, bool) {
	if len(s.pending) == 0 {
		return Token{}, false
	}
	tok := textToken(s.pending)
	s.pending = nil
	return tok, true
}
