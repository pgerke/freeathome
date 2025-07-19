package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pgerke/freeathome/pkg/freeathome"
	"github.com/pgerke/freeathome/pkg/models"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var setupFunc = setup

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Default to info level for unknown values
		return slog.LevelInfo
	}
}

// handleSysApError provides consistent error handling for system access point operations
func handleSysApError(err error, operation string, tlsEnabled, skipTLSVerify bool) error {
	if err == nil {
		return nil
	}

	// Provide helpful error message for TLS issues
	if tlsEnabled && !skipTLSVerify {
		return fmt.Errorf("failed to %s: %w\n\nIf you're getting TLS certificate errors, try:\n  - Using --skip-tls-verify flag\n  - Using --tls=false to use HTTP instead of HTTPS", operation, err)
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// outputJSON provides consistent JSON output formatting for system access point operations
func outputJSON(data any, dataType string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal %s to JSON: %w", dataType, err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func setup(v *viper.Viper, configFile string, tlsEnabled, skipTLSVerify bool, logLevel string) (*freeathome.SystemAccessPoint, error) {
	// Load configuration
	cfg, err := load(v, configFile)
	if err != nil {
		return nil, err
	}

	// Check if configuration is complete
	if cfg.Hostname == "" {
		return nil, fmt.Errorf("hostname not configured. Run 'free@home configure' first")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("username not configured. Run 'free@home configure' first")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("password not configured. Run 'free@home configure' first")
	}

	// Create a new logger with the specified options
	// Use a colorized handler if the terminal supports colors
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		color.NoColor = true
	}
	handler := freeathome.NewColorHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(logLevel),
	})
	logger := freeathome.NewDefaultLogger(handler)

	// Create system access point client
	config := freeathome.NewConfig(cfg.Hostname, cfg.Username, cfg.Password)
	config.TLSEnabled = tlsEnabled
	config.SkipTLSVerify = skipTLSVerify
	config.Logger = logger
	sysAp := freeathome.NewSystemAccessPoint(config)

	return sysAp, nil
}

// GetDeviceList retrieves and displays the device list
func GetDeviceList(v *viper.Viper, tlsEnabled, skipTLSVerify bool, logLevel string, outputFormat string) error {
	// Setup system access point
	sysAp, err := setupFunc(v, "", tlsEnabled, skipTLSVerify, logLevel)
	if err != nil {
		return err
	}

	// Get device list
	deviceList, err := sysAp.GetDeviceList()
	if err != nil {
		return handleSysApError(err, "get device list", tlsEnabled, skipTLSVerify)
	}

	// Output depending on output format
	if outputFormat == "json" {
		return outputJSON(deviceList, "device list")
	}

	// Check if device list is empty
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

	// Output as plain text (one device per line)
	for _, deviceSerial := range devices {
		fmt.Println(deviceSerial)
	}

	return nil
}

// GetConfiguration retrieves and displays the configuration
func GetConfiguration(v *viper.Viper, tlsEnabled, skipTLSVerify bool, logLevel string, outputFormat string) error {
	// Setup system access point
	sysAp, err := setupFunc(v, "", tlsEnabled, skipTLSVerify, logLevel)
	if err != nil {
		return err
	}

	// Get configuration
	configuration, err := sysAp.GetConfiguration()
	if err != nil {
		return handleSysApError(err, "get configuration", tlsEnabled, skipTLSVerify)
	}

	// Output depending on output format
	if outputFormat == "json" {
		return outputJSON(configuration, "configuration")
	}

	// Check if configuration is empty
	if configuration == nil || len(*configuration) == 0 {
		fmt.Println("No configuration found")
		return nil
	}

	// Output as plain text (one system access point per line)
	for sysApID, sysAp := range *configuration {
		fmt.Printf("System Access Point ID: %s\n", sysApID)
		fmt.Printf("  Name: %s\n", sysAp.SysApName)
		fmt.Printf("  Devices: %d\n", len(sysAp.Devices))
		fmt.Printf("  Users: %d\n", len(sysAp.Users))
		fmt.Printf("  Floors: %d\n", len(sysAp.Floorplan.Floors))
		fmt.Println()
	}

	return nil
}

// GetDevice retrieves and displays a specific device by serial number
func GetDevice(v *viper.Viper, tlsEnabled, skipTLSVerify bool, logLevel string, outputFormat string, serial string) error {
	// Setup system access point
	sysAp, err := setupFunc(v, "", tlsEnabled, skipTLSVerify, logLevel)
	if err != nil {
		return err
	}

	// Get device
	device, err := sysAp.GetDevice(serial)
	if err != nil {
		return handleSysApError(err, "get device", tlsEnabled, skipTLSVerify)
	}

	// Output depending on output format
	if outputFormat == "json" {
		return outputJSON(device, "device")
	}

	// Check if device is empty
	if device == nil || len(*device) == 0 {
		fmt.Printf("No device found with serial: %s\n", serial)
		return nil
	}

	// Get devices for the system access point (using EmptyUUID as key)
	devices, exists := (*device)[models.EmptyUUID]
	if !exists {
		fmt.Printf("No device found with serial: %s\n", serial)
		return nil
	}

	// Check if the specific device exists
	deviceData, deviceExists := devices.Devices[serial]
	if !deviceExists {
		fmt.Printf("No device found with serial: %s\n", serial)
		return nil
	}

	// Output as plain text
	fmt.Printf("Device Serial: %s\n", serial)
	if deviceData.DisplayName != nil {
		fmt.Printf("  Display Name: %s\n", *deviceData.DisplayName)
	}
	if deviceData.Room != nil {
		fmt.Printf("  Room: %s\n", *deviceData.Room)
	}
	if deviceData.Floor != nil {
		fmt.Printf("  Floor: %s\n", *deviceData.Floor)
	}
	if deviceData.Interface != nil {
		fmt.Printf("  Interface: %s\n", *deviceData.Interface)
	}
	if deviceData.NativeID != nil {
		fmt.Printf("  Native ID: %s\n", *deviceData.NativeID)
	}
	if deviceData.Channels != nil {
		fmt.Printf("  Channels: %d\n", len(*deviceData.Channels))
	}
	if deviceData.Parameters != nil {
		fmt.Printf("  Parameters: %d\n", len(*deviceData.Parameters))
	}

	return nil
}
