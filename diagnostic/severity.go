package diagnostic

import (
	"encoding"
	"fmt"
)

// Severity represents the severity of a diagnostic message.
type Severity string

const (
	// SeverityError is used for error-level diagnostics
	SeverityError Severity = "error"

	// SeverityWarning is used for warning-level diagnostics
	SeverityWarning Severity = "warning"

	// SeverityNotice is used for notice-level diagnostics
	SeverityNotice Severity = "notice"

	// SeverityDebug is used for debug-level diagnostics
	SeverityDebug Severity = "debug"
)

// UnmarshalText unmarshals a severity from a string.
func (s *Severity) UnmarshalText(b []byte) error {
	switch string(b) {
	case string(SeverityError):
		*s = SeverityError
	case string(SeverityWarning):
		*s = SeverityWarning
	case string(SeverityNotice):
		*s = SeverityNotice
	case string(SeverityDebug):
		*s = SeverityDebug
	default:
		return fmt.Errorf("unknown severity: %s", b)
	}
	return nil
}

func (s *Severity) validate() error {
	switch *s {
	case SeverityError, SeverityWarning, SeverityNotice, SeverityDebug:
		return nil
	default:
		return fmt.Errorf("severity: unknown value '%s'", *s)
	}
}

var _ encoding.TextUnmarshaler = (*Severity)(nil)
