package diagslog_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/diagnostic/internal/diagslog"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func diagnosticRecord(level slog.Level, id, title, message, file string, lineStart, lineEnd, colStart, colEnd int64) slog.Record {
	rec := slog.NewRecord(time.Time{}, level, message, 0)
	attrs := []slog.Attr{
		slog.String("id", id),
		slog.String("title", title),
		slog.String("message", message),
		slog.String("severity", level.String()),
	}
	var locAttrs []any
	if file != "" {
		locAttrs = append(locAttrs, slog.String("file", file))
	}
	if lineStart != 0 {
		locAttrs = append(locAttrs, slog.Int64("line_start", lineStart))
	}
	if lineEnd != 0 {
		locAttrs = append(locAttrs, slog.Int64("line_end", lineEnd))
	}
	if colStart != 0 {
		locAttrs = append(locAttrs, slog.Int64("column_start", colStart))
	}
	if colEnd != 0 {
		locAttrs = append(locAttrs, slog.Int64("column_end", colEnd))
	}
	if len(locAttrs) > 0 {
		attrs = append(attrs, slog.Group("source", locAttrs...))
	}
	rec.AddAttrs(attrs...)
	return rec
}

// themeTag renders the richtext theme markup the handler emits for a styled
// span, mirroring the handler's internal themeFormat.
func themeTag(name, s string) string {
	return "[theme:" + name + "]" + s + "[/theme]"
}

// rawTag renders the richtext verbatim markup the handler emits for
// untrusted spans, mirroring the handler's internal noFormatter.
func rawTag(s string) string {
	return "[richtext:off]" + s + "[/richtext]"
}

// idHeader renders the leading "severity[id]" header markup for a diagnostic
// that carries an ID, styled with the given severity theme.
func idHeader(severityTheme, severityText, id string) string {
	return themeTag(severityTheme, severityText+"[") +
		themeTag("emphasis", rawTag(id)) +
		themeTag(severityTheme, "]")
}

