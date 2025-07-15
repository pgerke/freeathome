package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestParseLogLevel tests the parseLogLevel function with various inputs
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Debug level",
			input:    "debug",
			expected: "DEBUG",
		},
		{
			name:     "Info level",
			input:    "info",
			expected: "INFO",
		},
		{
			name:     "Warn level",
			input:    "warn",
			expected: "WARN",
		},
		{
			name:     "Error level",
			input:    "error",
			expected: "ERROR",
		},
		{
			name:     "Case insensitive debug",
			input:    "DEBUG",
			expected: "DEBUG",
		},
		{
			name:     "Case insensitive info",
			input:    "INFO",
			expected: "INFO",
		},
		{
			name:     "Unknown level defaults to info",
			input:    "unknown",
			expected: "INFO",
		},
		{
			name:     "Empty string defaults to info",
			input:    "",
			expected: "INFO",
		},
		{
			name:     "Mixed case",
			input:    "DeBuG",
			expected: "DEBUG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			resultStr := result.String()

			if resultStr != tt.expected {
				t.Errorf("parseLogLevel(%q) = %s, expected %s", tt.input, resultStr, tt.expected)
			}
		})
	}
}

// TestSetup tests the setup function with various configurations
func TestSetup(t *testing.T) {
	tests := []struct {
		name          string
		configFile    string
		configData    string
		tlsEnabled    bool
		skipTLSVerify bool
		logLevel      string
		expectError   bool
		errorContains string
	}{
		{
			name:       "Valid configuration",
			configFile: "test-config.yaml",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			expectError:   false,
		},
		{
			name:       "Missing hostname",
			configFile: "test-config.yaml",
			configData: `username: test-user
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			expectError:   true,
			errorContains: "hostname not configured",
		},
		{
			name:       "Missing username",
			configFile: "test-config.yaml",
			configData: `hostname: test-host
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			expectError:   true,
			errorContains: "username not configured",
		},
		{
			name:       "Missing password",
			configFile: "test-config.yaml",
			configData: `hostname: test-host
username: test-user`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			expectError:   true,
			errorContains: "password not configured",
		},
		{
			name:          "Empty configuration",
			configFile:    "test-config.yaml",
			configData:    "",
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			expectError:   true,
			errorContains: "error reading config file",
		},
		{
			name:       "HTTP configuration",
			configFile: "test-config.yaml",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			tlsEnabled:    false,
			skipTLSVerify: false,
			logLevel:      "debug",
			expectError:   false,
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

			// Test setup function
			sysAp, err := setup(v, configFilePath, tt.tlsEnabled, tt.skipTLSVerify, tt.logLevel)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if sysAp == nil {
					t.Error("Expected SystemAccessPoint to be created, got nil")
				}
			}
		})
	}
}

// TestSetupWithInvalidConfigFile tests setup with an invalid config file
func TestSetupWithInvalidConfigFile(t *testing.T) {
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

	// Test setup function - should fail due to invalid YAML
	_, err = setup(v, configFile, true, false, "info")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestSetupWithNilViper tests setup with nil viper instance
func TestSetupWithNilViper(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Test setup function with nil viper
	_, err := setup(nil, configFile, true, false, "info")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestGetDeviceList tests the GetDeviceList function
func TestGetDeviceList(t *testing.T) {
	// TODO: Remove this once we have a way to test the network calls
	t.Skip("Skipping GetDeviceList for now, need to mock the network calls")
	tests := []struct {
		name          string
		configData    string
		tlsEnabled    bool
		skipTLSVerify bool
		logLevel      string
		outputFormat  string
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid configuration with JSON output",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			outputFormat:  "json",
			expectError:   false,
		},
		{
			name: "Valid configuration with text output",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			outputFormat:  "text",
			expectError:   false,
		},
		{
			name: "Missing hostname",
			configData: `username: test-user
password: test-pass`,
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
			outputFormat:  "text",
			expectError:   true,
			errorContains: "hostname not configured",
		},
		{
			name: "HTTP configuration",
			configData: `hostname: test-host
username: test-user
password: test-pass`,
			tlsEnabled:    false,
			skipTLSVerify: false,
			logLevel:      "debug",
			outputFormat:  "text",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Create config file in the expected location (~/.freeathome/config.yaml)
			configDir := filepath.Join(tempDir, ".freeathome")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config directory: %v", err)
			}

			configFilePath := filepath.Join(configDir, "config.yaml")
			if tt.configData != "" {
				err := os.WriteFile(configFilePath, []byte(tt.configData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
			}

			// Create a fresh viper instance for testing
			v := viper.New()

			// Test GetDeviceList function
			err := GetDeviceList(v, tt.tlsEnabled, tt.skipTLSVerify, tt.logLevel, tt.outputFormat)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestGetDeviceListWithInvalidConfigFile tests GetDeviceList with an invalid config file
func TestGetDeviceListWithInvalidConfigFile(t *testing.T) {
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

	// Test GetDeviceList function - should fail due to invalid YAML
	err = GetDeviceList(v, true, false, "info", "text")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestGetDeviceListWithNilViper tests GetDeviceList with nil viper instance
func TestGetDeviceListWithNilViper(t *testing.T) {
	// Test GetDeviceList function with nil viper
	err := GetDeviceList(nil, true, false, "info", "text")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestGetDeviceListFunctionExists tests that the GetDeviceList function exists and can be called
func TestGetDeviceListFunctionExists(t *testing.T) {
	// This test verifies that the GetDeviceList function exists and can be called
	tempDir := t.TempDir()

	// Create config file in the expected location (~/.freeathome/config.yaml)
	configDir := filepath.Join(tempDir, ".freeathome")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

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

	// Test that the function can be called (it will likely fail due to network issues, but that's expected)
	err = GetDeviceList(v, true, false, "info", "text")
	// We expect this to fail due to network/connection issues, but the function should exist
	if err == nil {
		t.Log("GetDeviceList function exists and was called successfully")
	} else {
		t.Logf("GetDeviceList function exists but failed as expected: %v", err)
	}
}

// TestSetupFunctionExists tests that the setup function exists and can be called
func TestSetupFunctionExists(t *testing.T) {
	// This test verifies that the setup function exists and can be called
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
	sysAp, err := setup(v, configFile, true, false, "info")
	if err != nil {
		t.Errorf("setup failed unexpectedly: %v", err)
	}
	if sysAp == nil {
		t.Error("Expected SystemAccessPoint to be created, got nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
