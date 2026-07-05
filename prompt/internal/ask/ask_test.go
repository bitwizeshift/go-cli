package ask_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/internal/term/termtest"
	"github.com/bitwizeshift/go-cli/prompt/internal/ask"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// blockingReader is an io.Reader whose Read never returns, simulating input that
// never arrives so cancellation can be exercised.
type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) {
	<-make(chan struct{})
	return 0, nil
}

func TestAsker_Line(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		prompt  string
		want    string
		wantErr error
	}{
		{
			name:   "ReadsLine",
			input:  "hello\n",
			prompt: "Name: ",
			want:   "hello",
		}, {
			name:   "TrimsCarriageReturn",
			input:  "hello\r\n",
			prompt: "Name: ",
			want:   "hello",
		}, {
			name:   "LastLineWithoutNewline",
			input:  "hello",
			prompt: "Name: ",
			want:   "hello",
		}, {
			name:    "EmptyInput",
			input:   "",
			prompt:  "Name: ",
			want:    "",
			wantErr: io.EOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			out := &bytes.Buffer{}
			sut := &ask.Asker{Out: out, In: strings.NewReader(tc.input)}
			ctx := context.Background()

			// Act
			line, err := sut.Line(ctx, tc.prompt)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Asker.Line(...) error got %v, want %v", got, want)
			}
			if got, want := line, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Asker.Line(...) got %q, want %q", got, want)
			}
			if got, want := out.String(), tc.prompt; !cmp.Equal(got, want) {
				t.Errorf("Asker.Line(...) prompt written got %q, want %q", got, want)
			}
		})
	}
}

func TestAsker_Value(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		target  any
		want    any
		wantErr error
	}{
		{
			name:   "Int",
			input:  "7\n",
			target: new(int),
			want:   7,
		}, {
			name:   "Duration",
			input:  "2s\n",
			target: new(time.Duration),
			want:   2 * time.Second,
		}, {
			name:    "Invalid",
			input:   "nope\n",
			target:  new(int),
			want:    0,
			wantErr: strconv.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			out := &bytes.Buffer{}
			sut := &ask.Asker{Out: out, In: strings.NewReader(tc.input)}
			ctx := context.Background()

			// Act
			err := sut.Value(ctx, "> ", tc.target)
			value := reflect.ValueOf(tc.target).Elem().Interface()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Asker.Value(...) error got %v, want %v", got, want)
			}
			if got, want := value, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Asker.Value(...) value got %v, want %v", got, want)
			}
		})
	}
}

func TestAsker_Confirm(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		want        bool
		wantPrompts int
	}{
		{
			name:        "Yes",
			input:       "y\n",
			want:        true,
			wantPrompts: 1,
		}, {
			name:        "YesWord",
			input:       "yes\n",
			want:        true,
			wantPrompts: 1,
		}, {
			name:        "No",
			input:       "n\n",
			want:        false,
			wantPrompts: 1,
		}, {
			name:        "NoWord",
			input:       "no\n",
			want:        false,
			wantPrompts: 1,
		}, {
			name:        "CaseInsensitive",
			input:       "Y\n",
			want:        true,
			wantPrompts: 1,
		}, {
			name:        "RepromptsOnUnknown",
			input:       "maybe\ny\n",
			want:        true,
			wantPrompts: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			out := &bytes.Buffer{}
			sut := &ask.Asker{Out: out, In: strings.NewReader(tc.input)}
			ctx := context.Background()

			// Act
			answer, err := sut.Confirm(ctx, "Continue?")
			prompts := strings.Count(out.String(), "Continue?")

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Asker.Confirm(...) error got %v, want %v", got, want)
			}
			if got, want := answer, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Asker.Confirm(...) got %v, want %v", got, want)
			}
			if got, want := prompts, tc.wantPrompts; !cmp.Equal(got, want) {
				t.Errorf("Asker.Confirm(...) prompt count got %d, want %d", got, want)
			}
		})
	}
}

