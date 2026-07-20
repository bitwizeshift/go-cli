package spec

import (
	"encoding"
	"fmt"

	"go.yaml.in/yaml/v4"
)

// HostOS names a host operating system that an application id may be
// specialized for. [HostOSDefault] is the reserved name whose id applies to any
// host lacking its own entry.
type HostOS string

const (
	HostOSDefault HostOS = "default"
	HostOSWindows HostOS = "windows"
	HostOSMacOS   HostOS = "macos"
	HostOSLinux   HostOS = "linux"
	HostOSFreeBSD HostOS = "freebsd"
	HostOSOpenBSD HostOS = "openbsd"
	HostOSNetBSD  HostOS = "netbsd"
	HostOSIOS     HostOS = "ios"
	HostOSAndroid HostOS = "android"
	HostOSSolaris HostOS = "solaris"
	HostOSPlan9   HostOS = "plan9"
)

// hostOSNames holds every name a [HostOS] may be spelled as.
var hostOSNames = map[HostOS]struct{}{
	HostOSDefault: {},
	HostOSWindows: {},
	HostOSMacOS:   {},
	HostOSLinux:   {},
	HostOSFreeBSD: {},
	HostOSOpenBSD: {},
	HostOSNetBSD:  {},
	HostOSIOS:     {},
	HostOSAndroid: {},
	HostOSSolaris: {},
	HostOSPlan9:   {},
}

// goosHostOS maps the GOOS values that are spelled differently as a [HostOS].
// Any GOOS absent from this map is used verbatim.
var goosHostOS = map[string]HostOS{
	"darwin": HostOSMacOS,
}

// UnmarshalText decodes text as one of the named host operating systems.
//
// It returns [ErrUnknownHostOS] if text names no such host.
func (h *HostOS) UnmarshalText(text []byte) error {
	host := HostOS(text)
	if _, ok := hostOSNames[host]; !ok {
		return fmt.Errorf("%w: %q", ErrUnknownHostOS, text)
	}
	*h = host
	return nil
}

var _ encoding.TextUnmarshaler = (*HostOS)(nil)

// AppID is the identifier scoping an application's on-disk storage, held per
// host operating system so that each host may use the identifier its
// conventions call for.
type AppID map[HostOS]string

// For returns the identifier to use on goos, preferring the entry matching that
// host and falling back to the [HostOSDefault] entry. It returns the empty
// string when neither is held.
func (id AppID) For(goos string) string {
	host, ok := goosHostOS[goos]
	if !ok {
		host = HostOS(goos)
	}
	if appID, ok := id[host]; ok {
		return appID
	}
	return id[HostOSDefault]
}

// UnmarshalYAML decodes either a string, which becomes the [HostOSDefault]
// entry, or a mapping of host operating system to identifier.
//
// It returns [ErrInvalidAppID] if node is neither, or [ErrUnknownHostOS] if a
// mapping key names no known host.
func (id *AppID) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		*id = AppID{HostOSDefault: node.Value}
		return nil
	case yaml.MappingNode:
		hosts := make(map[HostOS]string)
		if err := node.Decode(&hosts); err != nil {
			return err
		}
		*id = hosts
		return nil
	default:
		return fmt.Errorf("%w: got kind %d", ErrInvalidAppID, node.Kind)
	}
}

var _ yaml.Unmarshaler = (*AppID)(nil)
