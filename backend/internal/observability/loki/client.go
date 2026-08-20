package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// Client mengimplementasikan domain.LogQueryAdapter untuk mengambil stream log dari Grafana Loki.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient membuat instance baru adapter Loki Client.
func NewClient(baseURL string) domain.LogQueryAdapter {
	if baseURL == "" {
		baseURL = "http://localhost:3100"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryLogs mengeksekusi query LogQL ke endpoint /loki/api/v1/query_range.
func (c *Client) QueryLogs(ctx context.Context, query string, start, end time.Time, limit int) ([]domain.LokiLogEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	endpoint := fmt.Sprintf(
		"%s/loki/api/v1/query_range?query=%s&start=%d&end=%d&limit=%d",
		c.baseURL,
		url.QueryEscape(query),
		start.UnixNano(),
		end.UnixNano(),
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create loki request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute loki query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loki returned status code %d", resp.StatusCode)
	}

	var responseData struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode loki response: %w", err)
	}

	entries := make([]domain.LokiLogEntry, 0)
	for _, res := range responseData.Data.Result {
		for _, val := range res.Values {
			if len(val) >= 2 {
				nanoTime, _ := strconv.ParseInt(val[0], 10, 64)
				entries = append(entries, domain.LokiLogEntry{
					Timestamp: time.Unix(0, nanoTime),
					Line:      val[1],
					Labels:    res.Stream,
				})
			}
		}
	}

	return entries, nil
}
