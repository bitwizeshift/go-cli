// Command progress-demo is a self-contained showcase of the go-cli progress
// package.
//
// It binds a runner per bar and spinner style to an embedded YAML specification,
// each accepting --steps and --delay flags, to demonstrate live-updating output
// rendered in place.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/progress"
)

//go:embed app.yaml
var configYAML []byte

// The go-cli runtime already cancels each runner's context on interrupt, so the
// runners below simply wait on that context to stop their indicators cleanly.
func main() {
	cli.FromBytes(configYAML,
		cli.BindRunner("root", &rootRunner{}),
		cli.BindRunner("bar-ascii", newBarRunner(progress.ASCIIBar)),
		cli.BindRunner("bar-arrow", newBarRunner(progress.ArrowBar)),
		cli.BindRunner("bar-shaded", newBarRunner(progress.ShadedBar)),
		cli.BindRunner("bar-block", newBarRunner(progress.BlockBar)),
		cli.BindRunner("bar-smooth", newBarRunner(progress.SmoothBar)),
		cli.BindRunner("bar-line", newBarRunner(progress.LineBar)),
		cli.BindRunner("bar-dot", newBarRunner(progress.DotBar)),
		cli.BindRunner("spin-line", newSpinRunner(progress.LineSpinner)),
		cli.BindRunner("spin-dot", newSpinRunner(progress.DotSpinner)),
		cli.BindRunner("spin-circle", newSpinRunner(progress.CircleSpinner)),
		cli.BindRunner("spin-moon", newSpinRunner(progress.MoonSpinner)),
		cli.BindRunner("spin-arrow", newSpinRunner(progress.ArrowSpinner)),
		cli.BindRunner("spin-pulse", newSpinRunner(progress.PulseSpinner)),
		cli.BindRunner("spin-square", newSpinRunner(progress.SquareSpinner)),
		cli.BindRunner("all", newAllRunner()),
	).Execute()
}

// rootRunner backs the top-level command and shows usage when invoked without a
// subcommand.
type rootRunner struct{}

func (*rootRunner) Run(context.Context, ...string) error {
	return cli.ErrUsage
}

// barRunner fills a copy of a preset [progress.Bar] over --steps increments,
// pausing --delay milliseconds between each.
type barRunner struct {
	bar   progress.Bar
	steps int
	delay int
}

func newBarRunner(bar progress.Bar) *barRunner {
	return &barRunner{bar: bar, steps: 20, delay: 50}
}

func (r *barRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "steps", &r.steps,
		flag.Shorthand("s"),
		flag.Usage("number of steps to fill the bar"),
	)
	flag.Add(registry, "delay", &r.delay,
		flag.Shorthand("d"),
		flag.Usage("pause between steps, in milliseconds"),
	)
}

func (r *barRunner) Run(ctx context.Context, _ ...string) error {
	bar := r.bar
	bar.Total = int64(r.steps)
	bar.ShowPercent = true
	if !term.DefaultEnabler.EnableColour(os.Stdout) {
		bar.FillColour, bar.EmptyColour = "", ""
	}
	delay := time.Duration(r.delay) * time.Millisecond

	w := &progress.Writer{W: os.Stdout}
	for step := 0; step <= r.steps; step++ {
		bar.Current = int64(step)
		if err := w.Update(&bar); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return w.Done()
		case <-time.After(delay):
		}
	}
	return w.Done()
}

// spinRunner animates a copy of a preset [progress.Spinner] for --steps frames,
// pacing each frame by --delay milliseconds.
type spinRunner struct {
	spinner progress.Spinner
	steps   int
	delay   int
}

func newSpinRunner(spinner progress.Spinner) *spinRunner {
	return &spinRunner{spinner: spinner, steps: 30, delay: 80}
}

func (r *spinRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "steps", &r.steps,
		flag.Shorthand("s"),
		flag.Usage("number of frames to animate"),
	)
	flag.Add(registry, "delay", &r.delay,
		flag.Shorthand("d"),
		flag.Usage("pause between frames, in milliseconds"),
	)
}

