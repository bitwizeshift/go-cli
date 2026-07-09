package cursor

import "fmt"

// ClearDown is the escape sequence that erases from the cursor to the end of
// the screen.
const ClearDown string = "\x1b[0J"

// CursorUp returns the escape sequence that moves the cursor up n lines. It
// returns the empty string when n <= 0.
func CursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA", n)
}
