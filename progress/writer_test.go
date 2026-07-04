package progress_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/progress"
)

// alwaysOK is a flakyWriter budget large enough that no write fails.
const alwaysOK = 1 << 30

type staticRenderer string

func (s staticRenderer) Render() string { return string(s) }

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

func TestWriter_Update(t *testing.T) {
	t.Parallel()

	testErr := errors.New("write refused")
	testCases := []struct {
		name    string
		disable bool
		updates []string
		ok      int
		want    string
		wantErr error
	}{
		{
			name:    "EnabledRedrawsInPlace",
			disable: false,
			updates: []string{"aaa", "bb"},
			ok:      alwaysOK,
			want:    "aaa\r\x1b[0Jbb\n",
			wantErr: nil,
		}, {
			name:    "EnabledSingleUpdate",
			disable: false,
			updates: []string{"x"},
			ok:      alwaysOK,
			want:    "x\n",
			wantErr: nil,
		}, {
			name:    "DisabledEmitsOnlyFinalFrame",
			disable: true,
			updates: []string{"aaa", "bb"},
			ok:      alwaysOK,
			want:    "bb\n",
			wantErr: nil,
		}, {
			name:    "NoUpdatesWritesNothing",
			disable: false,
			updates: nil,
			ok:      alwaysOK,
			want:    "",
			wantErr: nil,
		}, {
			name:    "EnabledUpdateWriteFails",
			disable: false,
			updates: []string{"x"},
			ok:      0,
			want:    "",
			wantErr: testErr,
		}, {
			name:    "DisabledFinalDrawWriteFails",
			disable: true,
			updates: []string{"x"},
			ok:      0,
			want:    "",
			wantErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fw := &flakyWriter{ok: tc.ok, err: testErr}
			sut := &progress.Writer{
				W:                fw,
				DisableAnimation: tc.disable,
			}

			// Act
			var updateErr error
			for _, s := range tc.updates {
				updateErr = sut.Update(staticRenderer(s))
			}
			doneErr := sut.Done()
			err := errors.Join(updateErr, doneErr)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Writer.Update/Done(...) = %v, want %v", got, want)
			}
			if got, want := fw.buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Writer output = %q, want %q", got, want)
			}
		})
	}
}
