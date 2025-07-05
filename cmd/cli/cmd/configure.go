package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Configuration file path
	cfgFile string

	// Configuration values
	hostname string
	username string
	password string

	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure connection settings for free@home system access point",
		Long: `Configure the connection settings for your free@home system access point.
This command allows you to set the hostname, username, and password
using flags, environment variables, or a YAML configuration file.

Examples:
  free@home configure --hostname 192.168.1.100 --username admin --password mypass
  free@home configure --config ~/.freeathome/config.yaml
  export FREEATHOME_HOSTNAME=192.168.1.100
  export FREEATHOME_USERNAME=admin
  export FREEATHOME_PASSWORD=mypass
  free@home configure`,
		RunE: runConfigure,
	}

	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration settings",
		Long:  `Display the current configuration settings for the free@home system access point.`,
		RunE:  runShow,
	}
)

func init() {
	rootCmd.AddCommand(configureCmd)

	// Add subcommands
	configureCmd.AddCommand(showCmd)

	// Add flags
	configureCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.freeathome/config.yaml)")
	configureCmd.Flags().StringVar(&hostname, "hostname", "", "free@home system hostname or IP address")
	configureCmd.Flags().StringVar(&username, "username", "", "username for authentication")
	configureCmd.Flags().StringVar(&password, "password", "", "password for authentication")

	// Bind flags to viper
	_ = viper.BindPFlag("hostname", configureCmd.Flags().Lookup("hostname"))
	_ = viper.BindPFlag("username", configureCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("password", configureCmd.Flags().Lookup("password"))
}

func runConfigure(cmd *cobra.Command, args []string) error {
	// Initialize configuration
	initConfig()

	// Override config file if specified
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Get current values from flags, env vars, or existing config
	currentHostname := viper.GetString("hostname")
	currentUsername := viper.GetString("username")
	currentPassword := viper.GetString("password")

	// Get new values from flags
	newHostname := hostname
	newUsername := username
	newPassword := password

	// Interactive prompts with current value display (AWS CLI style)
	if newHostname == "" {
		if currentHostname != "" {
			fmt.Printf("Hostname/IP address [%s]: ", currentHostname)
		} else {
			fmt.Print("Hostname/IP address: ")
		}
		_, err := fmt.Scanln(&newHostname)
		if err != nil {
			return fmt.Errorf("error reading hostname: %w", err)
		}
		if newHostname == "" {
			newHostname = currentHostname
		}
		viper.Set("hostname", newHostname)
	}

	if newUsername == "" {
		if currentUsername != "" {
			fmt.Printf("Username [%s]: ", currentUsername)
		} else {
			fmt.Print("Username: ")
		}
		_, err := fmt.Scanln(&newUsername)
		if err != nil {
			return fmt.Errorf("error reading username: %w", err)
		}
		if newUsername == "" {
			newUsername = currentUsername
		}
		viper.Set("username", newUsername)
	}

	if newPassword == "" {
		if currentPassword != "" {
			fmt.Print("Password [***]: ")
		} else {
			fmt.Print("Password: ")
		}
		_, err := fmt.Scanln(&newPassword)
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		if newPassword == "" {
			newPassword = currentPassword
		}
		viper.Set("password", newPassword)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(viper.ConfigFileUsed())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write configuration to file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("Configuration saved to: %s\n", viper.ConfigFileUsed())
	fmt.Printf("Hostname: %s\n", newHostname)
	fmt.Printf("Username: %s\n", newUsername)
	if newPassword != "" {
		fmt.Printf("Password: %s\n", "***")
	}

	return nil
}

func runShow(cmd *cobra.Command, args []string) error {
	// Initialize configuration
	initConfig()

	// Override config file if specified
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Get current values
	currentHostname := viper.GetString("hostname")
	currentUsername := viper.GetString("username")
	currentPassword := viper.GetString("password")

	fmt.Println("Current configuration:")
	fmt.Printf("  Hostname: %s\n", currentHostname)
	fmt.Printf("  Username: %s\n", currentUsername)
	if currentPassword != "" {
		fmt.Printf("  Password: %s\n", "***")
	} else {
		fmt.Printf("  Password: %s\n", "(not set)")
	}

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("Config file: (not found)")
	}

	return nil
}

// GetConfig returns the current configuration values
func GetConfig() (string, string, string) {
	return viper.GetString("hostname"), viper.GetString("username"), viper.GetString("password")
}
