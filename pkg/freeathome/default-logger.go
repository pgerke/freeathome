package freeathome

import (
	"log/slog"
)

// DefaultLogger is a default implementation of the Logger interface that logs messages to the console.
type DefaultLogger struct {
	logger *slog.Logger
}

// NewDefaultLogger creates a new DefaultLogger instance with the specified slog.Handler.
func NewDefaultLogger(handler slog.Handler) *DefaultLogger {
	if handler == nil {
		handler = slog.Default().Handler()
	}

	return &DefaultLogger{
		logger: slog.New(handler),
	}
}

// Debug logs a debug message with optional parameters.
func (l *DefaultLogger) Debug(message string, optionalParams ...any) {
	l.logger.Debug(message, optionalParams...)
}

// Error logs an error message with optional parameters.
func (l *DefaultLogger) Error(message string, optionalParams ...any) {
	l.logger.Error(message, optionalParams...)
}

// Log logs a general message with optional parameters.
func (l *DefaultLogger) Log(message string, optionalParams ...any) {
	l.logger.Info(message, optionalParams...)
}

// Warn logs a warning message with optional parameters.
func (l *DefaultLogger) Warn(message string, optionalParams ...any) {
	l.logger.Warn(message, optionalParams...)
}
