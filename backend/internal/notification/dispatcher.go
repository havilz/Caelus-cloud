package notification

import (
	"context"
	"fmt"

	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// Dispatcher mendefinisikan kontrak terpadu untuk pengiriman notifikasi berbagai saluran.
type Dispatcher interface {
	// SendWebhook mengirimkan payload webhook ke URL tujuan.
	SendWebhook(ctx context.Context, url string, payload webhook.WebhookPayload) error

	// SendEmail mengirimkan email notifikasi ke alamat tujuan.
	SendEmail(ctx context.Context, msg email.EmailMessage) error
}

// UnifiedDispatcher mengimplementasikan Dispatcher menggunakan adapter webhook dan email.
type UnifiedDispatcher struct {
	webhookClient *webhook.Client
	emailClient   *email.Client
}

// NewUnifiedDispatcher membuat instance UnifiedDispatcher baru.
// Parameter webhookClient merupakan adapter pengirim webhook.
// Parameter emailClient merupakan adapter pengirim email SMTP.
// Mengembalikan pointer *UnifiedDispatcher.
func NewUnifiedDispatcher(webhookClient *webhook.Client, emailClient *email.Client) *UnifiedDispatcher {
	return &UnifiedDispatcher{
		webhookClient: webhookClient,
		emailClient:   emailClient,
	}
}

// SendWebhook mengirim notifikasi via HTTP POST ke endpoint eksternal.
func (d *UnifiedDispatcher) SendWebhook(ctx context.Context, url string, payload webhook.WebhookPayload) error {
	if d.webhookClient == nil {
		return fmt.Errorf("webhook client is not initialized")
	}
	logger.Info("Mengirimkan notifikasi webhook", "target_url", url, "event_type", payload.EventType)
	return d.webhookClient.SendWebhook(ctx, url, payload)
}

// SendEmail mengirim notifikasi via SMTP email.
func (d *UnifiedDispatcher) SendEmail(ctx context.Context, msg email.EmailMessage) error {
	if d.emailClient == nil {
		return fmt.Errorf("email client is not initialized")
	}
	logger.Info("Mengirimkan notifikasi email", "to", msg.To, "subject", msg.Subject)
	return d.emailClient.SendEmail(ctx, msg)
}