func TestAsker_Secret(t *testing.T) {
	t.Parallel()

	errNotTTY := errors.New("not a terminal")
	errRestore := errors.New("restore failed")

	testCases := []struct {
		name     string
		input    string
		disabler term.EchoDisabler
		want     string
		wantOut  string
		wantErr  error
	}{
		{
			name:     "ReadsMasked",
			input:    "hunter2\n",
			disabler: termtest.NoOpEchoDisabler(),
			want:     "hunter2",
			wantOut:  "*******\r\n",
		}, {
			name:     "Backspace",
			input:    "ab\x7fc\n",
			disabler: termtest.NoOpEchoDisabler(),
			want:     "ac",
			wantOut:  "**\b \b*\r\n",
		}, {
			name:     "Interrupt",
			input:    "secret\x03",
			disabler: termtest.NoOpEchoDisabler(),
			want:     "",
			wantOut:  "******",
			wantErr:  ask.ErrInterrupted,
		}, {
			name:     "NotATerminal",
			input:    "secret\n",
			disabler: termtest.ErrEchoDisabler(errNotTTY),
			want:     "",
			wantOut:  "",
			wantErr:  errNotTTY,
		}, {
			name:     "RestoreError",
			input:    "pw\n",
			disabler: termtest.ErrRestoreDisabler(errRestore),
			want:     "pw",
			wantOut:  "**\r\n",
			wantErr:  errRestore,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			out := &bytes.Buffer{}
			sut := &ask.Asker{
				Out:          out,
				In:           strings.NewReader(tc.input),
				EchoDisabler: tc.disabler,
				HiddenChar:   '*',
			}
			ctx := context.Background()

			// Act
			secret, err := sut.Secret(ctx, "")

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Asker.Secret(...) error got %v, want %v", got, want)
			}
			if got, want := secret, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Asker.Secret(...) got %q, want %q", got, want)
			}
			if got, want := out.String(), tc.wantOut; !cmp.Equal(got, want) {
				t.Errorf("Asker.Secret(...) output got %q, want %q", got, want)
			}
		})
	}
}

func TestAsker_Line_Cancelled(t *testing.T) {
	t.Parallel()

	// Arrange
	out := &bytes.Buffer{}
	sut := &ask.Asker{Out: out, In: blockingReader{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	line, err := sut.Line(ctx, "Name: ")

	// Assert
	if got, want := err, context.Canceled; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Asker.Line(...) error got %v, want %v", got, want)
	}
	if got, want := line, ""; !cmp.Equal(got, want) {
		t.Errorf("Asker.Line(...) got %q, want %q", got, want)
	}
	if got, want := out.String(), "Name: \n"; !cmp.Equal(got, want) {
		t.Errorf("Asker.Line(...) output got %q, want %q", got, want)
	}
}

func TestAsker_Secret_CancelledRestoresEcho(t *testing.T) {
	t.Parallel()

	// Arrange
	out := &bytes.Buffer{}
	recorder := &termtest.Recorder{}
	sut := &ask.Asker{
		Out:          out,
		In:           blockingReader{},
		EchoDisabler: recorder,
		HiddenChar:   '*',
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	secret, err := sut.Secret(ctx, "")

	// Assert
	if got, want := err, context.Canceled; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Asker.Secret(...) error got %v, want %v", got, want)
	}
	if got, want := secret, ""; !cmp.Equal(got, want) {
		t.Errorf("Asker.Secret(...) got %q, want %q", got, want)
	}
	if got, want := recorder.Disabled, false; !cmp.Equal(got, want) {
		t.Errorf("echo disabled after cancel got %v, want %v (echo not restored)", got, want)
	}
	if got, want := out.String(), "\r\n"; !cmp.Equal(got, want) {
		t.Errorf("Asker.Secret(...) output got %q, want %q", got, want)
	}
}
