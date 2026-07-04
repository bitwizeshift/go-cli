package progresstest

import (
	"time"

	"github.com/bitwizeshift/go-cli/progress"
)

// Ticker is a manually driven [progress.Ticker] for tests. Each [Ticker.Send]
// delivers one tick and blocks until it is received, so the caller can advance
// an animation a precise number of frames.
type Ticker struct {
	ch chan time.Time
}

// NewTicker returns a Ticker ready to drive a [progress.Animator].
func NewTicker() *Ticker {
	return &Ticker{ch: make(chan time.Time)}
}

// Ticks returns the channel on which [Ticker.Send] delivers ticks.
func (t *Ticker) Ticks() <-chan time.Time { return t.ch }

// Stop is a no-op; a Ticker holds no resources.
func (t *Ticker) Stop() {}

// Send delivers a single tick, blocking until the consumer receives it.
func (t *Ticker) Send() { t.ch <- time.Now() }

var _ progress.Ticker = (*Ticker)(nil)
