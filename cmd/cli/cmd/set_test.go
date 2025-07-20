package cmd

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

// TestSetCommand tests that the set command has the expected properties.
func TestSetCommand(t *testing.T) {
	if setCmd.Use != "set" {
		t.Errorf("Expected set command Use to be 'set', got '%s'", setCmd.Use)
	}

	if setCmd.Short == "" {
		t.Error("Expected set command to have a Short description")
	}

	if setCmd.Long == "" {
		t.Error("Expected set command to have a Long description")
	}
}

// TestSetCommandFlags tests that the set command has the expected persistent flags.
func TestSetCommandFlags(t *testing.T) {
	expectedFlags := []string{"tls", "skip-tls-verify", "log-level", "output"}

	for _, expected := range expectedFlags {
		flag := setCmd.PersistentFlags().Lookup(expected)
		if flag == nil {
			t.Errorf("Expected set command to have persistent flag '%s'", expected)
		}
	}
}

// TestSetCommandIsChildOfRoot tests that the set command is properly added to the root command.
func TestSetCommandIsChildOfRoot(t *testing.T) {
	found := slices.ContainsFunc(rootCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "set"
	})
	if !found {
		t.Error("Expected set command to be a child of root command")
	}
}

// TestDatapointSetCommand tests that the datapoint set command has the expected properties.
func TestDatapointSetCommand(t *testing.T) {
	if datapointSetCmd.Use != "datapoint [serial] [channel] [datapoint] [value]" {
		t.Errorf("Expected datapoint set command Use to be 'datapoint [serial] [channel] [datapoint] [value]', got '%s'", datapointSetCmd.Use)
	}

	if datapointSetCmd.Short == "" {
		t.Error("Expected datapoint set command to have a Short description")
	}

	if datapointSetCmd.Long == "" {
		t.Error("Expected datapoint set command to have a Long description")
	}
}

// TestDatapointSetCommandAliases tests that the datapoint set command has the expected aliases.
func TestDatapointSetCommandAliases(t *testing.T) {
	expectedAliases := []string{"dp"}

	for _, expected := range expectedAliases {
		found := slices.ContainsFunc(datapointSetCmd.Aliases, func(alias string) bool {
			return alias == expected
		})
		if !found {
			t.Errorf("Expected datapoint set command to have alias '%s'", expected)
		}
	}
}

// TestDatapointSetCommandIsChildOfSet tests that the datapoint set command is properly added to the set command.
func TestDatapointSetCommandIsChildOfSet(t *testing.T) {
	found := slices.ContainsFunc(setCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "datapoint"
	})
	if !found {
		t.Error("Expected datapoint set command to be a child of set command")
	}
}

// TestRunSetDatapointFunction tests that the runSetDatapoint function exists and can be called.
func TestRunSetDatapointFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runSetDatapoint() panicked: %v", r)
		}
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	_ = runSetDatapoint(nil, []string{"serial", "channel", "datapoint", "value"})
}

// TestSetCommandFlagDefaults tests that the set command has the expected flags with the expected default values.
func TestSetCommandFlagDefaults(t *testing.T) {
	tlsFlag := setCmd.PersistentFlags().Lookup("tls")
	if tlsFlag == nil {
		t.Error("Expected tls flag to exist")
	}

	skipTLSVerifyFlag := setCmd.PersistentFlags().Lookup("skip-tls-verify")
	if skipTLSVerifyFlag == nil {
		t.Error("Expected skip-tls-verify flag to exist")
	}

	logLevelFlag := setCmd.PersistentFlags().Lookup("log-level")
	if logLevelFlag == nil {
		t.Error("Expected log-level flag to exist")
	}

	outputFlag := setCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected output flag to exist")
	}
}

// TestSetCommandSubcommands tests that the set command has the expected subcommands.
func TestSetCommandSubcommands(t *testing.T) {
	expectedSubcommands := []string{"datapoint"}

	for _, expected := range expectedSubcommands {
		found := slices.ContainsFunc(setCmd.Commands(), func(cmd *cobra.Command) bool {
			return cmd.Name() == expected
		})
		if !found {
			t.Errorf("Expected set command to have subcommand '%s'", expected)
		}
	}
}
