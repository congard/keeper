package transport

import (
	"encoding/json"
	"testing"
)

func TestStatus_JSONRoundTrip(t *testing.T) {
	orig := Status{
		Type:    StatusTypeError,
		Message: "fail",
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var restored Status
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if restored.Type != orig.Type {
		t.Errorf("Type = %v, want %v", restored.Type, orig.Type)
	}
	if restored.Message != orig.Message {
		t.Errorf("Message = %q, want %q", restored.Message, orig.Message)
	}
}
