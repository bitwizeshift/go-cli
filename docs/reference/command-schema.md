# Command Schema

The `go-cli` library leverages configuration files to set up commands, rather
than having things done inline. There are several reasons for this:

* Inline strings for descriptions are _monstrous_ to deal with in code. For long
  descriptions, you may have lines that are hundreds of characters long, or you
  may use raw-string literals which result in the output descriptions having
  strange newline positions.

* Creating the nested commands becomes a pain after a certain point, especially
  when dealing with command groups.

* etc

Overall, `go-cli` chose to abstract this to a configuration format that gets
bound at runtime instead. This way 95% of the command is a simple `go:embed`ded
YAML that is easy to read, normalizes newlines, leaving the 5% to be the stuff
that needs to be tested.

## Schema

This configuration format is defined in [JSONSchema] format, which allows
IDE tools and linters to help better validate this. See the
[`command.schema.json`][schema] file at the root of the repo.

[JSONSchema]: https://json-schema.org/specification
[schema]: https://github.com/bitwizeshift/go-cli/blob/master/command.schema.json

## Structure

The configuration format is largely just a "root" command config node, followed
by recursively nested "command" nodes.

### Root

The root config is the top-level of the YAML file. The root contains all the
field of [Command nodes](#command), as well as some additional settings.

* `issue-url`: A URL that consumers can file issues at if an error occurs. This
  is provided to the user if the program ever panics without being `recover`ed
  before it reaches the end of the command.

For the rest of the fields, read below.

### Command

Most of the configuration consists of "Command" node objects, most of which map
1:1 with the `cobra.Command` configuration fields.

The fields are:

* `id` (required): an identifier that can be used to uniquely identify this
  command object. This identifier is what is used to bind runners to the command
  later.

* `use` (required): maps to the `cobra.Command.Use` field. This is necessary for
  knowing how to execute the program.

* `aliases`: A list of optional aliases that can be used instead of the command
  that will invoke the same command. Maps to `cobra.Command.Aliases`.

* `examples`: A list of optional examples to demonstrate how the command is used.
  This maps to `cobra.Command.Example`, except `go-cli` enables showing multiple.

* `summary`: The short summary that is provided when seen from a `--help` menu
  of the parent command. Optional, but recommended. This maps to the
  `cobra.Command.Short` field.

* `description`: The long description of how this command is used, seen when
  `--help` is provided to that command. Optional, but recommended. This maps to
  the `cobra.Command.Long` field.

* `version`: The version of this command. Ideally this should only be specified
  on the root, but is available on any command. Enables the `--version` flag.
  This maps to the `cobra.Command.Version` field.

* `hidden`: A boolean to indicate that the command is hidden, and won't appear
  in `--help` menus. Maps to the `cobra.Command.Hidden` field.

* `deprecated`: A string to indicate that the command is deprecated. When set,
  a message is indicated that the command is deprecated. Maps to the
  `cobra.Command.Deprecated` field.

* `arity`: An optional field indicating the range of positional arguments that
  can be provided to this command. See [Arity](#arity) for more details.
  This configures the `cobra.Command.PositionalArgs` settings.

* `commands`: This is where composition occurs. This field is a mapping of
  `<group-name>: [<commands>]`, where `<group-name>` is the title of a command
  group. Use `default` if no group is desired. Each command in the list of
  `<command>`s is a recursive member of this schema.

#### Arity

The `arity` field supports a custom DSL for indicating the range of arguments that are allowed.
This is nothing too complex; the expression can be one of 3 primary things:

* An exact value (e.g. `3`, `5`, etc)
* An open/unbounded range using on of `>=`, `>`, `<`, or `<=` -- in the form of
  `<op><number>` (e.g. `>=5` expects 5 or more positional arguments)
* A range with an _exclusive_ upper bound using `<start>..<end>` syntax (e.g.
  `3..5` expects 3 or 4 arguments)
* A range with an _inclusive_ upper bound using `<start>..=<end>` syntax (e.g.
  `3..=5` expects between 3 to 5 arguments)

In addition to the above, multiple clauses can be concatenated with `,` provided
the ranges don't overlap. For example, `3,5,>=7` enables 3, 5, or greater than 7
arguments.

### Example

Excerpt from [this example](../../examples/simple/app.yaml):

```yaml
id: root
use: example-cli <command>
version: 1.4.0
summary: A fictional workspace manager that showcases the go-cli library
description: |
  example-cli manages a local "vault" of items and the remotes they synchronize
  with.

  It is not a real tool; it exists to exercise the go-cli library end to end,
  demonstrating grouped commands, nested subcommands, positional-argument arity,
  flags with and without values, flag constraints, and crash handling.
examples:
  - example-cli init --path ./vault
  - example-cli add "Buy milk" --priority 2 --tag chores --tag home
  - example-cli remote add origin https://example.test --token s3cret
issue-url: https://github.com/bitwizeshift/go-cli/issues

commands:
  default:
    - id: init
      use: init [--path <dir>] [--force]
      summary: Initializes a new vault
      description: |
        Creates a new, empty vault in the target directory.

        An existing vault is left untouched unless --force is given, in which
        case it is recreated from scratch and its contents are discarded.
      arity: 0
```

### Runner IDs

The `id` field of the command nodes corresponds to the ID used for runners when
constructing a CLI application.

In the above [example](#example), the id `init` can be bound to real behavior
by using:

```go
var app = cli.FromReader(embeddedYAML,
  cli.BindRunner("init", ...), // bind 'init' ID to a 'Runner'
  ...
)
```
