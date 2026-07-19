// Command example-cli is a self-contained showcase of the go-cli library.
//
// It models a fictional workspace manager, binding runners to an embedded YAML
// specification to demonstrate grouped commands, nested subcommands, required
// and optional positional arguments, flags with and without values, named flag
// groups, flag constraints, and panic handling.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/arg"
)

//go:embed app.yaml
var configYAML []byte

func main() {
	cli.FromBytes(configYAML,
		cli.BindRunner("example-cli", &rootRunner{}),
		cli.BindRunner("example-cli.init", &initRunner{path: "."}),
		cli.BindRunner("example-cli.status", &statusRunner{}),
		cli.BindRunner("example-cli.panic", &panicRunner{}),
		cli.BindRunner("example-cli.add", &itemAddRunner{}),
		cli.BindRunner("example-cli.list", &itemListRunner{}),
		cli.BindRunner("example-cli.remove", &itemRemoveRunner{}),
		cli.BindRunner("example-cli.config.get", &configGetRunner{}),
		cli.BindRunner("example-cli.config.set", &configSetRunner{}),
		cli.BindRunner("example-cli.config.list", &configListRunner{}),
		cli.BindRunner("example-cli.remote.add", &remoteAddRunner{}),
		cli.BindRunner("example-cli.remote.list", &remoteListRunner{}),
		cli.BindRunner("example-cli.remote.remove", &remoteRemoveRunner{}),
	).Execute()
}

// rootRunner backs the top-level command. It hosts the subcommands and shows
// usage when invoked without one. The config and remote commands are, by
// contrast, left unbound so cobra prints their help directly.
type rootRunner struct{}

func (*rootRunner) Run(context.Context) error {
	return cli.ErrUsage
}

// initRunner backs "init" and demonstrates a string flag with a value and a
// bare boolean flag, each with a shorthand, collected into one named group.
type initRunner struct {
	path  string
	force bool
}

func (ir *initRunner) RegisterArgs(cl *arg.CommandLine) {
	path := arg.Flag("path", &ir.path,
		arg.Shorthand("p"),
		arg.Usage("directory to create the vault in"),
	)
	force := arg.Flag("force", &ir.force,
		arg.Shorthand("f"),
		arg.Usage("recreate the vault if one already exists"),
	)
	cl.Add(path, force)
	arg.Group("Vault Flags", path, force)
}

func (ir *initRunner) Run(context.Context) error {
	fmt.Printf("initialized vault at %q (force=%t)\n", ir.path, ir.force)
	return nil
}

// statusRunner backs "status" and takes no flags or arguments.
type statusRunner struct{}

func (*statusRunner) Run(context.Context) error {
	fmt.Println("vault: ready (0 items, 0 remotes)")
	return nil
}

// panicRunner backs the hidden "panic" command and always crashes.
type panicRunner struct{}

func (*panicRunner) Run(context.Context) error {
	panic("the vault collapsed in on itself")
}

// itemAddRunner backs "item add" and demonstrates a value flag, a repeatable
// slice flag, and a bare boolean spread across two named flag groups, with a
// mutually-exclusive constraint between two of them.
type itemAddRunner struct {
	name     string
	priority int
	tags     []string
	pin      bool
}

func (iar *itemAddRunner) RegisterArgs(cl *arg.CommandLine) {
	name := arg.Positional("name", 0, &iar.name,
		arg.Usage("name of the item to add"),
	)
	priority := arg.Flag("priority", &iar.priority,
		arg.Shorthand("p"),
		arg.Usage("priority from 1 (highest) to 9"),
	)
	pin := arg.Flag("pin", &iar.pin,
		arg.Usage("pin the item to the top of the list"),
	)
	tag := arg.Flag("tag", &iar.tags,
		arg.Shorthand("t"), arg.Usage("attach a tag; may be repeated"),
	)
	cl.Add(name, priority, pin, tag)
	arg.Group("Scheduling Flags", priority, pin)
	arg.Group("Metadata Flags", tag)
	arg.MarkMutuallyExclusive(priority, pin)
}

func (iar *itemAddRunner) Run(context.Context) error {
	name := iar.name
	if name == "" {
		name = "untitled"
	}
	fmt.Printf("added %q (priority=%d, pinned=%t, tags=[%s])\n",
		name, iar.priority, iar.pin, strings.Join(iar.tags, ", "))
	return nil
}

// itemListRunner backs "item list" and demonstrates several boolean and value
// flags split between a filtering group and an output group, two of which are
// mutually exclusive.
type itemListRunner struct {
	all    bool
	status string
	limit  int
	asJSON bool
}

func (ilr *itemListRunner) RegisterArgs(cl *arg.CommandLine) {
	all := arg.Flag("all", &ilr.all,
		arg.Shorthand("a"),
		arg.Usage("include completed items"),
	)
	status := arg.Flag("status", &ilr.status,
		arg.Shorthand("s"),
		arg.Usage("show only items with this status"),
	)
	limit := arg.Flag("limit", &ilr.limit,
		arg.Shorthand("n"),
		arg.Usage("maximum number of items to show"),
	)
	asJSON := arg.Flag("json", &ilr.asJSON,
		arg.Usage("emit the listing as JSON"),
	)
	cl.Add(all, status, limit, asJSON)
	arg.Group("Filtering Flags", all, status, limit)
	arg.Group("Output Flags", asJSON)
	arg.MarkMutuallyExclusive(all, status)
}

func (ilr *itemListRunner) Run(context.Context) error {
	fmt.Printf("listing items (all=%t, status=%q, limit=%d, json=%t)\n",
		ilr.all, ilr.status, ilr.limit, ilr.asJSON)
	return nil
}

