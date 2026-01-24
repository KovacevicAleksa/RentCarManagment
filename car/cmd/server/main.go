package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	
	broker := os.Getenv("MQTT_BROKER")
	
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("car-sender")
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
	
	log.Println("✅ Car service connected to MQTT broker")
	
	topic := "test/topic"
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	counter := 0
	for range ticker.C {
		counter++
		message := fmt.Sprintf("Car update #%d - timestamp: %s", counter, time.Now().Format(time.RFC3339))
		
		token := client.Publish(topic, 0, false, message)
		token.Wait()
		
		log.Printf("📤 Sent: %s", message)
	}
}