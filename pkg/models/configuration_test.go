package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDeserializeValid(t *testing.T) {
	serialized := `{"Test":{"devices":{},"floorplan":{"floors":{}},"sysapName":"Test System Access Point","users":{}}}`
	var config Configuration
	err := json.Unmarshal([]byte(serialized), &config)

	if err != nil {
		t.Fatalf("failed to deserialize JSON: %v", err)
	}

	expectedName := "Test System Access Point"
	if config["Test"].SysApName != expectedName {
		t.Errorf("Expected SysapName to be %q, got %q", expectedName, config["Test"].SysApName)
	}
}

func TestDeserializeFromFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "configuration.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read JSON test file: %v", err)
	}

	var config Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// // Assertions
	// if device.ID != "12345" {
	// 	t.Errorf("expected ID to be '12345', got '%s'", device.ID)
	// }

	// if device.Name == nil || *device.Name != "Stehlampe" {
	// 	t.Errorf("expected name to be 'Stehlampe', got '%v'", device.Name)
	// }

	// if device.Parameters == nil {
	// 	t.Fatal("expected parameters to be set, got nil")
	// }
}
