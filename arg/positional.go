package arg

import (
	"reflect"

	"github.com/bitwizeshift/go-cli/internal/argreg"
)

// PositionalArg is a positional-argument binding produced by [Positional]. It is
// registered on a [CommandLine] with [CommandLine.Add].
type PositionalArg struct {
	positional *argreg.Positional
}

// Positional constructs a positional argument at index bound to v, decoding it
// the same way as [Flag]. name is the label shown for the argument in help
// output. The returned [PositionalArg] is registered on a [CommandLine] with
// [CommandLine.Add].
//
// Positional arguments are drawn from the command line after flags are parsed.
// If no argument occupies index when the command runs, v is left unchanged; the
// command's arity governs how many arguments are permitted.
//
// By default the value is decoded with [Unmarshal] and reports a kebab-case type
// name derived from T; both may be adjusted with [Option] values. Options that
// concern flags alone, such as [Shorthand] or [Repeatable], have no effect.
func Positional[T any](name string, index int, v *T, options ...Option) *PositionalArg {
	cfg := newConfig(options...)
	return &PositionalArg{positional: &argreg.Positional{
		Index: index,
		Name:  name,
		Type:  cfg.typeName(v),
		Usage: cfg.usage,
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
	argreg.AddPositional((*argreg.CommandLine)(cl), p.positional)
}

var _ Arg = (*PositionalArg)(nil)
