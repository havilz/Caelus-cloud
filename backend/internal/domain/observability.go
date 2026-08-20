package domain

import (
	"context"
	"time"
)

// LokiLogEntry merepresentasikan satu baris log terstruktur dari Grafana Loki.
type LokiLogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// MetricsQueryAdapter mendefinisikan kontrak adapter query ke engine Prometheus.
type MetricsQueryAdapter interface {
	QueryInstant(ctx context.Context, query string) (any, error)
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (any, error)
}

// LogQueryAdapter mendefinisikan kontrak adapter query ke engine Grafana Loki.
type LogQueryAdapter interface {
	QueryLogs(ctx context.Context, query string, start, end time.Time, limit int) ([]LokiLogEntry, error)
}
