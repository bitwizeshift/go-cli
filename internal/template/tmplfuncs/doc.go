// Package tmplfuncs centralizes the functions available to the CLI's output
// templates.
//
// It exists so that every renderer shares one vocabulary of template calls
// instead of each defining ad-hoc closures. Functions are exposed as methods on
// small objects, gathered into a single map by [NewFunc]; templates invoke them
// as {{ build.VCS }} or {{ text.Wrap … }}. Because the functions are ordinary
// methods, they are unit-tested directly rather than through rendered output.
package tmplfuncs
