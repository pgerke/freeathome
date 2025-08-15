package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var configFileDir, _ = os.UserHomeDir()

// GetExecutableName returns the name of the executable
func GetExecutableName() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Base(executablePath), nil
}

// MustExecutableName returns the name of the executable or panics if it cannot be determined
func MustExecutableName() string {
	executableName, err := GetExecutableName()
	if err != nil {
		panic(err)
	}

	return executableName
}

// Config represents the CLI configuration settings
type Config struct {
	Executable string `yaml:"-"`
	Hostname   string `mapstructure:"hostname" yaml:"hostname"`
	Username   string `mapstructure:"username" yaml:"username"`
	Password   string `mapstructure:"password" yaml:"password"`
}

// CommandConfig represents the basic configuration for a command
type CommandConfig struct {
	Viper         *viper.Viper
	TLSEnabled    bool
	SkipTLSVerify bool
	LogLevel      string
}

// load loads the configuration from file and environment variables
func load(v *viper.Viper, configFile string) (*Config, error) {
	if v == nil {
		return nil, fmt.Errorf("viper is nil")
	}

	executableName, err := GetExecutableName()
	if err != nil {
		return nil, fmt.Errorf("error getting executable name: %w", err)
	}

	// Initialize viper configuration
	initConfig(v)

	// Override config file if specified
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Create config struct and unmarshal
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	cfg.Executable = executableName
	return &cfg, nil
}

// save saves the configuration to file
func (c *Config) save(v *viper.Viper) error {
	configDir := filepath.Dir(v.ConfigFileUsed())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Set values in viper
	v.Set("hostname", c.Hostname)
	v.Set("username", c.Username)
	v.Set("password", c.Password)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

// update updates the configuration with new values
func (c *Config) update(hostname, username, password string) {
	if hostname != "" {
		c.Hostname = hostname
	}
	if username != "" {
		c.Username = username
	}
	if password != "" {
		c.Password = password
	}
}

// printSummary prints a summary of the current configuration
func (c *Config) printSummary(v *viper.Viper) {
	fmt.Println("Current configuration:")
	fmt.Printf("  Hostname: %s\n", c.Hostname)
	fmt.Printf("  Username: %s\n", c.Username)
	if c.Password != "" {
		fmt.Printf("  Password: %s\n", "***")
	} else {
		fmt.Printf("  Password: %s\n", "(not set)")
	}

	if v.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n", v.ConfigFileUsed())
	} else {
		fmt.Println("Config file: (not found)")
	}
}

// initConfig initializes viper configuration
func initConfig(v *viper.Viper) {
	// Set config file name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Set default config file location
	configDir := filepath.Join(configFileDir, ".freeathome")
	v.AddConfigPath(configDir)
	v.SetConfigFile(filepath.Join(configDir, "config.yaml"))

	// Set environment variable prefix
	v.SetEnvPrefix("FREEATHOME")

	// Map environment variables to config keys
	_ = v.BindEnv("hostname")
	_ = v.BindEnv("username")
	_ = v.BindEnv("password")

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}
