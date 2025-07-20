package cmd

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

// TestConfigureCommand tests that the configure command has the expected properties.
func TestConfigureCommand(t *testing.T) {
	if configureCmd.Use != "configure" {
		t.Errorf("Expected configure command Use to be 'configure', got '%s'", configureCmd.Use)
	}

	if configureCmd.Short == "" {
		t.Error("Expected configure command to have a Short description")
	}

	if configureCmd.Long == "" {
		t.Error("Expected configure command to have a Long description")
	}
}

// TestConfigureCommandFlags tests that the configure command has the expected flags.
func TestConfigureCommandFlags(t *testing.T) {
	expectedFlags := []string{"config", "hostname", "username", "password"}

	for _, expected := range expectedFlags {
		flag := configureCmd.Flags().Lookup(expected)
		if flag == nil {
			t.Errorf("Expected configure command to have flag '%s'", expected)
		}
	}
}

// TestConfigureCommandIsChildOfRoot tests that the configure command is properly added to the root command.
func TestConfigureCommandIsChildOfRoot(t *testing.T) {
	found := slices.ContainsFunc(rootCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "configure"
	})
	if !found {
		t.Error("Expected configure command to be a child of root command")
	}
}

// TestRunConfigureFunction tests that the runConfigure function exists and can be called.
func TestRunConfigureFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runConfigure() panicked: %v", r)
		}
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	_ = runConfigure(nil, []string{})
}

// TestShowCommand tests that the show command has the expected properties.
func TestShowCommand(t *testing.T) {
	if showCmd.Use != "show" {
		t.Errorf("Expected show command Use to be 'show', got '%s'", showCmd.Use)
	}

	if showCmd.Short == "" {
		t.Error("Expected show command to have a Short description")
	}

	if showCmd.Long == "" {
		t.Error("Expected show command to have a Long description")
	}
}

// TestShowCommandIsChildOfConfigure tests that the show command is properly added to the configure command.
func TestShowCommandIsChildOfConfigure(t *testing.T) {
	found := slices.ContainsFunc(configureCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "show"
	})
	if !found {
		t.Error("Expected show command to be a child of configure command")
	}
}

// TestRunShowFunction tests that the runShow function exists and can be called.
func TestRunShowFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runShow() panicked: %v", r)
		}
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	_ = runShow(nil, []string{})
}

// TestConfigureCommandFlagBindings tests that the configure command flags are properly bound to viper.
// This is a structural test, not testing the actual binding logic.
func TestConfigureCommandFlagBindings(t *testing.T) {
	configFlag := configureCmd.Flags().Lookup("config")
	if configFlag == nil {
		t.Error("Expected config flag to exist")
	}

	hostnameFlag := configureCmd.Flags().Lookup("hostname")
	if hostnameFlag == nil {
		t.Error("Expected hostname flag to exist")
	}

	usernameFlag := configureCmd.Flags().Lookup("username")
	if usernameFlag == nil {
		t.Error("Expected username flag to exist")
	}

	passwordFlag := configureCmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Error("Expected password flag to exist")
	}
}
