// Package ask reads answers to interactive prompts from an input stream.
//
// It is the engine behind the public prompt package: an [Asker] writes a prompt,
// reads a line or a masked secret while honouring a cancellation context, and can
// decode a typed answer. Secret entry requires a terminal, surfacing the echo
// controller's error when one is not present.
package ask
