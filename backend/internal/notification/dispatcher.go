package notification

import (
	"context"
	"fmt"

	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type Dispatcher interface {
	SendWebhook(ctx context.Context, url string, payload webhook.WebhookPayload) error

	SendEmail(ctx context.Context, msg email.EmailMessage) error
}

type UnifiedDispatcher struct {
	webhookClient *webhook.Client
	emailClient   *email.Client
}

func NewUnifiedDispatcher(webhookClient *webhook.Client, emailClient *email.Client) *UnifiedDispatcher {
	return &UnifiedDispatcher{
		webhookClient: webhookClient,
		emailClient:   emailClient,
	}
}

func (d *UnifiedDispatcher) SendWebhook(ctx context.Context, url string, payload webhook.WebhookPayload) error {
	if d.webhookClient == nil {
		return fmt.Errorf("webhook client is not initialized")
	}
	logger.Info("Sending webhook notification", "target_url", url, "event_type", payload.EventType)
	return d.webhookClient.SendWebhook(ctx, url, payload)
}

func (d *UnifiedDispatcher) SendEmail(ctx context.Context, msg email.EmailMessage) error {
	if d.emailClient == nil {
		return fmt.Errorf("email client is not initialized")
	}
	logger.Info("Sending email notification", "to", msg.To, "subject", msg.Subject)
	return d.emailClient.SendEmail(ctx, msg)
}
