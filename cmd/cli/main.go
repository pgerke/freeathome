package main

import (
	"fmt"
	"os"

	"github.com/pgerke/freeathome/cmd/cli/cmd"
	internal "github.com/pgerke/freeathome/internal"
)

// version is the version of the application. The value will be overridden by the linker during the build process.
var version = "debug"

// commit is the commit hash of the application. The value will be overridden by the linker during the build process.
var commit = "unknown"

func main() {
	internal.Version = version
	internal.Commit = commit

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
