package exit_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"testing"

	"github.com/bitwizeshift/go-cli/exit"
	"github.com/google/go-cmp/cmp"
)

// constantClassifier returns a [exit.Classifier] that always classifies errors
// as code.
func constantClassifier(code exit.Code) exit.Classifier {
	return exit.ClassifierFunc(func(error) exit.Code {
		return code
	})
}

// runtimeError returns a [runtime.Error] produced by a failed type-assertion.
func runtimeError() (err error) {
	defer func() {
		err, _ = recover().(error)
	}()
	var value any = "not an int"
	_ = value.(int)
	return nil
}

func TestPOSIXClassifier(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		err  error
		want exit.Code
	}{
		{
			name: "NilError",
			err:  nil,
			want: exit.CodeSuccess,
		}, {
			name: "UnrecognizedError",
			err:  errors.New("something went wrong"),
			want: exit.CodeUnknown,
		}, {
			name: "FileNotExist",
			err: &fs.PathError{
				Op:   "open",
				Path: "input.txt",
				Err:  fs.ErrNotExist,
			},
			want: exit.CodeNoInput,
		}, {
			name: "FileAlreadyExists",
			err:  fmt.Errorf("create: %w", fs.ErrExist),
			want: exit.CodeCanNotCreate,
		}, {
			name: "PermissionDenied",
			err:  fmt.Errorf("open: %w", fs.ErrPermission),
			want: exit.CodeNoPerm,
		}, {
			name: "NumberSyntax",
			err: &strconv.NumError{
				Func: "Atoi",
				Num:  "twelve",
				Err:  strconv.ErrSyntax,
			},
			want: exit.CodeDataErr,
		}, {
			name: "NumberRange",
			err: &strconv.NumError{
				Func: "Atoi",
				Num:  "99999999999999999999",
				Err:  strconv.ErrRange,
			},
			want: exit.CodeDataErr,
		}, {
			name: "ContextCanceled",
			err:  fmt.Errorf("fetch: %w", context.Canceled),
			want: exit.CodeTempFail,
		}, {
			name: "ContextDeadlineExceeded",
			err:  context.DeadlineExceeded,
			want: exit.CodeTempFail,
		}, {
			name: "UnexpectedEOF",
			err:  fmt.Errorf("decode: %w", io.ErrUnexpectedEOF),
			want: exit.CodeIOErr,
		}, {
			name: "ExecutableNotFound",
			err: &exec.Error{
				Name: "cc",
				Err:  exec.ErrNotFound,
			},
			want: exit.CodeUnavailable,
		}, {
			name: "RuntimeError",
			err:  runtimeError(),
			want: exit.CodeSoftware,
		}, {
			name: "UnknownUser",
			err:  user.UnknownUserError("nobody"),
			want: exit.CodeNoUser,
		}, {
			name: "UnknownUserID",
			err:  user.UnknownUserIdError(1234),
			want: exit.CodeNoUser,
		}, {
			name: "DNSError",
			err: &net.DNSError{
				Err:  "no such host",
				Name: "example.invalid",
			},
			want: exit.CodeNoHost,
		}, {
			name: "NetworkError",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: errors.New("connection refused"),
			},
			want: exit.CodeUnavailable,
		}, {
			name: "AddressError",
			err: &net.AddrError{
				Err:  "missing port in address",
				Addr: "localhost",
			},
			want: exit.CodeUnavailable,
		}, {
			name: "JSONSyntaxError",
			err:  &json.SyntaxError{Offset: 4},
			want: exit.CodeDataErr,
		}, {
			name: "JSONTypeError",
			err: &json.UnmarshalTypeError{
				Value: "string",
				Field: "count",
			},
			want: exit.CodeDataErr,
		}, {
			name: "SystemError",
			err:  fmt.Errorf("mmap: %w", syscall.ENOMEM),
			want: exit.CodeOSErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := exit.POSIXClassifier

			// Act
			code := sut.ClassifyError(tc.err)

			// Assert
			if got, want := code, tc.want; !cmp.Equal(got, want) {
				t.Errorf("POSIXClassifier.ClassifyError(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestFallbackClassifier(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		sut  exit.FallbackClassifier
		err  error
		want exit.Code
	}{
		{
			name: "NoClassifiers",
			sut:  nil,
			err:  errors.New("something went wrong"),
			want: exit.CodeUnknown,
		}, {
			name: "NoClassifierRecognizesError",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeUnknown),
				constantClassifier(exit.CodeUnknown),
			},
			err:  errors.New("something went wrong"),
			want: exit.CodeUnknown,
		}, {
			name: "FirstClassifierRecognizesError",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeConfig),
				constantClassifier(exit.CodeIOErr),
			},
			err:  errors.New("something went wrong"),
			want: exit.CodeConfig,
		}, {
			name: "LaterClassifierRecognizesError",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeUnknown),
				constantClassifier(exit.CodeIOErr),
				constantClassifier(exit.CodeConfig),
			},
			err:  errors.New("something went wrong"),
			want: exit.CodeIOErr,
		}, {
			name: "ClassifierYieldsSuccess",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeSuccess),
				constantClassifier(exit.CodeIOErr),
			},
			err:  errors.New("something went wrong"),
			want: exit.CodeSuccess,
		}, {
			name: "POSIXClassifierRecognizesError",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeUnknown),
				exit.POSIXClassifier,
			},
			err:  fmt.Errorf("open: %w", fs.ErrNotExist),
			want: exit.CodeNoInput,
		}, {
			name: "PriorClassifierOverridesPOSIXClassifier",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeConfig),
				exit.POSIXClassifier,
			},
			err:  fmt.Errorf("open: %w", fs.ErrNotExist),
			want: exit.CodeConfig,
		}, {
			name: "POSIXClassifierDoesNotRecognizeError",
			sut: exit.FallbackClassifier{
				constantClassifier(exit.CodeUnknown),
				exit.POSIXClassifier,
			},
			err:  errors.New("something went wrong"),
			want: exit.CodeUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.sut

			// Act
			code := sut.ClassifyError(tc.err)

			// Assert
			if got, want := code, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FallbackClassifier.ClassifyError(...) = %v, want %v", got, want)
			}
		})
	}
}
