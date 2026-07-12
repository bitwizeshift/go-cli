package diagnostic

import (
	"log/slog"
	"time"

	"github.com/bitwizeshift/go-cli/diagnostic/internal/diagslog"
)

// Location represents the location of a diagnostic in the source code.
type Location struct {
	File        string `json:"file,omitzero"`
	LineStart   int    `json:"line_start,omitzero"`
	LineEnd     int    `json:"line_end,omitzero"`
	ColumnStart int    `json:"column_start,omitzero"`
	ColumnEnd   int    `json:"column_end,omitzero"`
}

// Diagnostic represents a diagnostic message, such as an error or warning, that
// can be reported to the user.
type Diagnostic struct {
	ID      string `json:"id,omitzero"`
	Title   string `json:"title,omitzero"`
	Message string `json:"message,omitzero"`

	Location *Location `json:"location,omitempty"`

	// Err is a sentinel error that can be used to indicate the type of error this
	// diagnostic is associated with.
	Err error `json:"-"`
}

func (d *Diagnostic) toRecord(level slog.Level) slog.Record {
	attrs := []slog.Attr{
		slog.String(diagslog.AttrID, d.ID),
		slog.String(diagslog.AttrTitle, d.Title),
		slog.String(diagslog.AttrMessage, d.Message),
		slog.String(diagslog.AttrSeverity, level.String()),
	}
	attrs = append(attrs, d.locationToAttr(d.Location)...)

	r := slog.NewRecord(
		time.Now(),
		level,
		d.Message,
		0,
	)

	r.AddAttrs(attrs...)
	return r
}

func (d *Diagnostic) locationToAttr(loc *Location) []slog.Attr {
	if loc == nil {
		return nil
	}
	var attrs []any
	if loc.File != "" {
		attrs = append(attrs, slog.String(diagslog.AttrFile, loc.File))
	}
	if loc.LineStart != 0 {
		attrs = append(attrs, slog.Int(diagslog.AttrLineStart, loc.LineStart))
	}
	if loc.LineEnd != 0 {
		attrs = append(attrs, slog.Int(diagslog.AttrLineEnd, loc.LineEnd))
	}
	if loc.ColumnStart != 0 {
		attrs = append(attrs, slog.Int(diagslog.AttrColumnStart, loc.ColumnStart))
	}
	if loc.ColumnEnd != 0 {
		attrs = append(attrs, slog.Int(diagslog.AttrColumnEnd, loc.ColumnEnd))
	}
	if len(attrs) == 0 {
		return nil
	}
	return []slog.Attr{slog.Group(diagslog.AttrSource, attrs...)}
}
