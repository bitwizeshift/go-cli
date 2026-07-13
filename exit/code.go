package exit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
)

// Code is a process exit status produced by running a [cli.CLI].
//
// The pre-defined codes mirror the POSIX-style conventions from sysexits.h,
// which reserve the range [64,78] for well-known failure classes. Codes 0 and
// [1,63] carry no sysexits.h meaning; 0 signals success, and the rest are fair
// game for application-specific use.
//
// Codes at or above 126 are reserved by the shell — 126 and 127 report that a
// command could not be executed or found, and 128+N reports termination by
// signal N — so applications should not produce them directly.
type Code int

const (
	// CodeUnknown indicates a [Classifier] could not classify an error, and that
	// classification should defer to another [Classifier].
	//
	// It is a classification sentinel, not a status: it is negative, and so is
	// never a valid value to exit a process with.
	CodeUnknown Code = -1

	// CodeSuccess indicates the command completed without error.
	CodeSuccess Code = 0

	// CodeUsage indicates the command was used incorrectly: a bad flag, the
	// wrong number of arguments, or otherwise malformed invocation.
	CodeUsage Code = 64

	// CodeDataErr indicates the user's input data was incorrect: it was
	// well-formed enough to read, but did not satisfy the format the command
	// requires.
	CodeDataErr Code = 65

	// CodeNoInput indicates a named input file did not exist or could not be
	// read.
	CodeNoInput Code = 66

	// CodeNoUser indicates a named user did not exist.
	CodeNoUser Code = 67

	// CodeNoHost indicates a named host did not exist.
	CodeNoHost Code = 68

	// CodeUnavailable indicates a required service is unavailable. This is the
	// catch-all for failures that have no more specific classification.
	CodeUnavailable Code = 69

	// CodeSoftware indicates an internal software error occurred, such as a
	// panic, that is unrelated to the operating system or the user's input.
	CodeSoftware Code = 70

	// CodeOSErr indicates an operating-system error occurred, such as being
	// unable to fork a process.
	CodeOSErr Code = 71

	// CodeOSFile indicates a system file was missing, could not be opened, or
	// contained an error.
	CodeOSFile Code = 72

	// CodeCanNotCreate indicates a user-specified output file could not be
	// created.
	CodeCanNotCreate Code = 73

	// CodeIOErr indicates an error occurred while doing I/O on a file.
	CodeIOErr Code = 74

	// CodeTempFail indicates a temporary failure that is not necessarily an
	// error; the request may succeed if retried later.
	CodeTempFail Code = 75

	// CodeProtocol indicates a remote system returned something that violated
	// the protocol being spoken.
	CodeProtocol Code = 76

	// CodeNoPerm indicates the operation was not permitted for reasons other
	// than a missing file; this is not for filesystem permission errors, which
	// are [CodeNoInput] or [CodeCanNotCreate].
	CodeNoPerm Code = 77

	// CodeConfig indicates something was misconfigured.
	CodeConfig Code = 78
)

// Exit terminates the process with the receiver's status via [os.Exit] and does
// not return.
func (c Code) Exit() {
	os.Exit(int(c))
}

// Classifier classifies an error as the [Code] that a command failing with that
// error should exit with.
//
// A classifier that does not recognize an error returns [CodeUnknown], which
// defers the decision to another classifier; see [FallbackClassifier].
type Classifier interface {
	ClassifyError(error) Code
}

// ClassifierFunc is a simple function that implements [Classifier].
type ClassifierFunc func(error) Code

// ClassifyError forwards the classification to the underlying func
func (cf ClassifierFunc) ClassifyError(err error) Code {
	return cf(err)
}

var _ Classifier = (*ClassifierFunc)(nil)

// FallbackClassifier is a [Classifier] that consults each of its classifiers in
// order, yielding the first result other than [CodeUnknown].
//
// A classifier that yields [CodeUnknown] has no opinion on the error, and so the
// next classifier is consulted. This makes it possible to layer
// application-specific classification on top of [POSIXClassifier].
type FallbackClassifier []Classifier

// ClassifyError returns the first [Code] other than [CodeUnknown] produced by
// the classifiers, or [CodeUnknown] if no classifier recognizes err.
func (fc FallbackClassifier) ClassifyError(err error) Code {
	for _, c := range fc {
		if code := c.ClassifyError(err); code != CodeUnknown {
			return code
		}
	}
	return CodeUnknown
}

var _ Classifier = (FallbackClassifier)(nil)

// POSIXClassifier is a [Classifier] that translates errors produced by the
// standard library into the sysexits.h POSIX [Code] that most closely describes
// them.
var POSIXClassifier = ClassifierFunc(classifyError)

func classifyError(err error) Code {
	if err == nil {
		return CodeSuccess
	}
	if code := classifySentinel(err); code != CodeUnknown {
		return code
	}
	return classifyType(err)
}

// classifySentinel classifies err against the sentinel errors of the standard
// library. These are checked before the error types, since types like
// [fs.PathError] carry a sentinel that describes them more precisely than the
// type does.
func classifySentinel(err error) Code {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return CodeNoInput
	case errors.Is(err, fs.ErrExist):
		return CodeCanNotCreate
	case errors.Is(err, fs.ErrPermission):
		return CodeNoPerm
	case errors.Is(err, strconv.ErrSyntax), errors.Is(err, strconv.ErrRange):
		return CodeDataErr
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, os.ErrDeadlineExceeded):
		return CodeTempFail
	case errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, io.ErrShortWrite),
		errors.Is(err, fs.ErrClosed):
		return CodeIOErr
	case errors.Is(err, exec.ErrNotFound):
		return CodeUnavailable
	}
	return CodeUnknown
}

// classifyType classifies err against the error types of the standard library.
//
// The network errors are checked before [syscall.Errno], since a network error
// commonly wraps the errno that produced it, and since [syscall.Errno] itself
// satisfies [net.Error].
func classifyType(err error) Code {
	switch {
	case isType[runtime.Error](err):
		return CodeSoftware
	case isType[user.UnknownUserError](err), isType[user.UnknownUserIdError](err):
		return CodeNoUser
	case isType[*net.DNSError](err):
		return CodeNoHost
	case isType[*net.OpError](err):
		return CodeUnavailable
	case isType[*json.SyntaxError](err), isType[*json.UnmarshalTypeError](err):
		return CodeDataErr
	case isType[syscall.Errno](err):
		return CodeOSErr
	case isType[net.Error](err):
		return CodeUnavailable
	}
	return CodeUnknown
}

// isType reports whether err's tree contains an error of type E.
func isType[E error](err error) bool {
	_, ok := errors.AsType[E](err)
	return ok
}
