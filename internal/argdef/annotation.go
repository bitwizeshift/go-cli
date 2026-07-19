package argdef

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// AnnotationRequired is the pflag annotation for setting a flag as required
	AnnotationRequired = "annotation://cli.required"

	// AnnotationRequiredTogether is the pflag annotation for marking flags as
	// being required together
	AnnotationRequiredTogether = "annotation://cli.required_together"

	// AnnotationMutuallyExclusive is the pflag annotation for marking flags as
	// being mutually exclusive
	AnnotationMutuallyExclusive = "annotation://cli.mutually_exclusive"

	// AnnotationOneRequired is the pflag annotation for marking a single flag as
	// being required
	AnnotationOneRequired = "annotation://cli.one_required"

	// AnnotationFlagGroup is an annotation for grouping flags into a named
	// group
	AnnotationFlagGroup = "annotation://cli.flag_group"

	// AnnotationENVFallback is an annotation for assigning an ENV variable that
	// can be used as a fallback value for flags.
	AnnotationENVFallback = "annnotation://cli.flag_env_fallback"

	// AnnotationFuncFallback is an annotation for assigning a func to compute
	// a flag's default value.
	AnnotationFuncFallback = "annotation://cli.flag_func_fallback"

	// AnnotationIssueURL is the pflag annotation for assigning a single flag's
	// issue URL for filing bugs.
	AnnotationIssueURL = "annotation://cli.cmd_issue_url"
)

// groupSeparator joins the members of a constraint group into a single stable
// annotation value. A space is safe because [pflag] flag names may not contain
// spaces.
const groupSeparator = "\x00"

// MarkRequired sets each flag in 'flags' to be required by assigning the
// [AnnotationRequired] annotation.
func MarkRequired(flags ...*pflag.Flag) {
	for _, flag := range flags {
		setAnnotation(flag, AnnotationRequired, "true")
	}
}

// MarkRequiredTogether sets each flag in 'flags' to be required together when
// specified by assigning the [AnnotationRequiredTogether] annotation.
func MarkRequiredTogether(flags ...*pflag.Flag) {
	markGroup(AnnotationRequiredTogether, flags)
}

// MarkMutuallyExclusive sets each flag in 'flags' to be mutually exclusive from
// one another by assigning the [AnnotationMutuallyExclusive] annotation.
func MarkMutuallyExclusive(flags ...*pflag.Flag) {
	markGroup(AnnotationMutuallyExclusive, flags)
}

// MarkOneRequired sets each flag in 'flags' to require at least one of them to
// be specified at a time, by assigning the [AnnotationOneRequired] annotation.
func MarkOneRequired(flags ...*pflag.Flag) {
	markGroup(AnnotationOneRequired, flags)
}

// AddToGroup adds all specified flags to be part of the same named flag group
// by assigning the [AnnotationFlagGroup] to each flag
func AddToGroup(name string, flags ...*pflag.Flag) {
	for _, flag := range flags {
		setAnnotation(flag, AnnotationFlagGroup, name)
	}
}

// Group returns the name of the group the flag is part of, if it was specified.
// Otherwise returns an empty string.
func Group(f *pflag.Flag) string {
	return strings.Join(f.Annotations[AnnotationFlagGroup], " ")
}

// IsRequired reports whether f has been marked required via [MarkRequired].
func IsRequired(f *pflag.Flag) bool {
	_, ok := f.Annotations[AnnotationRequired]
	return ok
}

// RequiredTogether returns the sorted, de-duplicated set of flag names that f
// is required together with, as recorded by [MarkRequiredTogether]. The result
// includes f itself. It is empty if f participates in no such group.
func RequiredTogether(f *pflag.Flag) []string {
	return groupMembers(f, AnnotationRequiredTogether)
}

