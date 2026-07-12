# Formatted Output with `richtext`

This tutorial covers producing styled terminal output: writing markup, defining a
theme, safely printing text you did not author, and testing the result.

It assumes you have read [Your First `go-cli` Application][first-app]. The
[richtext reference][richtext-ref] documents every tag and its semantics; this
tutorial does not repeat that list, it shows how to use them.

**Goal:** Write `richtext` markup both within go-cli runners, and outside.

## Writing markup

`richtext` is BBCode-inspired markup embedded in the strings you print. Tags open
with `[namespace:field]` and close with `[/namespace]`.

Inside a runner, `cli.OutStream(ctx)` is already a `*richtext.Writer`, so markup
works with nothing more than `fmt.Fprintf`:

```go
func (r *buildRunner) Run(ctx context.Context, args ...string) error {
  w := cli.OutStream(ctx)

  fmt.Fprintf(w, "[theme:title]Build complete[/theme]\n")
  fmt.Fprintf(w, "[theme:label]target:[/theme] [theme:value]%s[/theme]\n", r.target)
  fmt.Fprintf(w, "[fg:green][attr:bold]OK[/attr][/fg]\n")
  return nil
}
```

Nesting behaves the way CSS does: an inner tag applies for its extent, then the
previous style is restored. `attr` tags compound (bold inside italic gives you
both), while `fg` and `bg` replace.

Two failure modes to know about. Tags must close in the reverse of the order they
were opened, or `Write` returns a `*richtext.TagError` wrapping
`richtext.ErrUnbalancedTag`. A tag that is opened and never closed reports
`richtext.ErrUnclosedTag` at flush time.

An unrecognised *namespace* is not a tag at all, and is printed verbatim, so
`[INFO]` in your output stays `[INFO]`. An unrecognised *field* in a known
namespace (`[fg:chartreuse]`) is treated as a reset until its closing tag.

## Using richtext outside a command

The framework builds and flushes its writers for you. If you want one somewhere
else -- a library, a test, a tool that is not a `go-cli` command -- build it
yourself and flush it yourself:

```go
w := richtext.NewWriter(os.Stdout, richtext.DefaultTheme)
defer func() { _ = w.Flush() }()

io.WriteString(w, "[theme:heading]Results[/theme]\n")
```

`Flush` is not optional. It emits any trailing partial tag and is what reports
`ErrUnclosedTag`. Skip it and you can silently lose the tail of your output.

## Prefer themes over raw colours

`[fg:red]` hard-codes a decision. `[theme:error]` names an intent, and stays
correct when someone reconfigures the CLI's colours.

`richtext.DefaultTheme` ships with the names the library itself uses -- such as
`title`, `heading`, `label`, etc. Reach for those first.

Define your own by tagging `style.Style` fields on a struct:

```go
import (
  "github.com/bitwizeshift/go-cli/richtext"
  "github.com/bitwizeshift/go-cli/richtext/style"
)

theme, err := richtext.DefaultTheme.New(struct {
  Title   style.Style `theme:"title"`   // overrides the default
  Package style.Style `theme:"package"` // a new name
}{
  Title:   style.Style{Foreground: style.Cyan, Attributes: style.Bold},
  Package: style.Style{Foreground: style.RGB(120, 120, 255), Attributes: style.Italic},
})
if err != nil {
  return err
}

cli.FromBytes(configYAML,
  cli.Theme(theme),
  cli.BindRunner("build", &buildRunner{}),
).Execute()
```

`DefaultTheme.New` *derives*: names you declare win, and anything you leave out
falls back to the parent, so `[theme:error]` keeps working. Use
`richtext.NewTheme` instead if you want a theme with no inheritance.

Only exported fields carrying a `theme:"name"` tag are considered, and each must
be exactly a `style.Style`. Anything else is a `*richtext.ThemeFieldError`.

One behavioural difference from the format tags: a theme tag replaces the *entire*
active style rather than merging into it. `[theme:error]` inside `[attr:bold]` is
not bold unless the theme says it is.

## Third-party text: do not let it become markup

Any text you did not author is a formatting hazard. Error strings, filenames,
branch names, API responses, and user input can all contain `[` and `]`. A file
literally named `[fg:red]` should print as `[fg:red]`, not turn the rest of your
output red.

The mechanism is the `richtext:off` passthrough region. Content inside it is
written verbatim, never scanned for tags:

