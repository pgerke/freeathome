package models

// Room represents a room in a building with a name.
type Room struct {
	Name string `json:"name"`
}

// Rooms represents a map of rooms identified by their key.
type Rooms map[string]Room
