package freeathome

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// TestSystemAccessPointDefaultLogger tests the default logger functionality of SystemAccessPoint.
func TestSystemAccessPointDefaultLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	slog.SetDefault(slog.New(handler))

	// Create a SystemAccessPoint with the default logger
	NewSystemAccessPointWithDefaults("localhost", "user", "password")

	// Check if the log output contains the expected message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "No logger provided for SystemAccessPoint. Using default logger.") {
		t.Errorf("Expected log output to contain 'No logger provided for SystemAccessPoint. Using default logger.', got: %s", logOutput)
	}
}

// TestNoConfigErrors tests that an error is returned when a nil config is passed to NewSystemAccessPoint.
func TestNoConfigErrors(t *testing.T) {
	sysap, err := NewSystemAccessPoint(nil)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if sysap != nil {
		t.Errorf("Expected nil, got %v", sysap)
	}
}

// TestMustNewSystemAccessPoint tests that MustNewSystemAccessPoint panics when a nil config is passed to NewSystemAccessPoint.
func TestMustNewSystemAccessPoint(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got nil")
		}
	}()
	MustNewSystemAccessPoint(nil)
}

// TestSystemAccessPointGetHostName tests the GetHostName method of SystemAccessPoint.
func TestSystemAccessPointGetHostName(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)
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

// TestSystemAccessPointGetTlsEnabled tests the GetTlsEnabled method of SystemAccessPoint.
func TestSystemAccessPointGetTlsEnabled(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)
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

// TestSystemAccessPointGetSkipTLSVerify tests the GetSkipTLSVerify method of SystemAccessPoint.
func TestSystemAccessPointGetSkipTLSVerify(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)
	expected := false

	actual := sysAp.GetSkipTLSVerify()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual skip TLS verify status matches the expected status
	if actual != expected {
		t.Errorf("Expected skip TLS verify '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetSkipTLSVerifyEnabled tests the GetSkipTLSVerify method of SystemAccessPoint when skip TLS verify is enabled.
func TestSystemAccessPointGetSkipTLSVerifyEnabled(t *testing.T) {
	sysAp, buf, _ := setup(t, true, true)
	expected := true

	actual := sysAp.GetSkipTLSVerify()

	// Check if the log output is empty
	logOutput := buf.String()
	if !strings.Contains(logOutput, "this is not recommended") {
		t.Errorf("Expected log output to contain 'this is not recommended', got: %s", logOutput)
	}

	// Check if the actual skip TLS verify status matches the expected status
	if actual != expected {
		t.Errorf("Expected skip TLS verify '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetVerboseErrors tests the GetVerboseErrors method of SystemAccessPoint.
func TestSystemAccessPointGetVerboseErrors(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)
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

// TestSystemAccessPointGetUrlWithoutTls tests the GetUrl method of SystemAccessPoint without TLS.
func TestSystemAccessPointGetUrlWithoutTls(t *testing.T) {
	sysAp, buf, _ := setup(t, false, false)
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

// TestSystemAccessPointGetUrlWithTls tests the GetUrl method of SystemAccessPoint with TLS.
func TestSystemAccessPointGetUrlWithTls(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)

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

// TestSystemAccessPointGetWebSocketUrlWithoutTls tests the getWebSocketUrl method of SystemAccessPoint without TLS.
func TestSystemAccessPointGetWsUrlWithoutTls(t *testing.T) {
	sysAp, buf, _ := setup(t, false, false)

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

// TestSystemAccessPointGetWebSocketUrlWithTls tests the getWebSocketUrl method of SystemAccessPoint with TLS.
func TestSystemAccessPointGetWsUrlWithTls(t *testing.T) {
	sysAp, buf, _ := setup(t, true, false)

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

// TestSystemAccessPointCalculateBackoffDuration tests the calculateBackoffDuration method of SystemAccessPoint.
func TestSystemAccessPointCalculateBackoffDuration(t *testing.T) {
	sysAp, _, _ := setup(t, true, false)

	// Test exponential backoff calculation
	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 2 * time.Second},   // 1s * 2^1 = 2s
		{2, 4 * time.Second},   // 1s * 2^2 = 4s
		{3, 8 * time.Second},   // 1s * 2^3 = 8s
		{4, 16 * time.Second},  // 1s * 2^4 = 16s
		{5, 30 * time.Second},  // Capped at 30s
		{6, 30 * time.Second},  // Capped at 30s
		{10, 30 * time.Second}, // Capped at 30s
	}

	for _, tc := range testCases {
		result := sysAp.calculateBackoffDuration(tc.attempt)
		if result != tc.expected {
			t.Errorf("For attempt %d, expected %v, got %v", tc.attempt, tc.expected, result)
		}
	}
}