// MutuallyExclusive returns the sorted, de-duplicated set of flag names that f
// is mutually exclusive with, as recorded by [MarkMutuallyExclusive]. The
// result includes f itself. It is empty if f participates in no such group.
func MutuallyExclusive(f *pflag.Flag) []string {
	return groupMembers(f, AnnotationMutuallyExclusive)
}

// OneRequired returns the sorted, de-duplicated set of flag names in which at
// least one is required, as recorded by [MarkOneRequired]. The result includes
// f itself. It is empty if f participates in no such group.
func OneRequired(f *pflag.Flag) []string {
	return groupMembers(f, AnnotationOneRequired)
}

// ConfigureFlags walks through all flags to convert our local annotations into
// cobra's mechanism for upholding the flag requirements.
func ConfigureFlags(cmd *cobra.Command) {
	requiredTogether := map[string][]string{}
	mutuallyExclusive := map[string][]string{}
	oneRequired := map[string][]string{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if IsRequired(f) {
			// The flag is guaranteed to exist since it is being visited from
			// this same set, so this error cannot occur.
			_ = cmd.MarkFlagRequired(f.Name)
		}
		collectGroups(f, AnnotationRequiredTogether, requiredTogether)
		collectGroups(f, AnnotationMutuallyExclusive, mutuallyExclusive)
		collectGroups(f, AnnotationOneRequired, oneRequired)
	})

	for _, names := range requiredTogether {
		cmd.MarkFlagsRequiredTogether(names...)
	}
	for _, names := range mutuallyExclusive {
		cmd.MarkFlagsMutuallyExclusive(names...)
	}
	for _, names := range oneRequired {
		cmd.MarkFlagsOneRequired(names...)
	}
}

// IssueURL retrieves the New Issue URL from the given [cobra.Command]
func IssueURL(cmd *cobra.Command) string {
	return cmd.Annotations[AnnotationIssueURL]
}

// AddIssueURL adds the 'new issue' URL to the given [cobra.Command]. This will
// be assigned to all command objects recursively.
func AddIssueURL(cmd *cobra.Command, issueURL string) {
	addIssueURL(cmd, issueURL, map[*cobra.Command]struct{}{})
}

func addIssueURL(cmd *cobra.Command, issueURL string, visited map[*cobra.Command]struct{}) {
	if _, ok := visited[cmd]; ok {
		return
	}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[AnnotationIssueURL] = issueURL
	visited[cmd] = struct{}{}
	for _, cmd := range cmd.Commands() {
		addIssueURL(cmd, issueURL, visited)
	}
}

// setAnnotation assigns value as the sole value of key on f, initializing the
// annotation map if necessary.
func setAnnotation(f *pflag.Flag, key, value string) {
	if f.Annotations == nil {
		f.Annotations = map[string][]string{}
	}
	f.Annotations[key] = []string{value}
}

// appendAnnotation adds value to the values recorded under key on f,
// initializing the annotation map if necessary.
func appendAnnotation(f *pflag.Flag, key, value string) {
	if f.Annotations == nil {
		f.Annotations = map[string][]string{}
	}
	f.Annotations[key] = append(f.Annotations[key], value)
}

// markGroup records the membership of flags as a single stable group under key
// on every member.
func markGroup(key string, flags []*pflag.Flag) {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, flag.Name)
	}
	slices.Sort(names)
	group := strings.Join(names, groupSeparator)

	for _, flag := range flags {
		if flag.Annotations == nil {
			flag.Annotations = map[string][]string{}
		}
		flag.Annotations[key] = append(flag.Annotations[key], group)
	}
}

// groupMembers flattens every group recorded under key on f into a sorted,
// de-duplicated list of member names.
func groupMembers(f *pflag.Flag, key string) []string {
	var members []string
	for _, group := range f.Annotations[key] {
		members = append(members, strings.Split(group, groupSeparator)...)
	}
	slices.Sort(members)
	return slices.Compact(members)
}

