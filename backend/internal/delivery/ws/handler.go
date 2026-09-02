package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true // Diizinkan untuk komunikasi dashboard
	},
}

// Handler menangani endpoint koneksi masuk WebSocket dan Server-Sent Events (SSE).
type Handler struct {
	hub        *Hub
	jwtManager jwt.Manager
}

// NewHandler membuat instance baru WebSocket & SSE Handler.
func NewHandler(hub *Hub, jwtManager jwt.Manager) *Handler {
	return &Handler{
		hub:        hub,
		jwtManager: jwtManager,
	}
}

// HandleWebSocket meng-upgrade request HTTP ke protokol duplex WebSocket.
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or missing authentication token", nil)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	orgID := uuid.Nil
	if claims.OrganizationID != nil {
		orgID = *claims.OrganizationID
	}

	clientID := uuid.New().String()
	client := NewClient(clientID, claims.UserID, orgID, h.hub)
	h.hub.Register(client)

	// Otomatis subscribe ke topik organisasi jika ada
	if orgID != uuid.Nil {
		h.hub.Subscribe(client, fmt.Sprintf("org:%s", orgID))
	}

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

// HandleSSE menangani streaming data telemetri satu arah via Server-Sent Events (SSE).
// Endpoint ini dilindungi middleware JWT Authenticate — hanya user terautentikasi yang bisa akses (C-1).
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid server ID format", nil)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "streaming unsupported by client", nil)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Header Access-Control-Allow-Origin dikelola secara terpusat oleh middleware CORS global (L-1)

	clientID := uuid.New().String()
	client := NewClient(clientID, uuid.Nil, uuid.Nil, h.hub)
	h.hub.Register(client)
	h.hub.Subscribe(client, fmt.Sprintf("server:%s", serverID))

	defer func() {
		h.hub.Unregister(client)
	}()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", string(msg))
			flusher.Flush()
		}
	}
}

// readPump membaca pesan subscribe/unsubscribe masuk dari client WebSocket.
func (h *Handler) readPump(conn *websocket.Conn, client *Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = conn.Close()
	}()

	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var action struct {
			Type  string `json:"type"`
			Topic string `json:"topic"`
		}
		if err := json.Unmarshal(message, &action); err == nil {
			if action.Type == "subscribe" && action.Topic != "" {
				h.hub.Subscribe(client, action.Topic)
			} else if action.Type == "unsubscribe" && action.Topic != "" {
				h.hub.Unsubscribe(client, action.Topic)
			}
		}
	}
}

// writePump mendistribusikan pesan dari channel Send client ke koneksi WebSocket aktif.
func (h *Handler) writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(msg)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
