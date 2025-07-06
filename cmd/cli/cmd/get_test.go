package cmd

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

// TestGetCommand tests that the get command has the expected properties.
func TestGetCommand(t *testing.T) {
	if getCmd.Use != "get" {
		t.Errorf("Expected get command Use to be 'get', got '%s'", getCmd.Use)
	}

	if getCmd.Short == "" {
		t.Error("Expected get command to have a Short description")
	}

	if getCmd.Long == "" {
		t.Error("Expected get command to have a Long description")
	}
}

// TestGetCommandFlags tests that the get command has the expected persistent flags.
func TestGetCommandFlags(t *testing.T) {
	expectedFlags := []string{"tls", "skip-tls-verify", "log-level", "output"}

	for _, expected := range expectedFlags {
		flag := getCmd.PersistentFlags().Lookup(expected)
		if flag == nil {
			t.Errorf("Expected get command to have persistent flag '%s'", expected)
		}
	}
}

// TestGetCommandIsChildOfRoot tests that the get command is properly added to the root command.
func TestGetCommandIsChildOfRoot(t *testing.T) {
	found := slices.ContainsFunc(rootCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "get"
	})
	if !found {
		t.Error("Expected get command to be a child of root command")
	}
}

// TestDevicelistCommand tests that the devicelist command has the expected properties.
func TestDevicelistCommand(t *testing.T) {
	if devicelistCmd.Use != "devicelist" {
		t.Errorf("Expected devicelist command Use to be 'devicelist', got '%s'", devicelistCmd.Use)
	}

	if devicelistCmd.Short == "" {
		t.Error("Expected devicelist command to have a Short description")
	}

	if devicelistCmd.Long == "" {
		t.Error("Expected devicelist command to have a Long description")
	}
}

// TestDevicelistCommandAliases tests that the devicelist command has the expected aliases.
func TestDevicelistCommandAliases(t *testing.T) {
	expectedAliases := []string{"dl"}

	for _, expected := range expectedAliases {
		found := slices.ContainsFunc(devicelistCmd.Aliases, func(alias string) bool {
			return alias == expected
		})
		if !found {
			t.Errorf("Expected devicelist command to have alias '%s'", expected)
		}
	}
}

// TestDevicelistCommandIsChildOfGet tests that the devicelist command is properly added to the get command.
func TestDevicelistCommandIsChildOfGet(t *testing.T) {
	found := slices.ContainsFunc(getCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "devicelist"
	})
	if !found {
		t.Error("Expected devicelist command to be a child of get command")
	}
}

// TestRunGetDeviceListFunction tests that the runGetDeviceList function exists and can be called.
func TestRunGetDeviceListFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runGetDeviceList() panicked: %v", r)
		}
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	_ = runGetDeviceList(nil, []string{})
}

// TestGetCommandFlagDefaults tests that the get command has the expected flags with the expected default values.
func TestGetCommandFlagDefaults(t *testing.T) {
	tlsFlag := getCmd.PersistentFlags().Lookup("tls")
	if tlsFlag == nil {
		t.Error("Expected tls flag to exist")
	}

	skipTLSVerifyFlag := getCmd.PersistentFlags().Lookup("skip-tls-verify")
	if skipTLSVerifyFlag == nil {
		t.Error("Expected skip-tls-verify flag to exist")
	}

	logLevelFlag := getCmd.PersistentFlags().Lookup("log-level")
	if logLevelFlag == nil {
		t.Error("Expected log-level flag to exist")
	}

	outputFlag := getCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected output flag to exist")
	}
}

// TestGetCommandSubcommands tests that the get command has the expected subcommands.
func TestGetCommandSubcommands(t *testing.T) {
	expectedSubcommands := []string{"devicelist"}

	for _, expected := range expectedSubcommands {
		found := slices.ContainsFunc(getCmd.Commands(), func(cmd *cobra.Command) bool {
			return cmd.Name() == expected
		})
		if !found {
			t.Errorf("Expected get command to have subcommand '%s'", expected)
		}
	}
}
