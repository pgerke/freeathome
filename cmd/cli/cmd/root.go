package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "freehome",
	Short: "Interact with ABB free@home devices using the local API",
	Long:  `A CLI tool to interact with ABB free@home devices using the local API.`,
}

func Execute() error {
	return rootCmd.Execute()
}
