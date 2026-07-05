// Package prompt reads answers to interactive prompts from the terminal.
//
// It offers plain, confirmation, masked-secret, and typed reads, each taking a
// context so a prompt can be cancelled. A [Prompter] binds the streams to read
// and write; the package-level functions use [DefaultPrompter] over the process
// standard streams. Secret reads require an interactive terminal and error
// otherwise.
package prompt
