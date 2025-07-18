package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestConfigStruct tests the Config struct fields
func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		Hostname: "test-host",
		Username: "test-user",
		Password: "test-pass",
	}

	if cfg.Hostname != "test-host" {
		t.Errorf("Expected Hostname to be 'test-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "test-user" {
		t.Errorf("Expected Username to be 'test-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "test-pass" {
		t.Errorf("Expected Password to be 'test-pass', got '%s'", cfg.Password)
	}
}

// TestConfigUpdate tests the update method
func TestConfigUpdate(t *testing.T) {
	cfg := &Config{
		Hostname: "old-host",
		Username: "old-user",
		Password: "old-pass",
	}

	// Test updating with new values
	cfg.update("new-host", "new-user", "new-pass")

	if cfg.Hostname != "new-host" {
		t.Errorf("Expected Hostname to be 'new-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "new-user" {
		t.Errorf("Expected Username to be 'new-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "new-pass" {
		t.Errorf("Expected Password to be 'new-pass', got '%s'", cfg.Password)
	}
}

// TestConfigUpdateWithEmptyValues tests that empty values don't overwrite existing values
func TestConfigUpdateWithEmptyValues(t *testing.T) {
	cfg := &Config{
		Hostname: "existing-host",
		Username: "existing-user",
		Password: "existing-pass",
	}

	// Test updating with empty values
	cfg.update("", "", "")

	if cfg.Hostname != "existing-host" {
		t.Errorf("Expected Hostname to remain 'existing-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "existing-user" {
		t.Errorf("Expected Username to remain 'existing-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "existing-pass" {
		t.Errorf("Expected Password to remain 'existing-pass', got '%s'", cfg.Password)
	}
}

// TestConfigUpdatePartial tests updating only some fields
func TestConfigUpdatePartial(t *testing.T) {
	cfg := &Config{
		Hostname: "old-host",
		Username: "old-user",
		Password: "old-pass",
	}

	// Test updating only hostname
	cfg.update("new-host", "", "")

	if cfg.Hostname != "new-host" {
		t.Errorf("Expected Hostname to be 'new-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "old-user" {
		t.Errorf("Expected Username to remain 'old-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "old-pass" {
		t.Errorf("Expected Password to remain 'old-pass', got '%s'", cfg.Password)
	}
}

// TestLoadWithNilViper tests loading configuration with nil viper instance
func TestLoadWithNilViper(t *testing.T) {
	cfg, err := load(nil, "config.yaml")
	if err == nil {
		t.Fatalf("Expected error when viper is nil, got nil")
	}
	if cfg != nil {
		t.Fatalf("Expected nil config when viper is nil, got: %v", cfg)
	}
}

// TestLoadWithValidFile tests loading configuration from a valid file
func TestLoadWithValidFile(t *testing.T) {
	configData := `hostname: test-host
username: test-user
password: test-pass`

	v := createViperWithConfig(t, configData)

	// Load configuration
	cfg, err := load(v, v.ConfigFileUsed())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	assertConfigValues(t, cfg, "test-host", "test-user", "test-pass")
}

// TestLoadWithNonExistentFile tests loading configuration when file doesn't exist
func TestLoadWithNonExistentFile(t *testing.T) {
	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration with non-existent file
	cfg, err := load(v, "/non/existent/file.yaml")
	if err == nil {
		t.Fatalf("Expected error when file doesn't exist, got nil")
	}
	if cfg != nil {
		t.Fatalf("Expected nil config when file doesn't exist, got: %v", cfg)
	}
}

// TestLoadWithEmptyString tests loading configuration with empty file path
func TestLoadWithEmptyString(t *testing.T) {
	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration with empty file path
	cfg, err := load(v, "")
	if err != nil {
		t.Fatalf("Expected no error with empty file path, got: %v", err)
	}

	// Should return empty config
	if cfg.Hostname != "" {
		t.Errorf("Expected empty Hostname, got '%s'", cfg.Hostname)
	}

	if cfg.Username != "" {
		t.Errorf("Expected empty Username, got '%s'", cfg.Username)
	}

	if cfg.Password != "" {
		t.Errorf("Expected empty Password, got '%s'", cfg.Password)
	}
}

// TestLoadWithEnvironmentVariables tests loading configuration from environment variables
func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("FREEATHOME_HOSTNAME", "env-host")
	_ = os.Setenv("FREEATHOME_USERNAME", "env-user")
	_ = os.Setenv("FREEATHOME_PASSWORD", "env-pass")
	defer func() {
		_ = os.Unsetenv("FREEATHOME_HOSTNAME")
		_ = os.Unsetenv("FREEATHOME_USERNAME")
		_ = os.Unsetenv("FREEATHOME_PASSWORD")
	}()

	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration
	cfg, err := load(v, "")
	if err != nil {
		t.Fatalf("Failed to load config from environment: %v", err)
	}

	if cfg.Hostname != "env-host" {
		t.Errorf("Expected Hostname to be 'env-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "env-user" {
		t.Errorf("Expected Username to be 'env-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "env-pass" {
		t.Errorf("Expected Password to be 'env-pass', got '%s'", cfg.Password)
	}
}

// TestLoadWithFileAndEnvironment tests that environment variables override file values
func TestLoadWithFileAndEnvironment(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("FREEATHOME_HOSTNAME", "env-host")
	_ = os.Setenv("FREEATHOME_USERNAME", "env-user")
	_ = os.Setenv("FREEATHOME_PASSWORD", "env-pass")
	defer func() {
		_ = os.Unsetenv("FREEATHOME_HOSTNAME")
		_ = os.Unsetenv("FREEATHOME_USERNAME")
		_ = os.Unsetenv("FREEATHOME_PASSWORD")
	}()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `hostname: file-host
username: file-user
password: file-pass`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration
	cfg, err := load(v, configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment variables should override file values
	if cfg.Hostname != "env-host" {
		t.Errorf("Expected Hostname to be 'env-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "env-user" {
		t.Errorf("Expected Username to be 'env-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "env-pass" {
		t.Errorf("Expected Password to be 'env-pass', got '%s'", cfg.Password)
	}
}

// TestLoadWithInvalidYAML tests loading configuration with invalid YAML
func TestLoadWithInvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	invalidYAML := `hostname: test-host
username: test-user
password: test-pass
invalid: [unclosed bracket`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration should fail
	_, err = load(v, configFile)
	if err == nil {
		t.Error("Expected error when loading invalid YAML, got none")
	}
}

// TestLoadWithUnmarshallingError tests loading configuration with YAML that cannot be unmarshalled to the Config struct
func TestLoadWithUnmarshallingError(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	invalidYAML := `hostname: test-host
username:
  - test-user
  - test-user2
password: test-pass`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration should fail
	_, err = load(v, configFile)
	if err == nil {
		t.Error("Expected error when loading invalid YAML, got none")
	}
}

// TestLoadWithPartialFile tests loading configuration with partial values in file
func TestLoadWithPartialFile(t *testing.T) {
	// Create a temporary config file with partial values
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `hostname: test-host
# username and password are missing`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Load configuration
	cfg, err := load(v, configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Hostname != "test-host" {
		t.Errorf("Expected Hostname to be 'test-host', got '%s'", cfg.Hostname)
	}

	if cfg.Username != "" {
		t.Errorf("Expected empty Username, got '%s'", cfg.Username)
	}

	if cfg.Password != "" {
		t.Errorf("Expected empty Password, got '%s'", cfg.Password)
	}
}

// TestSaveConfig tests saving configuration to file
func TestSaveConfig(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create a fresh viper instance for testing
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	cfg := &Config{
		Hostname: "test-host",
		Username: "test-user",
		Password: "test-pass",
	}

	// Save configuration
	err := cfg.save(v)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read back the config to verify
	v2 := viper.New()
	v2.SetConfigFile(configFile)
	v2.SetConfigType("yaml")

	if err := v2.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if v2.GetString("hostname") != "test-host" {
		t.Errorf("Expected saved hostname to be 'test-host', got '%s'", v2.GetString("hostname"))
	}

	if v2.GetString("username") != "test-user" {
		t.Errorf("Expected saved username to be 'test-user', got '%s'", v2.GetString("username"))
	}

	if v2.GetString("password") != "test-pass" {
		t.Errorf("Expected saved password to be 'test-pass', got '%s'", v2.GetString("password"))
	}
}

// TestSaveConfigWithEmptyValues tests saving configuration with empty values
func TestSaveConfigWithEmptyValues(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create a fresh viper instance for testing
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	cfg := &Config{
		Hostname: "",
		Username: "",
		Password: "",
	}

	// Save configuration
	err := cfg.save(v)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

// TestPrintSummary tests the printSummary method
func TestPrintSummary(t *testing.T) {
	cfg := &Config{
		Hostname: "test-host",
		Username: "test-user",
		Password: "test-pass",
	}

	// Create a fresh viper instance for testing
	v := viper.New()
	v.SetConfigFile("/test/config.yaml")

	// This test just ensures the method doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printSummary() panicked: %v", r)
		}
	}()

	cfg.printSummary(v)
}

// TestPrintSummaryWithEmptyPassword tests printSummary with empty password
func TestPrintSummaryWithEmptyPassword(t *testing.T) {
	cfg := &Config{
		Hostname: "test-host",
		Username: "test-user",
		Password: "",
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// This test just ensures the method doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printSummary() panicked: %v", r)
		}
	}()

	cfg.printSummary(v)
}

// TestInitConfig tests the initConfig function
func TestInitConfig(t *testing.T) {
	// Create a fresh viper instance for testing
	v := viper.New()

	// This test just ensures the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initConfig() panicked: %v", r)
		}
	}()

	initConfig(v)

	// Verify that viper was configured with expected settings
	if v.GetEnvPrefix() != "FREEATHOME" {
		t.Error("Expected viper env prefix to be set to FREEATHOME but got ", v.GetEnvPrefix())
	}
}
