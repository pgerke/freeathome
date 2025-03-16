package models

type SysAP struct {
	Devices map[string]Device `json:"devices"`

	Floorplan Floorplan `json:"floorplan"`

	SysApName string `json:"sysapName"`

	Users Users `json:"users"`

	Error *Error `json:"error,omitempty"`
}
