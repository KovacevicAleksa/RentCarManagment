package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	client *Client
}

type MessageHandler mqtt.MessageHandler

func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Start(topics map[string]MessageHandler) error {
	if err := s.client.Connect(); err != nil {
		return err
	}

	for topic, handler := range topics {
		if err := s.client.Subscribe(topic, mqtt.MessageHandler(handler)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) PublishMessage(topic string, payload interface{}) error {
	return s.client.Publish(topic, payload)
}

func (s *Service) Stop() {
	s.client.Disconnect()
}