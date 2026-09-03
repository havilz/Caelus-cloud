package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/provider/mock"
)

func createTestServerHelper(driver domain.ProviderDriver, cred *domain.Credential) (*domain.ProviderServer, error) {
	req := domain.CreateServerRequest{
		Name:     "web-prod-1",
		Region:   "ap-southeast-1",
		OSType:   "ubuntu-22.04",
		PlanID:   "std-2vcpu-4gb",
		CPUCores: 2,
		MemoryMB: 4096,
		DiskGB:   50,
	}
	return driver.CreateServer(context.Background(), cred, req)
}

func TestMockDriver_CreateAndGetServer(t *testing.T) {
	driver := mock.NewMockDriver()
	cred := &domain.Credential{ID: uuid.New(), OrganizationID: uuid.New(), ProviderID: uuid.New(), Name: "Mock Creds"}

	server, err := createTestServerHelper(driver, cred)
	if err != nil {
		t.Fatalf("gagal membuat server tiruan: %v", err)
	}

	if server.ExternalID == "" || server.PublicIP == "" || server.Status != domain.ServerStatusRunning {
		t.Errorf("data server hasil provisioning tidak sesuai: %+v", server)
	}

	fetched, err := driver.GetServer(context.Background(), cred, server.ExternalID)
	if err != nil || fetched.Name != "web-prod-1" {
		t.Fatalf("gagal mengambil server berdasarkan external ID: %v", err)
	}
}

func TestMockDriver_ListServers(t *testing.T) {
	driver := mock.NewMockDriver()
	cred := &domain.Credential{ID: uuid.New(), OrganizationID: uuid.New(), ProviderID: uuid.New(), Name: "Mock Creds"}

	_, _ = createTestServerHelper(driver, cred)

	list, err := driver.ListServers(context.Background(), cred)
	if err != nil || len(list) != 1 {
		t.Fatalf("jumlah server pada list tidak sesuai, didapat: %d", len(list))
	}
}

func TestMockDriver_PowerControls(t *testing.T) {
	driver := mock.NewMockDriver()
	cred := &domain.Credential{ID: uuid.New(), OrganizationID: uuid.New(), ProviderID: uuid.New(), Name: "Mock Creds"}
	ctx := context.Background()

	server, _ := createTestServerHelper(driver, cred)

	if err := driver.ShutdownServer(ctx, cred, server.ExternalID); err != nil {
		t.Fatalf("gagal mematikan server: %v", err)
	}
	stopped, _ := driver.GetServer(ctx, cred, server.ExternalID)
	if stopped.Status != domain.ServerStatusStopped {
		t.Errorf("status server harus stopped, didapat: %s", stopped.Status)
	}

	if err := driver.StartServer(ctx, cred, server.ExternalID); err != nil {
		t.Fatalf("gagal menyalakan server: %v", err)
	}
	started, _ := driver.GetServer(ctx, cred, server.ExternalID)
	if started.Status != domain.ServerStatusRunning {
		t.Errorf("status server harus running, didapat: %s", started.Status)
	}

	if err := driver.RebootServer(ctx, cred, server.ExternalID); err != nil {
		t.Fatalf("gagal reboot server: %v", err)
	}
}

func TestMockDriver_ResizeServer(t *testing.T) {
	driver := mock.NewMockDriver()
	cred := &domain.Credential{ID: uuid.New(), OrganizationID: uuid.New(), ProviderID: uuid.New(), Name: "Mock Creds"}
	ctx := context.Background()

	server, _ := createTestServerHelper(driver, cred)

	resizeReq := domain.ResizeServerRequest{
		ExternalID: server.ExternalID,
		CPUCores:   4,
		MemoryMB:   8192,
		DiskGB:     100,
	}
	if err := driver.ResizeServer(ctx, cred, resizeReq); err != nil {
		t.Fatalf("gagal resize server: %v", err)
	}

	resized, _ := driver.GetServer(ctx, cred, server.ExternalID)
	if resized.CPUCores != 4 || resized.MemoryMB != 8192 || resized.DiskGB != 100 {
		t.Errorf("spesifikasi resize tidak sesuai: %+v", resized)
	}
}

func TestMockDriver_DeleteServer(t *testing.T) {
	driver := mock.NewMockDriver()
	cred := &domain.Credential{ID: uuid.New(), OrganizationID: uuid.New(), ProviderID: uuid.New(), Name: "Mock Creds"}
	ctx := context.Background()

	server, _ := createTestServerHelper(driver, cred)

	if err := driver.DeleteServer(ctx, cred, server.ExternalID); err != nil {
		t.Fatalf("gagal menghapus server: %v", err)
	}

	_, err := driver.GetServer(ctx, cred, server.ExternalID)
	if err != domain.ErrNotFound {
		t.Errorf("server yang telah dihapus harus mengembalikan ErrNotFound, didapat: %v", err)
	}
}
