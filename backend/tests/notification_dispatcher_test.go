package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
)

func TestWebhook_DeliveryAndSignature(t *testing.T) {
	var receivedSignature string
	var receivedPayload webhook.WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Caelus-Signature-256")
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := webhook.NewClient("test_secret_key_123")

	payload := webhook.WebhookPayload{
		EventID:   "evt-1234",
		EventType: "metric_threshold",
		RuleName:  "High CPU Alert",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: map[string]any{
			"cpu_usage_percent": 98.2,
		},
	}

	err := client.SendWebhook(context.Background(), server.URL, payload)
	if err != nil {
		t.Fatalf("failed to send webhook: %v", err)
	}

	if receivedSignature == "" {
		t.Errorf("expected HMAC signature in request header, got empty")
	}
	if receivedPayload.EventID != "evt-1234" {
		t.Errorf("expected EventID 'evt-1234', got '%s'", receivedPayload.EventID)
	}
}

func TestEmail_HTMLTemplate(t *testing.T) {
	html := email.BuildAlertHTMLTemplate("Critical Alert", "Auto-Heal DB Server", "server_status_changed", "Server stopped responding")
	if len(html) == 0 {
		t.Errorf("expected non-empty HTML template")
	}
}

func TestUnifiedDispatcher_Delegation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhookClient := webhook.NewClient("secret")
	emailClient := email.NewClient(email.Config{})
	dispatcher := notification.NewUnifiedDispatcher(webhookClient, emailClient)

	err := dispatcher.SendWebhook(context.Background(), server.URL, webhook.WebhookPayload{
		EventType: "test",
	})
	if err != nil {
		t.Fatalf("dispatcher failed to send webhook: %v", err)
	}

	err = dispatcher.SendEmail(context.Background(), email.EmailMessage{
		To:      "admin@example.com",
		Subject: "Test Alert",
	})
	if err != nil {
		t.Fatalf("dispatcher failed to send simulated email: %v", err)
	}
}