func TestTextHandler_Handle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		level   slog.Level
		id      string
		title   string
		message string
		want    string
	}{
		{
			name:    "ErrorIDTitleEqualsMessage",
			level:   slog.LevelError,
			id:      "E1",
			title:   "boom",
			message: "boom",
			want:    idHeader("error", "error", "E1") + ": boom\n\n",
		}, {
			name:    "WarnNoID",
			level:   slog.LevelWarn,
			id:      "",
			title:   "t",
			message: "t",
			want:    themeTag("warning", "warn") + ": t\n\n",
		}, {
			name:    "ErrorTitleMessageDistinct",
			level:   slog.LevelError,
			id:      "E2",
			title:   "headline",
			message: "details",
			want: idHeader("error", "error", "E2") + ": headline\n" +
				"   " + themeTag("error", "|") + " " + rawTag("details") + "\n\n",
		}, {
			name:    "InfoTitleEmpty",
			level:   slog.LevelInfo,
			id:      "I1",
			title:   "",
			message: "m",
			want:    idHeader("info", "info", "I1") + ": m\n\n",
		}, {
			name:    "InfoMessageEmpty",
			level:   slog.LevelInfo,
			id:      "I2",
			title:   "t",
			message: "",
			want:    idHeader("info", "info", "I2") + ": t\n\n",
		}, {
			name:    "ErrorAllEmpty",
			level:   slog.LevelError,
			id:      "",
			title:   "",
			message: "",
			want:    themeTag("error", "error") + ": \n\n",
		}, {
			name:    "DebugSeverityUsesDebugTheme",
			level:   slog.LevelDebug,
			id:      "D1",
			title:   "t",
			message: "t",
			want:    idHeader("debug", "debug", "D1") + ": t\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := diagslog.NewTextHandler(&buf, term.FixedSizer(80))
			rec := diagnosticRecord(tc.level, tc.id, tc.title, tc.message, "", 0, 0, 0, 0)
			ctx := context.Background()

			// Act
			err := sut.Handle(ctx, rec)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("TextHandler.Handle(...) err = %v, want %v", got, want)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

func TestTextHandler_Handle_Location(t *testing.T) {
	t.Parallel()

	base := idHeader("error", "error", "E1") + ": t\n"
	locLine := func(loc string) string {
		return "  " + themeTag("error", "-->") + " " + themeTag("url", loc) + "\n"
	}

	testCases := []struct {
		name      string
		file      string
		lineStart int64
		lineEnd   int64
		colStart  int64
		colEnd    int64
		want      string
	}{
		{
			name:      "NoLocation",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      base + "\n",
		}, {
			name:      "FileOnly",
			file:      "a.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      base + locLine("a.cpp") + "\n",
		}, {
			name:      "FileAndLineStart",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      base + locLine("a.cpp:10") + "\n",
		}, {
			name:      "FileAndLineRangeEqual",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   10,
			colStart:  0,
			colEnd:    0,
			want:      base + locLine("a.cpp:10") + "\n",
		}, {
			name:      "FileAndLineRangeDifferent",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   12,
			colStart:  0,
			colEnd:    0,
			want:      base + locLine("a.cpp:10-12") + "\n",
		}, {
			name:      "FileLineColumnStart",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    0,
			want:      base + locLine("a.cpp:10:5") + "\n",
		}, {
			name:      "FileLineColumnRangeEqual",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    5,
			want:      base + locLine("a.cpp:10:5") + "\n",
		}, {
			name:      "FileLineColumnRangeDifferent",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    8,
			want:      base + locLine("a.cpp:10:5-8") + "\n",
		}, {
			name:      "FileLineRangeColumnRange",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   12,
			colStart:  5,
			colEnd:    8,
			want:      base + locLine("a.cpp:10-12:5-8") + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := diagslog.NewTextHandler(&buf, term.FixedSizer(80))
			rec := diagnosticRecord(
				slog.LevelError, "E1", "t", "t",
				tc.file, tc.lineStart, tc.lineEnd, tc.colStart, tc.colEnd,
			)
			ctx := context.Background()

			// Act
			err := sut.Handle(ctx, rec)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("TextHandler.Handle(...) err = %v, want %v", got, want)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

func TestTextHandler_Handle_WrapsMessageAcrossColumns(t *testing.T) {
	t.Parallel()

	// Arrange
	first := strings.Repeat("a", 40)
	second := strings.Repeat("b", 40)
	var buf bytes.Buffer
	sut := diagslog.NewTextHandler(&buf, term.FixedSizer(80))
	rec := diagnosticRecord(slog.LevelError, "E1", "t", first+" "+second, "", 0, 0, 0, 0)
	ctx := context.Background()
	want := idHeader("error", "error", "E1") + ": t\n" +
		"   " + themeTag("error", "|") + " " + rawTag(first) + "\n" +
		"   " + themeTag("error", "|") + " " + rawTag(second) + "\n\n"

	// Act
	err := sut.Handle(ctx, rec)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("TextHandler.Handle(...) err = %v, want %v", got, want)
	}
	if got, want := buf.String(), want; !cmp.Equal(got, want) {
		t.Errorf("TextHandler.Handle(...) = mismatch (-got +want):\n%s", cmp.Diff(want, got))
	}
}

func TestTextHandler_Enabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{
			name:  "Debug",
			level: slog.LevelDebug,
			want:  true,
		}, {
			name:  "Info",
			level: slog.LevelInfo,
			want:  true,
		}, {
			name:  "Warn",
			level: slog.LevelWarn,
			want:  true,
		}, {
			name:  "Error",
			level: slog.LevelError,
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := diagslog.NewTextHandler(&bytes.Buffer{}, term.FixedSizer(80))
			ctx := context.Background()

			// Act
			got := sut.Enabled(ctx, tc.level)

			// Assert
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Enabled(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestTextHandler_WithAttrs_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := diagslog.NewTextHandler(&bytes.Buffer{}, term.FixedSizer(80))

	// Act
	got := sut.WithAttrs([]slog.Attr{slog.String("k", "v")})

	// Assert
	if got, want := got, slog.Handler(sut); got != want {
		t.Errorf("TextHandler.WithAttrs(...) got %v, want %v", got, want)
	}
}

func TestTextHandler_WithGroup_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := diagslog.NewTextHandler(&bytes.Buffer{}, term.FixedSizer(80))

	// Act
	got := sut.WithGroup("g")

	// Assert
	if got, want := got, slog.Handler(sut); got != want {
		t.Errorf("TextHandler.WithGroup(...) got %v, want %v", got, want)
	}
}
