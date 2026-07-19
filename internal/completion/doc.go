/*
Package completion owns the shell-completion candidates offered for command-line
arguments, and their translation into cobra's completion machinery.

Completions are supplied as cobra-free functions so that the arg package and its
callers never name a cobra type. A flag carries its completion as a pflag
annotation referencing a process-wide registry, since a [pflag.Flag] is
reachable without its command. Positional arguments have no such carrier, and
cobra permits only a single completion function per command for them, so their
completers are instead reconciled by the index of the argument being completed.
*/
package completion
