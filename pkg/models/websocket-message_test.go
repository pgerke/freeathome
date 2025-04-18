package models

import (
	"encoding/json"
	"testing"
)

func TestDeserializeValidWebSocketMessage(t *testing.T) {
	serialized := `{"00000000-0000-0000-0000-000000000000": {"datapoints": {"ABB7F59451FB/ch0000/odp0000": "0"},"parameters": {},"devices": {},"devicesAdded": [],"devicesRemoved": [],"scenesTriggered": {}}}`
	var message WebSocketMessage
	err := json.Unmarshal([]byte(serialized), &message)

	if err != nil {
		t.Fatalf("failed to deserialize JSON: %v", err)
	}

	if len(message) != 1 {
		t.Errorf("Expected message to contain one system access point, got %d", len(message))
	}

	if len(message[EmptyUUID].Datapoints) != 1 {
		t.Errorf("Expected message to contain one datapoint, got %d", len(message[EmptyUUID].Datapoints))
	}

	if message[EmptyUUID].Datapoints["ABB7F59451FB/ch0000/odp0000"] != "0" {
		t.Errorf("Expected datapoint value to be '0', got '%s'", message[EmptyUUID].Datapoints["ABB7F59451FB/ch0000/odp0000"])
	}
}
