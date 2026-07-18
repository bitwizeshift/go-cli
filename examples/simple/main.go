// Command example-cli is a self-contained showcase of the go-cli library.
//
// It models a fictional workspace manager, binding runners to an embedded YAML
// specification to demonstrate grouped commands, nested subcommands, positional
// arity, flags with and without values, named flag groups, flag constraints,
// and panic handling.
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
		cli.BindRunner("root", &rootRunner{}),
		cli.BindRunner("init", &initRunner{path: "."}),
		cli.BindRunner("status", &statusRunner{}),
		cli.BindRunner("panic", &panicRunner{}),
		cli.BindRunner("item-add", &itemAddRunner{}),
		cli.BindRunner("item-list", &itemListRunner{}),
		cli.BindRunner("item-remove", &itemRemoveRunner{}),
		cli.BindRunner("config-get", &configGetRunner{}),
		cli.BindRunner("config-set", &configSetRunner{}),
		cli.BindRunner("config-list", &configListRunner{}),
		cli.BindRunner("remote-add", &remoteAddRunner{}),
		cli.BindRunner("remote-list", &remoteListRunner{}),
		cli.BindRunner("remote-remove", &remoteRemoveRunner{}),
	).Execute()
}

// rootRunner backs the top-level command. It hosts the subcommands and shows
// usage when invoked without one. The config and remote commands are, by
// contrast, left unbound so cobra prints their help directly.
type rootRunner struct{}

func (*rootRunner) Run(context.Context, ...string) error {
	return cli.ErrUsage
}

// initRunner backs "init" and demonstrates a string flag with a value and a
// bare boolean flag, each with a shorthand, collected into one named group.
type initRunner struct {
	path  string
	force bool
}

func (ir *initRunner) RegisterArgs(cl *arg.CommandLine) {
	path := arg.AddFlag(cl, "path", &ir.path,
		arg.Shorthand("p"),
		arg.Usage("directory to create the vault in"),
	)
	force := arg.AddFlag(cl, "force", &ir.force,
		arg.Shorthand("f"),
		arg.Usage("recreate the vault if one already exists"),
	)
	arg.AddToGroup("Vault Flags", path, force)
}

func (ir *initRunner) Run(context.Context, ...string) error {
	fmt.Printf("initialized vault at %q (force=%t)\n", ir.path, ir.force)
	return nil
}

// statusRunner backs "status" and takes no flags or arguments.
type statusRunner struct{}

func (*statusRunner) Run(context.Context, ...string) error {
	fmt.Println("vault: ready (0 items, 0 remotes)")
	return nil
}

// panicRunner backs the hidden "panic" command and always crashes.
type panicRunner struct{}

func (*panicRunner) Run(context.Context, ...string) error {
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
	arg.Positional(cl, "name", 0, &iar.name,
		arg.Usage("name of the item to add"),
	)
	priority := arg.AddFlag(cl, "priority", &iar.priority,
		arg.Shorthand("p"),
		arg.Usage("priority from 1 (highest) to 9"),
	)
	pin := arg.AddFlag(cl, "pin", &iar.pin,
		arg.Usage("pin the item to the top of the list"),
	)
	tag := arg.AddFlag(cl, "tag", &iar.tags,
		arg.Shorthand("t"), arg.Usage("attach a tag; may be repeated"),
	)
	arg.AddToGroup("Scheduling Flags", priority, pin)
	arg.AddToGroup("Metadata Flags", tag)
	arg.MarkMutuallyExclusive(priority, pin)
}

func (iar *itemAddRunner) Run(context.Context, ...string) error {
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
	all := arg.AddFlag(cl, "all", &ilr.all,
		arg.Shorthand("a"),
		arg.Usage("include completed items"),
	)
	status := arg.AddFlag(cl, "status", &ilr.status,
		arg.Shorthand("s"),
		arg.Usage("show only items with this status"),
	)
	limit := arg.AddFlag(cl, "limit", &ilr.limit,
		arg.Shorthand("n"),
		arg.Usage("maximum number of items to show"),
	)
	asJSON := arg.AddFlag(cl, "json", &ilr.asJSON,
		arg.Usage("emit the listing as JSON"),
	)
	arg.AddToGroup("Filtering Flags", all, status, limit)
	arg.AddToGroup("Output Flags", asJSON)
	arg.MarkMutuallyExclusive(all, status)
}

