package progress_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/term/ansi"
	"github.com/bitwizeshift/go-cli/progress"
)

func TestBar_Render(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		bar  progress.Bar
		want string
	}{
		{
			name: "ZeroValueUsesAsciiFallbackAndDefaultWidth",
			bar: progress.Bar{
				Total:   10,
				Current: 5,
			},
			want: "##########          ",
		}, {
			name: "AsciiHalf",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"#"},
				Empty:   " ",
				Width:   30,
				Total:   100,
				Current: 50,
			},
			want: "[##############              ]",
		}, {
			name: "AsciiEmpty",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"#"},
				Empty:   " ",
				Width:   30,
				Total:   100,
				Current: 0,
			},
			want: "[                            ]",
		}, {
			name: "AsciiFull",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"#"},
				Empty:   " ",
				Width:   30,
				Total:   100,
				Current: 100,
			},
			want: "[############################]",
		}, {
			name: "ShowPercent",
			bar: progress.Bar{
				Left:        "[",
				Right:       "]",
				Fill:        []string{"#"},
				Empty:       " ",
				Width:       30,
				ShowPercent: true,
				Total:       100,
				Current:     50,
			},
			want: "[############            ] 50%",
		}, {
			name: "SuffixTruncatedToWidth",
			bar: progress.Bar{
				Left:        "[",
				Right:       "]",
				Fill:        []string{"#"},
				Empty:       " ",
				Width:       30,
				Suffix:      "downloading files",
				SuffixWidth: 6,
				Total:       100,
				Current:     0,
			},
			want: "[                     ] downl…",
		}, {
			name: "SubCellFrontier",
			bar: progress.Bar{
				Fill:    []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
				Empty:   " ",
				Width:   8,
				Total:   16,
				Current: 3,
			},
			want: "█▋      ",
		}, {
			name: "SubCellFrontierColoured",
			bar: progress.Bar{
				Fill:       []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
				Empty:      " ",
				Width:      8,
				FillColour: ansi.Green,
				Total:      16,
				Current:    3,
			},
			want: "\x1b[32m█▋\x1b[0m      ",
		}, {
			name: "SubCellOverShadedBackground",
			bar: progress.Bar{
				Fill:    []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
				Empty:   "▒",
				Width:   8,
				Total:   16,
				Current: 3,
			},
			want: "█▋▒▒▒▒▒▒",
		}, {
			name: "ArrowHead",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"="},
				Head:    ">",
				Empty:   " ",
				Width:   30,
				Total:   100,
				Current: 50,
			},
			want: "[==============>             ]",
		}, {
			name: "ArrowFullHasNoHead",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"="},
				Head:    ">",
				Empty:   " ",
				Width:   30,
				Total:   100,
				Current: 100,
			},
			want: "[============================]",
		}, {
			name: "WidthTooSmallCollapsesTrack",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Width:   2,
				Total:   10,
				Current: 5,
			},
			want: "[]",
		}, {
			name: "WidthBelowDecorationsCollapsesTrack",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Width:   1,
				Total:   10,
				Current: 5,
			},
			want: "[]",
		}, {
			name: "TotalZeroRendersEmptyTrack",
			bar: progress.Bar{
				Left:    "[",
				Right:   "]",
				Fill:    []string{"#"},
				Empty:   " ",
				Width:   12,
				Total:   0,
				Current: 5,
			},
			want: "[          ]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.bar

			// Act
			rendered := sut.Render()

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Bar.Render() = %q, want %q", got, want)
			}
		})
	}
}

func TestBar_Add(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		bar   progress.Bar
		delta int64
		want  int64
	}{
		{
			name: "IncrementsFromZero",
			bar: progress.Bar{
				Total: 100,
			},
			delta: 15,
			want:  15,
		}, {
			name: "AccumulatesOntoExisting",
			bar: progress.Bar{
				Total:   100,
				Current: 10,
			},
			delta: 15,
			want:  25,
		}, {
			name: "NegativeDeltaDecrements",
			bar: progress.Bar{
				Total:   100,
				Current: 10,
			},
			delta: -4,
			want:  6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.bar

			// Act
			sut.Add(tc.delta)

			// Assert
			if got, want := sut.Current, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Bar.Add(%d) Current = %d, want %d", tc.delta, got, want)
			}
		})
	}
}

func TestBar_Fraction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		bar  progress.Bar
		want float64
	}{
		{
			name: "TotalZero",
			bar: progress.Bar{
				Total:   0,
				Current: 5,
			},
			want: 0,
		}, {
			name: "NegativeCurrentClampsToZero",
			bar: progress.Bar{
				Total:   10,
				Current: -5,
			},
			want: 0,
		}, {
			name: "Half",
			bar: progress.Bar{
				Total:   10,
				Current: 5,
			},
			want: 0.5,
		}, {
			name: "OverflowClampsToOne",
			bar: progress.Bar{
				Total:   10,
				Current: 15,
			},
			want: 1,
		}, {
			name: "Quarter",
			bar: progress.Bar{
				Total:   4,
				Current: 1,
			},
			want: 0.25,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.bar

			// Act
			fraction := sut.Fraction()

			// Assert
			if got, want := fraction, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Bar.Fraction() = %v, want %v", got, want)
			}
		})
	}
}
