/*
Package diagnostic provides a standardized [diagnostic.Logger] that can be
used from CLI operations. The diagnostics logging mechanism is modeled after
Rust's lint diagnostic outputs.

Diagnostics can contain various optional pieces of data:

  - An ID, useful for linting
  - The file source location
  - The line/column range within the source file
  - A title

Formatting is dynamic and reflexive. It will shape itself based on the underlying
terminal, and it'll selectively disable itself based on NO_COLOR or output writer
formatters.
*/
package diagnostic
