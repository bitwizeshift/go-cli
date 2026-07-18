package arg

import "github.com/bitwizeshift/go-cli/internal/argreg"

// UnmatchedArg is an unmatched-argument binding produced by [Unmatched]. It is
// registered on a [CommandLine] with [CommandLine.Add].
type UnmatchedArg struct {
	unmatched *argreg.Unmatched
}

// Unmatched constructs a binding for every positional argument not claimed by a
// [Positional], assigning them to out in command-line order when the command
// runs. The returned [UnmatchedArg] is registered on a [CommandLine] with
// [CommandLine.Add].
func Unmatched(out *[]string) *UnmatchedArg {
	return &UnmatchedArg{unmatched: &argreg.Unmatched{
		Set: func(values []string) error {
			*out = values
			return nil
		},
	}}
}

// register records the unmatched-argument binding on cl, replacing any binding
// previously added.
func (u *UnmatchedArg) register(cl *CommandLine) {
	argreg.SetUnmatched((*argreg.CommandLine)(cl), u.unmatched)
}

var _ Arg = (*UnmatchedArg)(nil)
