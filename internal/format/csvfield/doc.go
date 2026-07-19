/*
Package csvfield splits a single command-line value into its comma-separated
fields.

A value that feeds a set of arguments -- a slice flag, or the fallback for an
unmatched-argument binding -- is written as one comma-separated string. Quoting
follows encoding/csv, so a field containing a comma may be written as "a,b".

This package exists so that every such value is split identically, no matter
which binding consumes it.
*/
package csvfield
