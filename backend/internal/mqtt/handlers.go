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
	Timestamp         time.Time `json:"timestamp"`
}

type TelemetryBroadcaster interface {
	BroadcastTelemetry(data interface{})
}

type HistorySaver interface {
	SaveTelemetry(carID string, fuel, lat, lon float64) error
}

var broadcaster TelemetryBroadcaster
var historySaver HistorySaver

func SetBroadcaster(b TelemetryBroadcaster) {
	broadcaster = b
	log.Println("MQTT Broadcaster set")
}

func SetHistorySaver(h HistorySaver) {
	historySaver = h
	log.Println("MQTT History Saver set")
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

	if broadcaster != nil {
		broadcaster.BroadcastTelemetry(telemetry)
	}
}