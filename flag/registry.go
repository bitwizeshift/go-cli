package flag

import (
	"github.com/bitwizeshift/go-cli/internal/flagreg"
)

// Registry is the opaque flag destination threaded through registration. It wraps
// a [pflag.Registry] and records which [Registrar] instances have already been
// registered so that a shared instance registers its flags only once. [Add] is
// its only mutator.
type Registry flagreg.Registry
