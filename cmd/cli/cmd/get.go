package cmd

import (
	"github.com/spf13/cobra"

	"github.com/pgerke/freeathome/internal/cli"
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
	return cli.GetDeviceList(tlsEnabled, skipTLSVerify)
}
