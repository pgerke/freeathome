package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// MonitorCommandConfig is a struct that contains the configuration for the monitor command
type MonitorCommandConfig struct {
	CommandConfig
	Timeout                 int
	MaxReconnectionAttempts int
	ExponentialBackoff      bool
}

// Monitor connects to the free@home system access point via WebSocket and monitors real-time events
func Monitor(config MonitorCommandConfig) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

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

	// Set the maximum reconnection attempts
	sysAp.SetMaxReconnectionAttempts(config.MaxReconnectionAttempts)

	// Set the exponential backoff setting
	sysAp.SetExponentialBackoffEnabled(config.ExponentialBackoff)

	// Connect to the system access point websocket
	timeout := time.Duration(config.Timeout) * time.Second
	sysAp.ConnectWebSocket(ctx, timeout)

	return nil
}
