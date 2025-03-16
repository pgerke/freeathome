package models

// Error represents a structured error message with a code, detail, and title.
// It is used to provide detailed error information in a standardized format.
type Error struct {
	// Code represents the error code in JSON format. Example: "2010"
	Code string `json:"code"`

	// Detail provides additional information about the error in a human-readable format. Example: "FreeAtHome connection timeout"
	Detail string `json:"detail"`

	// Title represents the title of the error message in JSON format. Example: "Connection Error"
	Title string `json:"title"`
}
