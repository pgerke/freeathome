package models

// SetDataPointResponse represents a map of the responses to the requested data points identified by system access point UUID.
type SetDataPointResponse map[string]SetDataPoint

// SetDataPoint represents a map of the responses to the requested data points identified by their key.
type SetDataPoint map[string]string
