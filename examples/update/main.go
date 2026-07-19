// Command update-demo shows go-cli's update-availability advisory.
//
// It registers a fake update.Provider that reports a fixed, newer version and
// configures the running build as an older version, so the root --help output
// ends with an advisory. The advisory is shown only on the root command; the
// "greet" subcommand's --help omits it.
package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/update"
)

//go:embed app.yaml
var configYAML []byte

func main() {
	cli.FromBytes(configYAML,
		cli.BindRunner("update-demo.greet", &greetRunner{}),
		cli.Version("v1.0.0"),
		cli.BuildSource("demo"),
		cli.UpdateProvider("demo", &fixedProvider{version: "v2.5.0"}),
	).Execute()
}

// greetRunner backs the "greet" subcommand.
type greetRunner struct {
	name string
}

func (r *greetRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(arg.Positional("name", 0, &r.name,
		arg.Usage("Optional name to greet"),
	))
}

func (r *greetRunner) Run(_ context.Context) error {
	name := "world"
	if r.name != "" {
		name = r.name
	}
	fmt.Printf("hello, %s\n", name)
	return nil
}

// fixedProvider is an [update.Provider] that always reports a fixed version. It
// stands in for a real distribution channel so the example needs no network.
type fixedProvider struct {
	version string
}

func (p *fixedProvider) LatestVersion(context.Context) (string, error) {
	return p.version, nil
}

var (
	_ cli.Runner      = (*greetRunner)(nil)
	_ update.Provider = (*fixedProvider)(nil)
)
