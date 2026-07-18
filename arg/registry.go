package arg

import (
	"github.com/bitwizeshift/go-cli/internal/argreg"
	"github.com/spf13/pflag"
)

// CommandLine is the opaque flag destination threaded through registration. It wraps
// a [pflag.CommandLine] and records which [Registrar] instances have already been
// registered so that a shared instance registers its flags only once. [AddFlag] is
// its only mutator.
type CommandLine argreg.CommandLine

// FlagSet is an escape-hatch to allow direct access to the underlying
// [flag.FlagSet]. In general, this should _only_ be used as a transitional
// mechanism -- but should otherwise aim to avoid relying on this.
func (r *CommandLine) FlagSet() *pflag.FlagSet {
	return argreg.Flags((*argreg.CommandLine)(r))
}

// Flags returns every [Flag] registered in this registry, in lexical order.
func (r *CommandLine) Flags() []*Flag {
	return flagsOf(r.FlagSet())
}
