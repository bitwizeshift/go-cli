package spec_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
	"github.com/bitwizeshift/go-cli/update"
	"github.com/bitwizeshift/go-cli/update/updatetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
)

func TestBuild_UpdateSourceConfig_ConfiguresProvider(t *testing.T) {
	t.Parallel()

	// Arrange
	provider := &update.GitHubProvider{}
	input := strings.NewReader(yamlDoc(
		"id: root",
		"use: root",
		"update-sources:",
		"  github:",
		"    owner: bitwizeshift",
		"    repo: go-cli",
	))
	opts := spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": spectest.NoOpRunner()}),
		Update: spec.UpdateOptions{
			Version:   "v1.0.0",
			Source:    "github",
			Providers: map[string]update.Provider{"github": provider},
		},
	}

	// Act
	_, err := spec.Build(input, opts)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Build() = %v, want %v", got, want)
	}
	if got, want := *provider, (update.GitHubProvider{Owner: "bitwizeshift", Repo: "go-cli"}); !cmp.Equal(got, want) {
		t.Errorf("configured provider = %+v, want %+v", got, want)
	}
}

func TestBuild_UndecodableUpdateSource_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	input := strings.NewReader(yamlDoc(
		"id: root",
		"use: root",
		"update-sources:",
		"  github: not-a-mapping",
	))
	opts := spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": spectest.NoOpRunner()}),
		Update: spec.UpdateOptions{
			Version:   "v1.0.0",
			Source:    "github",
			Providers: map[string]update.Provider{"github": &update.GitHubProvider{}},
		},
	}

	// Act
	_, err := spec.Build(input, opts)

	// Assert
	if got, want := err, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Build() = %v, want %v", got, want)
	}
}

func TestBuild_UpdateAdvisory(t *testing.T) {
	// os.UserCacheDir is forced to fail so the cache never touches the real
	// filesystem; the update check still runs against the in-memory provider.
	// The backing environment variable is OS-specific: HOME/XDG_CACHE_HOME on
	// Unix, LocalAppData on Windows. All are cleared so the cache is disabled on
	// every platform, otherwise a working cache leaks state between subtests.
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LocalAppData", "")

	const (
		currentVersion = "v1.0.0"
		latestVersion  = "v9.9.9"
	)
	errLookup := errors.New("release lookup")

	testCases := []struct {
		name         string
		lines        []string
		update       spec.UpdateOptions
		wantAdvisory bool
	}{
		{
			name: "Available",
			lines: []string{
				"id: root",
				"use: root",
			},
			update: spec.UpdateOptions{
				Version:   currentVersion,
				Source:    "github",
				TTL:       time.Hour,
				Providers: map[string]update.Provider{"github": updatetest.Provider(latestVersion)},
			},
			wantAdvisory: true,
		}, {
			name: "ExplicitAppID",
			lines: []string{
				"id: root",
				"use: root",
				"app-id: myapp",
			},
			update: spec.UpdateOptions{
				Version:   currentVersion,
				Source:    "github",
				Providers: map[string]update.Provider{"github": updatetest.Provider(latestVersion)},
			},
			wantAdvisory: true,
		}, {
			name: "DerivedAppID",
			lines: []string{
				"id: root",
			},
			update: spec.UpdateOptions{
				Version:   currentVersion,
				Source:    "github",
				Providers: map[string]update.Provider{"github": updatetest.Provider(latestVersion)},
			},
			wantAdvisory: true,
		}, {
			name: "UpToDate",
			lines: []string{
				"id: root",
				"use: root",
			},
			update: spec.UpdateOptions{
				Version:   latestVersion,
				Source:    "github",
				Providers: map[string]update.Provider{"github": updatetest.Provider(latestVersion)},
			},
			wantAdvisory: false,
		}, {
			name: "ProviderError",
			lines: []string{
				"id: root",
				"use: root",
			},
			update: spec.UpdateOptions{
				Version:   currentVersion,
				Source:    "github",
				Providers: map[string]update.Provider{"github": updatetest.ErrProvider(errLookup)},
			},
			wantAdvisory: false,
		}, {
			name: "Disabled",
			lines: []string{
				"id: root",
				"use: root",
			},
			update:       spec.UpdateOptions{},
			wantAdvisory: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := strings.NewReader(yamlDoc(tc.lines...))
			opts := spec.Options{
				Builders: toBuilders(map[string]spec.Runner{"root": spectest.NoOpRunner()}),
				Update:   tc.update,
			}
			cmd, buildErr := spec.Build(input, opts)

			// Act
			rendered, renderErr := renderHelp(cmd)

			// Assert
			if got, want := buildErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("spec.Build() = %v, want %v", got, want)
			}
			if got, want := renderErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("renderHelp() = %v, want %v", got, want)
			}
			surfacesLatest := strings.Contains(rendered, latestVersion)
			if got, want := surfacesLatest, tc.wantAdvisory; !cmp.Equal(got, want) {
				t.Errorf("help surfaces %q = %v, want %v", latestVersion, got, want)
			}
		})
	}
}

// yamlDoc joins lines into a single YAML document, one element per line.
func yamlDoc(lines ...string) string {
	return strings.Join(lines, "\n")
}

// renderHelp renders cmd's help to a buffer and returns the plain-text output.
func renderHelp(cmd *cobra.Command) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(context.Background())
	if err := cmd.Help(); err != nil {
		return "", err
	}
	return plain.Render(buf.String())
}
