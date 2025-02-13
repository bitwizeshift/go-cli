package diagnostic

import "fmt"

// Position represents the (line, column) pair location in a file or source
// content.
type Position struct {
	// Line is the line number in the file or source content. Begins at 1,
	// 0 if unknown.
	Line int `json:"line,omitempty"`

	// Column is the column number in the file or source content. Begins at 1,
	// 0 if unknown.
	Column int `json:"column,omitempty"`
}

// Diagnostic represents a diagnostic message that can be reported to the user.
type Diagnostic struct {
	// Severity is the severity of the diagnostic message. This field is
	// required.
	Severity Severity `json:"severity"`

	// Code is an optional 'code' for the diagnostic being emitted that will
	// appear connected to the severity (e,g, `error[E1234]`). This field is
	// optional.
	Code string `json:"code,omitempty"`

	// Title is an optional title for the diagnostic message.
	// This field is optional.
	Title string `json:"title,omitempty"`

	// Message is the actual diagnostic message to be reported. This field is
	// required.
	Message string `json:"message"`

	// File is the file path where the diagnostic message is located. This field
	// is optional.
	File string `json:"file,omitempty"`

	// Start is the starting position of the diagnostic message. This field is
	// optional.
	Start *Position `json:"position,omitempty"`

	// End is the ending position of the diagnostic message. This field is
	// optional, but if provided, Start must also be provided.
	End *Position `json:"end,omitempty"`
}

// New creates a new diagnostic message with the provided severity, format, and
// arguments.
func New(severity Severity, message string) *Diagnostic {
	return &Diagnostic{
		Severity: severity,
		Message:  message,
	}
}

// Errorf creates a new error-level diagnostic message with the provided format
// and arguments.
func Errorf(format string, args ...any) *Diagnostic {
	return New(SeverityError, fmt.Sprintf(format, args...))
}

// Warningf creates a new warning-level diagnostic message with the provided
// format and arguments.
func Warningf(format string, args ...any) *Diagnostic {
	return New(SeverityWarning, fmt.Sprintf(format, args...))
}

// Noticef creates a new notice-level diagnostic message with the provided
// format and arguments.
func Noticef(format string, args ...any) *Diagnostic {
	return New(SeverityNotice, fmt.Sprintf(format, args...))
}

// Debugf creates a new debug-level diagnostic message with the provided format
// and arguments.
func Debugf(format string, args ...any) *Diagnostic {
	return New(SeverityDebug, fmt.Sprintf(format, args...))
}

// With adds the provided annotations to the diagnostic message.
func (d *Diagnostic) With(annotations ...Annotation) *Diagnostic {
	for _, annotation := range annotations {
		annotation.annotate(d)
	}
	return d
}

func (d *Diagnostic) validate() error {
	if d == nil {
		return fmt.Errorf("diagnostic: nil diagnostic")
	}
	if d.Severity == "" {
		return fmt.Errorf("diagnostic: required field 'severity' is missing")
	}
	if d.Message == "" {
		return fmt.Errorf("diagnostic: required field 'message' is missing")
	}
	if err := d.Severity.validate(); err != nil {
		return err
	}
	return nil
}
