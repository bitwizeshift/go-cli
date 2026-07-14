# Developer Docs

This section provides guides and resources for developers looking to leverage
the `go-cli` project, or looking to understand contributor processes.

## Tutorials

These are task-oriented walkthroughs for developers new to `go-cli`:

* [Your First Application](./tutorial/first-application.md): Build, wire up, and
  test a `go-cli` application from scratch.
* [Custom Reusable Flag Types](./tutorial/custom-flags.md): Write flag components
  that register themselves, expose high-level objects, and are testable with
  `flagtest`.
* [Formatted Output with Richtext](./tutorial/richtext-output.md): Style terminal
  output, define themes, and safely print text you did not author.
* [Customizing Exit Codes](./tutorial/custom-exit-codes.md): Translate your
  application's errors into your own exit codes, while keeping the framework's
  translation of standard-library errors.

## References

This section provides reference material for understanding parts of the public
`go-cli` library surface area:

* [Richtext](./reference/richtext.md): Details about `go-cli`'s `richtext` format.
* [Command Schema](./reference/command-schema.md): Information about the custom
  command/app schema format.

## Contributor Resources

The links below primarily are directed at contributors or anyone trying to
understand the processes used in this project.

* [Commit Standards](commit-standards.md)
* ["AI" Disclosure](ai-disclosure.md)
