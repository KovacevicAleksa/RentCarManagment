package mqtt

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CarTelemetry struct {
	EngineTemperature float64   `json:"engine_temperature"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	GasThrottle       float64   `json:"gas_throttle"`
	EngineCoolantTemp float64   `json:"engine_coolant_temp"`
	FuelLevel         float64   `json:"fuel_level"`
	Timestamp         time.Time `json:"timestamp"`
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
		return
	}
	
	log.Printf("📥 Car Telemetry - Engine: %.1f°C, Coolant: %.1f°C, Throttle: %.2f, Fuel: %.1f%%, Pos: (%.6f, %.6f)",
		telemetry.EngineTemperature,
		telemetry.EngineCoolantTemp,
		telemetry.GasThrottle,
		telemetry.FuelLevel,
		telemetry.Latitude,
		telemetry.Longitude)
}