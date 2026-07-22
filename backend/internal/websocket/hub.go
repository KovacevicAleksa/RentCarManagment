package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("✅ WebSocket client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("❌ WebSocket client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := make([]*Client, 0, len(h.clients))
			for client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()

			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
		}
	}
}

// wsMessage is the envelope every broadcast is wrapped in, so clients can tell
// telemetry updates apart from notifications on the single WebSocket stream.
type wsMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (h *Hub) BroadcastTelemetry(data interface{}) {
	h.broadcastEnvelope("telemetry", data)
}

func (h *Hub) BroadcastNotification(data interface{}) {
	h.broadcastEnvelope("notification", data)
}

func (h *Hub) broadcastEnvelope(msgType string, payload interface{}) {
	jsonData, err := json.Marshal(wsMessage{Type: msgType, Payload: payload})
	if err != nil {
		log.Printf("❌ Error marshaling %s message: %v", msgType, err)
		return
	}

	select {
	case h.broadcast <- jsonData:
	default:
		log.Printf("⚠️ Broadcast channel full, dropping %s message", msgType)
	}
}