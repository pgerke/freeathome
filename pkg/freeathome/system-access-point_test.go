package freeathome

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestSystemAccessPoint_DefaultLogger tests the default logger functionality of SystemAccessPoint.
func TestSystemAccessPoint_DefaultLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	slog.SetDefault(slog.New(handler))

	// Create a SystemAccessPoint with the default logger
	NewSystemAccessPoint("localhost", "user", "password", false, false, nil)

	// Check if the log output contains the expected message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "No logger provided for SystemAccessPoint. Using default logger.") {
		t.Errorf("Expected log output to contain 'No logger provided for SystemAccessPoint. Using default logger.', got: %s", logOutput)
	}
}

// TestSystemAccessPoint_GetHostName tests the GetHostName method of SystemAccessPoint.
func TestSystemAccessPoint_GetHostName(t *testing.T) {
	sysAp, buf := setup(t, true)
	expected := "localhost"

	actual := sysAp.GetHostName()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual host name matches the expected host name
	if actual != expected {
		t.Errorf("Expected host name '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPoint_GetTlsEnabled tests the GetTlsEnabled method of SystemAccessPoint.
func TestSystemAccessPoint_GetTlsEnabled(t *testing.T) {
	sysAp, buf := setup(t, true)
	expected := true

	actual := sysAp.GetTlsEnabled()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual TLS enabled status matches the expected status
	if actual != expected {
		t.Errorf("Expected TLS enabled '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPoint_GetVerboseErrors tests the GetVerboseErrors method of SystemAccessPoint.
func TestSystemAccessPoint_GetVerboseErrors(t *testing.T) {
	sysAp, buf := setup(t, true)
	expected := false

	actual := sysAp.GetVerboseErrors()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual verbose errors status matches the expected status
	if actual != expected {
		t.Errorf("Expected verbose errors '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPoint_GetUrlWithoutTls tests the GetUrl method of SystemAccessPoint without TLS.
func TestSystemAccessPoint_GetUrlWithoutTls(t *testing.T) {
	sysAp, buf := setup(t, false)
	expected := "http://localhost/fhapi/v1/api/rest/test123"
	actual := sysAp.GetUrl("test123")

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPoint_GetUrlWithTls tests the GetUrl method of SystemAccessPoint with TLS.
func TestSystemAccessPoint_GetUrlWithTls(t *testing.T) {
	sysAp, buf := setup(t, true)

	actual := sysAp.GetUrl("test123")
	expected := "https://localhost/fhapi/v1/api/rest/test123"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPoint_GetWebSocketUrlWithoutTls tests the getWebSocketUrl method of SystemAccessPoint without TLS.
func TestSystemAccessPoint_GetWsUrlWithoutTls(t *testing.T) {
	sysAp, buf := setup(t, false)

	actual := sysAp.getWebSocketUrl()
	expected := "ws://localhost/fhapi/v1/api/ws"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPoint_GetWebSocketUrlWithTls tests the getWebSocketUrl method of SystemAccessPoint with TLS.
func TestSystemAccessPoint_GetWsUrlWithTls(t *testing.T) {
	sysAp, buf := setup(t, true)

	actual := sysAp.getWebSocketUrl()
	expected := "wss://localhost/fhapi/v1/api/ws"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}
