package argdef

import (
	"context"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/bitwizeshift/go-cli/internal/completion"
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
	Set func(values []string) error
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

// SetUnmatched records u as the unmatched-argument binding on reg, replacing any
// previously registered binding.
func SetUnmatched(reg *CommandLine, u *Unmatched) {
	reg.unmatched = u
}

// GetUnmatched returns the unmatched-argument binding registered on reg, or nil
// if none was registered.
func GetUnmatched(reg *CommandLine) *Unmatched {
	return reg.unmatched
}

// Bind assigns args to the positional and unmatched bindings registered on reg.
// A [Positional] whose index falls outside args is skipped, leaving its
// destination unchanged. Every argument not claimed by a [Positional] is passed,
// in command-line order, to an [Unmatched] binding.
//
// It returns the first error reported by a binding.
func Bind(ctx context.Context, reg *CommandLine, args []string) error {
	claimed := make(map[int]struct{})
	for _, p := range reg.positionals {
		if p.Index < 0 || p.Index >= len(args) {
			if err := setFallback(ctx, p); err != nil {
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
	rest := make([]string, 0, len(args))
	for i, arg := range args {
		if _, ok := claimed[i]; ok {
			continue
		}
		rest = append(rest, arg)
	}
	return reg.unmatched.Set(rest)
}

func setFallback(ctx context.Context, p *Positional) error {
	visited, err := envFallback(p)
	if visited || err != nil {
		return err
	}
	_, err = funcFallback(ctx, p)
	return err
}

func envFallback(p *Positional) (visited bool, err error) {
	for _, key := range p.EnvFallbacks {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			err = p.Set(value)
			visited = true
			if err != nil {
				err = fmt.Errorf("%w: $%v: %w", ErrSettingEnvFlag, key, err)
			}
			return
		}
	}
	return false, nil
}

func funcFallback(ctx context.Context, p *Positional) (visited bool, err error) {
	for _, fn := range p.FuncFallbacks {
		value, ferr := fn(ctx)
		if ferr != nil {
			return true, fmt.Errorf("%w: %w", ErrComputingFuncFlag, ferr)
		}
		if value == "" {
			continue
		}
		err = p.Set(value)
		visited = true
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrSettingFuncFlag, err)
		}
		return
	}
	return
}
