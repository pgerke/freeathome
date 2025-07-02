package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("free@home-cli v%s-%s\n", version, commit)
	},
	// Aliases:                []string{},
	// SuggestFor:             []string{},
	// GroupID:                "",
	// Long:                   "",
	// Example:                "",
	// ValidArgs:              []cobra.Completion{},
	// ValidArgsFunction:      nil,
	// Args:                   nil,
	// ArgAliases:             []string{},
	// BashCompletionFunction: "",
	// Deprecated:             "",
	// Annotations:            map[string]string{},
	// Version:                "",
	// PersistentPreRun: func(cmd *cobra.Command, args []string) {
	// 	panic("TODO")
	// },
	// PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
	// 	panic("TODO")
	// },
	// PreRun: func(cmd *cobra.Command, args []string) {
	// 	panic("TODO")
	// },
	// PreRunE: func(cmd *cobra.Command, args []string) error {
	// 	panic("TODO")
	// },
	// RunE: func(cmd *cobra.Command, args []string) error {
	// 	panic("TODO")
	// },
	// PostRun: func(cmd *cobra.Command, args []string) {
	// 	panic("TODO")
	// },
	// PostRunE: func(cmd *cobra.Command, args []string) error {
	// 	panic("TODO")
	// },
	// PersistentPostRun: func(cmd *cobra.Command, args []string) {
	// 	panic("TODO")
	// },
	// PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
	// 	panic("TODO")
	// },
	// FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	// CompletionOptions:          cobra.CompletionOptions{},
	// TraverseChildren:           false,
	// Hidden:                     false,
	// SilenceErrors:              false,
	// SilenceUsage:               false,
	// DisableFlagParsing:         false,
	// DisableAutoGenTag:          false,
	// DisableFlagsInUseLine:      false,
	// DisableSuggestions:         false,
	// SuggestionsMinimumDistance: 0,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
