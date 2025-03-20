package models

// Output represents an output channel with a value and a pairing ID.
type Output struct {
	Value     string `json:"value"`
	PairingId uint   `json:"pairingId"`
}

// Scene represents a scene with a list of channels.
type Scene struct {
	Channels map[string]Output `json:"channels"`
}

// ScenesTriggered represents a map of scenes triggered by their key.
type ScenesTriggered map[string]Scene
