package mqtt

import (
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
	broker string
}

func NewClient() *Client {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://localhost:1883" // default
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("backend-receiver")

	client := mqtt.NewClient(opts)

	return &Client{
		client: client,
		broker: broker,
	}
}

func (c *Client) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("⚠️  MQTT connection failed: %v", token.Error())
		return token.Error()
	}
	log.Println("✅ MQTT Receiver connected")
	return nil
}

func (c *Client) Subscribe(topic string, handler mqtt.MessageHandler) error {
	if token := c.client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		log.Printf("⚠️  MQTT subscribe failed: %v", token.Error())
		return token.Error()
	}
	log.Printf("👂 Listening for MQTT messages on topic: %s", topic)
	return nil
}

func (c *Client) Publish(topic string, payload interface{}) error {
	token := c.client.Publish(topic, 0, false, payload)
	token.Wait()
	return token.Error()
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
	log.Println("🔌 MQTT Client disconnected")
}