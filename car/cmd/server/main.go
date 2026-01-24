package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type CarTelemetry struct {
	CarID              string    `json:"car_id"`
	EngineTemperature  float64   `json:"engine_temperature"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	GasThrottle        float64   `json:"gas_throttle"`
	EngineCoolantTemp  float64   `json:"engine_coolant_temp"`
	FuelLevel          float64   `json:"fuel_level"`
	Timestamp          time.Time `json:"timestamp"`
}

type CarSimulator struct {
	carID       string
	latitude    float64
	longitude   float64
	heading     float64
	speed       float64
	fuelLevel   float64
	engineTemp  float64
	coolantTemp float64
	throttle    float64
}

func NewCarSimulator(carID string, startLat, startLon float64) *CarSimulator {
	return &CarSimulator{
		carID:       carID,
		latitude:    startLat,
		longitude:   startLon,
		heading:     rand.Float64() * 360,
		speed:       0,
		fuelLevel:   100.0,
		engineTemp:  20.0,
		coolantTemp: 20.0,
		throttle:    0.0,
	}
}

func (s *CarSimulator) GetNextTelemetry() CarTelemetry {
	targetThrottle := rand.Float64()
	s.throttle += (targetThrottle - s.throttle) * 0.3
	s.throttle = math.Max(0, math.Min(1, s.throttle))
	
	targetSpeed := s.throttle * 120
	s.speed += (targetSpeed - s.speed) * 0.2
	
	targetEngineTemp := 60 + s.throttle*60 + rand.Float64()*15
	s.engineTemp += (targetEngineTemp - s.engineTemp) * 0.1
	
	targetCoolantTemp := s.engineTemp - 10 + rand.Float64()*5
	s.coolantTemp += (targetCoolantTemp - s.coolantTemp) * 0.15
	
	fuelConsumption := s.throttle * 0.02
	s.fuelLevel -= fuelConsumption
	if s.fuelLevel < 0 {
		s.fuelLevel = 0
	}
	
	if s.speed > 5 {
		s.heading += (rand.Float64() - 0.5) * 10
		
		distanceKm := (s.speed / 3600) * 5
		s.latitude += (distanceKm * 0.009) * math.Cos(s.heading*math.Pi/180)
		s.longitude += (distanceKm * 0.009) * math.Sin(s.heading*math.Pi/180)
	}
	
	return CarTelemetry{
		CarID:             s.carID,
		EngineTemperature: s.engineTemp,
		Latitude:          s.latitude,
		Longitude:         s.longitude,
		GasThrottle:       s.throttle,
		EngineCoolantTemp: s.coolantTemp,
		FuelLevel:         s.fuelLevel,
		Timestamp:         time.Now(),
	}
}

func simulateCar(client mqtt.Client, carID string, startLat, startLon float64, wg *sync.WaitGroup) {
	defer wg.Done()
	
	sim := NewCarSimulator(carID, startLat, startLon)
	topic := fmt.Sprintf("car/%s/telemetry", carID)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	log.Printf("🚗 Car %s started simulation at (%.6f, %.6f)", carID, startLat, startLon)
	
	for range ticker.C {
		telemetry := sim.GetNextTelemetry()
		
		jsonData, err := json.Marshal(telemetry)
		if err != nil {
			log.Printf("❌ [%s] Error marshaling JSON: %v", carID, err)
			continue
		}
		
		token := client.Publish(topic, 0, false, jsonData)
		token.Wait()
		
		log.Printf("📤 [%s] Engine: %.1f°C, Coolant: %.1f°C, Throttle: %.2f, Fuel: %.1f%%, Pos: (%.6f, %.6f)", 
			carID,
			telemetry.EngineTemperature,
			telemetry.EngineCoolantTemp, 
			telemetry.GasThrottle, 
			telemetry.FuelLevel,
			telemetry.Latitude,
			telemetry.Longitude)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://localhost:1883"
	}
	
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("car-fleet-sender")
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	
	client := mqtt.NewClient(opts)
	
	log.Printf("Connecting to MQTT broker: %s", broker)
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("⚠️  Failed to connect to MQTT: %v, retrying in 5s...", token.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	
	log.Println("✅ Car fleet service connected to MQTT broker")
	
	cars := []struct {
		id  string
		lat float64
		lon float64
	}{
		{"CAR001", 44.7866, 20.4489},
		{"CAR002", 44.8125, 20.4612},
		{"CAR003", 44.8023, 20.4781},
		{"CAR004", 44.7689, 20.4567},
		{"CAR005", 44.8234, 20.4423},
	}
	
	var wg sync.WaitGroup
	
	for _, car := range cars {
		wg.Add(1)
		go simulateCar(client, car.id, car.lat, car.lon, &wg)
	}
	
	wg.Wait()
}