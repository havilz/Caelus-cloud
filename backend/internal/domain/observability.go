package domain

import (
	"context"
	"time"
)

type LokiLogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type MetricsQueryAdapter interface {
	QueryInstant(ctx context.Context, query string) (any, error)
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (any, error)
}

type LogQueryAdapter interface {
	QueryLogs(ctx context.Context, query string, start, end time.Time, limit int) ([]LokiLogEntry, error)
}
