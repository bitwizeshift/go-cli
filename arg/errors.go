package arg

import "errors"

// ErrAlreadySet indicates a non-accumulating flag was specified more than once.
// Only unnamed slice flags (such as []string) accumulate across occurrences by
// default; any other flag reports this error on its second occurrence unless it
// was made repeatable with [Repeatable] or [RepeatableUpTo].
var ErrAlreadySet = errors.New("flag already specified")

// ErrTooManyOccurrences indicates a flag was specified more times than its
// [RepeatableUpTo] cap allows.
var ErrTooManyOccurrences = errors.New("flag specified too many times")

// errDecoderType indicates that a decoder supplied via [UnmarshalWith] produces
// a value whose type does not match the flag's destination.
var errDecoderType = errors.New("decoder type does not match flag value")

// errCallbackType indicates a flag's decoded value is not assignable or
// convertible to a [Callback] function's parameter type.
var errCallbackType = errors.New("flag value not convertible to callback argument")
