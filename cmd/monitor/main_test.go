package main

import (
	"os"
	"testing"
)

// TestMonitor_Main tests the main function of the monitor package if the environment variable RUN_MAIN is set to "1".
func TestMonitor_Main(t *testing.T) {
	if os.Getenv("RUN_MAIN") != "1" {
		t.Skip("Skipping main test")
		return
	}
	main()
}

// TestMonitor_LookupEnvs tests the lookupEnvs function to ensure it correctly retrieves and validates environment variables.
func TestMonitor_LookupEnvs(t *testing.T) {
	// Backup original environment variables
	originalEnv := map[string]string{
		"SYSAP_HOST":     os.Getenv("SYSAP_HOST"),
		"SYSAP_USER_ID":  os.Getenv("SYSAP_USER_ID"),
		"SYSAP_PASSWORD": os.Getenv("SYSAP_PASSWORD"),
	}
	// Restore environment variables after test
	defer func() {
		for k, v := range originalEnv {
			var err error
			if v == "" {
				err = os.Unsetenv(k)
			} else {
				err = os.Setenv(k, v)
			}
			if err != nil {
				t.Fatalf("failed to restore environment variable %s: %v", k, err)
			}
		}
	}()

	// Define test cases
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedHost   string
		expectedUser   string
		expectedPass   string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "All environment variables set",
			envVars: map[string]string{
				"SYSAP_HOST":     "localhost",
				"SYSAP_USER_ID":  "admin",
				"SYSAP_PASSWORD": "password",
			},
			expectedHost: "localhost",
			expectedUser: "admin",
			expectedPass: "password",
			expectError:  false,
		},
		{
			name: "Missing SYSAP_HOST",
			envVars: map[string]string{
				"SYSAP_USER_ID":  "admin",
				"SYSAP_PASSWORD": "password",
			},
			expectError:    true,
			expectedErrMsg: "SYSAP_HOST variable is not set",
		},
		{
			name: "Missing SYSAP_USER_ID",
			envVars: map[string]string{
				"SYSAP_HOST":     "localhost",
				"SYSAP_PASSWORD": "password",
			},
			expectError:    true,
			expectedErrMsg: "SYSAP_USER_ID variable is not set",
		},
		{
			name: "Missing SYSAP_PASSWORD",
			envVars: map[string]string{
				"SYSAP_HOST":    "localhost",
				"SYSAP_USER_ID": "admin",
			},
			expectError:    true,
			expectedErrMsg: "SYSAP_PASSWORD variable is not set",
		},
	}

	// Run the tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range test.envVars {
				err := os.Setenv(key, value)
				if err != nil {
					t.Fatalf("failed to set environment variable %s: %v", key, err)
				}
			}

			// Unset environment variables after test
			defer func() {
				for key := range test.envVars {
					err := os.Unsetenv(key)
					if err != nil {
						t.Fatalf("failed to unset environment variable %s: %v", key, err)
					}
				}
			}()

			host, user, password, err := lookupEnvs()

			if test.expectError {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				if err.Error() != test.expectedErrMsg {
					t.Fatalf("expected error message '%s', got '%s'", test.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect an error but got: %v", err)
				}
				if host != test.expectedHost {
					t.Errorf("expected host '%s', got '%s'", test.expectedHost, host)
				}
				if user != test.expectedUser {
					t.Errorf("expected user '%s', got '%s'", test.expectedUser, user)
				}
				if password != test.expectedPass {
					t.Errorf("expected password '%s', got '%s'", test.expectedPass, password)
				}
			}
		})
	}
}
