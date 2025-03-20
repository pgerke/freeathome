package models

// InOutPut describes an input or output.
type InOutPut struct {
	// Value represents an optional string value that can be serialized to JSON.
	// The field is omitted from the JSON output if it is nil.
	Value *string `json:"value,omitempty"`

	// PairingID represents the unique identifier for pairing. It is an optional field
	// The field is omitted from the JSON output if it is nil.
	PairingID *uint `json:"pairingId,omitempty"`
}
