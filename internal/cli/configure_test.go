package cli

import (
	"fmt"
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

// TestConfigureInteractiveMode tests Configure in interactive mode with various scenarios
func TestConfigureInteractiveMode(t *testing.T) {
	tests := []struct {
		name           string
		hostname       string
		username       string
		password       string
		scanResponses  []string
		expectError    bool
		errorContains  string
		expectHostname string
		expectUsername string
		expectPassword string
	}{
		{
			name:           "Interactive mode with all values provided via prompts",
			hostname:       "",
			username:       "",
			password:       "",
			scanResponses:  []string{"new-host", "new-user", "new-pass"},
			expectError:    false,
			expectHostname: "new-host",
			expectUsername: "new-user",
			expectPassword: "new-pass",
		},
		{
			name:           "Interactive mode with partial values, complete via prompts",
			hostname:       "existing-host",
			username:       "",
			password:       "existing-pass",
			scanResponses:  []string{"", "new-user", ""},
			expectError:    false,
			expectHostname: "existing-host",
			expectUsername: "new-user",
			expectPassword: "existing-pass",
		},
		{
			name:           "Interactive mode with empty responses (keep existing)",
			hostname:       "existing-host",
			username:       "existing-user",
			password:       "existing-pass",
			scanResponses:  []string{"", "", ""},
			expectError:    false,
			expectHostname: "existing-host",
			expectUsername: "existing-user",
			expectPassword: "existing-pass",
		},
		{
			name:          "Interactive mode with scan error",
			hostname:      "existing-host",
			username:      "existing-user",
			password:      "existing-pass",
			scanResponses: []string{},
			expectError:   true,
			errorContains: "error reading input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file with initial data
			initialConfig := fmt.Sprintf(`hostname: %s
username: %s
password: %s`, tt.hostname, tt.username, tt.password)
			configFile := createTestConfigFile(t, initialConfig)

			// Create a fresh viper instance for testing
			v := viper.New()

			// Mock the scan function
			originalScanFunc := scanFunc
			defer func() { scanFunc = originalScanFunc }()

			responseIndex := 0
			scanFunc = func(a ...any) (n int, err error) {
				if responseIndex >= len(tt.scanResponses) {
					return 0, fmt.Errorf("mock scan error")
				}
				response := tt.scanResponses[responseIndex]
				responseIndex++

				// Simulate fmt.Scanln behavior
				if ptr, ok := a[0].(*string); ok {
					*ptr = response
					return 1, nil
				}
				return 0, fmt.Errorf("invalid argument type")
			}

			// Test Configure function in interactive mode
			err := Configure(v, configFile, tt.hostname, tt.username, tt.password, false)

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

				// Verify the configuration was saved with expected values
				cfg, err := load(v, configFile)
				if err != nil {
					t.Fatalf("Failed to load saved config: %v", err)
				}

				if cfg.Hostname != tt.expectHostname {
					t.Errorf("Expected hostname '%s', got '%s'", tt.expectHostname, cfg.Hostname)
				}
				if cfg.Username != tt.expectUsername {
					t.Errorf("Expected username '%s', got '%s'", tt.expectUsername, cfg.Username)
				}
				if cfg.Password != tt.expectPassword {
					t.Errorf("Expected password '%s', got '%s'", tt.expectPassword, cfg.Password)
				}
			}
		})
	}
}

