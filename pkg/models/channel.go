package models

// Channel describes a device channel.
type Channel struct {
	// DisplayName represents the display name of the channel.
	DisplayName *string `json:"displayName,omitempty"`

	// FunctionID represents the function identifier as defined in the Busch+Jaeger documentation.
	FunctionID *string `json:"functionId,omitempty"`

	// Room represents the room identifier.
	Room *string `json:"roomId,omitempty"`

	// Floor represents the floor identifier.
	Floor *string `json:"floorId,omitempty"`

	// Inputs represents the channel's inputs.
	Inputs []*InOutPut `json:"inputs,omitempty"`

	// Outputs represents the channel's outputs.
	Outputs []*InOutPut `json:"outputs,omitempty"`

	// Parameters represents the channel's parameters.
	Parameters *map[string]string `json:"parameters,omitempty"`

	// Type represents the channel type.
	Type *string `json:"type,omitempty"`
}
