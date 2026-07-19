package argdef

import (
	"context"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/bitwizeshift/go-cli/internal/completion"
	"github.com/bitwizeshift/go-cli/internal/format/csvfield"
	"github.com/spf13/pflag"
)

var (
	// ErrSettingEnvFlag indicates that a value sourced from an environment
	// fallback could not be assigned to its flag.
	ErrSettingEnvFlag = errors.New("setting flag env default")

	// ErrComputingFuncFlag indicates that a fallback function returned an error
	// while computing a flag's default.
	ErrComputingFuncFlag = errors.New("computing flag custom default")

	// ErrSettingFuncFlag indicates that a value produced by a fallback function
	// could not be assigned to its flag.
	ErrSettingFuncFlag = errors.New("setting flag custom default")
)

// FallbackFunc computes a fallback default for an arg that was not set
// during the CLI invocation, returning the value to assign or an error.
type FallbackFunc = func(ctx context.Context) (string, error)

// CommandLine is the opaque command-line destination threaded through
// registration.
//
// This type forms the base type for the proper [arg.CommandLine]
type CommandLine struct {
	flags       *pflag.FlagSet
	visited     map[unsafe.Pointer]struct{}
	positionals []*Positional
	unmatched   *Unmatched
}

// Positional is a registered positional-argument binding.
type Positional struct {
	Index int
	Name  string
	Type  string
	Usage string
	Set   func(value string) error

	EnvFallbacks  []string
	FuncFallbacks []FallbackFunc

	// Complete offers shell-completion candidates for this argument, or is nil
	// when the argument offers none.
	Complete completion.Func
}

// Unmatched is a registered binding for every argument not claimed by a
// [Positional].
type Unmatched struct {
	Type  string
	Usage string
	Set   func(values []string) error

	EnvFallbacks  []string
	FuncFallbacks []FallbackFunc

	// Complete offers shell-completion candidates for every argument index no
	// [Positional] claims, or is nil when the binding offers none.
	Complete completion.Func
}

// New returns a newly constructed [CommandLine]. This is to enable creating
// registries for testing purposes.
func New() *CommandLine {
	flags := pflag.NewFlagSet("registry", pflag.ContinueOnError)
	return FromFlagSet(flags)
}

// FromFlagSet constructs a [CommandLine] from a [pflag.FlagSet]. This is used in
// the real CLI construction.
func FromFlagSet(flags *pflag.FlagSet) *CommandLine {
	return &CommandLine{flags: flags, visited: map[unsafe.Pointer]struct{}{}}
}

// The below functions exist so that other exported packages are able to access
// unexported fields in their implementation.

// Flags returns the flags for this registry.
func Flags(reg *CommandLine) *pflag.FlagSet {
	return reg.flags
}

// Visited returns the map for visited entries for this registry.
func Visited(reg *CommandLine) map[unsafe.Pointer]struct{} {
	return reg.visited
}

// AddPositional records p as a positional-argument binding on reg.
func AddPositional(reg *CommandLine, p *Positional) {
	reg.positionals = append(reg.positionals, p)
}

// Positionals returns the positional-argument bindings registered on reg, in
// registration order.
func Positionals(reg *CommandLine) []*Positional {
	return reg.positionals
}

// PositionalCompletions returns the completion function of every [Positional]
// registered on reg that has one, keyed by the index it completes. A later
// registration at an index already claimed replaces the earlier one.
func PositionalCompletions(reg *CommandLine) map[int]completion.Func {
	result := map[int]completion.Func{}
	for _, p := range reg.positionals {
		if p.Complete == nil {
			continue
		}
		result[p.Index] = p.Complete
	}
	return result
}

// SetUnmatched records u as the unmatched-argument binding on reg.
//
// It panics if reg already carries an unmatched-argument binding, since the
// arguments a second binding would claim are already spoken for.
func SetUnmatched(reg *CommandLine, u *Unmatched) {
	if reg.unmatched != nil {
		panic("arg: unmatched argument bound more than once")
	}
	reg.unmatched = u
}

