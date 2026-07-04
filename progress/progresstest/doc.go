// Package progresstest provides test doubles for exercising code that drives
// the progress package's animations.
//
// Its [Ticker] is a manually pumped [progress.Ticker]: each [Ticker.Send]
// advances a running [progress.Animator] by exactly one frame, so animation
// tests run deterministically without real time.
package progresstest
