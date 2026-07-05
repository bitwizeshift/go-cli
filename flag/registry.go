package flag

import (
	"unsafe"

	"github.com/spf13/pflag"
)

// Registry is the opaque flag destination threaded through registration. It wraps
// a [pflag.Registry] and records which [Registrar] instances have already been
// registered so that a shared instance registers its flags only once. [Add] and
// [AddCallback] are its only mutators.
type Registry struct {
	flags   *pflag.FlagSet
	visited map[unsafe.Pointer]struct{}
}

// NewRegistry wraps flags so that it can be passed to [Add], [AddCallback], and
// [Register].
func NewRegistry(flags *pflag.FlagSet) *Registry {
	return &Registry{flags: flags, visited: map[unsafe.Pointer]struct{}{}}
}
