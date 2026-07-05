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
	"github.com/bitwizeshift/go-cli/flag"
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

func (ir *initRunner) RegisterFlags(registry *flag.Registry) {
	path := flag.Add(registry, "path", &ir.path,
		flag.Shorthand("p"),
		flag.Usage("directory to create the vault in"),
	)
	force := flag.Add(registry, "force", &ir.force,
		flag.Shorthand("f"),
		flag.Usage("recreate the vault if one already exists"),
	)
	flag.AddToGroup("Vault Flags", path, force)
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
	priority int
	tags     []string
	pin      bool
}

func (iar *itemAddRunner) RegisterFlags(registry *flag.Registry) {
	priority := flag.Add(registry, "priority", &iar.priority,
		flag.Shorthand("p"),
		flag.Usage("priority from 1 (highest) to 9"),
	)
	pin := flag.Add(registry, "pin", &iar.pin,
		flag.Usage("pin the item to the top of the list"),
	)
	tag := flag.Add(registry, "tag", &iar.tags,
		flag.Shorthand("t"), flag.Usage("attach a tag; may be repeated"),
	)
	flag.AddToGroup("Scheduling Flags", priority, pin)
	flag.AddToGroup("Metadata Flags", tag)
	flag.MarkMutuallyExclusive(priority, pin)
}

func (iar *itemAddRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("added %q (priority=%d, pinned=%t, tags=[%s])\n",
		args[0], iar.priority, iar.pin, strings.Join(iar.tags, ", "))
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

func (ilr *itemListRunner) RegisterFlags(registry *flag.Registry) {
	all := flag.Add(registry, "all", &ilr.all,
		flag.Shorthand("a"),
		flag.Usage("include completed items"),
	)
	status := flag.Add(registry, "status", &ilr.status,
		flag.Shorthand("s"),
		flag.Usage("show only items with this status"),
	)
	limit := flag.Add(registry, "limit", &ilr.limit,
		flag.Shorthand("n"),
		flag.Usage("maximum number of items to show"),
	)
	asJSON := flag.Add(registry, "json", &ilr.asJSON,
		flag.Usage("emit the listing as JSON"),
	)
	flag.AddToGroup("Filtering Flags", all, status, limit)
	flag.AddToGroup("Output Flags", asJSON)
	flag.MarkMutuallyExclusive(all, status)
}

func (ilr *itemListRunner) Run(context.Context, ...string) error {
	fmt.Printf("listing items (all=%t, status=%q, limit=%d, json=%t)\n",
		ilr.all, ilr.status, ilr.limit, ilr.asJSON)
	return nil
}

// itemRemoveRunner backs "item remove" and demonstrates a single ungrouped
// confirmation flag alongside a required positional argument.
type itemRemoveRunner struct {
	yes bool
}

func (irr *itemRemoveRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "yes", &irr.yes,
		flag.Shorthand("y"),
		flag.Usage("skip the confirmation prompt"),
	)
}

func (irr *itemRemoveRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("removed %q (confirmed=%t)\n", args[0], irr.yes)
	return nil
}

// configGetRunner backs "config get".
type configGetRunner struct{}

func (*configGetRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("%s = <unset>\n", args[0])
	return nil
}

// configSetRunner backs "config set".
type configSetRunner struct{}

func (*configSetRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("set %s = %s\n", args[0], args[1])
	return nil
}

// configListRunner backs "config list" and demonstrates a lone ungrouped
// boolean flag on a nested subcommand.
type configListRunner struct {
	asJSON bool
}

func (clr *configListRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "json", &clr.asJSON,
		flag.Usage("emit the settings as JSON"),
	)
}

func (clr *configListRunner) Run(context.Context, ...string) error {
	fmt.Printf("no configuration values set (json=%t)\n", clr.asJSON)
	return nil
}

// remoteAddRunner backs "remote add" and demonstrates a required value flag
// grouped alongside a bare boolean.
type remoteAddRunner struct {
	token    string
	insecure bool
}

func (rar *remoteAddRunner) RegisterFlags(registry *flag.Registry) {
	token := flag.Add(registry, "token", &rar.token,
		flag.Shorthand("k"),
		flag.Usage("authentication token for the remote"),
	)
	insecure := flag.Add(registry, "insecure", &rar.insecure,
		flag.Usage("permit insecure TLS connections"),
	)
	flag.AddToGroup("Connection Flags", token, insecure)
	flag.MarkRequired(token)
}

func (rar *remoteAddRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("added remote %q -> %s (insecure=%t)\n", args[0], args[1], rar.insecure)
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
	yes bool
}

func (rrr *remoteRemoveRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "yes", &rrr.yes,
		flag.Shorthand("y"),
		flag.Usage("skip the confirmation prompt"),
	)
}

func (rrr *remoteRemoveRunner) Run(_ context.Context, args ...string) error {
	fmt.Printf("removed remote %q (confirmed=%t)\n", args[0], rrr.yes)
	return nil
}

var (
	_ cli.Runner     = (*rootRunner)(nil)
	_ cli.Runner     = (*statusRunner)(nil)
	_ cli.Runner     = (*panicRunner)(nil)
	_ cli.Runner     = (*configGetRunner)(nil)
	_ cli.Runner     = (*configSetRunner)(nil)
	_ cli.Runner     = (*remoteListRunner)(nil)
	_ flag.Registrar = (*initRunner)(nil)
	_ flag.Registrar = (*itemAddRunner)(nil)
	_ flag.Registrar = (*itemListRunner)(nil)
	_ flag.Registrar = (*itemRemoveRunner)(nil)
	_ flag.Registrar = (*configListRunner)(nil)
	_ flag.Registrar = (*remoteAddRunner)(nil)
	_ flag.Registrar = (*remoteRemoveRunner)(nil)
)
