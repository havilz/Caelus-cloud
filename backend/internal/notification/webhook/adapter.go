package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookPayload struct {
	EventID        string         `json:"event_id"`
	EventType      string         `json:"event_type"`
	OrganizationID string         `json:"organization_id"`
	RuleID         string         `json:"rule_id,omitempty"`
	RuleName       string         `json:"rule_name,omitempty"`
	Timestamp      string         `json:"timestamp"`
	Data           map[string]any `json:"data"`
}

type Client struct {
	httpClient *http.Client
	secretKey  string
}

func NewClient(secretKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		secretKey: secretKey,
	}
}

func (c *Client) SendWebhook(ctx context.Context, targetURL string, payload WebhookPayload) error {
	if targetURL == "" {
		return fmt.Errorf("target webhook URL cannot be empty")
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Caelus-Cloud-Webhook/1.0")

	if c.secretKey != "" {
		mac := hmac.New(sha256.New, []byte(c.secretKey))
		mac.Write(bodyBytes)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Caelus-Signature-256", signature)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint responded with HTTP status %d", resp.StatusCode)
	}

	return nil
}
