package arg

import (
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/spf13/pflag"
)

// MarkRequired marks that all of the specified flags must be required when
// parsing command lines.
func MarkRequired(flags ...*Flag) {
	annotation.MarkRequired(pflags(flags)...)
}

// MarkRequiredTogether marks that all flags must be specified together when any
// one flag is specified. Note that this does not mean that all flags are always
// required; it's all or none. If all flags are always required, then
// [MarkRequired] should be used.
func MarkRequiredTogether(flags ...*Flag) {
	annotation.MarkRequiredTogether(pflags(flags)...)
}

// MarkMutuallyExclusive marks that all flags must be mutually exclusive with
// each other, and will generate an error when parsing flags that have both set.
func MarkMutuallyExclusive(flags ...*Flag) {
	annotation.MarkMutuallyExclusive(pflags(flags)...)
}

// MarkOneRequired marks that at least one of the specified flags is required
// when parsing command lines.
func MarkOneRequired(flags ...*Flag) {
	annotation.MarkOneRequired(pflags(flags)...)
}

// pflags unwraps flags into the representation the annotations are recorded on.
func pflags(flags []*Flag) []*pflag.Flag {
	result := make([]*pflag.Flag, 0, len(flags))
	for _, f := range flags {
		result = append(result, f.Flag())
	}
	return result
}
