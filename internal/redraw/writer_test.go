package redraw_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/redraw"
)

// alwaysOK is a flakyWriter budget large enough that no write fails.
const alwaysOK = 1 << 30

// flakyWriter records writes and fails after ok successful writes.
type flakyWriter struct {
	buf   bytes.Buffer
	ok    int
	count int
	err   error
}

func (f *flakyWriter) Write(b []byte) (int, error) {
	f.count++
	if f.count > f.ok {
		return 0, f.err
	}
	return f.buf.Write(b)
}

// drawAll performs the prerequisite draws for a test, failing if any errors.
func drawAll(t *testing.T, sut *redraw.Writer, draws []string) {
	t.Helper()
	for _, s := range draws {
		if err := sut.Draw(s); err != nil {
			t.Fatalf("Draw(%q) = %v, want nil", s, err)
		}
	}
}

func TestWriter_Draw(t *testing.T) {
	t.Parallel()

	testErr := errors.New("write refused")
	testCases := []struct {
		name    string
		draws   []string
		ok      int
		want    string
		wantErr error
	}{
		{
			name:    "FirstDrawWritesVerbatim",
			draws:   []string{"loading"},
			ok:      alwaysOK,
			want:    "loading",
			wantErr: nil,
		}, {
			name:    "SecondDrawErasesSingleLineBlock",
			draws:   []string{"a", "bb"},
			ok:      alwaysOK,
			want:    "a\r\x1b[0Jbb",
			wantErr: nil,
		}, {
			name:    "MultiLineDrawWritesVerbatim",
			draws:   []string{"x\ny"},
			ok:      alwaysOK,
			want:    "x\ny",
			wantErr: nil,
		}, {
			name:    "RedrawErasesMultiLineBlock",
			draws:   []string{"x\ny", "z"},
			ok:      alwaysOK,
			want:    "x\ny\r\x1b[1A\x1b[0Jz",
			wantErr: nil,
		}, {
			name:    "FirstWriteFails",
			draws:   []string{"anything"},
			ok:      0,
			want:    "",
			wantErr: testErr,
		}, {
			name:    "EraseWriteFails",
			draws:   []string{"a", "b"},
			ok:      1,
			want:    "a",
			wantErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fw := &flakyWriter{ok: tc.ok, err: testErr}
			sut := redraw.NewWriter(fw)

			// Act
			var err error
			for _, s := range tc.draws {
				err = sut.Draw(s)
			}

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Draw(...) = %v, want %v", got, want)
			}
			if got, want := fw.buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Draw(...) output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriter_Flush(t *testing.T) {
	t.Parallel()

	testErr := errors.New("write refused")
	testCases := []struct {
		name    string
		draws   []string
		ok      int
		want    string
		wantErr error
	}{
		{
			name:    "TerminatesBlockWithNewline",
			draws:   []string{"done"},
			ok:      alwaysOK,
			want:    "done\n",
			wantErr: nil,
		}, {
			name:    "WithoutPriorDrawIsNoOp",
			draws:   nil,
			ok:      alwaysOK,
			want:    "",
			wantErr: nil,
		}, {
			name:    "NewlineWriteFails",
			draws:   []string{"x"},
			ok:      1,
			want:    "x",
			wantErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fw := &flakyWriter{ok: tc.ok, err: testErr}
			sut := redraw.NewWriter(fw)
			drawAll(t, sut, tc.draws)

			// Act
			err := sut.Flush()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Flush() = %v, want %v", got, want)
			}
			if got, want := fw.buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Flush() output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriter_Clear(t *testing.T) {
	t.Parallel()

	testErr := errors.New("write refused")
	testCases := []struct {
		name    string
		draws   []string
		ok      int
		want    string
		wantErr error
	}{
		{
			name:    "ErasesBlock",
			draws:   []string{"scratch"},
			ok:      alwaysOK,
			want:    "scratch\r\x1b[0J",
			wantErr: nil,
		}, {
			name:    "WithoutPriorDrawIsNoOp",
			draws:   nil,
			ok:      alwaysOK,
			want:    "",
			wantErr: nil,
		}, {
			name:    "EraseWriteFails",
			draws:   []string{"a"},
			ok:      1,
			want:    "a",
			wantErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fw := &flakyWriter{ok: tc.ok, err: testErr}
			sut := redraw.NewWriter(fw)
			drawAll(t, sut, tc.draws)

			// Act
			err := sut.Clear()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Clear() = %v, want %v", got, want)
			}
			if got, want := fw.buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Clear() output = %q, want %q", got, want)
			}
		})
	}
}
