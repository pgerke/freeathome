package models

// VirtualDeviceResponse represents a map of created virtual devices identified the system access point UUID.
type VirtualDeviceResponse map[string]CreatedVirtualDevices

// CreatedVirtualDevices represents a map of created virtual devices identified by their key.
type CreatedVirtualDevices struct {
	Devices map[string]CreatedVirtualDevice `json:"devices"`
}

// CreatedVirtualDevice represents a created virtual device with a serial number.
type CreatedVirtualDevice struct {
	Serial string `json:"serial"`
}
