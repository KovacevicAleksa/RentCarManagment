package mqtt

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func DefaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("📩 MQTT Poruka primljena!")
	log.Printf("   Topic: %s", msg.Topic())
	log.Printf("   Poruka: %s", string(msg.Payload()))
}
