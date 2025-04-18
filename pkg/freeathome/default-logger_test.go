package freeathome

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

var testParams = []any{"key1", "value1", "key2", 123, "key3", true}

// TestDefaultLogger tests the default logger functionality.
func TestDefaultLogger_DefaultHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	slog.SetDefault(slog.New(handler))
	logger := NewDefaultLogger(nil)

	// Test message and parameters
	message := "Test message"

	// Call the Log method
	logger.Log(message, testParams...)

	// Check if the log output contains the expected message and parameters
	evaluateLogOutput(t, buf.String(), message, testParams)
}

// TestDefaultLogger_Debug tests the Debug method of the default logger.
func TestDefaultLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(t, &buf)
	message := "Debug message"
	logger.Debug(message, testParams...)
	evaluateLogOutput(t, buf.String(), message, testParams)
}

// TestDefaultLogger_Error tests the Error method of the default logger.
func TestDefaultLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(t, &buf)
	message := "Error message"
	logger.Error(message, testParams...)
	evaluateLogOutput(t, buf.String(), message, testParams)
}

// TestDefaultLogger_Log tests the Log method of the default logger.
func TestDefaultLogger_Log(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(t, &buf)
	message := "Log message"
	logger.Log(message, testParams...)
	evaluateLogOutput(t, buf.String(), message, testParams)
}

// TestDefaultLogger_Warn tests the Warn method of the default logger.
func TestDefaultLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(t, &buf)
	message := "Warn message"
	logger.Warn(message, testParams...)
	evaluateLogOutput(t, buf.String(), message, testParams)
}

// createTestLogger creates a test logger with a buffer to capture log output.
func createTestLogger(t *testing.T, buf *bytes.Buffer) *DefaultLogger {
	t.Helper()

	handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug})

	return NewDefaultLogger(handler)
}

// evaluateLogOutput checks if the log output contains the expected message and parameters.
func evaluateLogOutput(t *testing.T, logOutput string, message string, params []any) {
	t.Helper()

	if !strings.Contains(logOutput, message) {
		t.Errorf("Expected log output to contain message '%s', but got '%s'", message, logOutput)
	}
	for _, param := range params {
		if !strings.Contains(logOutput, fmt.Sprint(param)) {
			t.Errorf("Expected log output to contain parameter '%v', but got '%s'", param, logOutput)
		}
	}
}