// itemRemoveRunner backs "item remove" and demonstrates a single ungrouped
// confirmation flag alongside [arg.Unmatched], which collects every remaining
// argument so the command can remove several items at once. The unmatched
// binding carries the same options a flag or positional does, so it documents
// itself in help, offers completion, and falls back to $SIMPLE_ITEMS when no
// item is named.
type itemRemoveRunner struct {
	names []string
	yes   bool
}

func (irr *itemRemoveRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(
		arg.Unmatched("names", &irr.names,
			arg.Usage("names of the items to remove"),
			arg.Type("name"),
			arg.CompleteFrom("alpha", "beta", "gamma"),
			arg.DefaultFromEnv("SIMPLE_ITEMS"),
		),
		arg.Flag("yes", &irr.yes,
			arg.Shorthand("y"),
			arg.Usage("skip the confirmation prompt"),
		),
	)
}

func (irr *itemRemoveRunner) Run(context.Context) error {
	fmt.Printf("removed %s (confirmed=%t)\n", strings.Join(irr.names, ", "), irr.yes)
	return nil
}

// configGetRunner backs "config get" and reads its single key from a positional
// argument.
type configGetRunner struct {
	key string
}

func (cgr *configGetRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(arg.Positional("key", 0, &cgr.key,
		arg.Usage("configuration key to read"),
		arg.Required(),
	))
}

func (cgr *configGetRunner) Run(context.Context) error {
	fmt.Printf("%s = <unset>\n", cgr.key)
	return nil
}

// configSetRunner backs "config set" and demonstrates two ordered positional
// arguments decoded into distinct fields.
type configSetRunner struct {
	key   string
	value string
}

func (csr *configSetRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(
		arg.Positional("key", 0, &csr.key,
			arg.Usage("configuration key to set"),
			arg.Required(),
		),
		arg.Positional("value", 1, &csr.value,
			arg.Usage("value to assign to the key"),
			arg.Required(),
		),
	)
}

func (csr *configSetRunner) Run(context.Context) error {
	fmt.Printf("set %s = %s\n", csr.key, csr.value)
	return nil
}

// configListRunner backs "config list" and demonstrates a lone ungrouped
// boolean flag on a nested subcommand.
type configListRunner struct {
	asJSON bool
}

func (clr *configListRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(arg.Flag("json", &clr.asJSON,
		arg.Usage("emit the settings as JSON"),
	))
}

func (clr *configListRunner) Run(context.Context) error {
	fmt.Printf("no configuration values set (json=%t)\n", clr.asJSON)
	return nil
}

// remoteAddRunner backs "remote add" and demonstrates a required value flag
// grouped alongside a bare boolean.
type remoteAddRunner struct {
	name     string
	url      string
	token    string
	insecure bool
}

func (rar *remoteAddRunner) RegisterArgs(cl *arg.CommandLine) {
	name := arg.Positional("name", 0, &rar.name,
		arg.Usage("name for the remote"),
		arg.Required(),
	)
	url := arg.Positional("url", 1, &rar.url,
		arg.Usage("URL of the remote"),
		arg.Required(),
	)
	token := arg.Flag("token", &rar.token,
		arg.Shorthand("k"),
		arg.Usage("authentication token for the remote"),
	)
	insecure := arg.Flag("insecure", &rar.insecure,
		arg.Usage("permit insecure TLS connections"),
	)
	cl.Add(name, url, token, insecure)
	arg.Group("Connection Flags", token, insecure)
	arg.MarkRequired(token)
}

func (rar *remoteAddRunner) Run(context.Context) error {
	fmt.Printf("added remote %q -> %s (insecure=%t)\n", rar.name, rar.url, rar.insecure)
	return nil
}

// remoteListRunner backs "remote list".
type remoteListRunner struct{}

func (*remoteListRunner) Run(context.Context) error {
	fmt.Println("no remotes configured")
	return nil
}

// remoteRemoveRunner backs "remote remove".
type remoteRemoveRunner struct {
	name string
	yes  bool
}

func (rrr *remoteRemoveRunner) RegisterArgs(cl *arg.CommandLine) {
	name := arg.Positional("name", 0, &rrr.name,
		arg.Usage("name of the remote to remove"),
		arg.Required(),
	)
	yes := arg.Flag("yes", &rrr.yes,
		arg.Shorthand("y"),
		arg.Usage("skip the confirmation prompt"),
	)
	cl.Add(name, yes)
}

func (rrr *remoteRemoveRunner) Run(context.Context) error {
	fmt.Printf("removed remote %q (confirmed=%t)\n", rrr.name, rrr.yes)
	return nil
}

var (
	_ cli.Runner    = (*rootRunner)(nil)
	_ cli.Runner    = (*statusRunner)(nil)
	_ cli.Runner    = (*panicRunner)(nil)
	_ cli.Runner    = (*remoteListRunner)(nil)
	_ arg.Registrar = (*initRunner)(nil)
	_ arg.Registrar = (*itemAddRunner)(nil)
	_ arg.Registrar = (*itemListRunner)(nil)
	_ arg.Registrar = (*itemRemoveRunner)(nil)
	_ arg.Registrar = (*configGetRunner)(nil)
	_ arg.Registrar = (*configSetRunner)(nil)
	_ arg.Registrar = (*configListRunner)(nil)
	_ arg.Registrar = (*remoteAddRunner)(nil)
	_ arg.Registrar = (*remoteRemoveRunner)(nil)
)
