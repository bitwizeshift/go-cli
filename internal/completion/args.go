package completion

import "github.com/spf13/cobra"

// ForArgs builds the single completion function cobra permits per command out of
// the completers registered against individual argument indices, keyed by the
// index each one completes, and the completer bound to every unclaimed index.
//
// The returned function completes the argument currently being typed, which is
// the argument at the index following those already supplied. An index with no
// registered completer defers to unmatched, or to the shell's default file
// completion when unmatched is nil, as an uncompleted command would. It returns
// nil when no completer is supplied at all, leaving such a command exactly as
// cobra found it.
func ForArgs(fns map[int]Func, unmatched Func) cobra.CompletionFunc {
	if len(fns) == 0 && unmatched == nil {
		return nil
	}
	return func(_ *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		fn, ok := fns[len(args)]
		if !ok {
			fn = unmatched
		}
		if fn == nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		values, directive := fn(toComplete)
		return values, cobraDirective(directive)
	}
}
