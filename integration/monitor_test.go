//go:build integration

package integration

import (
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const bin = "./../cli-integration.test"
const coverageDirectory = "./../coverage-cli"

// TestMonitorMissingEnvs verifies the monitor's behavior when required environment variables are missing.
func TestMonitorMissingEnvs(t *testing.T) {
	// Run with missing env
	run := exec.Command(bin, "monitor")
	run.Env = append(os.Environ(), "GOCOVERDIR="+coverageDirectory)

	output, err := run.CombinedOutput()
	t.Logf("output:\n%s", output)

	// Check the exit code
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("could not get exit code: %v", err)
		}
	}

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

// TestMonitorSuccessfulRunInputKeypress tests that the monitor runs successfully when the user presses the 'q' key.
func TestMonitorSuccessfulRunInputKeypress(t *testing.T) {
	// Set up the test server
	addr, shutdown, sendMessage := startTestWebSocketServer(t)
	defer shutdown()

	// Run the monitor
	run := exec.Command(
		bin,
		"monitor",
		"--log-level=debug",
		"--tls=false",
	)

	run.Env = append(os.Environ(),
		"GOCOVERDIR="+coverageDirectory,
		"FREEATHOME_HOSTNAME="+addr,
		"FREEATHOME_USERNAME=admin",
		"FREEATHOME_PASSWORD=password",
	)

	// Set up pipes for keypress input and output
	var stdin io.WriteCloser
	var stdout, stderr io.ReadCloser
	var err error
	stdin, err = run.StdinPipe()
	if err != nil {
		t.Fatalf("could not get stdin pipe: %v", err)
	}
	stdout, err = run.StdoutPipe()
	if err != nil {
		t.Fatalf("could not get stdout pipe: %v", err)
	}
	stderr, err = run.StderrPipe()
	if err != nil {
		t.Fatalf("could not get stderr pipe: %v", err)
	}
	go io.Copy(t.Output(), stdout)
	go io.Copy(t.Output(), stderr)

	// Start the monitor
	if err := run.Start(); err != nil {
		t.Fatalf("could not start monitor: %v", err)
	}

	t.Logf("Waiting for monitor to connect...")
	time.Sleep(500 * time.Millisecond) // Allow some time for the monitor to connect

	// Send 'q' keypress to trigger graceful shutdown
	t.Logf("Sending 'q' keypress to trigger graceful shutdown...")
	if _, err := stdin.Write([]byte("q\n")); err != nil {
		t.Errorf("could not send 'q' keypress: %v", err)
	}
	stdin.Close() // Close stdin after sending the keypress

	// Send a message to the test server to trigger a message to the monitor
	sendMessage()

	// Wait for the monitor to finish
	err = run.Wait()

	// Check the exit code
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			t.Logf("Process exited with code: %d", exitCode)
		} else {
			t.Fatalf("could not get exit code: %v", err)
		}
	} else {
		t.Logf("Process completed successfully with exit code: %d", exitCode)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

// TestMonitorSuccessfulRunInterrupt tests that the monitor runs successfully when the user presses the 'q' key.
func TestMonitorSuccessfulRunInterrupt(t *testing.T) {
	// Set up the test server
	addr, shutdown, sendMessage := startTestWebSocketServer(t)
	defer shutdown()

	// Run the monitor
	run := exec.Command(
		bin,
		"monitor",
		"--log-level=debug",
		"--tls=false",
	)

	run.Env = append(os.Environ(),
		"GOCOVERDIR="+coverageDirectory,
		"FREEATHOME_HOSTNAME="+addr,
		"FREEATHOME_USERNAME=admin",
		"FREEATHOME_PASSWORD=password",
	)

	// Set up pipes for output
	var stdout, stderr io.ReadCloser
	var err error
	stdout, err = run.StdoutPipe()
	if err != nil {
		t.Fatalf("could not get stdout pipe: %v", err)
	}
	stderr, err = run.StderrPipe()
	if err != nil {
		t.Fatalf("could not get stderr pipe: %v", err)
	}
	go io.Copy(t.Output(), stdout)
	go io.Copy(t.Output(), stderr)

	// Start the monitor
	if err := run.Start(); err != nil {
		t.Fatalf("could not start monitor: %v", err)
	}

	t.Logf("Waiting for monitor to connect...")
	time.Sleep(500 * time.Millisecond) // Allow some time for the monitor to connect

	// Send an interrupt signal to the monitor process
	if err := run.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send interrupt signal: %v", err)
	}

	// Send a message to the test server to trigger a message to the monitor
	sendMessage()

	// Wait for the monitor to finish
	err = run.Wait()

	// Check the exit code
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			t.Logf("Process exited with code: %d", exitCode)
		} else {
			t.Fatalf("could not get exit code: %v", err)
		}
	} else {
		t.Logf("Process completed successfully with exit code: %d", exitCode)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

// TestMonitorSuccessfulRunForcedExit tests that the monitor runs successfully when the user presses the 'q' key.
func TestMonitorSuccessfulRunForcedExit(t *testing.T) {
	// Set up the test server
	addr, shutdown, _ := startTestWebSocketServer(t)
	defer shutdown()

	// Run the monitor
	run := exec.Command(
		bin,
		"monitor",
		"--log-level=debug",
		"--tls=false",
	)

	run.Env = append(os.Environ(),
		"GOCOVERDIR="+coverageDirectory,
		"FREEATHOME_HOSTNAME="+addr,
		"FREEATHOME_USERNAME=admin",
		"FREEATHOME_PASSWORD=password",
	)

	// Set up pipes for output
	var stdout, stderr io.ReadCloser
	var err error
	stdout, err = run.StdoutPipe()
	if err != nil {
		t.Fatalf("could not get stdout pipe: %v", err)
	}
	stderr, err = run.StderrPipe()
	if err != nil {
		t.Fatalf("could not get stderr pipe: %v", err)
	}
	go io.Copy(t.Output(), stdout)
	go io.Copy(t.Output(), stderr)

	// Start the monitor
	if err := run.Start(); err != nil {
		t.Fatalf("could not start monitor: %v", err)
	}

	t.Logf("Waiting for monitor to connect...")
	time.Sleep(500 * time.Millisecond) // Allow some time for the monitor to connect

	// Send interrupt signals to the monitor process
	if err := run.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send first interrupt signal: %v", err)
	}
	time.Sleep(250 * time.Millisecond)
	if err := run.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send second interrupt signal: %v", err)
	}

	// Wait for the monitor to finish
	err = run.Wait()

	// Check the exit code
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			t.Logf("Process exited with code: %d", exitCode)
		} else {
			t.Fatalf("could not get exit code: %v", err)
		}
	} else {
		t.Logf("Process completed successfully with exit code: %d", exitCode)
	}

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func startTestWebSocketServer(t *testing.T) (addr string, shutdown func(), sendMessage func()) {
	t.Helper()
	upgrader := websocket.Upgrader{}

	var conn *websocket.Conn
	mux := http.NewServeMux()
	mux.HandleFunc("/fhapi/v1/api/ws", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Upgrading WebSocket")
		var err error
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Failed to upgrade WebSocket: %v", err)
			return
		}
		defer conn.Close()

		// Echo everything
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			t.Logf("Received message: %s", string(msg))
			_ = conn.WriteMessage(mt, msg)
		}
	})

	// Listen on a random port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	server := &http.Server{
		Handler: mux,
	}

	go server.Serve(ln)

	// Return address, sendMessage and shutdown functions
	return ln.Addr().String(), func() {
			server.Close()
		}, func() {
			conn.WriteMessage(websocket.TextMessage, []byte("{}"))
		}
}
