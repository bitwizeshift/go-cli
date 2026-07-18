package arg

// Arg is an argument that can be registered on a [CommandLine] with
// [CommandLine.Add].
type Arg interface {
	register(cl *CommandLine)
}
