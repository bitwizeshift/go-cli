package panichandler

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
)

// tracePipe is the prefix that indents each stack-trace line.
const tracePipe = " | "

// data is the template model for a panic report.
type data struct {
	Message  string
	Trace    []string
	IssueURL string
}

// newData derives the template model from ctx.
func newData(ctx PanicContext) data {
	return data{
		Message:  fmt.Sprintf("%v", ctx.Err),
		Trace:    splitStack(ctx.Stack),
		IssueURL: ctx.IssueURL,
	}
}

// splitStack breaks a raw stack trace into its lines, dropping the trailing
// newline left by [runtime.Stack].
func splitStack(stack []byte) []string {
	return strings.Split(strings.TrimRight(string(stack), "\n"), "\n")
}

// traceBlock renders the stack-trace lines, prefixing each with a coloured
// gutter.
func traceBlock(p palette.Palette, lines []string) string {
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, p.Gutter(tracePipe)+p.Quote(line))
	}
	return strings.Join(rendered, "\n")
}

// funcs builds the template function map using palette p. It extends the shared
// [tmplfuncs.NewFunc] set with the stack-trace block layout function.
func funcs(p palette.Palette) template.FuncMap {
	f := tmplfuncs.NewFunc(p)
	f["traceBlock"] = func(lines []string) string { return traceBlock(p, lines) }
	return f
}
