package cli

import (
	"bufio"
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

	// Create error channel for the shutdown
	shutdown := make(chan error, 1)

	// Setup keypress handling for graceful shutdown
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				char, _, err := reader.ReadRune()
				if err != nil {
					break
				}
				if char == 'q' || char == 'Q' {
					// Send SIGINT to trigger graceful shutdown
					sigs <- syscall.SIGINT
					return
				}
			}
		}
	}()

	go func() {
		// First signal triggers graceful shutdown
		<-sigs
		fmt.Println("Exit signal received, shutting down gracefully...")
		fmt.Println("Press Ctrl+C to force exit")
		cancel()

		// Second signal triggers immediate, forced shutdown
		<-sigs
		fmt.Println("\nSecond exit signal received, shutting down immediately...")
		shutdown <- fmt.Errorf("forced shutdown requested")
	}()

	fmt.Println("Press 'q' or Ctrl+C to exit")

	// Set the maximum reconnection attempts
	sysAp.SetMaxReconnectionAttempts(config.MaxReconnectionAttempts)

	// Set the exponential backoff setting
	sysAp.SetExponentialBackoffEnabled(config.ExponentialBackoff)

	// Connect to the system access point websocket
	timeout := time.Duration(config.Timeout) * time.Second
	go func() {
		shutdown <- sysAp.ConnectWebSocket(ctx, timeout)
	}()

	// Handle both forced shutdown and WebSocket connection errors
	err = <-shutdown
	if err != nil && err != context.Canceled {
		return err
	}

	return nil
}
