package spec_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.yaml.in/yaml/v4"
)

func TestGroupCommandsUnmarshalYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    spec.GroupCommands
		wantErr error
	}{
		{
			name: "preserves document order",
			input: `
Zebra Commands:
  - id: z
    use: z
Apple Commands:
  - id: a
    use: a
`,
			want: spec.GroupCommands{
				{Name: "Zebra Commands", Commands: []spec.CommandInfo{{ID: "z", Use: "z"}}},
				{Name: "Apple Commands", Commands: []spec.CommandInfo{{ID: "a", Use: "a"}}},
			},
		},
		{
			name: "retains default group",
			input: `
default:
  - id: d
    use: d
Named Commands:
  - id: n
    use: n
`,
			want: spec.GroupCommands{
				{Name: "default", Commands: []spec.CommandInfo{{ID: "d", Use: "d"}}},
				{Name: "Named Commands", Commands: []spec.CommandInfo{{ID: "n", Use: "n"}}},
			},
		},
		{
			name:    "rejects non-mapping node",
			input:   "- id: x\n",
			wantErr: spec.ErrNotMapping,
		},
		{
			name:    "rejects non-list group value",
			input:   "Named:\n  scalar\n",
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut spec.GroupCommands

			// Act
			err := yaml.Unmarshal([]byte(tc.input), &sut)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("yaml.Unmarshal(...) = %v, want %v", got, want)
			}
			if got, want := sut, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("yaml.Unmarshal(...) diff (-got +want):\n%s", cmp.Diff(got, want))
			}
		})
	}
}
