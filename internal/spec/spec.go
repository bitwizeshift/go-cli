package spec

import (
	"fmt"
	"os"
	"path/filepath"

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

	// AppID scopes the application's on-disk storage, selected for the host the
	// application runs on. It is derived from the command's Name, and failing
	// that the running binary's name, when the host resolves no identifier.
	AppID AppID `yaml:"app-id,omitempty"`

	// UpdateSources holds per-source configuration for update checking, keyed by
	// the source name. Each value is decoded into the update provider registered
	// under that name.
	UpdateSources map[string]yaml.Node `yaml:"update-sources"`
}

// resolveAppID returns the effective application id used to scope storage on
// goos: the [Application.AppID] selected for that host, else the command's Name,
// else the base name of the running binary.
func (a *Application) resolveAppID(goos string) string {
	if appID := a.AppID.For(goos); appID != "" {
		return appID
	}
	if a.Name != "" {
		return a.Name
	}
	return filepath.Base(os.Args[0])
}

// CommandInfo describes a single command in a plain-text, easily edited YAML
// form that mirrors the fields of a [github.com/spf13/cobra.Command].
type CommandInfo struct {
	Name        string   `yaml:"name"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Examples    []string `yaml:"examples,omitempty"`
	Summary     string   `yaml:"summary,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty"`
	Deprecated  string   `yaml:"deprecated,omitempty"`

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
