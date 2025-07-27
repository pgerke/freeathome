//go:build integration

package integration

import (
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
	t.Skip("Unsure if still needed!")
	// Run with missing env
	run := exec.Command(bin, "-test.run=TestCLIMain", "monitor")
	run.Env = append(os.Environ(), "RUN_MAIN=1", "GOCOVERDIR="+coverageDirectory)

	output, err := run.CombinedOutput()
	t.Logf("output:\n%s", output)

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

func TestMonitorSuccessfulRun(t *testing.T) {
	// TODO: Fix this test
	t.Skip("TODO: Fix this test")
	// Set up the test server
	addr, shutdown := startTestWebSocketServer(t)
	defer shutdown()

	// Run the monitor
	run := exec.Command(bin, "-test.run=TestCLIMain", "monitor")
	run.Env = append(os.Environ(),
		"RUN_MAIN=1",
		"GOCOVERDIR="+coverageDirectory,
		"SYSAP_HOST="+addr,
		"SYSAP_USER_ID=admin",
		"SYSAP_PASSWORD=password",
	)

	if err := run.Start(); err != nil {
		t.Fatalf("could not start monitor: %v", err)
	}

	time.Sleep(2500 * time.Millisecond) // Allow some time for the monitor to connect

	// Send an interrupt signal to the monitor process
	if err := run.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send interrupt signal: %v", err)
	}

	// Wait for the monitor to finish
	err := run.Wait()
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("could not get exit code: %v", err)
		}
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func startTestWebSocketServer(t *testing.T) (addr string, shutdown func()) {
	upgrader := websocket.Upgrader{}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
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

	// Return address and shutdown function
	return ln.Addr().String(), func() {
		server.Close()
	}
}
