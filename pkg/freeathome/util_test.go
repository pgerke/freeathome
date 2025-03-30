package freeathome

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setup initializes a SystemAccessPoint with a mock logger and returns it along with a buffer to capture log output.
func setup(t *testing.T, tlsEnabled bool) (*SystemAccessPoint, *bytes.Buffer) {
	t.Helper()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := NewDefaultLogger(handler)

	// Create a SystemAccessPoint with the default logger
	return NewSystemAccessPoint("localhost", "user", "password", tlsEnabled, false, logger), &buf
}

// MockRoundTripper is a mock implementation of http.RoundTripper for testing purposes.
type MockRoundTripper struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}

// RoundTrip executes a single HTTP transaction and returns the response.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.Request = req
	return m.Response, m.Err
}

// It captures the request and response for testing purposes.
func loadTestResponseBody(t *testing.T, filename string) io.ReadCloser {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read test file %s: %v", filename, err)
	}

	content := string(data)
	return io.NopCloser(strings.NewReader(content))
}
