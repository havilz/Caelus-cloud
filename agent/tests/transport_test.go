package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

// TestTransport_SendReportSuccess memverifikasi keberhasilan pengiriman payload metrik dan validitas header.
func TestTransport_SendReportSuccess(t *testing.T) {
	serverID := uuid.New()
	secret := "test-agent-secret"

	var receivedHeaders http.Header
	var receivedPayload transport.AgentReportPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := transport.NewHTTPClient(server.URL, serverID, secret, true)
	ctx := context.Background()

	payload := &transport.AgentReportPayload{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
		Host: transport.HostMetrics{
			CPUUsagePct: 25.5,
			CPUCores:    4,
		},
		DockerAvailable: true,
	}

	err := client.SendReport(ctx, payload)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if receivedHeaders.Get("Authorization") != "Bearer "+secret {
		t.Errorf("expected auth header 'Bearer %s', got '%s'", secret, receivedHeaders.Get("Authorization"))
	}
	if receivedHeaders.Get("X-Server-ID") != serverID.String() {
		t.Errorf("expected X-Server-ID '%s', got '%s'", serverID.String(), receivedHeaders.Get("X-Server-ID"))
	}
	if receivedPayload.Host.CPUUsagePct != 25.5 {
		t.Errorf("expected CPUUsagePct 25.5, got %f", receivedPayload.Host.CPUUsagePct)
	}
}

// TestTransport_SendReportUnauthorized memverifikasi penanganan error 401 Unauthorized dari server.
func TestTransport_SendReportUnauthorized(t *testing.T) {
	serverID := uuid.New()
	secret := "wrong-secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	client := transport.NewHTTPClient(server.URL, serverID, secret, true)
	ctx := context.Background()

	payload := &transport.AgentReportPayload{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
	}

	err := client.SendReport(ctx, payload)
	if err != transport.ErrUnauthorizedPayload {
		t.Fatalf("expected ErrUnauthorizedPayload, got: %v", err)
	}
}

// TestTransport_SendReportRetrySuccess memverifikasi mekanisme retry hingga berhasil ketika server sempat mengalami error sementara.
func TestTransport_SendReportRetrySuccess(t *testing.T) {
	serverID := uuid.New()
	secret := "test-secret"

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		curr := atomic.AddInt32(&attempts, 1)
		if curr < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"temporary outage"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := transport.NewHTTPClient(server.URL, serverID, secret, true)
	ctx := context.Background()

	payload := &transport.AgentReportPayload{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
	}

	err := client.SendReport(ctx, payload)
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}
