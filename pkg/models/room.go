package models

type Room struct {
	Name string `json:"name"`
}

type Rooms map[string]*Room