// GetUnmatched returns the unmatched-argument binding registered on reg, or nil
// if none was registered.
func GetUnmatched(reg *CommandLine) *Unmatched {
	return reg.unmatched
}

// UnmatchedCompletion returns the completion function of the [Unmatched]
// binding registered on reg, or nil if no binding was registered or it offers
// no completion.
func UnmatchedCompletion(reg *CommandLine) completion.Func {
	if reg.unmatched == nil {
		return nil
	}
	return reg.unmatched.Complete
}

// Bind assigns args to the positional and unmatched bindings registered on reg.
// A [Positional] whose index falls outside args is skipped, leaving its
// destination unchanged. Every argument not claimed by a [Positional] is passed,
// in command-line order, to an [Unmatched] binding.
//
// A binding left without a value falls back to the first of its environment
// variables that is set, then to the first of its fallback functions that
// yields a value. An [Unmatched] binding's fallback value carries the whole set
// as comma-separated fields, so it applies only when no argument went
// unclaimed.
//
// It returns the first error reported by a binding.
func Bind(ctx context.Context, reg *CommandLine, args []string) error {
	claimed := make(map[int]struct{})
	for _, p := range reg.positionals {
		if p.Index < 0 || p.Index >= len(args) {
			if _, err := setFallback(ctx, p.EnvFallbacks, p.FuncFallbacks, p.Set); err != nil {
				return err
			}
			continue
		}
		claimed[p.Index] = struct{}{}
		if err := p.Set(args[p.Index]); err != nil {
			return err
		}
	}
	if reg.unmatched == nil {
		return nil
	}
	return bindUnmatched(ctx, reg.unmatched, unclaimed(args, claimed))
}

// unclaimed returns the arguments whose index is absent from claimed, in
// command-line order.
func unclaimed(args []string, claimed map[int]struct{}) []string {
	rest := make([]string, 0, len(args))
	for i, arg := range args {
		if _, ok := claimed[i]; ok {
			continue
		}
		rest = append(rest, arg)
	}
	return rest
}

// bindUnmatched assigns rest to u, sourcing a fallback set instead when no
// argument went unclaimed.
func bindUnmatched(ctx context.Context, u *Unmatched, rest []string) error {
	if len(rest) == 0 {
		visited, err := setFallback(ctx, u.EnvFallbacks, u.FuncFallbacks, u.setFields)
		if visited || err != nil {
			return err
		}
	}
	return u.Set(rest)
}

// setFields assigns value to u as the comma-separated fields it holds.
func (u *Unmatched) setFields(value string) error {
	fields, err := csvfield.Split(value)
	if err != nil {
		return err
	}
	return u.Set(fields)
}

// setFallback assigns the first available fallback value to set, preferring an
// environment variable over a fallback function. It reports whether a fallback
// supplied a value.
func setFallback(ctx context.Context, envs []string, funcs []FallbackFunc, set func(string) error) (visited bool, err error) {
	visited, err = envFallback(envs, set)
	if visited || err != nil {
		return visited, err
	}
	return funcFallback(ctx, funcs, set)
}

func envFallback(envs []string, set func(string) error) (visited bool, err error) {
	for _, key := range envs {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			err = set(value)
			visited = true
			if err != nil {
				err = fmt.Errorf("%w: $%v: %w", ErrSettingEnvFlag, key, err)
			}
			return
		}
	}
	return false, nil
}

func funcFallback(ctx context.Context, funcs []FallbackFunc, set func(string) error) (visited bool, err error) {
	for _, fn := range funcs {
		value, ferr := fn(ctx)
		if ferr != nil {
			return true, fmt.Errorf("%w: %w", ErrComputingFuncFlag, ferr)
		}
		if value == "" {
			continue
		}
		err = set(value)
		visited = true
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrSettingFuncFlag, err)
		}
		return
	}
	return
}
