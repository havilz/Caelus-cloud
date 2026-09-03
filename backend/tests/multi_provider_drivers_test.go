package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	"github.com/havilz/caelus-cloud/backend/internal/provider/aws"
	"github.com/havilz/caelus-cloud/backend/internal/provider/contabo"
	"github.com/havilz/caelus-cloud/backend/internal/provider/digitalocean"
	"github.com/havilz/caelus-cloud/backend/internal/provider/hetzner"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

func TestMultiProviderDrivers(t *testing.T) {
	ctx := context.Background()
	encKey := []byte("12345678901234567890123456789012")

	encAPIKey, _ := encryptor.Encrypt("test-api-key", encKey)
	encSecretKey, _ := encryptor.Encrypt("test-api-secret", encKey)

	cred := &domain.Credential{
		ID:                 uuid.New(),
		Name:               "Test Credential",
		EncryptedAPIKey:    &encAPIKey,
		EncryptedAPISecret: &encSecretKey,
		Metadata: map[string]any{
			"region": "us-east-1",
		},
	}

	t.Run("AWS EC2 Driver Full Lifecycle", func(t *testing.T) {
		driver := aws.NewEC2Driver(encKey)

		srv, err := driver.CreateServer(ctx, cred, domain.CreateServerRequest{
			Name:     "aws-web-node",
			Region:   "us-east-1",
			CPUCores: 2,
			MemoryMB: 4096,
			DiskGB:   50,
		})
		if err != nil || srv == nil {
			t.Fatalf("expected create server to succeed, got %v", err)
		}
		if srv.Status != domain.ServerStatusRunning {
			t.Errorf("expected running status, got %s", srv.Status)
		}

		got, err := driver.GetServer(ctx, cred, srv.ExternalID)
		if err != nil || got.ExternalID != srv.ExternalID {
			t.Fatalf("expected get server to succeed, got %v", err)
		}

		list, err := driver.ListServers(ctx, cred)
		if err != nil || len(list) != 1 {
			t.Fatalf("expected list of 1 server, got %d, err: %v", len(list), err)
		}

		if err := driver.ShutdownServer(ctx, cred, srv.ExternalID); err != nil {
			t.Fatalf("expected shutdown to succeed, got %v", err)
		}
		if err := driver.StartServer(ctx, cred, srv.ExternalID); err != nil {
			t.Fatalf("expected start to succeed, got %v", err)
		}
		if err := driver.RebootServer(ctx, cred, srv.ExternalID); err != nil {
			t.Fatalf("expected reboot to succeed, got %v", err)
		}

		if err := driver.ResizeServer(ctx, cred, domain.ResizeServerRequest{
			ExternalID: srv.ExternalID,
			CPUCores:   4,
			MemoryMB:   8192,
		}); err != nil {
			t.Fatalf("expected resize to succeed, got %v", err)
		}

		if err := driver.DeleteServer(ctx, cred, srv.ExternalID); err != nil {
			t.Fatalf("expected delete to succeed, got %v", err)
		}
		listAfter, _ := driver.ListServers(ctx, cred)
		if len(listAfter) != 0 {
			t.Fatalf("expected 0 servers after delete, got %d", len(listAfter))
		}
	})

	t.Run("Hetzner Driver Full Lifecycle", func(t *testing.T) {
		driver := hetzner.NewHetznerDriver(encKey)

		srv, err := driver.CreateServer(ctx, cred, domain.CreateServerRequest{
			Name:     "hetzner-web-node",
			Region:   "fsn1",
			CPUCores: 2,
			MemoryMB: 4096,
			DiskGB:   40,
		})
		if err != nil || srv == nil {
			t.Fatalf("expected create server to succeed, got %v", err)
		}

		got, err := driver.GetServer(ctx, cred, srv.ExternalID)
		if err != nil || got.ExternalID != srv.ExternalID {
			t.Fatalf("expected get server to succeed, got %v", err)
		}

		_ = driver.ShutdownServer(ctx, cred, srv.ExternalID)
		_ = driver.StartServer(ctx, cred, srv.ExternalID)
		_ = driver.RebootServer(ctx, cred, srv.ExternalID)
		_ = driver.ResizeServer(ctx, cred, domain.ResizeServerRequest{
			ExternalID: srv.ExternalID,
			CPUCores:   4,
		})
		_ = driver.DeleteServer(ctx, cred, srv.ExternalID)
	})

	t.Run("DigitalOcean Driver Full Lifecycle", func(t *testing.T) {
		driver := digitalocean.NewDigitalOceanDriver(encKey)

		srv, err := driver.CreateServer(ctx, cred, domain.CreateServerRequest{
			Name:     "do-droplet-1",
			Region:   "sgp1",
			CPUCores: 2,
			MemoryMB: 4096,
			DiskGB:   50,
		})
		if err != nil || srv == nil {
			t.Fatalf("expected create server to succeed, got %v", err)
		}

		got, err := driver.GetServer(ctx, cred, srv.ExternalID)
		if err != nil || got.ExternalID != srv.ExternalID {
			t.Fatalf("expected get server to succeed, got %v", err)
		}

		_ = driver.ShutdownServer(ctx, cred, srv.ExternalID)
		_ = driver.StartServer(ctx, cred, srv.ExternalID)
		_ = driver.RebootServer(ctx, cred, srv.ExternalID)
		_ = driver.ResizeServer(ctx, cred, domain.ResizeServerRequest{
			ExternalID: srv.ExternalID,
			CPUCores:   8,
		})
		_ = driver.DeleteServer(ctx, cred, srv.ExternalID)
	})

	t.Run("Contabo Driver Full Lifecycle", func(t *testing.T) {
		driver := contabo.NewContaboDriver(encKey)

		srv, err := driver.CreateServer(ctx, cred, domain.CreateServerRequest{
			Name:     "contabo-node-1",
			Region:   "EU",
			CPUCores: 4,
			MemoryMB: 8192,
			DiskGB:   200,
		})
		if err != nil || srv == nil {
			t.Fatalf("expected create server to succeed, got %v", err)
		}

		got, err := driver.GetServer(ctx, cred, srv.ExternalID)
		if err != nil || got.ExternalID != srv.ExternalID {
			t.Fatalf("expected get server to succeed, got %v", err)
		}

		_ = driver.ShutdownServer(ctx, cred, srv.ExternalID)
		_ = driver.StartServer(ctx, cred, srv.ExternalID)
		_ = driver.RebootServer(ctx, cred, srv.ExternalID)
		_ = driver.ResizeServer(ctx, cred, domain.ResizeServerRequest{
			ExternalID: srv.ExternalID,
			CPUCores:   8,
		})
		_ = driver.DeleteServer(ctx, cred, srv.ExternalID)
	})

	t.Run("Driver Factory Registration", func(t *testing.T) {
		factory := provFactory.NewDriverFactoryWithKey(encKey)

		for _, slug := range []string{"mock", "custom", "aws", "hetzner", "digitalocean", "contabo"} {
			driver, err := factory.GetDriver(slug)
			if err != nil || driver == nil {
				t.Errorf("expected driver for %s to be found in factory, got error %v", slug, err)
			}
		}

		_, err := factory.GetDriver("unknown-provider")
		if err != domain.ErrNotFound {
			t.Errorf("expected ErrNotFound for unknown provider, got %v", err)
		}
	})
}
