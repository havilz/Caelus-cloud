package ws

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// EventMessage merepresentasikan payload data real-time yang didistribusikan ke client.
type EventMessage struct {
	Topic string `json:"topic"`
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// Client merepresentasikan satu koneksi client aktif (WebSocket atau SSE stream).
type Client struct {
	ID     string
	UserID uuid.UUID
	OrgID  uuid.UUID
	Send   chan []byte
	topics map[string]bool
	mu     sync.RWMutex
	hub    *Hub
}

// NewClient membuat instance baru Client stream terdaftar.
func NewClient(id string, userID, orgID uuid.UUID, hub *Hub) *Client {
	return &Client{
		ID:     id,
		UserID: userID,
		OrgID:  orgID,
		Send:   make(chan []byte, 256),
		topics: make(map[string]bool),
		hub:    hub,
	}
}

// Hub mengelola registrasi client dan mendistribusikan pesan ke topik saluran yang relevan.
type Hub struct {
	clients    map[*Client]bool
	topicMap   map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *EventMessage
	mu         sync.RWMutex
}

// NewHub menginisialisasi broker pesan terpusat untuk komunikasi real-time client.
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		topicMap:   make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *EventMessage, 1024),
	}
	go h.run()
	return h
}

// run mengelola siklus hidup registrasi, deregistrasi, dan pengiriman pesan secara konkruen.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				for topic := range client.topics {
					if clientsInTopic, exists := h.topicMap[topic]; exists {
						delete(clientsInTopic, client)
						if len(clientsInTopic) == 0 {
							delete(h.topicMap, topic)
						}
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			msgBytes, err := json.Marshal(msg)
			if err == nil {
				if clientsInTopic, ok := h.topicMap[msg.Topic]; ok {
					for client := range clientsInTopic {
						select {
						case client.Send <- msgBytes:
						default:
							close(client.Send)
							delete(h.clients, client)
							delete(clientsInTopic, client)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register mendaftarkan client baru ke dalam Hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister menghapus client dan membersihkan seluruh topik langganan.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe mendaftarkan client ke topik spesifik (misal: "server:<id>").
func (h *Hub) Subscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.topics[topic] = true
	client.mu.Unlock()

	if _, ok := h.topicMap[topic]; !ok {
		h.topicMap[topic] = make(map[*Client]bool)
	}
	h.topicMap[topic][client] = true
}

// Unsubscribe menghapus client dari topik spesifik.
func (h *Hub) Unsubscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	delete(client.topics, topic)
	client.mu.Unlock()

	if clientsInTopic, ok := h.topicMap[topic]; ok {
		delete(clientsInTopic, client)
		if len(clientsInTopic) == 0 {
			delete(h.topicMap, topic)
		}
	}
}

// BroadcastToServer mengirimkan pembaruan metrik ke seluruh client yang memantau server tertentu.
func (h *Hub) BroadcastToServer(serverID uuid.UUID, event string, data any) {
	topic := fmt.Sprintf("server:%s", serverID)
	h.broadcast <- &EventMessage{
		Topic: topic,
		Event: event,
		Data:  data,
	}
}

// BroadcastToOrg mengirimkan pembaruan sistem/alert ke seluruh client di organisasi tertentu.
func (h *Hub) BroadcastToOrg(orgID uuid.UUID, event string, data any) {
	topic := fmt.Sprintf("org:%s", orgID)
	h.broadcast <- &EventMessage{
		Topic: topic,
		Event: event,
		Data:  data,
	}
}

// BroadcastDeployment mengirimkan event deployment ke client yang mendengarkan deployment tertentu.
func (h *Hub) BroadcastDeployment(deploymentID uuid.UUID, event string, data any) {
	topic := fmt.Sprintf("deployment:%s", deploymentID)
	h.broadcast <- &EventMessage{
		Topic: topic,
		Event: event,
		Data:  data,
	}
}

// BroadcastDeploymentLog mengirimkan baris log deployment secara realtime.
func (h *Hub) BroadcastDeploymentLog(deploymentID uuid.UUID, log any) {
	h.BroadcastDeployment(deploymentID, "deployment.log", log)
}

// LogActiveClients mencatat jumlah koneksi aktif untuk keperluan pemantauan.
func (h *Hub) LogActiveClients(logger *slog.Logger) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if logger != nil {
		logger.Debug("active real-time stream connections", "total_clients", len(h.clients))
	}
}