```go
fmt.Fprintf(w, "[theme:error]error:[/theme] [richtext:off]%s[/richtext]\n", err)
```

This is safe because of *when* substitution happens. `Fprintf` expands `%s`
before any byte reaches the writer, so the untrusted text is already sitting
inside the passthrough region by the time the parser sees it. The surrounding
style still applies -- the text is styled, just not parsed.

Never pass untrusted text as the format string itself:

```go
fmt.Fprintf(w, untrusted)             // wrong: also a vet error
fmt.Fprintf(w, "%s", untrusted)       // right, but unfenced: tags are parsed
fmt.Fprintf(w, "[richtext:off]%s[/richtext]", untrusted) // right, and fenced
```

### The caveat

A passthrough region ends at the first literal `[/richtext]` in the stream. Text
that itself contains `[/richtext]` therefore closes the region early, and
everything after it is parsed as markup again. There is no escape sequence for
this.

For text that could contain genuinely anything, bypass the parser instead of
fencing it off. The writer exposes its underlying destination:

```go
w := cli.OutStream(ctx)

fmt.Fprintf(w, "[theme:error]error:[/theme] ")
if rw, ok := w.(*richtext.Writer); ok {
  io.WriteString(rw.Writer(), untrusted) // never parsed, at all
} else {
  io.WriteString(w, untrusted)
}
fmt.Fprintln(w)
```

Bytes written to `rw.Writer()` skip the tag scanner entirely. The style you set
beforehand is still in effect, because that was already emitted to the terminal
as an escape sequence.

Use `[richtext:off]` for ordinary untrusted text, which is the common case and
what the library does internally for error messages. Reach for the bypass when
the input is genuinely adversarial.

## Aligning styled text

Markup has no width. `len("[theme:label]name[/theme]")` is 24, but the terminal
shows 4 characters, so any column arithmetic done with `len` will be wrong by the
length of the tags.

`richtext.Len` counts only the visible runes:

```go
label := "[theme:label]" + name + "[/theme]"
padding := strings.Repeat(" ", width-richtext.Len(label))
fmt.Fprintf(w, "%s%s%s\n", label, padding, value)
```

Pair it with `cli.StreamColumns(ctx, w)` when the width should follow the
terminal.

## Colour is conditional, and you get that for free

You do not check for a TTY. Colour is emitted only when the destination is a
terminal and `NO_COLOR` is not set, so piping your CLI into a file produces clean
text with no escape codes, and the same markup you already wrote.

`cli.ForceColour()` and `cli.DisableColour()` override the detection. They are
mutually exclusive, and setting a colour mode more than once panics.

## Testing styled output

`clitest.WithCaptureWriters` captures into a `strings.Builder`, which is not a
terminal, so colour is off and tags resolve away. Assertions compare plain text:

```go
func TestBuildRunner_Run(t *testing.T) {
  t.Parallel()

  // Arrange
  ctx, output := clitest.WithCaptureWriters(context.Background())
  sut := &buildRunner{target: "./cmd/greeter"}

  // Act
  err := sut.Run(ctx)

  // Assert
  if err != nil {
    t.Fatalf("sut.Run(...) = %v, want nil", err)
  }
  want := "Build complete\ntarget: ./cmd/greeter\n"
  if got, want := output.Stdout.String(), want; !cmp.Equal(got, want) {
    t.Errorf("sut.Run(...) stdout = %q, want %q", got, want)
  }
}
```

This is the useful property: your tests assert on *content*, and never on escape
sequences. Restyling a command does not break its tests.

To assert that the passthrough is doing its job, feed a runner input that looks
like markup and check it survives verbatim:

```go
err := sut.Run(ctx, "[fg:red]not-a-tag[/fg]")
// output.Stdout.String() contains "[fg:red]not-a-tag[/fg]" literally
```

If you need to test the escape sequences themselves, construct a writer directly
and call `ForceColour`:

```go
var buf strings.Builder
w := richtext.NewWriter(&buf, richtext.DefaultTheme)
w.ForceColour()
```

## Where to go next

* [Richtext reference][richtext-ref] is the complete tag, colour, and attribute
  catalogue, plus the format-stack semantics.
* [`examples/progress`](../../examples/progress) uses `richtext.Len` for
  width-aware layout.

[first-app]: ./first-application.md
[richtext-ref]: ../reference/richtext.md
