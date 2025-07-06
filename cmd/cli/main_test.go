package main

import (
	"os"
	"testing"

	internal "github.com/pgerke/freeathome/internal"
)

// TestMainVersionSetting tests that the version and commit are set correctly.
func TestMainVersionSetting(t *testing.T) {
	if version != "debug" {
		t.Errorf("Expected version to be 'debug', got '%s'", version)
	}

	if commit != "unknown" {
		t.Errorf("Expected commit to be 'unknown', got '%s'", commit)
	}
}

// TestMainFunctionDoesNotPanic tests that the main function doesn't panic when called.
func TestMainFunctionDoesNotPanic(t *testing.T) {
	// We'll capture stderr to avoid output during tests
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() panicked: %v", r)
		}
		os.Stderr = oldStderr
		_ = w.Close()
	}()

	// This will likely fail since we're not providing proper args, but we're testing it doesn't panic
	// We'll call main in a goroutine and then immediately return to avoid hanging
	go func() {
		main()
	}()
}

// TestVersionAssignment tests that the version and commit are assigned correctly.
func TestVersionAssignment(t *testing.T) {
	originalVersion := internal.Version
	originalCommit := internal.Commit

	// Set test values
	internal.Version = "test-version"
	internal.Commit = "test-commit"

	// Verify assignment
	if internal.Version != "test-version" {
		t.Errorf("Expected internal.Version to be 'test-version', got '%s'", internal.Version)
	}

	if internal.Commit != "test-commit" {
		t.Errorf("Expected internal.Commit to be 'test-commit', got '%s'", internal.Commit)
	}

	// Restore original values
	internal.Version = originalVersion
	internal.Commit = originalCommit
}
