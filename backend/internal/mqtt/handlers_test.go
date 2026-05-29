package mqtt

import (
	"encoding/json"
	"testing"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// fakeMessage implements paho's mqtt.Message for handler tests.
type fakeMessage struct {
	topic   string
	payload []byte
}

func (m *fakeMessage) Duplicate() bool   { return false }
func (m *fakeMessage) Qos() byte         { return 0 }
func (m *fakeMessage) Retained() bool    { return false }
func (m *fakeMessage) Topic() string     { return m.topic }
func (m *fakeMessage) MessageID() uint16 { return 0 }
func (m *fakeMessage) Payload() []byte   { return m.payload }
func (m *fakeMessage) Ack()              {}

type savedTelemetry struct {
	carID         string
	fuel, lat, lon float64
}

type fakeSaver struct {
	calls []savedTelemetry
	err   error
}

func (s *fakeSaver) SaveTelemetry(carID string, fuel, lat, lon float64) error {
	s.calls = append(s.calls, savedTelemetry{carID, fuel, lat, lon})
	return s.err
}

type fakeBroadcaster struct {
	calls int
}

func (b *fakeBroadcaster) BroadcastTelemetry(data interface{}) {
	b.calls++
}

func withWiring(t *testing.T) (*fakeSaver, *fakeBroadcaster) {
	t.Helper()
	saver := &fakeSaver{}
	bc := &fakeBroadcaster{}
	historySaver = saver
	broadcaster = bc
	t.Cleanup(func() {
		historySaver = nil
		broadcaster = nil
	})
	return saver, bc
}

func msgFor(t *testing.T, tel CarTelemetry) pahomqtt.Message {
	t.Helper()
	payload, err := json.Marshal(tel)
	if err != nil {
		t.Fatalf("failed to marshal telemetry: %v", err)
	}
	return &fakeMessage{topic: "car/" + tel.CarID + "/telemetry", payload: payload}
}

func TestCarTelemetryHandlerSavesAndBroadcasts(t *testing.T) {
	saver, bc := withWiring(t)

	CarTelemetryHandler(nil, msgFor(t, CarTelemetry{
		CarID: "CAR001", FuelLevel: 42.5, Latitude: 44.7866, Longitude: 20.4489,
	}))

	if len(saver.calls) != 1 {
		t.Fatalf("expected 1 save call, got %d", len(saver.calls))
	}
	got := saver.calls[0]
	if got.carID != "CAR001" || got.fuel != 42.5 || got.lat != 44.7866 || got.lon != 20.4489 {
		t.Errorf("saved wrong telemetry: %+v", got)
	}
	if bc.calls != 1 {
		t.Errorf("expected 1 broadcast, got %d", bc.calls)
	}
}

func TestCarTelemetryHandlerClampsFuelHigh(t *testing.T) {
	saver, _ := withWiring(t)

	CarTelemetryHandler(nil, msgFor(t, CarTelemetry{CarID: "CAR001", FuelLevel: 150}))

	if saver.calls[0].fuel != 100 {
		t.Errorf("fuel = %f, want clamped to 100", saver.calls[0].fuel)
	}
}

func TestCarTelemetryHandlerClampsFuelLow(t *testing.T) {
	saver, _ := withWiring(t)

	CarTelemetryHandler(nil, msgFor(t, CarTelemetry{CarID: "CAR001", FuelLevel: -10}))

	if saver.calls[0].fuel != 0 {
		t.Errorf("fuel = %f, want clamped to 0", saver.calls[0].fuel)
	}
}

func TestCarTelemetryHandlerIgnoresMissingCarID(t *testing.T) {
	saver, bc := withWiring(t)

	CarTelemetryHandler(nil, msgFor(t, CarTelemetry{FuelLevel: 50}))

	if len(saver.calls) != 0 || bc.calls != 0 {
		t.Errorf("telemetry without car_id should be ignored, saves=%d broadcasts=%d", len(saver.calls), bc.calls)
	}
}

func TestCarTelemetryHandlerIgnoresInvalidJSON(t *testing.T) {
	saver, bc := withWiring(t)

	CarTelemetryHandler(nil, &fakeMessage{topic: "car/x/telemetry", payload: []byte("{not json")})

	if len(saver.calls) != 0 || bc.calls != 0 {
		t.Errorf("invalid JSON should be ignored, saves=%d broadcasts=%d", len(saver.calls), bc.calls)
	}
}
