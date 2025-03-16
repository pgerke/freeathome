package models

import (
	"encoding/json"
	"testing"
)

func TestDeserializeValid(t *testing.T) {
	serialized := `{"Test":{"devices":{},"floorplan":{"floors":{}},"sysapName":"Test System Access Point","users":{}}}`
	var config Configuration
	err := json.Unmarshal([]byte(serialized), &config)

	if err != nil {
		t.Fatalf("Failed to deserialize JSON: %v", err)
	}

	expectedName := "Test System Access Point"
	if config["Test"].SysApName != expectedName {
		t.Errorf("Expected SysapName to be %q, got %q", expectedName, config["Test"].SysApName)
	}
}