func (ilr *itemListRunner) Run(context.Context, ...string) error {
	fmt.Printf("listing items (all=%t, status=%q, limit=%d, json=%t)\n",
		ilr.all, ilr.status, ilr.limit, ilr.asJSON)
	return nil
}

// itemRemoveRunner backs "item remove" and demonstrates a single ungrouped
// confirmation flag alongside [arg.Unmatched], which collects every remaining
// argument so the command can remove several items at once.
type itemRemoveRunner struct {
	names []string
	yes   bool
}

func (irr *itemRemoveRunner) RegisterArgs(cl *arg.CommandLine) {
	arg.Unmatched(cl, &irr.names)
	arg.AddFlag(cl, "yes", &irr.yes,
		arg.Shorthand("y"),
		arg.Usage("skip the confirmation prompt"),
	)
}

func (irr *itemRemoveRunner) Run(context.Context, ...string) error {
	fmt.Printf("removed %s (confirmed=%t)\n", strings.Join(irr.names, ", "), irr.yes)
	return nil
}

// configGetRunner backs "config get" and reads its single key from a positional
// argument.
type configGetRunner struct {
	key string
}

func (cgr *configGetRunner) RegisterArgs(cl *arg.CommandLine) {
	arg.Positional(cl, "key", 0, &cgr.key,
		arg.Usage("configuration key to read"),
	)
}

func (cgr *configGetRunner) Run(context.Context, ...string) error {
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
	arg.Positional(cl, "key", 0, &csr.key,
		arg.Usage("configuration key to set"),
	)
	arg.Positional(cl, "value", 1, &csr.value,
		arg.Usage("value to assign to the key"),
	)
}

func (csr *configSetRunner) Run(context.Context, ...string) error {
	fmt.Printf("set %s = %s\n", csr.key, csr.value)
	return nil
}

// configListRunner backs "config list" and demonstrates a lone ungrouped
// boolean flag on a nested subcommand.
type configListRunner struct {
	asJSON bool
}

func (clr *configListRunner) RegisterArgs(cl *arg.CommandLine) {
	arg.AddFlag(cl, "json", &clr.asJSON,
		arg.Usage("emit the settings as JSON"),
	)
}

func (clr *configListRunner) Run(context.Context, ...string) error {
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
	arg.Positional(cl, "name", 0, &rar.name,
		arg.Usage("name for the remote"),
	)
	arg.Positional(cl, "url", 1, &rar.url,
		arg.Usage("URL of the remote"),
	)
	token := arg.AddFlag(cl, "token", &rar.token,
		arg.Shorthand("k"),
		arg.Usage("authentication token for the remote"),
	)
	insecure := arg.AddFlag(cl, "insecure", &rar.insecure,
		arg.Usage("permit insecure TLS connections"),
	)
	arg.AddToGroup("Connection Flags", token, insecure)
	arg.MarkRequired(token)
}

func (rar *remoteAddRunner) Run(context.Context, ...string) error {
	fmt.Printf("added remote %q -> %s (insecure=%t)\n", rar.name, rar.url, rar.insecure)
	return nil
}

// remoteListRunner backs "remote list".
type remoteListRunner struct{}

func (*remoteListRunner) Run(context.Context, ...string) error {
	fmt.Println("no remotes configured")
	return nil
}

// remoteRemoveRunner backs "remote remove".
type remoteRemoveRunner struct {
	name string
	yes  bool
}

func (rrr *remoteRemoveRunner) RegisterArgs(cl *arg.CommandLine) {
	arg.Positional(cl, "name", 0, &rrr.name,
		arg.Usage("name of the remote to remove"),
	)
	arg.AddFlag(cl, "yes", &rrr.yes,
		arg.Shorthand("y"),
		arg.Usage("skip the confirmation prompt"),
	)
}

func (rrr *remoteRemoveRunner) Run(context.Context, ...string) error {
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
