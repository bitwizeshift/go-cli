package diagnostic

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"rodusek.dev/pkg/cli/formatting/ansi"
	"rodusek.dev/pkg/cli/formatting/wrap"
)

// terminalEmitter is an [Emitter] that always prints object in ANSI format.
type terminalEmitter struct {
	writer   io.Writer
	basePath string
	wrapper  *wrap.Wrapper
}

// Emit emits a diagnostic message in a format that is suitable for ANSI
// terminals.
func (e *terminalEmitter) Emit(msg *Diagnostic) error {
	format := e.severityFormat(msg.Severity)
	if err := e.writeTitle(format, msg); err != nil {
		return err
	}
	if err := e.writeFile(format, msg); err != nil {
		return err
	}
	if err := e.writeExcerpt(format, msg); err != nil {
		return err
	}
	if err := e.writeUnderline(format, msg); err != nil {
		return err
	}
	if err := e.writeMessage(format, msg); err != nil {
		return err
	}
	return nil
}

func (e *terminalEmitter) writeTitle(format ansi.SGRControlSequence, msg *Diagnostic) error {
	if _, err := fmt.Fprintf(e.writer, "%v%v", format, msg.Severity); err != nil {
		return err
	}
	if msg.Code != "" {
		if _, err := fmt.Fprintf(e.writer, "[%v%v%v]", codeFormat, msg.Code, format); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(e.writer, ":%v", ansi.Reset); err != nil {
		return err
	}
	if msg.Title != "" {
		if _, err := fmt.Fprintf(e.writer, " %v%v%v\n", titleFormat, msg.Title, ansi.Reset); err != nil {
			return err
		}
	} else if msg.Message != "" {
		text := e.wrapper.String(msg.Message)
		if _, err := fmt.Fprintf(e.writer, " %v\n", text); err != nil {
			return err
		}
	}
	return nil
}

func (e *terminalEmitter) writeFile(format ansi.SGRControlSequence, msg *Diagnostic) error {
	prefix := e.format(msg)
	if msg.File != "" {
		if _, err := fmt.Fprintf(e.writer, " %s%v--->%v", prefix, format, ansi.Reset); err != nil {
			return err
		}

		source := e.source(msg)

		if _, err := fmt.Fprintf(e.writer, " %v%v%v\n", fileFormat, source, ansi.Reset); err != nil {
			return err
		}
	}
	return nil
}

func (e *terminalEmitter) source(msg *Diagnostic) string {
	var sb strings.Builder
	_, _ = sb.WriteString(e.relPath(msg.File))
	if msg.Start != nil && msg.Start.Line != 0 {
		_, _ = sb.WriteString(":")
		_, _ = sb.WriteString(fmt.Sprint(msg.Start.Line))
		if msg.End != nil && msg.Start.Line != 0 {
			_, _ = sb.WriteString("-")
			_, _ = sb.WriteString(fmt.Sprint(msg.End.Line))
			return sb.String()
		}
		_, _ = sb.WriteString(":")
		_, _ = sb.WriteString(fmt.Sprint(msg.Start.Column))
		if msg.End != nil {
			_, _ = sb.WriteString("-")
			_, _ = sb.WriteString(fmt.Sprint(msg.End.Column))
		}
	}
	return sb.String()
}

func (e *terminalEmitter) writeExcerpt(format ansi.SGRControlSequence, msg *Diagnostic) error {
	if msg.Start == nil || msg.Start.Line == 0 || msg.File == "" {
		return nil
	}
	start := msg.Start
	if start == nil || start.Line == 0 {
		return nil
	}
	data, err := readLineRange(msg.File, start, msg.End)
	if err != nil {
		// If we can't read the file for some reason, just skip the excerpt
		return nil
	}
	if data.start.line == 0 {
		return nil
	}
	if err := e.writeBlockQuotePrefix(format, msg); err != nil {
		return err
	}
	digits := max(countDigits(data.start.line), countDigits(data.end.line))
	startLine := strings.ReplaceAll(data.start.text, "\t", "   ")
	if _, err := fmt.Fprintf(e.writer, " %v%*d%v |%v   %v%v\n", codeFormat, digits, data.start.line, format, excerptFormat, startLine, ansi.Reset); err != nil {
		return err
	}
	if data.end.line == 0 || data.end.line == data.start.line {
		return nil
	}
	if !data.isContiguous() {
		prefix := e.format(msg)
		if _, err := fmt.Fprintf(e.writer, " %v%v |%v   ...%v\n", prefix, format, excerptFormat, ansi.Reset); err != nil {
			return err
		}
	}
	endLine := strings.ReplaceAll(data.end.text, "\t", "   ")
	if _, err := fmt.Fprintf(e.writer, " %v%*d%v |%v   %v%v\n", codeFormat, digits, data.end.line, format, excerptFormat, endLine, ansi.Reset); err != nil {
		return err
	}
	return nil
}

func (e *terminalEmitter) writeUnderline(format ansi.SGRControlSequence, msg *Diagnostic) error {
	if msg.Start == nil || msg.Start.Column == 0 {
		return nil
	}
	if msg.End != nil && (msg.Start.Line != msg.End.Line || msg.Start.Column > msg.End.Column) {
		return nil
	}

	data, err := readLineRange(msg.File, msg.Start, msg.End)
	if err != nil {
		// If we can't read the file for some reason, just skip the excerpt
		return nil
	}
	distance := 1
	if msg.End != nil {
		end := countPrefixSpaces(data.end.text, msg.End.Column)
		start := countPrefixSpaces(data.start.text, msg.Start.Column)
		distance = end - start
	}
	underline := "^"
	if distance > 1 {
		underline += strings.Repeat("~", distance)
	}
	spaces := strings.Repeat(" ", max(0, countPrefixSpaces(data.start.text, msg.Start.Column)-1))

	prefix := e.format(msg)
	if _, err := fmt.Fprintf(e.writer, " %v%v |   %s%v%v%v\n", prefix, format, spaces, underlineFormat, underline, ansi.Reset); err != nil {
		return err
	}

	return nil
}

func (e *terminalEmitter) writeMessage(format ansi.SGRControlSequence, msg *Diagnostic) error {
	if msg.Title == "" {
		return nil
	}
	prefix := e.format(msg)
	if msg.Message != "" {
		wrapper := e.wrapper
		if wrapper == nil {
			wrapper = &wrap.Wrapper{MaxWidth: 120}
		}
		wrapper.MaxWidth -= (len(prefix) + 3)
		if err := e.writeEmptyBlockQuoteLine(format, msg); err != nil {
			return err
		}
		lines := wrapper.Lines(msg.Message)
		for _, line := range lines {
			if _, err := fmt.Fprintf(e.writer, " %v%v%v | %v%v\n", format, prefix, format, ansi.Reset, line); err != nil {
				return err
			}
		}
		if len(lines) > 0 {
			if err := e.writeEmptyBlockQuoteLine(format, msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *terminalEmitter) writeBlockQuotePrefix(format ansi.SGRControlSequence, msg *Diagnostic) error {
	prefix := e.format(msg)
	if _, err := fmt.Fprintf(e.writer, " %v%v%v |%v\n", format, prefix, format, ansi.Reset); err != nil {
		return err
	}
	return nil
}

func (e *terminalEmitter) writeEmptyBlockQuoteLine(format ansi.SGRControlSequence, msg *Diagnostic) error {
	if err := e.writeBlockQuotePrefix(format, msg); err != nil {
		return err
	}
	return nil
}

func (e *terminalEmitter) format(msg *Diagnostic) string {
	indentAmount := 1
	if msg.Start != nil && msg.Start.Line > 0 {
		if digits := countDigits(msg.Start.Line); digits > indentAmount {
			indentAmount = digits
		}
	}
	if msg.End != nil && msg.End.Line > 0 {
		if digits := countDigits(msg.End.Line); digits > indentAmount {
			indentAmount = digits
		}
	}
	return strings.Repeat(" ", indentAmount)
}

var (
	titleFormat     = ansi.Format(ansi.FGBrightWhite)
	fileFormat      = ansi.Format(ansi.FGBrightWhite, ansi.Underline)
	underlineFormat = ansi.Format(ansi.FGGreen)
	quoteFormat     = ansi.Format(ansi.FGCyan)
	codeFormat      = ansi.Format(ansi.Bold, ansi.FGWhite)
	excerptFormat   = ansi.Format(ansi.FGBrightBlack)
)

func (e *terminalEmitter) severityFormat(severity Severity) ansi.SGRControlSequence {
	switch severity {
	case SeverityError:
		return ansi.Format(ansi.FGRed, ansi.Bold)
	case SeverityWarning:
		return ansi.Format(ansi.FGYellow)
	case SeverityNotice:
		return ansi.Format(ansi.FGCyan)
	case SeverityDebug:
		return ansi.Format(ansi.FGBrightBlack)
	default:
		return ansi.Format(ansi.FGBrightBlack)
	}
}

func (e *terminalEmitter) relPath(path string) string {
	root := e.basePath
	if root == "" {
		return path
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}

type lineData struct {
	// text is the raw line text after having translated any tabs to 3-spaces
	text string

	// line is the origin line number in the file or source content.
	line int

	// col is the column number in the file or source content, if it was specified
	// in the diagnostic message. Begins at 1, 0 if unspecified. The exact
	// location here is translated by the tab-to-space translation, counting '3'
	// for each tab character.
	col int
}

func readLineRange(file string, start, end *Position) (*lineRange, error) {
	lines, err := readLines(file)
	if err != nil {
		return nil, err
	}
	var lr lineRange
	if start != nil {
		lr.start = readLineData(lines, start)
	}
	if end != nil {
		lr.end = readLineData(lines, end)
	}
	return &lr, nil
}

func readLineData(lines []string, pos *Position) lineData {
	if pos.Line == 0 {
		return lineData{}
	}
	line := pos.Line - 1
	if line >= len(lines) {
		return lineData{}
	}
	col := pos.Column

	if col == 0 {
		return lineData{text: lines[line], line: pos.Line}
	}
	return lineData{text: lines[line], line: pos.Line, col: col}
}

func readLines(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

type lineRange struct {
	start lineData
	end   lineData
}

func (lr *lineRange) isContiguous() bool {
	return lr.start.line+1 == lr.end.line
}

func countPrefixSpaces(line string, col int) int {
	count := 0
	for i := 0; i < col; i++ {
		if len(line) <= i {
			break
		}
		if line[i] == '\t' {
			count += 3
		} else {
			count++
		}
	}
	return count
}

var _ Emitter = (*terminalEmitter)(nil)

func countPadding(msg *Diagnostic) int {
	padding := 1
	if msg.Start != nil && msg.Start.Line > 0 {
		padding = countDigits(msg.Start.Line)
	}
	if msg.End != nil && msg.End.Line > 0 {
		padding = countDigits(msg.End.Line)
	}
	return max(padding, 3)
}

func countDigits(i int) int {
	if i == 0 {
		return 1
	}
	count := 0
	for i != 0 {
		i /= 10
		count++
	}
	return count
}
