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
			name: "PreservesDocumentOrder",
			input: `
Zebra Commands:
  - name: z
Apple Commands:
  - name: a
`,
			want: spec.GroupCommands{
				{Name: "Zebra Commands", Commands: []spec.CommandInfo{{Name: "z"}}},
				{Name: "Apple Commands", Commands: []spec.CommandInfo{{Name: "a"}}},
			},
		},
		{
			name: "RetainsDefaultGroup",
			input: `
default:
  - name: d
Named Commands:
  - name: n
`,
			want: spec.GroupCommands{
				{Name: "default", Commands: []spec.CommandInfo{{Name: "d"}}},
				{Name: "Named Commands", Commands: []spec.CommandInfo{{Name: "n"}}},
			},
		},
		{
			name:    "RejectsNonMappingNode",
			input:   "- name: x\n",
			wantErr: spec.ErrNotMapping,
		},
		{
			name:    "RejectsNonListGroupValue",
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
