package spec

import "context"

// Runner is the generalized "run" behavior that each bound command executes.
//
// Implementations may additionally implement [github.com/bitwizeshift/go-cli/flag.Registrar],
// or expose reachable fields that do, so that flags are registered automatically
// when the command tree is built.
type Runner interface {
	// Run executes the command with the resolved positional arguments.
	Run(ctx context.Context, args ...string) error
}
