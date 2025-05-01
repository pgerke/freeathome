package integration

import (
	"os"
	"os/exec"
	"testing"
)

const bin = "./monitor-integration.test"
const coverageDirectory = "./../coverage-monitor"

// TestMonitor_MissingEnvs verifies the monitor's behavior when required environment variables are missing.
func TestMonitor_MissingEnvs(t *testing.T) {
	// Run with missing env
	run := exec.Command(bin, "-test.run=TestMonitor_Main")
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
