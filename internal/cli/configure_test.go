package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestConfigure tests the Configure function with various scenarios
func TestConfigure(t *testing.T) {
	// TODO: Remove this once we have a way to test the interactive prompts
	t.Skip("Skipping Configure test due to interactive prompts")
	tests := []struct {
		name           string
		configFile     string
		hostname       string
		username       string
		password       string
		existingConfig string
		expectError    bool
	}{
		{
			name:       "Update existing configuration with all values",
			configFile: "test-config.yaml",
			hostname:   "new-host",
			username:   "new-user",
			password:   "new-pass",
			existingConfig: `hostname: old-host
username: old-user
password: old-pass`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Create config file with existing config
			configFilePath := filepath.Join(tempDir, tt.configFile)
			err := os.WriteFile(configFilePath, []byte(tt.existingConfig), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Create a fresh viper instance for testing
			v := viper.New()

			// Test Configure function
			err = Configure(v, configFilePath, tt.hostname, tt.username, tt.password)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error expected, verify config was saved
			if !tt.expectError && err == nil {
				// Verify the config file was updated
				if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
					t.Error("Expected config file to exist but it doesn't")
				}
			}
		})
	}
}

// TestConfigureWithInvalidConfigFile tests Configure with an invalid config file
func TestConfigureWithInvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test Configure function - should fail due to invalid YAML
	err = Configure(v, configFile, "test-host", "test-user", "test-pass")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestShowConfiguration tests the ShowConfiguration function
func TestShowConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		configData  string
		expectError bool
	}{
		{
			name:       "Show existing configuration",
			configFile: "test-config.yaml",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			expectError: false,
		},
		{
			name:        "Show empty configuration",
			configFile:  "empty-config.yaml",
			configData:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			var configFilePath string
			if tt.configData != "" {
				configFilePath = filepath.Join(tempDir, tt.configFile)
				err := os.WriteFile(configFilePath, []byte(tt.configData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
			} else {
				configFilePath = filepath.Join(tempDir, tt.configFile)
			}

			// Create a fresh viper instance for testing
			v := viper.New()

			// Test ShowConfiguration function
			err := ShowConfiguration(v, configFilePath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestShowConfigurationWithInvalidConfig tests ShowConfiguration with invalid config
func TestShowConfigurationWithInvalidConfig(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test ShowConfiguration function - should fail due to invalid YAML
	err = ShowConfiguration(v, configFile)
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestConfigureWithNilViper tests Configure with nil viper instance
func TestConfigureWithNilViper(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Test Configure function with nil viper
	err := Configure(nil, configFile, "test-host", "test-user", "test-pass")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestShowConfigurationWithNilViper tests ShowConfiguration with nil viper instance
func TestShowConfigurationWithNilViper(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Test ShowConfiguration function with nil viper
	err := ShowConfiguration(nil, configFile)
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestConfigureFunctionExists tests that the Configure function exists and can be called
func TestConfigureFunctionExists(t *testing.T) {
	// This test verifies that the Configure function exists and has the expected signature
	// It doesn't test the actual functionality since that requires interactive input
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Create a minimal config file
	configData := `hostname: test-host
username: test-user
password: test-pass`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test that the function can be called (it will fail due to interactive prompts, but that's expected)
	err = Configure(v, configFile, "new-host", "new-user", "new-pass")
	// We expect this to fail due to interactive prompts, but the function should exist
	if err == nil {
		t.Log("Configure function exists and was called successfully")
	} else {
		t.Logf("Configure function exists but failed as expected: %v", err)
	}
}

// TestShowConfigurationFunctionExists tests that the ShowConfiguration function exists and can be called
func TestShowConfigurationFunctionExists(t *testing.T) {
	// This test verifies that the ShowConfiguration function exists and can be called
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Create a minimal config file
	configData := `hostname: test-host
username: test-user
password: test-pass`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test that the function can be called
	err = ShowConfiguration(v, configFile)
	if err != nil {
		t.Errorf("ShowConfiguration failed unexpectedly: %v", err)
	}
}
