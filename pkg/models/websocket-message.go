package models

// WebSocketMessage represents a message that can be sent over a web socket connection.
type WebSocketMessage map[string]Message

type Message struct {
	// Datapoints is a map of datapoint identifiers to their values.
	Datapoints map[string]WebSocketMessageDatapoint `json:"datapoints"`

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

// WebSocketMessageDatapoint represents a map of datapoints identified by their key.
type WebSocketMessageDatapoint struct {
	// Datapoints is a map of datapoint identifiers to their values.
	Datapoints map[string]string `json:"datapoints"`
}
