package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// Client mengimplementasikan domain.MetricsQueryAdapter untuk berinteraksi dengan HTTP API Prometheus.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient membuat instance baru adapter Prometheus Client.
func NewClient(baseURL string) domain.MetricsQueryAdapter {
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryInstant mengeksekusi instant PromQL query ke endpoint /api/v1/query.
func (c *Client) QueryInstant(ctx context.Context, query string) (any, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query?query=%s", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prometheus query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status code %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	return result, nil
}

// QueryRange mengeksekusi range PromQL query ke endpoint /api/v1/query_range.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (any, error) {
	stepSec := int(step.Seconds())
	if stepSec <= 0 {
		stepSec = 15
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%ds",
		c.baseURL,
		url.QueryEscape(query),
		start.Unix(),
		end.Unix(),
		stepSec,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus range request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prometheus range query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus range query returned status code %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus range response: %w", err)
	}

	return result, nil
}
