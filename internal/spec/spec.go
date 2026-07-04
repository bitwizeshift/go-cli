package spec

import (
	"fmt"

	"github.com/bitwizeshift/go-cli/internal/arity"
	"go.yaml.in/yaml/v4"
)

// DefaultGroup is the reserved group name whose commands are left ungrouped.
const DefaultGroup = "default"

// Application is the root of a command specification. It is a [CommandInfo] with
// additional settings that apply to the whole command hierarchy.
type Application struct {
	CommandInfo `yaml:",inline"`

	// IssueURL is the URL users are directed to for filing bugs, propagated to
	// every command in the hierarchy.
	IssueURL string `yaml:"issue-url"`
}

// CommandInfo describes a single command in a plain-text, easily edited YAML
// form that mirrors the fields of a [github.com/spf13/cobra.Command].
type CommandInfo struct {
	ID          string   `yaml:"id"`
	Use         string   `yaml:"use"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Examples    []string `yaml:"examples,omitempty"`
	Summary     string   `yaml:"summary,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Version     string   `yaml:"version,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty"`
	Deprecated  string   `yaml:"deprecated,omitempty"`

	Arity arity.ArityFunc `yaml:"arity"`

	Commands GroupCommands `yaml:"commands"`
}

// GroupCommandInfo is a named group of commands. A group named [DefaultGroup]
// denotes commands that belong to no group.
type GroupCommandInfo struct {
	Name     string
	Commands []CommandInfo
}

// GroupCommands is an ordered list of command groups. The order matches the
// order the groups appear in the YAML document.
type GroupCommands []GroupCommandInfo

// UnmarshalYAML decodes a YAML mapping of group name to command list into an
// ordered [GroupCommands], preserving the order the groups appear in the
// document.
//
// It returns [ErrNotMapping] if node is not a mapping.
func (gc *GroupCommands) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: got kind %d", ErrNotMapping, node.Kind)
	}
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		var commands []CommandInfo
		if err := value.Decode(&commands); err != nil {
			return err
		}
		*gc = append(*gc, GroupCommandInfo{
			Name:     key.Value,
			Commands: commands,
		})
	}
	return nil
}

var _ yaml.Unmarshaler = (*GroupCommands)(nil)
