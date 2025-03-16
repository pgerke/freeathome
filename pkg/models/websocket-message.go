package models

type Message struct {
	Datapoints map[string]string `json:"datapoints"`

	Devices map[string]Device `json:"devices"`

	DevicesAdded []string `json:"devicesAdded"`

	DevicesRemoved []string `json:"devicesRemoved"`

	// ScenesTriggered ScenesTriggered `json:"scenesTriggered"`

	Parameters *map[string]string `json:"parameters,omitempty"`
}

type WebSocketMessage map[string]Message
