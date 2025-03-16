package models

// Device represents a device in the system.
type Device struct {
	// DisplayName is the display name of the device.
	DisplayName *string `json:"displayName,omitempty"`

	// Room is the room where the device is located.
	Room *string `json:"room,omitempty"`

	// Floor is the floor where the device is located.
	Floor *string `json:"floor,omitempty"`

	// Interface is the interface type of the device.
	Interface *string `json:"interface,omitempty"`

	// NativeID is the native identifier of the device.
	NativeID *string `json:"nativeId,omitempty"`

	// Channels is a map of channel identifiers to Channel objects associated with the device.
	Channels *map[string]*Channel `json:"channels,omitempty"`

	// Parameters is a map of parameter names to their values for the device.
	Parameters *map[string]string `json:"parameters,omitempty"`
}
