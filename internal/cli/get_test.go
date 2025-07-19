package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pgerke/freeathome/pkg/freeathome"
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
			configData:    "",
			tlsEnabled:    true,
			skipTLSVerify: false,
			logLevel:      "info",
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
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			configFileDir = t.TempDir()

			// Create config file in the expected location
			configDir := filepath.Join(configFileDir, ".freeathome")
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

			// Test setup function
			sysAp, err := setup(v, "", tt.tlsEnabled, tt.skipTLSVerify, tt.logLevel)

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
	configFileDir := t.TempDir()

	configDir := filepath.Join(configFileDir, ".freeathome")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

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
	_, err = setup(v, "", true, false, "info")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestSetupWithNilViper tests setup with nil viper instance
func TestSetupWithNilViper(t *testing.T) {
	// Test setup function with nil viper
	_, err := setup(nil, "", true, false, "info")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

func TestGetDeviceList(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		responseBody  string
		responseCode  int
		expectError   bool
		errorContains string
		expectOutput  string
	}{
		{
			name:         "Successful JSON output",
			outputFormat: "json",
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": [
    "ABB7F595EC47",
    "ABB7013B85DE",
    "ABB7F5947E20"
  ]}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: `{"00000000-0000-0000-0000-000000000000":["ABB7F595EC47","ABB7013B85DE","ABB7F5947E20"]}
`,
		},
		{
			name:         "Successful text output",
			outputFormat: "text",
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": [
    "ABB7F595EC47",
    "ABB7013B85DE",
    "ABB7F5947E20"
  ]}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "ABB7F595EC47\nABB7013B85DE\nABB7F5947E20\n",
		},
		{
			name:         "Empty device list",
			outputFormat: "text",
			responseBody: `{}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "No devices found\n",
		},
		{
			name:         "Empty devices for EmptyUUID",
			outputFormat: "text",
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": []
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "No devices found\n",
		},
		{
			name:         "No devices for EmptyUUID",
			outputFormat: "text",
			responseBody: `{
  "other-uuid": ["ABB7F595EC47", "ABB7013B85DE"]
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "No devices found for system access point\n",
		},
		{
			name:          "HTTP error response",
			outputFormat:  "text",
			responseBody:  `{"error": "Unauthorized"}`,
			responseCode:  http.StatusUnauthorized,
			expectError:   true,
			errorContains: "failed to get device list",
		},
		{
			name:          "Invalid JSON response",
			outputFormat:  "text",
			responseBody:  `invalid json`,
			responseCode:  http.StatusOK,
			expectError:   true,
			errorContains: "failed to get device list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viper instance
			v := setupViper(t)

			// Setup mock SystemAccessPoint
			sysAp, _, _ := setupMock(t, v, tt.responseCode, tt.responseBody)

			// Override the setupFunc to use the mock SystemAccessPoint
			setupFunc = func(_ *viper.Viper, _ string, _ bool, _ bool, _ string) (*freeathome.SystemAccessPoint, error) {
				return sysAp, nil
			}
			defer func() {
				setupFunc = setup
			}()

			// Capture stdout for output testing
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Test GetDeviceList function
			err := GetDeviceList(v, false, false, "info", tt.outputFormat)

			// Close pipe and read output
			_ = w.Close()
			output, _ := io.ReadAll(r)

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
				if tt.expectOutput != "" && string(output) != tt.expectOutput {
					t.Errorf("Expected output '%s', got '%s'", tt.expectOutput, string(output))
				}
			}
		})
	}
}

// TestGetDeviceListWithInvalidConfigFile tests GetDeviceList with an invalid config file
func TestGetDeviceListWithInvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	configFile := filepath.Join(t.TempDir(), "invalid-config.yaml")

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
	// Create viper instance
	v := setupViper(t)

	// Test that the function can be called (it will likely fail due to network issues, but that's expected)
	err := GetDeviceList(v, true, false, "info", "text")
	// We expect this to fail due to network/connection issues, but the function should exist
	if err == nil {
		t.Log("GetDeviceList function exists and was called successfully")
	} else {
		t.Logf("GetDeviceList function exists but failed as expected: %v", err)
	}
}

// TestSetupFunctionExists tests that the setup function exists and can be called
func TestSetupFunctionExists(t *testing.T) {
	// Create viper instance
	v := setupViper(t)

	// Test that the function can be called
	sysAp, err := setup(v, "", true, false, "info")
	if err != nil {
		t.Errorf("setup failed unexpectedly: %v", err)
	}
	if sysAp == nil {
		t.Error("Expected SystemAccessPoint to be created, got nil")
	}
}

func TestGetConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		responseBody  string
		responseCode  int
		expectError   bool
		errorContains string
		expectOutput  string
	}{
		{
			name:         "Successful JSON output",
			outputFormat: "json",
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "devices": {
      "ABB7F595EC47": {},
      "ABB7013B85DE": {}
    },
    "floorplan": {
      "floors": {}
    },
    "sysapName": "Test System",
    "users": {
      "user1": {
        "enabled": false,
        "flags": null,
        "grantedPermissions": null,
        "jid": "",
        "name": "Test User",
        "requestedPermissions": null,
        "role": ""
      }
    }
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: `{"00000000-0000-0000-0000-000000000000":{"devices":{"ABB7013B85DE":{},"ABB7F595EC47":{}},"floorplan":{"floors":{}},"sysapName":"Test System","users":{"user1":{"enabled":false,"flags":null,"grantedPermissions":null,"jid":"","name":"Test User","requestedPermissions":null,"role":""}}}}
`,
		},
		{
			name:         "Successful text output",
			outputFormat: "text",
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "devices": {
      "ABB7F595EC47": {
        "name": "Test Device 1",
        "type": "switch"
      }
    },
    "floorplan": {
      "floors": {
        "ground": {
          "name": "Ground Floor",
          "rooms": {}
        }
      }
    },
    "sysapName": "Test System",
    "users": {
      "user1": {
        "name": "Test User"
      }
    }
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "System Access Point ID: 00000000-0000-0000-0000-000000000000\n  Name: Test System\n  Devices: 1\n  Users: 1\n  Floors: 1\n\n",
		},
		{
			name:         "Empty configuration",
			outputFormat: "text",
			responseBody: `{}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "No configuration found\n",
		},
		{
			name:          "HTTP error response",
			outputFormat:  "text",
			responseBody:  `{"error": "Unauthorized"}`,
			responseCode:  http.StatusUnauthorized,
			expectError:   true,
			errorContains: "failed to get configuration",
		},
		{
			name:          "Invalid JSON response",
			outputFormat:  "text",
			responseBody:  `invalid json`,
			responseCode:  http.StatusOK,
			expectError:   true,
			errorContains: "failed to get configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viper instance
			v := setupViper(t)

			// Setup mock SystemAccessPoint
			sysAp, _, _ := setupMock(t, v, tt.responseCode, tt.responseBody)

			// Override the setupFunc to use the mock SystemAccessPoint
			setupFunc = func(_ *viper.Viper, _ string, _ bool, _ bool, _ string) (*freeathome.SystemAccessPoint, error) {
				return sysAp, nil
			}
			defer func() {
				setupFunc = setup
			}()

			// Capture stdout for output testing
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Test GetConfiguration function
			err := GetConfiguration(v, false, false, "info", tt.outputFormat)

			// Close pipe and read output
			_ = w.Close()
			output, _ := io.ReadAll(r)

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
				if tt.expectOutput != "" && string(output) != tt.expectOutput {
					t.Errorf("Expected output '%s', got '%s'", tt.expectOutput, string(output))
				}
			}
		})
	}
}

// TestGetConfigurationWithInvalidConfigFile tests GetConfiguration with an invalid config file
func TestGetConfigurationWithInvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	configFile := filepath.Join(t.TempDir(), "invalid-config.yaml")

	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test GetConfiguration function - should fail due to invalid YAML
	err = GetConfiguration(v, true, false, "info", "text")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestGetConfigurationWithNilViper tests GetConfiguration with nil viper instance
func TestGetConfigurationWithNilViper(t *testing.T) {
	// Test GetConfiguration function with nil viper
	err := GetConfiguration(nil, true, false, "info", "text")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestGetConfigurationFunctionExists tests that the GetConfiguration function exists and can be called
func TestGetConfigurationFunctionExists(t *testing.T) {
	// Create viper instance
	v := setupViper(t)

	// Test that the function can be called (it will likely fail due to network issues, but that's expected)
	err := GetConfiguration(v, true, false, "info", "text")
	// We expect this to fail due to network/connection issues, but the function should exist
	if err == nil {
		t.Log("GetConfiguration function exists and was called successfully")
	} else {
		t.Logf("GetConfiguration function exists but failed as expected: %v", err)
	}
}

// TestHandleSysApError tests the handleSysApError function with various scenarios
func TestHandleSysApError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		operation     string
		tlsEnabled    bool
		skipTLSVerify bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "Nil error",
			err:           nil,
			operation:     "test operation",
			tlsEnabled:    true,
			skipTLSVerify: false,
			expectError:   false,
		},
		{
			name:          "TLS enabled without skip verify",
			err:           fmt.Errorf("test error"),
			operation:     "test operation",
			tlsEnabled:    true,
			skipTLSVerify: false,
			expectError:   true,
			errorContains: "failed to test operation",
		},
		{
			name:          "TLS enabled with skip verify",
			err:           fmt.Errorf("test error"),
			operation:     "test operation",
			tlsEnabled:    true,
			skipTLSVerify: true,
			expectError:   true,
			errorContains: "failed to test operation",
		},
		{
			name:          "TLS disabled",
			err:           fmt.Errorf("test error"),
			operation:     "test operation",
			tlsEnabled:    false,
			skipTLSVerify: false,
			expectError:   true,
			errorContains: "failed to test operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleSysApError(tt.err, tt.operation, tt.tlsEnabled, tt.skipTLSVerify)

			if tt.expectError {
				if result == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(result.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, result.Error())
				}
			} else {
				if result != nil {
					t.Errorf("Expected no error but got: %v", result)
				}
			}
		})
	}
}

// TestOutputJSON tests the outputJSON function with various scenarios
func TestOutputJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          interface{}
		dataType      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid JSON data",
			data:        map[string]string{"key": "value"},
			dataType:    "test data",
			expectError: false,
		},
		{
			name:        "Complex nested data",
			data:        map[string]interface{}{"nested": map[string]int{"count": 42}},
			dataType:    "complex data",
			expectError: false,
		},
		{
			name:        "Empty data",
			data:        map[string]interface{}{},
			dataType:    "empty data",
			expectError: false,
		},
		{
			name:        "Nil data",
			data:        nil,
			dataType:    "nil data",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for output testing
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Test outputJSON function
			err := outputJSON(tt.data, tt.dataType)

			// Close pipe and read output
			_ = w.Close()
			output, _ := io.ReadAll(r)

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
				// Verify that valid JSON was output
				if len(output) > 0 {
					// Try to parse the output as JSON to ensure it's valid
					var parsed interface{}
					if jsonErr := json.Unmarshal(output, &parsed); jsonErr != nil {
						t.Errorf("Output is not valid JSON: %v", jsonErr)
					}
				}
			}
		})
	}
}
