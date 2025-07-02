package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "free@home",
		Short: "Interact with ABB free@home devices using the local API",
		Long:  `A CLI tool to interact with ABB free@home devices using the local API.`,
	}

	// Package-level variables to store version information
	version string
	commit  string
)

func Execute(ver, com string) {
	version = ver
	commit = com

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
