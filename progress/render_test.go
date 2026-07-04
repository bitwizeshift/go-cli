package progress_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/progress"
)

func TestGroup_Render(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		group progress.Group
		want  string
	}{
		{
			name:  "Empty",
			group: progress.Group{},
			want:  "",
		}, {
			name:  "SingleMember",
			group: progress.Group{staticRenderer("only")},
			want:  "only",
		}, {
			name:  "StacksMembersWithNewlines",
			group: progress.Group{staticRenderer("first"), staticRenderer("second")},
			want:  "first\nsecond",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.group

			// Act
			rendered := sut.Render()

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Group.Render() = %q, want %q", got, want)
			}
		})
	}
}
