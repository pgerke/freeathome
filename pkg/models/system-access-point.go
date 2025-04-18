package models

// EmptyUUID is a constant representing an empty UUID. In the local (non-cloud) free@home API, the system access points ID is always the empty UUID.
const EmptyUUID = "00000000-0000-0000-0000-000000000000"

// SysAP represents a system access point with a name, a list of devices, a floorplan, a list of users, and an optional error.
type SysAP struct {
	// Devices represents a map of devices identified by their key.
	Devices map[string]Device `json:"devices"`

	// Floorplan represents the floorplan of the building.
	Floorplan Floorplan `json:"floorplan"`

	// SysApName represents the name of the system access point.
	SysApName string `json:"sysapName"`

	// Users represents a map of users identified by their key.
	Users Users `json:"users"`

	// Error is an optional field that can be used to indicate an error.
	Error *Error `json:"error,omitempty"`
}
