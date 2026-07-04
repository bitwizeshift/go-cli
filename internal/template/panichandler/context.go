package panichandler

// PanicContext is the input to a panic report.
type PanicContext struct {
	// Err is the recovered value. This is ually an error, but not required by
	// panic's contract.
	Err any

	// Stack is the raw stack trace, as produced by [runtime.Stack].
	Stack []byte

	// IssueURL is an optional link to file a bug report. When non-empty, a footer
	// inviting the user to file an issue is shown.
	IssueURL string
}
