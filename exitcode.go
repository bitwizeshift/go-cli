package cli

import "os"

// ExitCode is a process exit status produced by running a [CLI].
type ExitCode int

const (
	// ExitSuccess indicates the command completed without error.
	ExitSuccess ExitCode = 0

	// ExitError indicates the command returned an error.
	ExitError ExitCode = 1

	// ExitPanic indicates the command terminated by recovering from a panic.
	ExitPanic ExitCode = 2
)

// Exit terminates the process with the receiver's status via [os.Exit] and does
// not return.
func (e ExitCode) Exit() {
	os.Exit(int(e))
}
