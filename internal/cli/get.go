package cli

import (
	"fmt"

	"github.com/pgerke/freeathome/pkg/freeathome"
	"github.com/pgerke/freeathome/pkg/models"
)

// GetDeviceList retrieves and displays the device list
func GetDeviceList(tlsEnabled, skipTLSVerify bool) error {
	// Load configuration
	cfg, err := load("")
	if err != nil {
		return err
	}

	// Check if configuration is complete
	if cfg.Hostname == "" {
		return fmt.Errorf("hostname not configured. Run 'free@home configure' first")
	}
	if cfg.Username == "" {
		return fmt.Errorf("username not configured. Run 'free@home configure' first")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password not configured. Run 'free@home configure' first")
	}

	// Create system access point client
	sysAp := freeathome.NewSystemAccessPoint(cfg.Hostname, cfg.Username, cfg.Password, tlsEnabled, skipTLSVerify, false, nil)

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
