package plain

import (
	"io"
	"strings"

	"github.com/bitwizeshift/go-cli/richtext"
)

// Render returns markup with its styling tags removed, as a richtext writer
// would emit it with colour disabled. It reports an error when the markup's tags
// are unbalanced.
func Render(markup string) (string, error) {
	var buf strings.Builder
	w := richtext.NewWriter(&buf, nil)
	w.EnableColour(false)
	if _, err := io.WriteString(w, markup); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
