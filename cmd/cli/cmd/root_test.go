package cmd

import (
	"testing"
)

// TestRootCommand tests that the root command has the expected properties.
func TestRootCommand(t *testing.T) {
	// Test that root command has the expected properties
	// In test mode, the executable name might be different (e.g., "cmd.test")
	// So we just check that it's not empty
	if rootCmd.Use == "" {
		t.Error("Expected root command Use to not be empty")
	}

	if rootCmd.Short == "" {
		t.Error("Expected root command to have a Short description")
	}

	if rootCmd.Long == "" {
		t.Error("Expected root command to have a Long description")
	}
}

// TestRootCommandHasSubcommands tests that the root command has the expected subcommands.
func TestRootCommandHasSubcommands(t *testing.T) {
	// Test that root command has the expected subcommands
	expectedSubcommands := []string{"configure", "get", "version"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected root command to have subcommand '%s'", expected)
		}
	}
}

// TestExecute tests that the Execute function doesn't panic.
func TestExecute(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	_ = Execute()
}
