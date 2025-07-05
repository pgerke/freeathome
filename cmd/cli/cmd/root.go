package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "freehome",
	Short: "Interact with ABB free@home devices using the local API",
	Long:  `A CLI tool to interact with ABB free@home devices using the local API.`,
}

func Execute() error {
	return rootCmd.Execute()
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
	_ = viper.BindEnv("hostname", "FREEATHOME_HOSTNAME")
	_ = viper.BindEnv("username", "FREEATHOME_USERNAME")
	_ = viper.BindEnv("password", "FREEATHOME_PASSWORD")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}
