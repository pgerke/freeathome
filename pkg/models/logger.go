package models

// Logger is an interface that defines methods for logging messages at various levels of severity.
// Implementations of this interface should provide mechanisms to log debug, error, general log, and warning messages.
//
// Methods:
//   - Debug(message string, optionalParams ...interface{}): Logs a debug message with optional parameters.
//   - Error(message string, optionalParams ...interface{}): Logs an error message with optional parameters.
//   - Log(message string, optionalParams ...interface{}): Logs a general message with optional parameters.
//   - Warn(message string, optionalParams ...interface{}): Logs a warning message with optional parameters.
type Logger interface {
	Debug(message string, optionalParams ...interface{})
	Error(message string, optionalParams ...interface{})
	Log(message string, optionalParams ...interface{})
	Warn(message string, optionalParams ...interface{})
}
