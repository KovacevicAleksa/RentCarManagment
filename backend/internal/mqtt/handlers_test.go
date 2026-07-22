package mqtt

import (
	"encoding/json"
	"testing"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/KovacevicAleksa/rentcar/backend/internal/carstats"
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
	carID          string
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
	last  interface{}
}

func (b *fakeBroadcaster) BroadcastTelemetry(data interface{}) {
	b.calls++
	b.last = data
}

// fakeProcessor records the arguments it is called with and returns a fixed
// result so handler enrichment can be asserted.
type fakeProcessor struct {
	gotCarID       string
	gotEngineTemp  float64
	gotCheckEngine bool
	result         carstats.Result
}

func (p *fakeProcessor) Observe(carID string, engineTemp float64, checkEngine bool) carstats.Result {
	p.gotCarID = carID
	p.gotEngineTemp = engineTemp
	p.gotCheckEngine = checkEngine
	return p.result
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

func TestCarTelemetryHandlerEnrichesBroadcastWithReliability(t *testing.T) {
	_, bc := withWiring(t)
	proc := &fakeProcessor{result: carstats.Result{
		OverheatFrequency: 0.5,
		OverheatCount:     1,
		TemperatureScore:  90,
		CheckEngineCount:  3,
		CheckEngineScore:  85,
		Reliability:       88,
	}}
	processor = proc
	t.Cleanup(func() { processor = nil })

	CarTelemetryHandler(nil, msgFor(t, CarTelemetry{
		CarID: "CAR001", EngineTemperature: 112, CheckEngine: true,
	}))

	if proc.gotCarID != "CAR001" || proc.gotEngineTemp != 112 || !proc.gotCheckEngine {
		t.Fatalf("processor called with wrong args: %+v", proc)
	}

	tel, ok := bc.last.(CarTelemetry)
	if !ok {
		t.Fatalf("broadcast payload is not CarTelemetry: %T", bc.last)
	}
	if tel.Reliability != 88 {
		t.Errorf("reliability = %v, want 88", tel.Reliability)
	}
	if tel.TemperatureScore != 90 {
		t.Errorf("temperature_score = %v, want 90", tel.TemperatureScore)
	}
	if tel.CheckEngineScore != 85 {
		t.Errorf("check_engine_score = %v, want 85", tel.CheckEngineScore)
	}
	if tel.CheckEngineCount != 3 {
		t.Errorf("check_engine_count = %v, want 3", tel.CheckEngineCount)
	}
	if tel.OverheatCount != 1 {
		t.Errorf("overheat_count = %v, want 1", tel.OverheatCount)
	}
	if tel.OverheatFrequency != 0.5 {
		t.Errorf("overheat_frequency = %v, want 0.5", tel.OverheatFrequency)
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
