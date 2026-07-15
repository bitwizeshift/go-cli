# Rich Text

This project uses a custom in-house terminal-text markup system called
`richtext` (original, I know).

This document describes how it works and how it's used.

## Overview

The `richtext` concept is functionally simple; text is written with
BBCode-inspired tags like `[fg:red]content[/fg]`, and the terminal will render
this as one would expect -- with red foreground text.

The formatting behavior is similar to HTML: nested tags will alter the behavior
for the remainder of the nesting, but will unwind to the previous style once
the ending tag is reached. Unlike HTML, we have a concept of [themes](#themes)
that will overwrite _all_ states while active before popping back to the
previous state from the [format stack](#format-stack).

Markup is [conditional](#conditional-styling) based on the terminal.

## Conditional Styling

Whether or not the CLI is colored with ANSI escape-codes is conditional, and
comes down to several factors:

1. Is the terminal a valid TTY? If it's not, the default behavior will be to
   disable coloring so that escape codes aren't littered in the output. This
   means piping a CLI built with `go-cli`'s richtext system just "does the right
   thing".

2. Is the environment variable `NO_COLOR` set? This logic honours the `NO_COLOR`
   convention and will conditionally disable colour when set.

3. Is colour forced either off or on with [`cli.DisableColour`] or
   [`cli.ForceColour`]? These will override the base determinations from (1) or
   (2).

[`cli.DisableColour`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli#DisableColour
[`cli.ForceColour`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli#ForceColour

## Tags

> [!NOTE]
> Any tags encountered during parsing with an invalid namespace are treated as
  _text_, and emitted verbatim, instead of as an instruction.

All tags are formed using a `namespace:directive` format. There are currently
5 valid namespaces:

* **Format**
  * [`fg`]: changes foreground text color
  * [`bg`]: changes background text color
  * [`attr`]: changes text attributes
* **Meta**
  * [`richtext`]: meta tags for changing the engine's behavior
* **Aggregate**
  * [`theme`]: changes the theme state

The first three tags are treated as a joined set of "format" tags, and merge
styles when opened, undoing them on closure. See the [Format Stack](#format-stack)
for more details.

Unknown `directive`s for a known namespace are treated as a zero-signal and send
a reset code to the style until the end-tag is reached. This means `fg:invalid`
will render as the default terminal color.

Tags are opened with `[namespace:directive]` and closed with `[/namespace]`
(note: no directive). It is an error to miss a closing tag for a valid opening
tag.

[`fg`]: #fg-foreground-tags
[`bg`]: #bg-background-tags
[`attr`]: #attr-attribute-tags
[`richtext`]: #richtext-tags
[`theme`]: #theme-tags

### `fg` foreground tags

The `fg` directive changes the current foreground color of the terminal. Valid
`directive` values here follow the names of ANSI color-code name:

* `black`
* `red`
* `green`
* `yellow`
* `blue`
* `magenta`
* `cyan`
* `white`
* `brightblack`
* `brightred`
* `brightgreen`
* `brightyellow`
* `brightblue`
* `brightmagenta`
* `brightcyan`
* `brightwhite`

In addition to this, a special `rgb(r,g,b)` value is allowed to set true-color
output. Values outside of `0..=255` are treated as invalid values.

### `bg` background tags

The `bg` namespace changes the current background color of the terminal without
impacting the foreground color. Valid `directive` values are identical to the
ones defined in the [`fg` tag](#fg-foreground-tags).

### `attr` attribute tags

The `attr` namespace impacts the current _style_ of the text. This is orthogonal
to colour values by enabling features like italics, bolding, underlining, etc.
Support for each feature ranges based on terminal implementation.

Valid `directive` values here are:

* `bold`: Bolds the terminal text. In practice, frequently makes the text look
  brighter. Implementations vary based on terminal.
* `faint`: Makes the text look more faint. In practice, frequently dims the text
  color.
* `italic`: Slants the text.
* `underline`: Adds an underline below the text.
* `blink`: Cause text to flash. Not supported by all terminals.
* `reverse`: Reverses the text. Not supported by all terminals.
* `strike`: Adds a strikethrough the text. Not supported by all terminals.

Attributes can compound with other attributes (e.g. it's valid to form an
underlined and bold line of text, if the terminal supports it).

### `richtext` tags

The `richtext` namespace sets meta features of the rendering engine. Currently
only `richtext:off` is supported as a directive.

`richtext:off` allows the contents within the tags to be treated as plaintext
instead of being nested markup. This is good for handling user-input that may
look like tags:

```go
w := cli.OutStream(ctx)
fmt.Fprintf(w, "[richtext:off]")
fmt.Fprintf(w, userInput)
fmt.Fprintf(w, "[/richtext]")
```

If `userInput` is `[fg:red]testing[/fg]`

The output will be, verbatim:

```text
[fg:red]testing[/fg]
```

### `theme` tags

The `theme` namespace sets custom group of styles to apply at once. This
combine a `fg`, `bg`, and `attr`s together to allow for a cohesive set of color
schemes in your terminal.

Theme `directive` values are free-form, and are configured based on what is
configured at the time of CLI creation. See [themes](#themes) for more details.

## Format Stack

Styling with `richtext` codes attempts to be intuitive by applying a
"Format stack". The stack has 3 general operation modes, changed based on which
directives were last encountered.

* **Raw Style Mode**: This mode occurs on any valid `fg`, `bg`, or `attrs` mode.
  In this mode, the format stack has 3 substacks that decide the current styling.
  Every `fg` or `bg` tag is pushed onto the respective `fg`/`bg` stack to set
  the current foreground or background colour respectively. Every `attrs` is
  merged with the previous value and pushed on the stack. This means that an
  `italic` attr on top of a `bold` attr will form a bold-italic attr -- which is
  what would generally be intuitive as an author.

  That is:

  ```text
  [fg:red]
    [bg:black]
      [attr:italic]
        hello [fg:green][attr:bold]world[/attr][/fg]
      [/attr]
    [/bg]
  [/fg]
  ```

  will form "hello" with red foreground, black background, and italic style, and
  "world" with **green** foreground, black background, italic _and_ bold style.
  The `fg` overwrites, whereas the `attrs` merge.

  This is very similar to the behavior of HTML's CSS styling.

* **Theme Style Mode**: This mode occurs any time a `theme` is encountered.
  Themes always overwrite the complete styling at the point its encountered, so
  this resets the `fg`, `bg`, and `attrs` to what is defined in the theme for
  that directive.

  On closing tag, the previous state is retrieved

* **Error Style Mode**: This mode only occurs when an invalid directive is
  encountered for a valid namespace -- for example, `[fg:bad]`. This mode
  operates similarly to "theme" mode, except custom rendering is disabled at this
  state in the stack. This is done to prevent style flooding or other errors.

## Themes

Themes are sets of named styles that can be applied monolithically in richtext
codes. Themes are extensible, allowing for themes to either override or extend
features from an existing theme. Themes always override what is on the
[format stack](#format-stack).

Construction is done with either [`richtext.NewTheme`] for a new, base theme, or
[`richtext.Theme.New`] for derived. They are configured from `map` objects
binding string names to defined [`style.Style`]s:

```go
theme := richtext.NewTheme(map[string]{
  "title":   style.Style{Foreground: style.Cyan, Attributes: style.Bold},
  "heading": style.Style{Foreground: style.Green},
})
```

This theme would enable `[theme:title]` and `[theme:heading]` in the `richtext`
markup.

Themes can also derive and extend other themes, allowing for slight changes to
existing styles. Following from above:

```go
derived, err := theme.New(struct {
  Title style.Style `theme:"title"`
  Note  style.Style `theme:"note"`
}{
  Title: style.Style{Foreground: style.Red},
  Note:  style.Style{Foreground: style.Blue},
})
```

This would override `[theme:title]` to be red, and adds a new `[theme:note]`
style.

[`richtext.NewTheme`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/richtext#NewTheme
[`richtext.Theme.New`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/richtext#Theme.New
[`style.Style`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/richtext/style#Style
