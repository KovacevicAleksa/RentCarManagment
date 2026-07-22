package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBroadcastTelemetryQueuesEnvelope(t *testing.T) {
	h := NewHub()

	h.BroadcastTelemetry(map[string]any{"car_id": "CAR001", "fuel_level": 42.5})

	select {
	case raw := <-h.broadcast:
		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("queued payload is not valid JSON: %v", err)
		}
		if decoded["type"] != "telemetry" {
			t.Errorf("type = %v, want telemetry", decoded["type"])
		}
		payload, ok := decoded["payload"].(map[string]any)
		if !ok || payload["car_id"] != "CAR001" {
			t.Errorf("payload = %v, want car_id CAR001", decoded["payload"])
		}
	default:
		t.Fatal("expected a message to be queued on the broadcast channel")
	}
}

func TestBroadcastNotificationQueuesEnvelope(t *testing.T) {
	h := NewHub()

	h.BroadcastNotification(map[string]any{"id": "n1", "title": "Pregrevanje"})

	select {
	case raw := <-h.broadcast:
		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("queued payload is not valid JSON: %v", err)
		}
		if decoded["type"] != "notification" {
			t.Errorf("type = %v, want notification", decoded["type"])
		}
		payload, ok := decoded["payload"].(map[string]any)
		if !ok || payload["title"] != "Pregrevanje" {
			t.Errorf("payload = %v, want title Pregrevanje", decoded["payload"])
		}
	default:
		t.Fatal("expected a notification to be queued on the broadcast channel")
	}
}

func TestBroadcastTelemetrySkipsUnmarshalable(t *testing.T) {
	h := NewHub()

	// channels cannot be marshaled to JSON; BroadcastTelemetry must drop it.
	h.BroadcastTelemetry(make(chan int))

	select {
	case <-h.broadcast:
		t.Fatal("unmarshalable value should not be queued")
	default:
	}
}

func TestRunDeliversBroadcastToRegisteredClient(t *testing.T) {
	h := NewHub()
	go h.Run()

	client := &Client{hub: h, send: make(chan []byte, 1)}
	h.register <- client // unbuffered: returns only after Run registers it

	h.BroadcastTelemetry(map[string]any{"car_id": "CAR007"})

	select {
	case raw := <-client.send:
		var decoded map[string]any
		_ = json.Unmarshal(raw, &decoded)
		payload, _ := decoded["payload"].(map[string]any)
		if payload["car_id"] != "CAR007" {
			t.Errorf("delivered car_id = %v, want CAR007", decoded["payload"])
		}
	case <-time.After(time.Second):
		t.Fatal("registered client did not receive the broadcast within 1s")
	}
}
