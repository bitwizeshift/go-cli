package token

import "bytes"

// Scanner turns successive byte chunks into tokens. Its zero value is ready to
// use. It retains only an incomplete trailing tag between calls, never the whole
// input.
type Scanner struct {
	pending []byte // an unterminated "[..." fragment carried to the next Scan
}

// Scan consumes p and returns the tokens recognised so far. Any trailing bytes
// that form an incomplete tag are retained and reconsidered on the next call.
func (s *Scanner) Scan(p []byte) []Token {
	data := p
	if len(s.pending) > 0 {
		data = append(s.pending, p...)
		s.pending = nil
	}

	var tokens []Token
	textStart, i := 0, 0
	for i < len(data) {
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
		i = end + 1
		textStart = i
	}
	if textStart < len(data) {
		tokens = append(tokens, textToken(data[textStart:]))
	}
	return tokens
}

// Flush returns the buffered incomplete tag as a literal-text token, reporting
// false when nothing is pending. Call it once no further input will arrive.
func (s *Scanner) Flush() (Token, bool) {
	if len(s.pending) == 0 {
		return Token{}, false
	}
	tok := textToken(s.pending)
	s.pending = nil
	return tok, true
}
