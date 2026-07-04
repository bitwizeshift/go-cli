package format_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/format"
)

func TestGrid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		rows      []format.Row
		columns   int
		indent    int
		gap       int
		minMarker int
		want      string
	}{
		{
			name:    "NoRows",
			rows:    nil,
			columns: 80,
			indent:  2,
			gap:     3,
			want:    "",
		}, {
			name: "MinMarkerForcesWiderColumn",
			rows: []format.Row{
				{Marker: "-a", Description: "first"},
			},
			columns:   80,
			indent:    2,
			gap:       3,
			minMarker: 10,
			want:      "  -a           first",
		}, {
			name: "SingleRowFits",
			rows: []format.Row{
				{Marker: "-h", Description: "help"},
			},
			columns: 80,
			indent:  0,
			gap:     1,
			want:    "-h help",
		}, {
			name: "MultiRowAlignsToWidestMarker",
			rows: []format.Row{
				{Marker: "-a", Description: "first"},
				{Marker: "--bbbb", Description: "second"},
			},
			columns: 80,
			indent:  2,
			gap:     3,
			want:    "  -a       first\n  --bbbb   second",
		}, {
			name: "ColouredMarkerAlignsByVisibleWidth",
			rows: []format.Row{
				{Marker: "\x1b[36m-a\x1b[0m", MarkerWidth: 2, Description: "first"},
				{Marker: "--bbbb", Description: "second"},
			},
			columns: 80,
			indent:  2,
			gap:     3,
			want:    "  \x1b[36m-a\x1b[0m       first\n  --bbbb   second",
		}, {
			name: "DescriptionWrapsAndIndentsToDescriptionColumn",
			rows: []format.Row{
				{Marker: "-f", Description: "this is a long description that wraps"},
			},
			columns: 24,
			indent:  2,
			gap:     3,
			want:    "  -f   this is a long\n       description that\n       wraps",
		}, {
			name: "EmptyDescriptionEmitsMarkerOnly",
			rows: []format.Row{
				{Marker: "-x", Description: ""},
				{Marker: "--yy", Description: "has desc"},
			},
			columns: 80,
			indent:  2,
			gap:     3,
			want:    "  -x\n  --yy   has desc",
		}, {
			name: "NonPositiveColumnsEmitsUnwrappedDescription",
			rows: []format.Row{
				{Marker: "-f", Description: "a b c d e f g really long"},
			},
			columns: 0,
			indent:  2,
			gap:     3,
			want:    "  -f   a b c d e f g really long",
		}, {
			name: "ZeroIndentAndGap",
			rows: []format.Row{
				{Marker: "-a", Description: "x"},
				{Marker: "-bb", Description: "y"},
			},
			columns: 80,
			indent:  0,
			gap:     0,
			want:    "-a x\n-bby",
		}, {
			name: "MarkerWiderThanColumnsStillPadsSiblings",
			rows: []format.Row{
				{Marker: "--a-very-long-marker", Description: "desc"},
				{Marker: "-b", Description: "other"},
			},
			columns: 10,
			indent:  2,
			gap:     3,
			want:    "  --a-very-long-marker   desc\n  -b                     other",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rows := tc.rows
			columns := tc.columns
			indent := tc.indent
			gap := tc.gap
			minMarker := tc.minMarker

			// Act
			grid := format.Grid(rows, columns, indent, gap, minMarker)

			// Assert
			if got, want := grid, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Grid(...) = %q, want %q\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}
