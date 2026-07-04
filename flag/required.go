package flag

import (
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/spf13/pflag"
)

// MarkRequired marks that all of the specified flags must be required when
// parsing command lines.
func MarkRequired(flags ...*pflag.Flag) {
	annotation.MarkRequired(flags...)
}

// MarkRequiredTogether marks that all flags must be specified together when any
// one flag is specified. Note that this does not mean that all flags are always
// required; it's all or none. If all flags are always required, then
// [MarkRequired] should be used.
func MarkRequiredTogether(flags ...*pflag.Flag) {
	annotation.MarkRequiredTogether(flags...)
}

// MarkMutuallyExclusive marks that all flags must be mutually exclusive with
// each other, and will generate an error when parsing flags that have both set.
func MarkMutuallyExclusive(flags ...*pflag.Flag) {
	annotation.MarkMutuallyExclusive(flags...)
}

// MarkOneRequired marks that at least one of the specified flags is required
// when parsing command lines.
func MarkOneRequired(flags ...*pflag.Flag) {
	annotation.MarkOneRequired(flags...)
}
