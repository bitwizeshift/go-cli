package spec_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.yaml.in/yaml/v4"
)

func TestHostOS_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    spec.HostOS
		wantErr error
	}{
		{
			name:  "Default",
			input: "default",
			want:  spec.HostOSDefault,
		},
		{
			name:  "Windows",
			input: "windows",
			want:  spec.HostOSWindows,
		},
		{
			name:  "MacOS",
			input: "macos",
			want:  spec.HostOSMacOS,
		},
		{
			name:  "Linux",
			input: "linux",
			want:  spec.HostOSLinux,
		},
		{
			name:  "FreeBSD",
			input: "freebsd",
			want:  spec.HostOSFreeBSD,
		},
		{
			name:  "OpenBSD",
			input: "openbsd",
			want:  spec.HostOSOpenBSD,
		},
		{
			name:  "NetBSD",
			input: "netbsd",
			want:  spec.HostOSNetBSD,
		},
		{
			name:  "IOS",
			input: "ios",
			want:  spec.HostOSIOS,
		},
		{
			name:  "Android",
			input: "android",
			want:  spec.HostOSAndroid,
		},
		{
			name:  "Solaris",
			input: "solaris",
			want:  spec.HostOSSolaris,
		},
		{
			name:  "Plan9",
			input: "plan9",
			want:  spec.HostOSPlan9,
		},
		{
			name:    "RejectsGOOSSpelling",
			input:   "darwin",
			wantErr: spec.ErrUnknownHostOS,
		},
		{
			name:    "RejectsAbbreviation",
			input:   "mac",
			wantErr: spec.ErrUnknownHostOS,
		},
		{
			name:    "RejectsEmpty",
			input:   "",
			wantErr: spec.ErrUnknownHostOS,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut spec.HostOS

			// Act
			err := sut.UnmarshalText([]byte(tc.input))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("HostOS.UnmarshalText(...) = %v, want %v", got, want)
			}
			if got, want := sut, tc.want; !cmp.Equal(got, want) {
				t.Errorf("HostOS.UnmarshalText(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestAppID_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    spec.AppID
		wantErr error
	}{
		{
			name:  "ScalarBecomesDefault",
			input: "com.rodusek.example-cli\n",
			want: spec.AppID{
				spec.HostOSDefault: "com.rodusek.example-cli",
			},
		},
		{
			name:  "MappingRetainsEveryHost",
			input: "default: example-cli\nmacos: com.rodusek.example-cli\nwindows: ExampleCLI\n",
			want: spec.AppID{
				spec.HostOSDefault: "example-cli",
				spec.HostOSMacOS:   "com.rodusek.example-cli",
				spec.HostOSWindows: "ExampleCLI",
			},
		},
		{
			name:  "MappingWithoutDefault",
			input: "linux: example-cli\n",
			want: spec.AppID{
				spec.HostOSLinux: "example-cli",
			},
		},
		{
			name:    "RejectsUnknownHost",
			input:   "mac: com.rodusek.example-cli\n",
			wantErr: spec.ErrUnknownHostOS,
		},
		{
			name:    "RejectsSequenceNode",
			input:   "- example-cli\n",
			wantErr: spec.ErrInvalidAppID,
		},
		{
			name:    "RejectsNonStringValue",
			input:   "macos:\n  nested: value\n",
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut spec.AppID

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

func TestAppID_For(t *testing.T) {
	t.Parallel()

	fullAppID := spec.AppID{
		spec.HostOSDefault: "example-cli",
		spec.HostOSMacOS:   "com.rodusek.example-cli",
		spec.HostOSWindows: "ExampleCLI",
	}

	testCases := []struct {
		name  string
		appID spec.AppID
		goos  string
		want  string
	}{
		{
			name:  "DarwinSelectsMacOS",
			appID: fullAppID,
			goos:  "darwin",
			want:  "com.rodusek.example-cli",
		},
		{
			name:  "HostMatchWinsOverDefault",
			appID: fullAppID,
			goos:  "windows",
			want:  "ExampleCLI",
		},
		{
			name:  "UnlistedHostFallsBackToDefault",
			appID: fullAppID,
			goos:  "linux",
			want:  "example-cli",
		},
		{
			name:  "UnrecognizedGOOSFallsBackToDefault",
			appID: fullAppID,
			goos:  "wasip1",
			want:  "example-cli",
		},
		{
			name: "HostMatchWithoutDefault",
			appID: spec.AppID{
				spec.HostOSLinux: "example-cli",
			},
			goos: "linux",
			want: "example-cli",
		},
		{
			name: "UnlistedHostWithoutDefault",
			appID: spec.AppID{
				spec.HostOSLinux: "example-cli",
			},
			goos: "darwin",
			want: "",
		},
		{
			name:  "EmptyAppID",
			appID: spec.AppID{},
			goos:  "darwin",
			want:  "",
		},
		{
			name:  "NilAppID",
			appID: nil,
			goos:  "darwin",
			want:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.appID

			// Act
			appID := sut.For(tc.goos)

			// Assert
			if got, want := appID, tc.want; !cmp.Equal(got, want) {
				t.Errorf("AppID.For(%q) = %q, want %q", tc.goos, got, want)
			}
		})
	}
}
