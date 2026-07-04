package progress

import (
	"context"
	"errors"
	"sync"
	"time"
)

// defaultInterval is the tick interval used when [Animator.Interval] is unset.
const defaultInterval = 100 * time.Millisecond

// Ticker delivers periodic ticks. It abstracts [time.Ticker] so animations can
// be driven deterministically in tests.
type Ticker interface {
	// Ticks is the channel on which ticks are delivered.
	Ticks() <-chan time.Time
	// Stop releases the ticker's resources.
	Stop()
}

// TickRenderer is a [Renderer] whose state advances one frame per Tick.
type TickRenderer interface {
	Renderer
	// Tick advances the renderer to its next frame.
	Tick()
}

var _ TickRenderer = (*Spinner)(nil)

// Animator advances Target and redraws it on Writer once per tick, from [Start]
// until [Stop] or context cancellation. Target must not be mutated by other
// goroutines while the animator is running.
type Animator struct {
	// Writer and Target must be set.
	Writer *Writer
	Target TickRenderer
	// NewTicker builds the driving ticker. A nil NewTicker uses a [time.Ticker]
	// of Interval, or [defaultInterval] when Interval is not positive.
	NewTicker func() Ticker
	Interval  time.Duration

	mu     sync.Mutex
	err    error
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Start begins the animation, spawning a goroutine that ticks Target and
// redraws it on each tick. The goroutine runs until [Stop] is called or ctx is
// cancelled, whichever happens first.
func (a *Animator) Start(ctx context.Context) {
	ctx, a.cancel = context.WithCancel(ctx)
	ticker := a.newTicker()
	a.wg.Add(1)
	go a.loop(ctx, ticker)
}

// Stop halts the animation, waits for the goroutine to exit, and finalises the
// [Writer]. It returns the first redraw error observed, if any.
func (a *Animator) Stop() error {
	a.cancel()
	a.wg.Wait()

	doneErr := a.Writer.Done()

	a.mu.Lock()
	defer a.mu.Unlock()
	return errors.Join(a.err, doneErr)
}

func (a *Animator) loop(ctx context.Context, ticker Ticker) {
	defer a.wg.Done()
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.Ticks():
			a.Target.Tick()
			if err := a.Writer.Update(a.Target); err != nil {
				a.setErr(err)
			}
		}
	}
}

// setErr records the first redraw error observed by the loop.
func (a *Animator) setErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err == nil {
		a.err = err
	}
}

func (a *Animator) newTicker() Ticker {
	if a.NewTicker != nil {
		return a.NewTicker()
	}
	interval := a.Interval
	if interval <= 0 {
		interval = defaultInterval
	}
	return timeTicker{time.NewTicker(interval)}
}

// timeTicker adapts a [time.Ticker] to the [Ticker] interface.
type timeTicker struct {
	*time.Ticker
}

// Ticks returns the underlying ticker's channel.
func (t timeTicker) Ticks() <-chan time.Time { return t.C }

var _ Ticker = timeTicker{}
