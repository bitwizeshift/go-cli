package argdef

import (
	"strings"

	"github.com/spf13/pflag"
)

// flagsPlaceholder stands in for the optional flags a command accepts but that
// its synopsis does not spell out.
const flagsPlaceholder = "[flags]"

// Usage returns the POSIX-style synopsis of the arguments registered on cl,
// excluding the command name itself. Required arguments are delimited by angle
// brackets and optional ones by square brackets, so a command taking a required
// "src" and an optional "dst" reports "<src> [dst]".
//
// Only required flags are spelled out; the presence of any other visible flag
// is reported by a trailing "[flags]". operands are inserted between the flags
// and the registered positional arguments, and name the parts of the command
// line that cl does not model, such as a subcommand.
//
// It returns an empty string when the command accepts no arguments at all.
func Usage(cl *CommandLine, operands ...string) string {
	segments := requiredFlagUsage(cl)
	segments = append(segments, operands...)
	for _, p := range cl.positionals {
		segments = append(segments, positionalUsage(p))
	}
	if cl.unmatched != nil {
		segments = append(segments, unmatchedUsage(cl.unmatched))
	}
	if hasOptionalFlags(cl) {
		segments = append(segments, flagsPlaceholder)
	}
	return strings.Join(segments, " ")
}

// requiredFlagUsage returns a segment for every visible flag that must be
// specified. Flags belonging to a group of which at least one member is
// required contribute a single alternation segment, emitted in the position of
// the first member visited.
func requiredFlagUsage(cl *CommandLine) []string {
	var segments []string
	emitted := map[string]struct{}{}
	cl.flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		group := OneRequired(f)
		if len(group) == 0 {
			if IsRequired(f) {
				segments = append(segments, flagUsage(f))
			}
			return
		}
		key := strings.Join(group, ",")
		if _, ok := emitted[key]; ok {
			return
		}
		emitted[key] = struct{}{}
		segments = append(segments, groupUsage(cl, group))
	})
	return segments
}

// groupUsage returns the alternation of the visible flags named by group, of
// which at least one must be specified. A group is only ever rendered from a
// visible member, so at least one alternative always survives.
func groupUsage(cl *CommandLine, group []string) string {
	alternatives := make([]string, 0, len(group))
	for _, name := range group {
		f := cl.flags.Lookup(name)
		if f == nil || f.Hidden {
			continue
		}
		alternatives = append(alternatives, flagUsage(f))
	}
	if len(alternatives) == 1 {
		return alternatives[0]
	}
	return "(" + strings.Join(alternatives, " | ") + ")"
}

// hasOptionalFlags reports whether cl carries a visible flag that the synopsis
// does not spell out.
func hasOptionalFlags(cl *CommandLine) bool {
	optional := false
	cl.flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || IsRequired(f) || len(OneRequired(f)) > 0 {
			return
		}
		optional = true
	})
	return optional
}

// flagUsage returns the synopsis of a single flag. Boolean flags take no value,
// and so are named alone.
func flagUsage(f *pflag.Flag) string {
	if f.Value.Type() == "bool" {
		return "--" + f.Name
	}
	return "--" + f.Name + " <" + f.Value.Type() + ">"
}

// positionalUsage returns the synopsis of a single positional argument.
func positionalUsage(p *Positional) string {
	if p.Required {
		return "<" + p.Name + ">"
	}
	return "[" + p.Name + "]"
}

// unmatchedUsage returns the synopsis of the binding claiming every argument no
// positional claims, marked variadic by a trailing ellipsis.
func unmatchedUsage(u *Unmatched) string {
	if u.Required {
		return "<" + u.Name + ">..."
	}
	return "[" + u.Name + "...]"
}
