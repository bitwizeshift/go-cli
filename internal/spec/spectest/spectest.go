package spectest

import (
	"context"

	"github.com/bitwizeshift/go-cli/internal/spec"
)

// NoOpRunner returns a [spec.Runner] that succeeds without doing anything.
func NoOpRunner() spec.Runner {
	return runner(func(context.Context, ...string) error {
		return nil
	})
}

// Err returns a [spec.Runner] that always returns err.
func Err(err error) spec.Runner {
	return runner(func(context.Context, ...string) error {
		return err
	})
}

// UsageRunner returns a [spec.Runner] that always returns [spec.ErrUsage].
func UsageRunner() spec.Runner {
	return Err(spec.ErrUsage)
}

// PanicRunner returns a [spec.Runner] that always panics with value.
func PanicRunner(value any) spec.Runner {
	return runner(func(context.Context, ...string) error {
		panic(value)
	})
}

// runner is a behavior-backed [spec.Runner] used to build the doubles above.
type runner func(ctx context.Context, args ...string) error

// Run invokes the underlying behavior.
func (r runner) Run(ctx context.Context, args ...string) error {
	return r(ctx, args...)
}
