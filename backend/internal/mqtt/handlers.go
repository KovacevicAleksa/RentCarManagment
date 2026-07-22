package mqtt

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/KovacevicAleksa/rentcar/backend/internal/monitoring"
)

type CarTelemetry struct {
	CarID             string    `json:"car_id"`
	EngineTemperature float64   `json:"engine_temperature"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	GasThrottle       float64   `json:"gas_throttle"`
	EngineCoolantTemp float64   `json:"engine_coolant_temp"`
	FuelLevel         float64   `json:"fuel_level"`
	CheckEngine       bool      `json:"check_engine"`
	Timestamp         time.Time `json:"timestamp"`

	// OverheatFrequency is computed by the backend (not sent by the car) and
	// attached before broadcasting so clients get the live derived value.
	OverheatFrequency float64 `json:"overheat_frequency"`
}

type TelemetryBroadcaster interface {
	BroadcastTelemetry(data interface{})
}

type HistorySaver interface {
	SaveTelemetry(carID string, fuel, lat, lon float64) error
}

// TelemetryAlerter inspects incoming telemetry and may raise notifications
// (e.g. engine overheating). It is decoupled so the mqtt package does not
// depend on the notification package directly.
type TelemetryAlerter interface {
	CheckEngineTemp(carID string, engineTemp float64)
}

// TelemetryProcessor accumulates per-car statistics from the telemetry stream
// and returns a derived value (the car's current overheat frequency) to enrich
// the outgoing broadcast.
type TelemetryProcessor interface {
	Observe(carID string, engineTemp float64) float64
}

var broadcaster TelemetryBroadcaster
var historySaver HistorySaver
var alerter TelemetryAlerter
var processor TelemetryProcessor

func SetBroadcaster(b TelemetryBroadcaster) {
	broadcaster = b
	log.Println("MQTT Broadcaster set")
}

func SetHistorySaver(h HistorySaver) {
	historySaver = h
	log.Println("MQTT History Saver set")
}

func SetAlerter(a TelemetryAlerter) {
	alerter = a
	log.Println("MQTT Alerter set")
}

func SetProcessor(p TelemetryProcessor) {
	processor = p
	log.Println("MQTT Processor set")
}

func DefaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("MQTT message received")
	log.Printf("Topic: %s", msg.Topic())
	log.Printf("Payload: %s", string(msg.Payload()))
}

func CarTelemetryHandler(client mqtt.Client, msg mqtt.Message) {
	var telemetry CarTelemetry
	
	if err := json.Unmarshal(msg.Payload(), &telemetry); err != nil {
		log.Printf("Error parsing telemetry: %v", err)
		return
	}

	if telemetry.CarID == "" {
		log.Printf("Missing car_id")
		return
	}

	monitoring.CarTelemetryTotal.WithLabelValues(telemetry.CarID).Inc()

	if telemetry.Timestamp.IsZero() {
		telemetry.Timestamp = time.Now()
	}

	if telemetry.EngineTemperature < -50 || telemetry.EngineTemperature > 200 {
		log.Printf("[%s] Suspicious engine temp: %.1f°C",
			telemetry.CarID, telemetry.EngineTemperature)
	}

	if alerter != nil {
		alerter.CheckEngineTemp(telemetry.CarID, telemetry.EngineTemperature)
	}

	if telemetry.FuelLevel < 0 || telemetry.FuelLevel > 100 {
		if telemetry.FuelLevel < 0 {
			telemetry.FuelLevel = 0
		}
		if telemetry.FuelLevel > 100 {
			telemetry.FuelLevel = 100
		}
	}

	if historySaver != nil {
		if err := historySaver.SaveTelemetry(
			telemetry.CarID,
			telemetry.FuelLevel,
			telemetry.Latitude,
			telemetry.Longitude,
		); err != nil {
			log.Printf("Failed to save history for %s: %v", telemetry.CarID, err)
		}
	}

	if processor != nil {
		telemetry.OverheatFrequency = processor.Observe(telemetry.CarID, telemetry.EngineTemperature)
	}

	if broadcaster != nil {
		broadcaster.BroadcastTelemetry(telemetry)
	}
}