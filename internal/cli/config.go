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
func load(configFile string) (*Config, error) {
	// Initialize viper configuration
	initConfig()

	// Override config file if specified
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Create config struct and unmarshal
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// save saves the configuration to file
func (c *Config) save() error {
	configDir := filepath.Dir(viper.ConfigFileUsed())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Set values in viper
	viper.Set("hostname", c.Hostname)
	viper.Set("username", c.Username)
	viper.Set("password", c.Password)

	if err := viper.WriteConfig(); err != nil {
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
func (c *Config) printSummary() {
	fmt.Println("Current configuration:")
	fmt.Printf("  Hostname: %s\n", c.Hostname)
	fmt.Printf("  Username: %s\n", c.Username)
	if c.Password != "" {
		fmt.Printf("  Password: %s\n", "***")
	} else {
		fmt.Printf("  Password: %s\n", "(not set)")
	}

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("Config file: (not found)")
	}
}

// initConfig initializes viper configuration
func initConfig() {
	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set default config file location
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}
	configDir := filepath.Join(home, ".freeathome")
	viper.AddConfigPath(configDir)
	viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))

	// Set environment variable prefix
	viper.SetEnvPrefix("FREEATHOME")
	viper.AutomaticEnv()

	// Map environment variables to config keys
	_ = viper.BindEnv("hostname", "HOSTNAME")
	_ = viper.BindEnv("username", "USERNAME")
	_ = viper.BindEnv("password", "PASSWORD")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}
