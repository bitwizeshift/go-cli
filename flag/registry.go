package flag

import (
	"github.com/bitwizeshift/go-cli/internal/flagreg"
	"github.com/spf13/pflag"
)

// Registry is the opaque flag destination threaded through registration. It wraps
// a [pflag.Registry] and records which [Registrar] instances have already been
// registered so that a shared instance registers its flags only once. [Add] is
// its only mutator.
type Registry flagreg.Registry

// FlagSet is an escape-hatch to allow direct access to the underlying
// [flag.FlagSet]. In general, this should _only_ be used as a transitional
// mechanism -- but should otherwise aim to avoid relying on this.
func (r *Registry) FlagSet() *pflag.FlagSet {
	return flagreg.Flags((*flagreg.Registry)(r))
}

// Flags returns every [Flag] registered in this registry, in lexical order.
func (r *Registry) Flags() []*Flag {
	return flagsOf(r.FlagSet())
}
