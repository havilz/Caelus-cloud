package collector

import (
	"context"

	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

type Collector interface {
	Collect(ctx context.Context) (*transport.HostMetrics, error)
}
