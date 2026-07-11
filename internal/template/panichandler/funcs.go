package panichandler

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/tag"
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

// traceBlock renders the stack-trace lines, prefixing each with a styled gutter.
// Each line's content is wrapped in a passthrough region so bracketed sequences
// in the trace are never parsed as tags.
func traceBlock(lines []string) string {
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, tag.Themed("gutter", tracePipe)+tag.Themed("quote", tag.Raw(line)))
	}
	return strings.Join(rendered, "\n")
}

// funcs builds the template function map. It extends the shared
// [tmplfuncs.NewFunc] set with the stack-trace block layout function.
func funcs() template.FuncMap {
	f := tmplfuncs.NewFunc()
	f["traceBlock"] = func(lines []string) string { return traceBlock(lines) }
	return f
}
