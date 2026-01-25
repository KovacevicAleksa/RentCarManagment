package mqtt

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

var broadcaster TelemetryBroadcaster

func SetBroadcaster(b TelemetryBroadcaster) {
	broadcaster = b
	log.Println("✅ MQTT Broadcaster postavljen")
}

func DefaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("📩 MQTT Poruka primljena!")
	log.Printf("   Topic: %s", msg.Topic())
	log.Printf("   Poruka: %s", string(msg.Payload()))
}

func CarTelemetryHandler(client mqtt.Client, msg mqtt.Message) {
	var telemetry CarTelemetry
	
	if err := json.Unmarshal(msg.Payload(), &telemetry); err != nil {
		log.Printf("❌ Error parsing telemetry: %v", err)
		log.Printf("   Payload: %s", string(msg.Payload()))
		return
	}

	if telemetry.CarID == "" {
		log.Printf("⚠️ Telemetry missing car_id")
		return
	}

	if telemetry.Timestamp.IsZero() {
		telemetry.Timestamp = time.Now()
	}

	if telemetry.EngineTemperature < -50 || telemetry.EngineTemperature > 200 {
		log.Printf("⚠️ [%s] Suspicious engine temperature: %.1f°C", 
			telemetry.CarID, telemetry.EngineTemperature)
	}

	if telemetry.FuelLevel < 0 || telemetry.FuelLevel > 100 {
		log.Printf("⚠️ [%s] Invalid fuel level: %.1f%%", 
			telemetry.CarID, telemetry.FuelLevel)
		if telemetry.FuelLevel < 0 {
			telemetry.FuelLevel = 0
		}
		if telemetry.FuelLevel > 100 {
			telemetry.FuelLevel = 100
		}
	}

	log.Printf("📥 [%s] Engine: %.1f°C, Coolant: %.1f°C, Throttle: %.2f, Fuel: %.1f%%, Location: (%.4f, %.4f)",
		telemetry.CarID,
		telemetry.EngineTemperature,
		telemetry.EngineCoolantTemp,
		telemetry.GasThrottle,
		telemetry.FuelLevel,
		telemetry.Latitude,
		telemetry.Longitude)

	if broadcaster != nil {
		broadcaster.BroadcastTelemetry(telemetry)
	} else {
		log.Printf("⚠️ Broadcaster nije postavljen, telemetrija se ne emituje!")
	}
}