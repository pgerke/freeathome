package models

// Floor represents a floor in a building with a name and a list of rooms.
type Floor struct {
	Name  string `json:"name"`
	Rooms Rooms  `json:"rooms"`
}

// Floors represents a map of floors identified by their key.
type Floors map[string]Floor

// Floorplan represents a floorplan of a building with a list of floors.
type Floorplan struct {
	Floors Floors `json:"floors"`
}
