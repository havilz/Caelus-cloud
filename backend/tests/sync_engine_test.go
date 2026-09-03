package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	provSync "github.com/havilz/caelus-cloud/backend/internal/provider/sync"
)

func TestSyncEngine(t *testing.T) {
	ctx := context.Background()
	serverRepo := newMockServerRepo()
	providerRepo := newMockProviderRepo()
	credRepo := newMockCredRepo()
	factory := provFactory.NewDriverFactoryWithKey([]byte("12345678901234567890123456789012"))

	orgID := uuid.New()
	provID := uuid.New()
	providerRepo.providers[provID] = &domain.Provider{
		ID:       provID,
		Name:     "DigitalOcean",
		Slug:     "digitalocean",
		IsActive: true,
	}

	extID := "do-test-1234"
	oldIP := "100.100.100.1"
	srvID := uuid.New()

	serverRepo.servers[srvID] = &domain.Server{
		ID:               srvID,
		OrganizationID:   orgID,
		ProviderID:       provID,
		ExternalServerID: &extID,
		Name:             "remote-do-node",
		IPAddress:        &oldIP,
		Status:           domain.ServerStatusRunning,
	}

	var eventsEmitted []domain.SystemEvent
	eventPublisher := func(ctx context.Context, event domain.SystemEvent) {
		eventsEmitted = append(eventsEmitted, event)
	}

	engine := provSync.NewSyncEngine(
		serverRepo,
		providerRepo,
		credRepo,
		factory,
		eventPublisher,
		10*time.Second,
	)

	if err := engine.SyncOnce(ctx); err != nil {
		t.Fatalf("expected syncOnce to succeed, got %v", err)
	}

	updatedServer, err := serverRepo.GetByID(ctx, srvID)
	if err != nil {
		t.Fatalf("expected to get updated server, got %v", err)
	}
	if updatedServer.IPAddress == nil || *updatedServer.IPAddress == oldIP {
		t.Errorf("expected IP address to be synchronized, got %v", updatedServer.IPAddress)
	}

	if len(eventsEmitted) == 0 {
		t.Errorf("expected system event to be emitted on status/IP sync")
	}

	engine.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	engine.Stop()
}
