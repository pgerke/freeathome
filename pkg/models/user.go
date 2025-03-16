package models

type User struct {
	Enabled bool `json:"enabled"`

	Flags []string `json:"flags"`

	GrantedPermissions []string `json:"grantedPermissions"`

	JID string `json:"jid"`

	Name string `json:"name"`

	RequestedPermissions []string `json:"requestedPermissions"`

	Role string `json:"role"`
}

type Users map[string]*User
