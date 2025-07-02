package main

import "github.com/pgerke/freeathome/cmd/cli/cmd"

// version is the version of the application. The value will be overridden by the linker during the build process.
var version = "debug"

// commit is the commit hash of the application. The value will be overridden by the linker during the build process.
var commit = "unknown"

func main() {
	cmd.Execute(version, commit)
}
