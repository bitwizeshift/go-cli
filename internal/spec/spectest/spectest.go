package spectest

import (
	"context"

	"github.com/bitwizeshift/go-cli/internal/spec"
)

func PassThroughBuilder(runner spec.Runner) spec.Builder {
	return &passThroughBuilder{Runner: runner}
}

type passThroughBuilder struct {
	Runner spec.Runner
}

func (b *passThroughBuilder) Build(ctx context.Context) (spec.Runner, error) {
	return b.Runner, nil
}

func ErrBuilder(err error) spec.Builder {
	return builderFunc(func(context.Context) (spec.Runner, error) {
		return nil, err
	})
}

type builderFunc func(ctx context.Context) (spec.Runner, error)

func (bf builderFunc) Build(ctx context.Context) (spec.Runner, error) {
	return bf(ctx)
}

var _ spec.Builder = (*builderFunc)(nil)

// NoOpRunner returns a [spec.Runner] that succeeds without doing anything.
func NoOpRunner() spec.Runner {
	return runner(func(context.Context) error {
		return nil
	})
}

// Err returns a [spec.Runner] that always returns err.
func Err(err error) spec.Runner {
	return runner(func(context.Context) error {
		return err
	})
}

// UsageRunner returns a [spec.Runner] that always returns [spec.ErrUsage].
func UsageRunner() spec.Runner {
	return Err(spec.ErrUsage)
}

// PanicRunner returns a [spec.Runner] that always panics with value.
func PanicRunner(value any) spec.Runner {
	return runner(func(context.Context) error {
		panic(value)
	})
}

// runner is a behavior-backed [spec.Runner] used to build the doubles above.
type runner func(ctx context.Context) error

// Run invokes the underlying behavior.
func (r runner) Run(ctx context.Context) error {
	return r(ctx)
}
