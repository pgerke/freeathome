package models

// User represents a user with a name, JID, role, flags, granted permissions, requested permissions, and enabled status.
type User struct {
	// Enabled indicates whether the user is enabled.
	Enabled bool `json:"enabled"`

	// Flags represents the flags of the user.
	Flags []string `json:"flags"`

	// GrantedPermissions represents the granted permissions of the user.
	GrantedPermissions []string `json:"grantedPermissions"`

	// JID represents the JID of the user.
	JID string `json:"jid"`

	// Name represents the name of the user.
	Name string `json:"name"`

	// RequestedPermissions represents the requested permissions of the user.
	RequestedPermissions []string `json:"requestedPermissions"`

	// Role represents the role of the user.
	Role string `json:"role"`
}

// Users represents a map of users identified by their key.
type Users map[string]*User