func (r *spinRunner) Run(ctx context.Context, _ ...string) error {
	spinner := r.spinner
	spinner.Label = "working…"
	if !term.DefaultEnabler.EnableColour(os.Stdout) {
		spinner.Colour = ""
	}

	w := &progress.Writer{W: os.Stdout}
	delay := time.Duration(r.delay) * time.Millisecond
	animator := &progress.Animator{Writer: w, Target: &spinner, Interval: delay}

	animator.Start(ctx)
	timer := time.NewTimer(time.Duration(r.steps) * delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
	return animator.Stop()
}

// allRunner renders every bar and spinner together in one in-place block,
// advancing them in sync on each update.
type allRunner struct {
	steps int
	delay int
}

func newAllRunner() *allRunner {
	return &allRunner{steps: 40, delay: 60}
}

func (r *allRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "steps", &r.steps,
		flag.Shorthand("s"),
		flag.Usage("number of steps to run"),
	)
	flag.Add(registry, "delay", &r.delay,
		flag.Shorthand("d"),
		flag.Usage("pause between updates, in milliseconds"),
	)
}

func (r *allRunner) Run(ctx context.Context, _ ...string) error {
	coloured := term.DefaultEnabler.EnableColour(os.Stdout)
	indicators := r.indicators(coloured)
	group := indicators.group()
	delay := time.Duration(r.delay) * time.Millisecond

	w := &progress.Writer{W: os.Stdout}
	for step := 0; step <= r.steps; step++ {
		indicators.advance(step)
		if err := w.Update(group); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return w.Done()
		case <-time.After(delay):
		}
	}
	return w.Done()
}

// indicators builds the labelled bars and spinners shown by the "all" command,
// colouring them only when coloured is true.
func (r *allRunner) indicators(coloured bool) indicators {
	bar := func(preset progress.Bar) *progress.Bar {
		b := preset
		b.Total = int64(r.steps)
		b.ShowPercent = true
		if !coloured {
			b.FillColour, b.EmptyColour = "", ""
		}
		return &b
	}
	spin := func(preset progress.Spinner) *progress.Spinner {
		s := preset
		s.Label = "working…"
		if !coloured {
			s.Colour = ""
		}
		return &s
	}
	return indicators{
		barIndicator("bar-ascii", bar(progress.ASCIIBar)),
		barIndicator("bar-arrow", bar(progress.ArrowBar)),
		barIndicator("bar-shaded", bar(progress.ShadedBar)),
		barIndicator("bar-block", bar(progress.BlockBar)),
		barIndicator("bar-smooth", bar(progress.SmoothBar)),
		barIndicator("bar-line", bar(progress.LineBar)),
		barIndicator("bar-dot", bar(progress.DotBar)),
		spinIndicator("spin-line", spin(progress.LineSpinner)),
		spinIndicator("spin-dot", spin(progress.DotSpinner)),
		spinIndicator("spin-circle", spin(progress.CircleSpinner)),
		spinIndicator("spin-moon", spin(progress.MoonSpinner)),
		spinIndicator("spin-arrow", spin(progress.ArrowSpinner)),
		spinIndicator("spin-pulse", spin(progress.PulseSpinner)),
		spinIndicator("spin-square", spin(progress.SquareSpinner)),
	}
}

// indicator pairs a labelled renderer with how it advances on each step.
type indicator struct {
	label   string
	render  progress.Renderer
	advance func(step int)
}

func barIndicator(label string, bar *progress.Bar) indicator {
	return indicator{label: label, render: bar, advance: func(step int) { bar.Current = int64(step) }}
}

func spinIndicator(label string, spin *progress.Spinner) indicator {
	return indicator{label: label, render: spin, advance: func(int) { spin.Tick() }}
}

type indicators []indicator

// advance steps every indicator forward together, keeping them in sync.
func (in indicators) advance(step int) {
	for _, ind := range in {
		ind.advance(step)
	}
}

// group wraps each indicator as a [labeled] row, aligning labels to the widest
// so the indicators form a grid.
func (in indicators) group() progress.Group {
	width := 0
	for _, ind := range in {
		width = max(width, len(ind.label))
	}
	group := make(progress.Group, len(in))
	for i, ind := range in {
		group[i] = labeled{label: ind.label, width: width, inner: ind.render}
	}
	return group
}

// labeled renders inner on a single row indented by two spaces and prefixed by
// label padded to width, so a stack of rows aligns into a grid.
type labeled struct {
	label string
	width int
	inner progress.Renderer
}

func (l labeled) Render() string {
	return fmt.Sprintf("  %-*s  %s", l.width, l.label, l.inner.Render())
}

var (
	_ cli.Runner        = (*rootRunner)(nil)
	_ flag.Registrar    = (*barRunner)(nil)
	_ flag.Registrar    = (*spinRunner)(nil)
	_ flag.Registrar    = (*allRunner)(nil)
	_ progress.Renderer = labeled{}
)
