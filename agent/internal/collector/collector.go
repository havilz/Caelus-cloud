package collector

import (
	"context"

	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

// Collector mendefinisikan kontrak pengumpulan metrik sistem host lokal.
type Collector interface {
	Collect(ctx context.Context) (*transport.HostMetrics, error)
}
