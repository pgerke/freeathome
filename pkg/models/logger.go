package models

// Logger is an interface that defines methods for logging messages at various levels of severity.
// Implementations of this interface should provide mechanisms to log debug, error, general log, and warning messages.
type Logger interface {
	// Debug logs a debug message with optional parameters.
	Debug(message string, optionalParams ...any)

	// Error logs an error message with optional parameters.
	Error(message string, optionalParams ...any)

	// Log logs a general message with optional parameters.
	Log(message string, optionalParams ...any)

	// Warn logs a warning message with optional parameters.
	Warn(message string, optionalParams ...any)
}
