package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgerke/freeathome/pkg/freeathome"
	"github.com/pgerke/freeathome/pkg/models"
)

var (
	// TLS configuration flags
	tlsEnabled    bool
	skipTLSVerify bool

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get data from the free@home system access point",
		Long:  `Retrieve various types of data from the free@home system access point.`,
	}

	devicelistCmd = &cobra.Command{
		Use:     "devicelist",
		Aliases: []string{"dl"},
		Short:   "Get the list of devices from the system access point",
		Long:    `Retrieve and display the list of devices from the free@home system access point.`,
		RunE:    runGetDeviceList,
	}
)

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands
	getCmd.AddCommand(devicelistCmd)

	// Add TLS configuration flags
	getCmd.PersistentFlags().BoolVar(&tlsEnabled, "tls", true, "Enable TLS for connection")
	getCmd.PersistentFlags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")
}

func runGetDeviceList(cmd *cobra.Command, args []string) error {
	// Initialize configuration
	initConfig()

	// Get configuration values
	hostname := viper.GetString("hostname")
	username := viper.GetString("username")
	password := viper.GetString("password")

	// Check if configuration is complete
	if hostname == "" {
		return fmt.Errorf("hostname not configured. Run 'free@home configure' first")
	}
	if username == "" {
		return fmt.Errorf("username not configured. Run 'free@home configure' first")
	}
	if password == "" {
		return fmt.Errorf("password not configured. Run 'free@home configure' first")
	}

	// Create system access point client
	sysAp := freeathome.NewSystemAccessPoint(hostname, username, password, tlsEnabled, skipTLSVerify, false, nil)

	// Get device list
	deviceList, err := sysAp.GetDeviceList()
	if err != nil {
		// Provide helpful error message for TLS issues
		if tlsEnabled && !skipTLSVerify {
			return fmt.Errorf("failed to get device list: %w\n\nIf you're getting TLS certificate errors, try:\n  - Using --skip-tls-verify flag\n  - Using --tls=false to use HTTP instead of HTTPS", err)
		}
		return fmt.Errorf("failed to get device list: %w", err)
	}

	// Display the device list
	if deviceList == nil || len(*deviceList) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	// Get devices for the system access point (using EmptyUUID as key)
	devices, exists := (*deviceList)[models.EmptyUUID]
	if !exists {
		fmt.Println("No devices found for system access point")
		return nil
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	// Display device list
	fmt.Printf("Found %d devices:\n", len(devices))
	for i, deviceSerial := range devices {
		fmt.Printf("  %d. %s\n", i+1, deviceSerial)
	}

	return nil
}