// collectGroups records each group string recorded under key on f into groups,
// keyed by the group string so that identical groups shared across members are
// de-duplicated.
func collectGroups(f *pflag.Flag, key string, groups map[string][]string) {
	for _, group := range f.Annotations[key] {
		groups[group] = strings.Split(group, groupSeparator)
	}
}

// mu guards flagFuncs and current. flagFuncs is the process-wide registry of
// fallback functions, keyed by the identifier stored in a flag's
// [AnnotationFuncFallback] annotation; current is the counter used to mint
// those identifiers.
var (
	mu        sync.RWMutex
	current   int
	flagFuncs = map[string]FallbackFunc{}
)

// AddEnvFallback records env as an environment variable that may source a
// fallback default for f, as consumed by [SetFlagFallbacks].
func AddEnvFallback(f *pflag.Flag, env string) {
	appendAnnotation(f, AnnotationENVFallback, env)
}

// AddFuncFallback registers fallback as a source of a fallback default for f,
// as consumed by [SetFlagFallbacks]. The function is held in a process-wide
// registry; a best-effort [runtime.AddCleanup] removes it if f is collected,
// though this cleanup may never run.
func AddFuncFallback(f *pflag.Flag, fallback FallbackFunc) {
	mu.Lock()
	id := fmt.Sprintf("%d", current)
	current++
	flagFuncs[id] = fallback
	mu.Unlock()
	appendAnnotation(f, AnnotationFuncFallback, id)

	// Register a cleanup function in case the flag is deleted.
	// This _may never run_ in practice, as is documented by [runtime.AddCleanup],
	// but is done as a "best-effort" to reduce storage space of global values
	cleanup := func(id string) {
		mu.Lock()
		delete(flagFuncs, id)
		mu.Unlock()
	}
	runtime.AddCleanup(f, cleanup, id)
}

// SetFlagFallbacks goes through all unset flags and sets fallback values that
// come from either the Environment, or through fallback functions.
//
// Failures encountered while setting fallbacks are joined and returned together.
// Every flag is visited and computed in this process.
func SetFlagFallbacks(ctx context.Context, fs *pflag.FlagSet) error {
	var errs []error
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			return
		}
		visited, err := runEnvFlagFallback(f)
		errs = append(errs, err)
		if visited {
			return
		}
		_, err = runFuncFlagFallback(ctx, f)
		errs = append(errs, err)
	})
	return errors.Join(errs...)
}

// runEnvFlagFallback assigns f the value of the first present environment
// variable recorded on f via [AddEnvFallback]. It reports whether a variable
// was present, and wraps [ErrSettingEnvFlag] if assignment failed.
func runEnvFlagFallback(f *pflag.Flag) (visited bool, err error) {
	for _, key := range f.Annotations[AnnotationENVFallback] {
		if value, exists := os.LookupEnv(key); exists {
			err = f.Value.Set(value)
			visited = true
			if err != nil {
				err = fmt.Errorf("%w: $%v: %w", ErrSettingEnvFlag, key, err)
			}
			return
		}
	}
	return false, nil
}

// runFuncFlagFallback assigns f the first non-empty value produced by a
// fallback function recorded on f via [AddFuncFallback]. It reports whether a
// function supplied a value, and wraps [ErrComputingFuncFlag] if a function
// errored or [ErrSettingFuncFlag] if assignment failed.
func runFuncFlagFallback(ctx context.Context, f *pflag.Flag) (visited bool, err error) {
	for _, key := range f.Annotations[AnnotationFuncFallback] {
		mu.RLock()
		fn, ok := flagFuncs[key]
		mu.RUnlock()
		if !ok {
			continue
		}
		value, ferr := fn(ctx)
		if ferr != nil {
			return true, fmt.Errorf("%w: %w", ErrComputingFuncFlag, ferr)
		}
		if value == "" {
			continue
		}
		err = f.Value.Set(value)
		visited = true
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrSettingFuncFlag, err)
		}
		return
	}
	return false, nil
}
