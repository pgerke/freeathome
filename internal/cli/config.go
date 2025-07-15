package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration settings
type Config struct {
	Hostname string `mapstructure:"hostname" yaml:"hostname"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

// load loads the configuration from file and environment variables
func load(v *viper.Viper, configFile string) (*Config, error) {
	if v == nil {
		return nil, fmt.Errorf("viper is nil")
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
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}
	configDir := filepath.Join(home, ".freeathome")
	v.AddConfigPath(configDir)
	v.SetConfigFile(filepath.Join(configDir, "config.yaml"))

	// Set environment variable prefix
	v.SetEnvPrefix("FREEATHOME")
	v.AutomaticEnv()

	// Map environment variables to config keys
	_ = v.BindEnv("hostname", "HOSTNAME")
	_ = v.BindEnv("username", "USERNAME")
	_ = v.BindEnv("password", "PASSWORD")

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}
