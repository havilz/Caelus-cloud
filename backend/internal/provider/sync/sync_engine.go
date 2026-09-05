package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type SyncEngine struct {
	serverRepo     domain.ServerRepository
	providerRepo   domain.ProviderRepository
	credRepo       domain.CredentialRepository
	driverFactory  provFactory.Factory
	eventPublisher func(ctx context.Context, event domain.SystemEvent)
	interval       time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
	isRunning      bool
}

func NewSyncEngine(
	serverRepo domain.ServerRepository,
	providerRepo domain.ProviderRepository,
	credRepo domain.CredentialRepository,
	factory provFactory.Factory,
	eventPublisher func(ctx context.Context, event domain.SystemEvent),
	interval time.Duration,
) *SyncEngine {
	if interval < 10*time.Second {
		interval = 60 * time.Second
	}

	return &SyncEngine{
		serverRepo:     serverRepo,
		providerRepo:   providerRepo,
		credRepo:       credRepo,
		driverFactory:  factory,
		eventPublisher: eventPublisher,
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

func (e *SyncEngine) Start(ctx context.Context) {
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return
	}
	e.isRunning = true
	e.stopChan = make(chan struct{})
	e.mu.Unlock()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()

		logger.Info("Multi-Provider SyncEngine active", "interval", e.interval.String())

		for {
			select {
			case <-ctx.Done():
				return
			case <-e.stopChan:
				return
			case <-ticker.C:
				syncCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				if err := e.SyncOnce(syncCtx); err != nil {
					logger.Warn("Failed to synchronize multi-provider status", "error", err)
				}
				cancel()
			}
		}
	}()
}

func (e *SyncEngine) Stop() {
	e.mu.Lock()
	if !e.isRunning {
		e.mu.Unlock()
		return
	}
	e.isRunning = false
	close(e.stopChan)
	e.mu.Unlock()

	e.wg.Wait()
	logger.Info("Multi-Provider SyncEngine stopped successfully")
}

func (e *SyncEngine) SyncOnce(ctx context.Context) error {
	servers, err := e.serverRepo.ListAllRunning(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengambil server running: %w", err)
	}

	for _, srv := range servers {
		if srv.ExternalServerID == nil || *srv.ExternalServerID == "" {
			continue
		}

		prov, err := e.providerRepo.GetByID(ctx, srv.ProviderID)
		if err != nil {
			continue
		}

		if prov.Slug == "custom" {
			continue
		}

		driver, err := e.driverFactory.GetDriver(prov.Slug)
		if err != nil {
			continue
		}

		var cred *domain.Credential
		if srv.CredentialID != nil && *srv.CredentialID != uuid.Nil {
			cred, _ = e.credRepo.GetByID(ctx, *srv.CredentialID)
		}

		remoteServer, err := driver.GetServer(ctx, cred, *srv.ExternalServerID)
		if err != nil {
			logger.Warn("Failed to fetch remote server status from provider",
				"server_id", srv.ID,
				"external_id", *srv.ExternalServerID,
				"provider", prov.Slug,
				"error", err,
			)
			continue
		}

		if remoteServer == nil {
			continue
		}

		needsUpdate := false
		if remoteServer.Status != "" && remoteServer.Status != srv.Status {
			srv.Status = remoteServer.Status
			needsUpdate = true
		}

		if remoteServer.PublicIP != "" && (srv.IPAddress == nil || *srv.IPAddress != remoteServer.PublicIP) {
			ip := remoteServer.PublicIP
			srv.IPAddress = &ip
			needsUpdate = true
		}

		if needsUpdate {
			if err := e.serverRepo.Update(ctx, &srv); err != nil {
				logger.Error("Failed to update synchronized server status in database", "server_id", srv.ID, "error", err)
				continue
			}

			if e.eventPublisher != nil {
				e.eventPublisher(ctx, domain.SystemEvent{
					ID:             uuid.New(),
					OrganizationID: srv.OrganizationID,
					Type:           "server.status_changed",
					SourceResource: fmt.Sprintf("server:%s", srv.ID.String()),
					Data: map[string]any{
						"server_id":   srv.ID.String(),
						"name":        srv.Name,
						"status":      string(srv.Status),
						"ip_address":  srv.IPAddress,
						"provider":    prov.Slug,
						"external_id": *srv.ExternalServerID,
					},
					OccurredAt: time.Now().UTC(),
				})
			}
		}
	}

	return nil
}
