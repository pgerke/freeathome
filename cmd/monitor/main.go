package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/pgerke/freeathome/pkg/freeathome"
	"golang.org/x/term"
)

// version is the version of the application. The value will be overridden by the linker during the build process.
var version = "debug"

// commit is the commit hash of the application. The value will be overridden by the linker during the build process.
var commit = "unknown"

func main() {
	// Wait for interrupt signal to trigger graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// First signal triggers graceful shutdown
		<-sigs
		fmt.Println("Interrupt received, shutting down gracefully...")
		fmt.Println("Press Ctrl+C again to force exit")
		cancel()

		// Second signal triggers immediate, forced shutdown
		<-sigs
		fmt.Println("\nSecond interrupt received, shutting down immediately...")
		os.Exit(1)
	}()

	fmt.Println("Press Ctrl+C to exit")
	exitCode := run(ctx)
	fmt.Printf("Exiting with code %d\n", exitCode)
	os.Exit(exitCode)
}

func run(ctx context.Context) int {
	// Create a new logger with the specified options
	// Use a colorized handler if the terminal supports colors
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		color.NoColor = true
	}
	handler := freeathome.NewColorHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug, // TODO: make this configurable -> command line flag
	})
	logger := freeathome.NewDefaultLogger(handler)
	logger.Log("Starting free@home Monitor", "version", version, "commit", commit)

	// Retrieve the environment variables for the system access point
	host, user, password, err := lookupEnvs()
	if err != nil {
		logger.Error("Error looking up environment variables", "error", err)
		return 1
	}

	// Create a new system access point with the specified configuration
	config := freeathome.NewConfig(host, user, password)
	config.VerboseErrors = true
	config.Logger = logger
	sysAp := freeathome.NewSystemAccessPoint(config)

	// Connect to the system access point websocket
	sysAp.ConnectWebSocket(ctx, 30*time.Second) // TODO: make this configurable -> command line flag

	logger.Debug("shutdown complete, bye!")
	return 0
}

// lookupEnvs retrieves the environment variables required for connecting to the SYSAP system.
// It checks for the presence and non-emptiness of the following environment variables:
// - SYSAP_HOST: The hostname or IP address of the SYSAP system.
// - SYSAP_USER_ID: The user ID for authentication.
// - SYSAP_PASSWORD: The password for authentication.
//
// If any of these variables are not set or are empty, the function returns an error.
//
// Returns:
//
//	string - The value of the SYSAP_HOST environment variable.
//	string - The value of the SYSAP_USER_ID environment variable.
//	string - The value of the SYSAP_PASSWORD environment variable.
//	error  - An error if any of the required environment variables are missing or empty.
func lookupEnvs() (string, string, string, error) {
	host, ok := os.LookupEnv("SYSAP_HOST")
	if !ok || host == "" {
		return "", "", "", fmt.Errorf("SYSAP_HOST variable is not set")
	}

	user, ok := os.LookupEnv("SYSAP_USER_ID")
	if !ok || user == "" {
		return "", "", "", fmt.Errorf("SYSAP_USER_ID variable is not set")
	}

	password, ok := os.LookupEnv("SYSAP_PASSWORD")
	if !ok || password == "" {
		return "", "", "", fmt.Errorf("SYSAP_PASSWORD variable is not set")
	}

	return host, user, password, nil
}
