package usagetest

import "github.com/spf13/cobra"

// Case is a single named usage-rendering scenario.
type Case struct {
	Name    string
	Command *cobra.Command
}

// Cases returns the usage-rendering scenarios shared by the golden test and its
// generator: a root command and a nested subcommand.
func Cases() []Case {
	return []Case{
		{Name: "root.golden.txt", Command: root()},
		{Name: "sub.golden.txt", Command: subcommand()},
	}
}

// root returns a bare root command.
func root() *cobra.Command {
	return &cobra.Command{Use: "app"}
}

// subcommand returns a subcommand nested beneath a root command.
func subcommand() *cobra.Command {
	root := &cobra.Command{Use: "app"}
	sub := &cobra.Command{Use: "sub"}
	root.AddCommand(sub)
	return sub
}
