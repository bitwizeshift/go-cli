package arg

import (
	"reflect"

	"github.com/bitwizeshift/go-cli/internal/argdef"
)

// PositionalArg is a positional-argument binding produced by [Positional]. It is
// registered on a [CommandLine] with [CommandLine.Add].
type PositionalArg struct {
	positional *argdef.Positional
}

// Positional constructs a positional argument at index bound to v, decoding it
// the same way as [Flag]. name is the label shown for the argument in help
// output. The returned [PositionalArg] is registered on a [CommandLine] with
// [CommandLine.Add].
//
// Positional arguments are drawn from the command line after flags are parsed.
// If no argument occupies index when the command runs, v is left unchanged.
// Marking the argument [Required] instead demands that the command line reach
// index, so v is always assigned before the command runs.
//
// By default the value is decoded with [Unmarshal] and reports a kebab-case type
// name derived from T; both may be adjusted with [Option] values.
func Positional[T any](name string, index int, v *T, options ...Option) *PositionalArg {
	cfg := newConfig(options...)
	fallbackFuncs := make([]argdef.FallbackFunc, 0, len(cfg.custom))
	for _, f := range cfg.custom {
		fallbackFuncs = append(fallbackFuncs, f)
	}
	return &PositionalArg{positional: &argdef.Positional{
		Index:         index,
		Name:          name,
		Type:          cfg.typeName(v),
		Usage:         cfg.usage,
		Required:      cfg.required,
		Complete:      cfg.completer,
		EnvFallbacks:  cfg.envs,
		FuncFallbacks: fallbackFuncs,
		Set: func(s string) error {
			var tmp T
			if err := cfg.set(&tmp, []byte(s)); err != nil {
				return err
			}
			*v = tmp
			for _, cb := range cfg.callbacks {
				if err := invokeCallback(cb, reflect.ValueOf(tmp)); err != nil {
					return err
				}
			}
			return nil
		},
	}}
}

// register records the positional-argument binding on cl.
func (p *PositionalArg) register(cl *CommandLine) {
	argdef.AddPositional((*argdef.CommandLine)(cl), p.positional)
}

var _ Arg = (*PositionalArg)(nil)