// TestPromptForField tests the promptForField function directly
func TestPromptForField(t *testing.T) {
	tests := []struct {
		name          string
		displayName   string
		currentValue  string
		maskValue     bool
		scanResponse  string
		scanError     error
		expectError   bool
		errorContains string
		expectValue   string
	}{
		{
			name:         "Prompt with current value, user provides new value",
			displayName:  "Hostname",
			currentValue: "old-host",
			maskValue:    false,
			scanResponse: "new-host",
			expectValue:  "new-host",
		},
		{
			name:         "Prompt with current value, user keeps existing",
			displayName:  "Username",
			currentValue: "existing-user",
			maskValue:    false,
			scanResponse: "",
			expectValue:  "existing-user",
		},
		{
			name:         "Prompt with masked password",
			displayName:  "Password",
			currentValue: "old-pass",
			maskValue:    true,
			scanResponse: "new-pass",
			expectValue:  "new-pass",
		},
		{
			name:         "Prompt with no current value",
			displayName:  "Hostname",
			currentValue: "",
			maskValue:    false,
			scanResponse: "new-host",
			expectValue:  "new-host",
		},
		{
			name:          "Scan error",
			displayName:   "Hostname",
			currentValue:  "",
			maskValue:     false,
			scanError:     fmt.Errorf("scan error"),
			expectError:   true,
			errorContains: "error reading input",
		},
		{
			name:         "User confirms current value",
			displayName:  "Hostname",
			currentValue: "existing-host",
			maskValue:    false,
			scanError:    fmt.Errorf("unexpected newline"),
			expectValue:  "existing-host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the scan function
			originalScanFunc := scanFunc
			defer func() { scanFunc = originalScanFunc }()

			scanFunc = func(a ...any) (n int, err error) {
				if tt.scanError != nil {
					return 0, tt.scanError
				}

				if ptr, ok := a[0].(*string); ok {
					*ptr = tt.scanResponse
					return 1, nil
				}
				return 0, fmt.Errorf("invalid argument type")
			}

			// Test promptForField
			var result string
			setter := func(s string) { result = s }

			err := promptForField(tt.displayName, tt.currentValue, tt.maskValue, setter)

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
				if result != tt.expectValue {
					t.Errorf("Expected value '%s', got '%s'", tt.expectValue, result)
				}
			}
		})
	}
}

// TestPromptForValues tests the promptForValues function directly
func TestPromptForValues(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *Config
		scanResponses  []string
		expectError    bool
		errorContains  string
		expectHostname string
		expectUsername string
		expectPassword string
	}{
		{
			name: "Complete configuration via prompts",
			initialConfig: &Config{
				Hostname: "",
				Username: "",
				Password: "",
			},
			scanResponses:  []string{"test-host", "test-user", "test-pass"},
			expectHostname: "test-host",
			expectUsername: "test-user",
			expectPassword: "test-pass",
		},
		{
			name: "Partial configuration, complete via prompts",
			initialConfig: &Config{
				Hostname: "existing-host",
				Username: "",
				Password: "existing-pass",
			},
			scanResponses:  []string{"", "new-user", ""},
			expectHostname: "existing-host",
			expectUsername: "new-user",
			expectPassword: "existing-pass",
		},
		{
			name: "Keep existing values",
			initialConfig: &Config{
				Hostname: "existing-host",
				Username: "existing-user",
				Password: "existing-pass",
			},
			scanResponses:  []string{"", "", ""},
			expectHostname: "existing-host",
			expectUsername: "existing-user",
			expectPassword: "existing-pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the scan function
			originalScanFunc := scanFunc
			defer func() { scanFunc = originalScanFunc }()

			responseIndex := 0
			scanFunc = func(a ...any) (n int, err error) {
				if responseIndex >= len(tt.scanResponses) {
					return 0, fmt.Errorf("mock scan error")
				}
				response := tt.scanResponses[responseIndex]
				responseIndex++

				if ptr, ok := a[0].(*string); ok {
					*ptr = response
					return 1, nil
				}
				return 0, fmt.Errorf("invalid argument type")
			}

			// Test promptForValues
			err := promptForValues(tt.initialConfig)

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

				if tt.initialConfig.Hostname != tt.expectHostname {
					t.Errorf("Expected hostname '%s', got '%s'", tt.expectHostname, tt.initialConfig.Hostname)
				}
				if tt.initialConfig.Username != tt.expectUsername {
					t.Errorf("Expected username '%s', got '%s'", tt.expectUsername, tt.initialConfig.Username)
				}
				if tt.initialConfig.Password != tt.expectPassword {
					t.Errorf("Expected password '%s', got '%s'", tt.expectPassword, tt.initialConfig.Password)
				}
			}
		})
	}
}
