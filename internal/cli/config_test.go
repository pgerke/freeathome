package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestConfig_Update(t *testing.T) {
	cfg := &Config{
		Hostname: "old-host",
		Username: "old-user",
		Password: "old-pass",
	}

	// Test updating with new values
	cfg.update("new-host", "new-user", "new-pass")

	if cfg.Hostname != "new-host" {
		t.Errorf("Expected hostname to be 'new-host', got '%s'", cfg.Hostname)
	}
	if cfg.Username != "new-user" {
		t.Errorf("Expected username to be 'new-user', got '%s'", cfg.Username)
	}
	if cfg.Password != "new-pass" {
		t.Errorf("Expected password to be 'new-pass', got '%s'", cfg.Password)
	}
}

func TestConfig_Update_EmptyValues(t *testing.T) {
	cfg := &Config{
		Hostname: "old-host",
		Username: "old-user",
		Password: "old-pass",
	}

	// Test updating with empty values (should not change existing values)
	cfg.update("", "", "")

	if cfg.Hostname != "old-host" {
		t.Errorf("Expected hostname to remain 'old-host', got '%s'", cfg.Hostname)
	}
	if cfg.Username != "old-user" {
		t.Errorf("Expected username to remain 'old-user', got '%s'", cfg.Username)
	}
	if cfg.Password != "old-pass" {
		t.Errorf("Expected password to remain 'old-pass', got '%s'", cfg.Password)
	}
}

func TestLoad_WithTempConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Write test config
	configContent := `hostname: "test-host"
username: "test-user"
password: "test-pass"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load config
	cfg, err := load(viper.GetViper(), configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if cfg.Hostname != "test-host" {
		t.Errorf("Expected hostname to be 'test-host', got '%s'", cfg.Hostname)
	}
	if cfg.Username != "test-user" {
		t.Errorf("Expected username to be 'test-user', got '%s'", cfg.Username)
	}
	if cfg.Password != "test-pass" {
		t.Errorf("Expected password to be 'test-pass', got '%s'", cfg.Password)
	}
}
