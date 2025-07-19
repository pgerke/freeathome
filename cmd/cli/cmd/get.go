package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgerke/freeathome/internal/cli"
)

var (
	// TLS configuration flags
	tlsEnabled    bool
	skipTLSVerify bool
	// Logging configuration
	logLevel string
	// Output format configuration
	outputFormat string

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

	configurationCmd = &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config", "cfg"},
		Short:   "Get the configuration from the system access point",
		Long:    `Retrieve and display the configuration from the free@home system access point.`,
		RunE:    runGetConfiguration,
	}

	deviceCmd = &cobra.Command{
		Use:   "device [serial]",
		Short: "Get a specific device from the system access point",
		Long:  `Retrieve and display information about a specific device by its serial number.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runGetDevice,
	}
)

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands
	getCmd.AddCommand(devicelistCmd)
	getCmd.AddCommand(configurationCmd)
	getCmd.AddCommand(deviceCmd)

	// Add TLS configuration flags
	getCmd.PersistentFlags().BoolVar(&tlsEnabled, "tls", true, "Enable TLS for connection")
	getCmd.PersistentFlags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")

	// Add logging configuration flag
	getCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	// Add output format flag
	getCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "Set the output format (json, text)")
}

func runGetDeviceList(cmd *cobra.Command, args []string) error {
	return cli.GetDeviceList(viper.GetViper(), tlsEnabled, skipTLSVerify, logLevel, outputFormat)
}

func runGetConfiguration(cmd *cobra.Command, args []string) error {
	return cli.GetConfiguration(viper.GetViper(), tlsEnabled, skipTLSVerify, logLevel, outputFormat)
}

func runGetDevice(cmd *cobra.Command, args []string) error {
	return cli.GetDevice(viper.GetViper(), tlsEnabled, skipTLSVerify, logLevel, outputFormat, args[0])
}
