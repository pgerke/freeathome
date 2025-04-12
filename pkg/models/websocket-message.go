package models

// WebSocketMessage represents a message that can be sent over a web socket connection.
type WebSocketMessage map[string]Message

type Message struct {
	// Datapoints is a map of datapoint identifiers to their values.
	Datapoints map[string]string `json:"datapoints"`

	// Devices is a map of device identifiers to their values.
	Devices Devices `json:"devices"`

	// DevicesAdded is a list of devices that have been added.
	DevicesAdded []string `json:"devicesAdded"`

	// DevicesRemoved is a list of devices that have been removed.
	DevicesRemoved []string `json:"devicesRemoved"`

	// ScenesTriggered is a list of scenes that have been triggered.
	ScenesTriggered ScenesTriggered `json:"scenesTriggered"`

	// Parameters is a map of parameters that can be used to pass additional information.
	Parameters *map[string]string `json:"parameters,omitempty"`
}

// DatapointPattern is a regular expression pattern that matches the format of a datapoint identifier.
const DatapointPattern = `(?i)^([a-z0-9]{12})\/(ch[\da-f]{4})\/([io]dp\d{4})$`
