// Package template wires the CLI's output renderers to cobra.
//
// It exists so that callers install colour- and width-aware help, usage, panic,
// and version output without knowing how each renderer decides colour or size.
// A [RenderEngine] resolves those decisions from an [ansi.ColourEnabler] and a
// [term.Sizer] and hands back cobra-ready renderers, functions, and templates;
// use [DefaultRenderEngine] for the standard policy.
package template
