package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorizedPayload = errors.New("agent authentication failed: invalid secret or server ID")
	ErrServerError         = errors.New("caelus API server returned an error")
)

// Client mendefinisikan interface pengiriman payload telemetri ke control plane.
type Client interface {
	SendReport(ctx context.Context, payload *AgentReportPayload) error
}

// HTTPClient mengimplementasikan interface Client menggunakan protokol HTTP/HTTPS.
type HTTPClient struct {
	httpClient  *http.Client
	apiEndpoint string
	serverID    uuid.UUID
	agentSecret string
	maxRetries  int
}

// NewHTTPClient membuat instance baru HTTPClient dengan konfigurasi koneksi dan TLS.
func NewHTTPClient(apiEndpoint string, serverID uuid.UUID, agentSecret string, tlsSkipVerify bool) *HTTPClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkipVerify, //nolint:gosec
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression: false,
	}

	return &HTTPClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
		apiEndpoint: apiEndpoint,
		serverID:    serverID,
		agentSecret: agentSecret,
		maxRetries:  3,
	}
}

// SendReport mengirimkan payload data metrik ke endpoint API dengan mekanisme retry dan otentikasi header.
func (c *HTTPClient) SendReport(ctx context.Context, payload *AgentReportPayload) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal agent payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/telemetry/report", c.apiEndpoint)

	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create http request: %w", err)
		}

		c.applyHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		err = c.evaluateResponse(resp)
		_ = resp.Body.Close()

		if err == nil {
			return nil
		}

		if errors.Is(err, ErrUnauthorizedPayload) {
			return err
		}

		lastErr = err
		time.Sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("failed to send telemetry report after %d attempts: %w", c.maxRetries, lastErr)
}

// applyHeaders menambahkan header autentikasi dan metadata ke request HTTP.
func (c *HTTPClient) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.agentSecret))
	req.Header.Set("X-Server-ID", c.serverID.String())
	req.Header.Set("User-Agent", "caelus-agent/1.0.0")
}

// evaluateResponse memvalidasi status kode HTTP dari respon server.
func (c *HTTPClient) evaluateResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrUnauthorizedPayload
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("%w: status %d, body: %s", ErrServerError, resp.StatusCode, string(respBody))
}
