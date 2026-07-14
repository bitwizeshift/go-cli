package flagreg

import (
	"unsafe"

	"github.com/spf13/pflag"
)

// Registry is the opaque flag destination threaded through registration.
//
// This type forms the base type for the proper [flag.Registry]
type Registry struct {
	flags   *pflag.FlagSet
	visited map[unsafe.Pointer]struct{}
}

// New returns a newly constructed [Registry]. This is to enable creating
// registries for testing purposes.
func New() *Registry {
	flags := pflag.NewFlagSet("registry", pflag.ContinueOnError)
	return FromFlagSet(flags)
}

// FromFlagSet constructs a [Registry] from a [pflag.FlagSet]. This is used in
// the real CLI construction.
func FromFlagSet(flags *pflag.FlagSet) *Registry {
	return &Registry{flags: flags, visited: map[unsafe.Pointer]struct{}{}}
}

// The below functions exist so that other exported packages are able to access
// unexported fields in their implementation.

// Flags returns the flags for this registry.
func Flags(reg *Registry) *pflag.FlagSet {
	return reg.flags
}

// Visited returns the map for visited entries for this registry.
func Visited(reg *Registry) map[unsafe.Pointer]struct{} {
	return reg.visited
}
