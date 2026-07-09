package progress_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/term/ansi"
	"github.com/bitwizeshift/go-cli/progress"
)

func TestSpinner_Render(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spinner progress.Spinner
		want    string
	}{
		{
			name:    "ZeroValueUsesLineFrames",
			spinner: progress.Spinner{},
			want:    "-",
		}, {
			name: "FrameIndexSelectsGlyph",
			spinner: progress.Spinner{
				Frames: progress.LineFrames,
				Frame:  2,
			},
			want: "|",
		}, {
			name: "Coloured",
			spinner: progress.Spinner{
				Frames: progress.DotFrames,
				Colour: ansi.Cyan,
			},
			want: "\x1b[36m⠋\x1b[0m",
		}, {
			name: "WithLabel",
			spinner: progress.Spinner{
				Frames: progress.LineFrames,
				Frame:  1,
				Label:  "working",
			},
			want: "\\ working",
		}, {
			name: "LabelTruncatedToWidth",
			spinner: progress.Spinner{
				Frames:     progress.LineFrames,
				Label:      "loading assets",
				LabelWidth: 4,
			},
			want: "- loa…",
		}, {
			name: "FrameWrapsViaModulo",
			spinner: progress.Spinner{
				Frames: progress.CircleFrames,
				Frame:  5,
			},
			want: "◓",
		}, {
			name: "NegativeFrameWrapsPositive",
			spinner: progress.Spinner{
				Frames: progress.LineFrames,
				Frame:  -1,
			},
			want: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.spinner

			// Act
			rendered := sut.Render()

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Spinner.Render() = %q, want %q", got, want)
			}
		})
	}
}

func TestSpinner_Tick(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spinner progress.Spinner
		ticks   int
		want    int
	}{
		{
			name: "AdvancesFrame",
			spinner: progress.Spinner{
				Frames: progress.LineFrames,
				Frame:  0,
			},
			ticks: 1,
			want:  1,
		}, {
			name: "WrapsAtEnd",
			spinner: progress.Spinner{
				Frames: progress.LineFrames,
				Frame:  3,
			},
			ticks: 1,
			want:  0,
		}, {
			name:    "ZeroValueUsesLineFrames",
			spinner: progress.Spinner{},
			ticks:   1,
			want:    1,
		}, {
			name: "MultipleTicksWrapAround",
			spinner: progress.Spinner{
				Frames: progress.CircleFrames,
				Frame:  0,
			},
			ticks: 6,
			want:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.spinner

			// Act
			for range tc.ticks {
				sut.Tick()
			}

			// Assert
			if got, want := sut.Frame, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Spinner.Tick() Frame = %d, want %d", got, want)
			}
		})
	}
}
