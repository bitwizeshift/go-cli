package progress

// Spinner frame sets, cycled by [Spinner.Tick].
var (
	// LineFrames is a portable ASCII spinner: - \ | /.
	LineFrames = []string{`-`, `\`, `|`, `/`}

	// DotFrames is a Unicode braille spinner.
	DotFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// CircleFrames is a Unicode quartered-circle spinner.
	CircleFrames = []string{"◐", "◓", "◑", "◒"}

	// BounceFrames is a Unicode braille spinner that bounces a single dot.
	BounceFrames = []string{"⠁", "⠂", "⠄", "⠂"}

	// MoonFrames is a Unicode moon-phase spinner.
	MoonFrames = []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}

	// ArrowFrames is a Unicode rotating-arrow spinner.
	ArrowFrames = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}

	// PulseFrames is a Unicode vertical bar that grows and shrinks.
	PulseFrames = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}

	// SquareFrames is a Unicode rotating quartered-square spinner.
	SquareFrames = []string{"◰", "◳", "◲", "◱"}
)
