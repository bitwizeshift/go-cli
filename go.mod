module rodusek.dev/pkg/cli

go 1.22.0

toolchain go1.23.1

// Build Dependencies
require (
	github.com/google/uuid v1.6.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.6
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3
	golang.org/x/term v0.29.0
)

// Test dependencies
require github.com/google/go-cmp v0.6.0

// Indirect dependencies
require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)
