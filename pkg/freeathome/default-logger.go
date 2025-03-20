package freeathome

import "log/slog"

// DefaultLogger is a default implementation of the Logger interface that logs messages to the console.
type DefaultLogger struct{}

// Debug logs a debug message with optional parameters.
func (l *DefaultLogger) Debug(message string, optionalParams ...any) {
	slog.Debug(message, optionalParams...)
}

// Error logs an error message with optional parameters.
func (l *DefaultLogger) Error(message string, optionalParams ...any) {
	slog.Error(message, optionalParams...)
}

// Log logs a general message with optional parameters.
func (l *DefaultLogger) Log(message string, optionalParams ...any) {
	slog.Info(message, optionalParams...)
}

// Warn logs a warning message with optional parameters.
func (l *DefaultLogger) Warn(message string, optionalParams ...any) {
	slog.Warn(message, optionalParams...)
}
