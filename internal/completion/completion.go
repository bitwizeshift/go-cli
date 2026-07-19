package completion

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Annotation is the pflag annotation that records the identifier of a flag's
// registered completion function.
const Annotation = "annotation://cli.flag_completion"

// Directive instructs the shell how to treat the candidates returned by a
// [Func]. It is a cobra-free mirror of the relevant subset of cobra's shell
// completion directives.
type Directive int

const (
	// Default lets the shell perform its default file completion.
	Default Directive = iota

	// NoFileComp offers only the returned candidates, suppressing the shell's
	// file completion.
	NoFileComp

	// FilterFileExt treats the returned candidates as file extensions and
	// completes files matching them.
	FilterFileExt

	// FilterDirs completes directory names only.
	FilterDirs
)

// Func computes completion candidates for the partial word toComplete,
// returning the candidates and a [Directive] describing how the shell should
// treat them.
type Func = func(toComplete string) ([]string, Directive)

// mu guards funcs and current. funcs is the process-wide registry of completion
// functions, keyed by the identifier stored in a flag's [Annotation]; current is
// the counter used to mint those identifiers.
var (
	mu      sync.RWMutex
	current int
	funcs   = map[string]Func{}
)

// AddFlag registers fn as the completion function for f, as consumed by
// [RegisterFlags]. The function is held in a process-wide registry; a
// best-effort [runtime.AddCleanup] removes it if f is collected, though this
// cleanup may never run.
func AddFlag(f *pflag.Flag, fn Func) {
	mu.Lock()
	id := fmt.Sprintf("%d", current)
	current++
	funcs[id] = fn
	mu.Unlock()

	if f.Annotations == nil {
		f.Annotations = map[string][]string{}
	}
	f.Annotations[Annotation] = []string{id}

	cleanup := func(id string) {
		mu.Lock()
		delete(funcs, id)
		mu.Unlock()
	}
	runtime.AddCleanup(f, cleanup, id)
}

// FlagFunc returns the completion function registered on f via [AddFlag], or nil
// if f has none.
func FlagFunc(f *pflag.Flag) Func {
	fn, ok := lookup(f)
	if !ok {
		return nil
	}
	return fn
}

// RegisterFlags walks the flags of cmd and registers a cobra completion function
// for each flag annotated via [AddFlag], translating its [Directive] into the
// corresponding [cobra.ShellCompDirective].
func RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		fn, ok := lookup(f)
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

// lookup returns the completion function recorded on f via [AddFlag], reporting
// whether one was present.
func lookup(f *pflag.Flag) (Func, bool) {
	ids := f.Annotations[Annotation]
	if len(ids) == 0 {
		return nil, false
	}
	mu.RLock()
	fn, ok := funcs[ids[0]]
	mu.RUnlock()
	return fn, ok
}

// cobraDirective maps a [Directive] onto its [cobra.ShellCompDirective]
// equivalent.
func cobraDirective(directive Directive) cobra.ShellCompDirective {
	switch directive {
	case NoFileComp:
		return cobra.ShellCompDirectiveNoFileComp
	case FilterFileExt:
		return cobra.ShellCompDirectiveFilterFileExt
	case FilterDirs:
		return cobra.ShellCompDirectiveFilterDirs
	default:
		return cobra.ShellCompDirectiveDefault
	}
}
