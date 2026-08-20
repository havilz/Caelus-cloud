package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
)

// TestWSHub_RegisterAndBroadcast memverifikasi pengiriman pesan real-time ke subscriber topik server dan organisasi.
func TestWSHub_RegisterAndBroadcast(t *testing.T) {
	hub := ws.NewHub()
	userID := uuid.New()
	orgID := uuid.New()
	serverID := uuid.New()

	client := ws.NewClient("client-1", userID, orgID, hub)
	hub.Register(client)
	hub.Subscribe(client, "server:"+serverID.String())
	hub.Subscribe(client, "org:"+orgID.String())

	time.Sleep(50 * time.Millisecond)

	// Broadcast ke server
	hub.BroadcastToServer(serverID, "metrics.updated", map[string]any{"cpu": 55.0})

	select {
	case msg := <-client.Send:
		var event ws.EventMessage
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Event != "metrics.updated" {
			t.Errorf("expected event 'metrics.updated', got '%s'", event.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for server broadcast message")
	}

	// Broadcast ke organisasi
	hub.BroadcastToOrg(orgID, "alert.created", map[string]any{"title": "High CPU"})

	select {
	case msg := <-client.Send:
		var event ws.EventMessage
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Event != "alert.created" {
			t.Errorf("expected event 'alert.created', got '%s'", event.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for org broadcast message")
	}

	hub.Unregister(client)
}
