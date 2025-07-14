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
	sysAp := freeathome.NewSystemAccessPoint(cfg.Hostname, cfg.Username, cfg.Password, tlsEnabled, skipTLSVerify, false, logger)

	return sysAp, nil
}

// GetDeviceList retrieves and displays the device list
func GetDeviceList(v *viper.Viper, tlsEnabled, skipTLSVerify bool, logLevel string, outputFormat string) error {
	// Setup system access point
	sysAp, err := setup(v, "", tlsEnabled, skipTLSVerify, logLevel)
	if err != nil {
		return err
	}

	// Get device list
	deviceList, err := sysAp.GetDeviceList()
	if err != nil {
		// Provide helpful error message for TLS issues
		if tlsEnabled && !skipTLSVerify {
			return fmt.Errorf("failed to get device list: %w\n\nIf you're getting TLS certificate errors, try:\n  - Using --skip-tls-verify flag\n  - Using --tls=false to use HTTP instead of HTTPS", err)
		}
		return fmt.Errorf("failed to get device list: %w", err)
	}

	// Output depending on output format
	if outputFormat == "json" {
		jsonData, err := json.Marshal(deviceList)
		if err != nil {
			return fmt.Errorf("failed to marshal device list to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
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
