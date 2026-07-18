package arg

import (
	"reflect"

	"github.com/bitwizeshift/go-cli/internal/argreg"
)

// Positional binds the positional argument at index to v, decoding it the same
// way as [AddFlag]. name is the label shown for the argument in help output.
//
// Positional arguments are drawn from the command line after flags are parsed.
// If no argument occupies index when the command runs, v is left unchanged; the
// command's arity governs how many arguments are permitted.
//
// By default the value is decoded with [Unmarshal] and reports a kebab-case type
// name derived from T; both may be adjusted with [Option] values. Options that
// concern flags alone, such as [Shorthand] or [Repeatable], have no effect.
func Positional[T any](cl *CommandLine, name string, index int, v *T, options ...Option) {
	cfg := newConfig(options...)
	argreg.AddPositional((*argreg.CommandLine)(cl), &argreg.Positional{
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
	})
}

// Unmatched binds out to every positional argument not claimed by a [Positional].
// The arguments are assigned in command-line order when the command runs.
func Unmatched(cl *CommandLine, out *[]string) {
	argreg.SetUnmatched((*argreg.CommandLine)(cl), &argreg.Unmatched{
		Set: func(values []string) error {
			*out = values
			return nil
		},
	})
}
