package diagslog

import "log/slog"

// Record represents a diagnostic message, such as an error or warning, that
// can be reported to the user.
type Record struct {
	ID      string
	Title   string
	Message string

	File        string
	LineStart   int64
	LineEnd     int64
	ColumnStart int64
	ColumnEnd   int64
}

const (
	AttrID          = "id"
	AttrTitle       = "title"
	AttrMessage     = "message"
	AttrSeverity    = "severity"
	AttrFile        = "file"
	AttrLineStart   = "line_start"
	AttrLineEnd     = "line_end"
	AttrColumnStart = "column_start"
	AttrColumnEnd   = "column_end"
	AttrSource      = "source"
)

// RecordFromSlog converts a [slog.Record] into a [Record]
func RecordFromSlog(r slog.Record) *Record {
	var result Record
	fields := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindGroup {
			for _, inner := range a.Value.Group() {
				fields[inner.Key] = inner.Value.Any()
			}
		} else {
			fields[a.Key] = a.Value.Any()
		}
		return true
	})

	result.ID, _ = fields[AttrID].(string)
	result.Title, _ = fields[AttrTitle].(string)
	result.Message, _ = fields[AttrMessage].(string)

	result.File, _ = fields[AttrFile].(string)
	result.LineStart, _ = fields[AttrLineStart].(int64)
	result.LineEnd, _ = fields[AttrLineEnd].(int64)
	result.ColumnStart, _ = fields[AttrColumnStart].(int64)
	result.ColumnEnd, _ = fields[AttrColumnEnd].(int64)
	return &result
}
