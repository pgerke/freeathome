package models

// GetDataPointResponse describes the response to a query requesting a data point. It is a map of data point names to their values using the System Access Point's UUID as a key.
type GetDataPointResponse map[string]DataPoint

// DataPoint represents a data point in the system.
type DataPoint struct {
	Values []string `json:"values"`
}
