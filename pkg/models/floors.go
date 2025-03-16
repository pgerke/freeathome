package models

type Floors map[string]*Rooms

type Floorplan struct {
	Floors Floors `json:"floors"`
}
