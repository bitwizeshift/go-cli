package spec

import "context"

// Builder constructs a [Runner] to use for execution.
type Builder interface {
	// Build takes the application context and attempts to construct a [Runner].
	Build(ctx context.Context) (Runner, error)
}

// Runner is the generalized "run" behavior that each bound command executes.
type Runner interface {
	// Run executes the command with the resolved positional arguments.
	Run(ctx context.Context) error
}
