package arg

import (
	"reflect"

	"github.com/bitwizeshift/go-cli/internal/argdef"
)

// UnmatchedArg is an unmatched-argument binding produced by [Unmatched]. It is
// registered on a [CommandLine] with [CommandLine.Add].
type UnmatchedArg struct {
	unmatched *argdef.Unmatched
}

// Unmatched constructs a binding for every positional argument not claimed by a
// [Positional], decoding each one the same way as [Flag] and assigning them to
// out in command-line order when the command runs. The returned [UnmatchedArg]
// is registered on a [CommandLine] with [CommandLine.Add].
//
// By default each argument is decoded with [Unmarshal] and the set reports a
// kebab-case type name derived from T; both may be adjusted with [Option]
// values. A [DefaultFromEnv] or [DefaultFromFunc] fallback supplies the whole
// set as comma-separated fields, and so applies only when no argument went
// unclaimed. out is left unchanged if any argument fails to decode.
func Unmatched[T any](name string, out *[]T, options ...Option) *UnmatchedArg {
	cfg := newConfig(options...)
	fallbackFuncs := make([]argdef.FallbackFunc, 0, len(cfg.custom))
	for _, f := range cfg.custom {
		fallbackFuncs = append(fallbackFuncs, f)
	}
	return &UnmatchedArg{unmatched: &argdef.Unmatched{
		Name:          name,
		Type:          cfg.typeName(new(T)),
		Usage:         cfg.usage,
		Required:      cfg.required,
		Complete:      cfg.completer,
		EnvFallbacks:  cfg.envs,
		FuncFallbacks: fallbackFuncs,
		Set: func(values []string) error {
			result := make([]T, 0, len(values))
			for _, value := range values {
				var tmp T
				if err := cfg.set(&tmp, []byte(value)); err != nil {
					return err
				}
				for _, cb := range cfg.callbacks {
					if err := invokeCallback(cb, reflect.ValueOf(tmp)); err != nil {
						return err
					}
				}
				result = append(result, tmp)
			}
			*out = result
			return nil
		},
	}}
}

// register records the unmatched-argument binding on cl. It panics if cl
// already carries one.
func (u *UnmatchedArg) register(cl *CommandLine) {
	argdef.SetUnmatched((*argdef.CommandLine)(cl), u.unmatched)
}

var _ Arg = (*UnmatchedArg)(nil)
