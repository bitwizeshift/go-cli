package annotation

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AnnotationCompletion is the pflag annotation that records the identifier of a
// flag's registered completion function.
const AnnotationCompletion = "annotation://cli.flag_completion"

// CompletionDirective instructs the shell how to treat the candidates returned
// by a [CompletionFunc]. It is a cobra-free mirror of the relevant subset of
// cobra's shell completion directives.
type CompletionDirective int

const (
	// CompletionDefault lets the shell perform its default file completion.
	CompletionDefault CompletionDirective = iota

	// CompletionNoFileComp offers only the returned candidates, suppressing the
	// shell's file completion.
	CompletionNoFileComp

	// CompletionFilterFileExt treats the returned candidates as file extensions
	// and completes files matching them.
	CompletionFilterFileExt

	// CompletionFilterDirs completes directory names only.
	CompletionFilterDirs
)

// CompletionFunc computes completion candidates for the partial word toComplete,
// returning the candidates and a [CompletionDirective] describing how the shell
// should treat them.
type CompletionFunc = func(toComplete string) ([]string, CompletionDirective)

// completionMu guards completionFuncs and completionCurrent. completionFuncs is
// the process-wide registry of completion functions, keyed by the identifier
// stored in a flag's [AnnotationCompletion] annotation; completionCurrent is the
// counter used to mint those identifiers.
var (
	completionMu      sync.RWMutex
	completionCurrent int
	completionFuncs   = map[string]CompletionFunc{}
)

// AddCompletion registers fn as the completion function for f, as consumed by
// [RegisterFlagCompletions]. The function is held in a process-wide registry; a
// best-effort [runtime.AddCleanup] removes it if f is collected, though this
// cleanup may never run.
func AddCompletion(f *pflag.Flag, fn CompletionFunc) {
	completionMu.Lock()
	id := fmt.Sprintf("%d", completionCurrent)
	completionCurrent++
	completionFuncs[id] = fn
	completionMu.Unlock()
	setAnnotation(f, AnnotationCompletion, id)

	cleanup := func(id string) {
		completionMu.Lock()
		delete(completionFuncs, id)
		completionMu.Unlock()
	}
	runtime.AddCleanup(f, cleanup, id)
}

// GetCompletionFunc returns the completion function registered on f via
// [AddCompletion], or nil if f has none.
func GetCompletionFunc(f *pflag.Flag) CompletionFunc {
	fn, ok := lookupCompletion(f)
	if !ok {
		return nil
	}
	return fn
}

// RegisterFlagCompletions walks the flags of cmd and registers a cobra
// completion function for each flag annotated via [AddCompletion], translating
// its [CompletionDirective] into the corresponding [cobra.ShellCompDirective].
func RegisterFlagCompletions(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		fn, ok := lookupCompletion(f)
		if !ok {
			return
		}
		// The flag is guaranteed to exist and to have no prior completion, so
		// this error cannot occur.
		_ = cmd.RegisterFlagCompletionFunc(f.Name, func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			values, directive := fn(toComplete)
			return values, cobraDirective(directive)
		})
	})
}

// lookupCompletion returns the completion function recorded on f via
// [AddCompletion], reporting whether one was present.
func lookupCompletion(f *pflag.Flag) (CompletionFunc, bool) {
	ids := f.Annotations[AnnotationCompletion]
	if len(ids) == 0 {
		return nil, false
	}
	completionMu.RLock()
	fn, ok := completionFuncs[ids[0]]
	completionMu.RUnlock()
	return fn, ok
}

// cobraDirective maps a [CompletionDirective] onto its [cobra.ShellCompDirective]
// equivalent.
func cobraDirective(directive CompletionDirective) cobra.ShellCompDirective {
	switch directive {
	case CompletionNoFileComp:
		return cobra.ShellCompDirectiveNoFileComp
	case CompletionFilterFileExt:
		return cobra.ShellCompDirectiveFilterFileExt
	case CompletionFilterDirs:
		return cobra.ShellCompDirectiveFilterDirs
	default:
		return cobra.ShellCompDirectiveDefault
	}
}
