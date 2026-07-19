package completion

import "github.com/spf13/cobra"

// ForPositionals builds the single completion function cobra permits per command
// out of the completers registered against individual argument indices, keyed by
// the index each one completes.
//
// The returned function completes the argument currently being typed, which is
// the argument at the index following those already supplied. An index with no
// registered completer defers to the shell's default file completion, as an
// uncompleted command would. It returns nil when fns holds no completers,
// leaving such a command exactly as cobra found it.
func ForPositionals(fns map[int]Func) cobra.CompletionFunc {
	if len(fns) == 0 {
		return nil
	}
	return func(_ *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		fn, ok := fns[len(args)]
		if !ok {
			return nil, cobra.ShellCompDirectiveDefault
		}
		values, directive := fn(toComplete)
		return values, cobraDirective(directive)
	}
}
