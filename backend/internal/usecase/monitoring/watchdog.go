package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// HeartbeatWatchdog memantau liveness dan detak jantung (heartbeat) telemetri seluruh instance server.
type HeartbeatWatchdog struct {
	serverRepo     domain.ServerRepository
	metricRepo     domain.MetricRepository
	wsHub          *ws.Hub
	eventEmitter   func(ctx context.Context, event domain.SystemEvent)
	timeout        time.Duration
	checkInterval  time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
	isRunning      bool
}

// NewHeartbeatWatchdog membuat instance baru HeartbeatWatchdog.
// Parameter serverRepo merupakan repositori data server.
// Parameter metricRepo merupakan repositori metrik telemetri time-series.
// Parameter wsHub merupakan WebSocket Hub untuk real-time status broadcasting.
// Parameter eventEmitter merupakan callback pengirim event ke Central Event Dispatcher.
// Parameter timeout merupakan batas waktu tidak adanya metrik sebelum server ditandai offline/stopped.
func NewHeartbeatWatchdog(
	serverRepo domain.ServerRepository,
	metricRepo domain.MetricRepository,
	wsHub *ws.Hub,
	eventEmitter func(ctx context.Context, event domain.SystemEvent),
	timeout time.Duration,
) *HeartbeatWatchdog {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &HeartbeatWatchdog{
		serverRepo:    serverRepo,
		metricRepo:    metricRepo,
		wsHub:         wsHub,
		eventEmitter:  eventEmitter,
		timeout:       timeout,
		checkInterval: 10 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// Start menjalankan background loop watchdog secara asinkron.
func (w *HeartbeatWatchdog) Start() {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = true
	w.mu.Unlock()

	w.wg.Add(1)
	go w.runLoop()
	logger.Info("Heartbeat Liveness Watchdog berhasil dijalankan", "timeout", w.timeout.String())
}

// Stop menghentikan background loop watchdog secara anggun.
func (w *HeartbeatWatchdog) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = false
	close(w.stopChan)
	w.mu.Unlock()

	w.wg.Wait()
	logger.Info("Heartbeat Liveness Watchdog berhasil dihentikan")
}

// runLoop merupakan loop periodik pengecekan status detak jantung seluruh server aktif.
func (w *HeartbeatWatchdog) runLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.evaluateHeartbeats()
		}
	}
}

// evaluateHeartbeats memeriksa seluruh server yang berstatus running dan menandai server mati jika timeout.
func (w *HeartbeatWatchdog) evaluateHeartbeats() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	servers, err := w.serverRepo.ListAllRunning(ctx)
	if err != nil {
		logger.Error("Gagal mengambil daftar server running untuk watchdog", "error", err)
		return
	}

	now := time.Now().UTC()
	for _, srv := range servers {

		latestMetric, err := w.metricRepo.GetLatestByServerID(ctx, srv.ID)
		if err != nil || latestMetric == nil {
			continue
		}

		// Jika selisih waktu metrik terakhir melebihi batas timeout
		if now.Sub(latestMetric.RecordedAt) > w.timeout {
			logger.Warn("Server terdeteksi offline karena tidak mengirimkan heartbeat telemetri",
				"server_id", srv.ID,
				"server_name", srv.Name,
				"last_heartbeat", latestMetric.RecordedAt,
				"elapsed", now.Sub(latestMetric.RecordedAt).String(),
			)

			// Update status server di database
			if err := w.serverRepo.UpdateStatus(ctx, srv.ID, domain.ServerStatusStopped); err != nil {
				logger.Error("Gagal memperbarui status server menjadi stopped", "server_id", srv.ID, "error", err)
				continue
			}

			// Broadcast realtime event ke frontend via WebSocket
			if w.wsHub != nil {
				w.wsHub.BroadcastToOrg(srv.OrganizationID, "server.status_changed", map[string]any{
					"server_id":   srv.ID,
					"old_status":  "running",
					"new_status":  "stopped",
					"reason":      "heartbeat_timeout",
					"detected_at": now,
				})
			}

			// Dispatch system event ke Central Event Dispatcher (Rule Engine)
			if w.eventEmitter != nil {
				w.eventEmitter(ctx, domain.SystemEvent{
					ID:             uuid.New(),
					OrganizationID: srv.OrganizationID,
					Type:           "server_status_changed",
					SourceResource: "server:" + srv.ID.String(),
					Data: map[string]any{
						"old_status":  "running",
						"new_status":  "stopped",
						"reason":      "heartbeat_timeout",
						"server_name": srv.Name,
						"server_id":   srv.ID.String(),
					},
					OccurredAt: now,
				})
			}
		}
	}
}
