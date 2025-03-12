package main

import (
	"fmt"
	"log/slog"
	"os"
)

var version = "debug" // Version will be overridden by the linker during build

// main is the entry point for the application.
func main() {
	outLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	outLogger.Info("Starting free@home", "version", version)
	fmt.Println("Hello, free@home!")
}
