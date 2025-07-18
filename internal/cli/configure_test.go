package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestConfigure tests the Configure function with various scenarios
func TestConfigure(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		hostname       string
		username       string
		password       string
		nonInteractive bool
		existingConfig string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Update existing configuration with all values",
			configFile:     "test-config.yaml",
			hostname:       "new-host",
			username:       "new-user",
			password:       "new-pass",
			nonInteractive: true,
			existingConfig: `hostname: old-host
username: old-user
password: old-pass`,
			expectError: false,
		},
		{
			name:           "Non-interactive mode with missing hostname",
			configFile:     "test-config.yaml",
			hostname:       "",
			username:       "test-user",
			password:       "test-pass",
			nonInteractive: true,
			existingConfig: `username: old-user
password: old-pass`,
			expectError:   true,
			errorContains: "hostname is required but not provided",
		},
		{
			name:           "Non-interactive mode with missing username",
			configFile:     "test-config.yaml",
			hostname:       "test-host",
			username:       "",
			password:       "test-pass",
			nonInteractive: true,
			existingConfig: `hostname: old-host
password: old-pass`,
			expectError:   true,
			errorContains: "username is required but not provided",
		},
		{
			name:           "Non-interactive mode with missing password",
			configFile:     "test-config.yaml",
			hostname:       "test-host",
			username:       "test-user",
			password:       "",
			nonInteractive: true,
			existingConfig: `hostname: old-host
username: old-user`,
			expectError:   true,
			errorContains: "password is required but not provided",
		},
		{
			name:           "Non-interactive mode with all values provided",
			configFile:     "test-config.yaml",
			hostname:       "test-host",
			username:       "test-user",
			password:       "test-pass",
			nonInteractive: true,
			existingConfig: `hostname: old-host
username: old-user
password: old-pass`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file if existing config is provided
			var configFilePath string
			if tt.existingConfig != "" {
				configFilePath = createTestConfigFile(t, tt.existingConfig)
			} else if tt.configFile != "" {
				configFilePath = createTestConfigFile(t, "")
			}

			// Create a fresh viper instance for testing
			v := viper.New()

			// Test Configure function
			err := Configure(v, configFilePath, tt.hostname, tt.username, tt.password, tt.nonInteractive)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// If no error expected, verify config was saved
				if tt.configFile != "" {
					if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
						t.Error("Expected config file to exist but it doesn't")
					}
				}
			}
		})
	}
}

// TestConfigureWithInvalidConfigFile tests Configure with an invalid config file
func TestConfigureWithInvalidConfigFile(t *testing.T) {
	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	configFile := createTestConfigFile(t, invalidYAML)
	v := viper.New()

	// Test Configure function - should fail due to invalid YAML
	err := Configure(v, configFile, "test-host", "test-user", "test-pass", true)
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
			// Create config file
			configFilePath := createTestConfigFile(t, tt.configData)

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
	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	configFile := createTestConfigFile(t, invalidYAML)
	v := viper.New()

	// Test ShowConfiguration function - should fail due to invalid YAML
	err := ShowConfiguration(v, configFile)
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestConfigureWithNilViper tests Configure with nil viper instance
func TestConfigureWithNilViper(t *testing.T) {
	err := Configure(nil, "config.yaml", "test-host", "test-user", "test-pass", true)
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestShowConfigurationWithNilViper tests ShowConfiguration with nil viper instance
func TestShowConfigurationWithNilViper(t *testing.T) {
	err := ShowConfiguration(nil, "config.yaml")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestConfigureFunctionExists tests that the Configure function exists and can be called
func TestConfigureFunctionExists(t *testing.T) {
	configData := `hostname: test-host
username: test-user
password: test-pass`

	configFile := createTestConfigFile(t, configData)
	v := viper.New()

	// Test that the function can be called in non-interactive mode
	err := Configure(v, configFile, "new-host", "new-user", "new-pass", true)
	if err != nil {
		t.Errorf("Configure function failed unexpectedly: %v", err)
	}
}

// TestShowConfigurationFunctionExists tests that the ShowConfiguration function exists and can be called
func TestShowConfigurationFunctionExists(t *testing.T) {
	configData := `hostname: test-host
username: test-user
password: test-pass`

	configFile := createTestConfigFile(t, configData)
	v := viper.New()

	// Test that the function can be called
	err := ShowConfiguration(v, configFile)
	if err != nil {
		t.Errorf("ShowConfiguration failed unexpectedly: %v", err)
	}
}
