package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage"
	"github.com/havilz/caelus-cloud/backend/internal/storage/mock"
	backupUcPkg "github.com/havilz/caelus-cloud/backend/internal/usecase/backup"
	storageUcPkg "github.com/havilz/caelus-cloud/backend/internal/usecase/storage"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func TestStorageAndBackupHTTP_Endpoints(t *testing.T) {
	jwtManager := jwt.NewJWTManager(&config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}, "caelus-test")

	orgID := uuid.New()
	user := &domain.User{ID: uuid.New(), Email: "test@caelus.cloud", FullName: "Storage Admin", IsActive: true}

	tokenPair, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}
	token := tokenPair.AccessToken

	bucketRepo := newMockBucketRepo()
	backupRepo := newMockBackupRepo()
	serverRepo := newMockServerRepo()

	hostname := "storage-node-01.local"
	server := &domain.Server{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "storage-node-01",
		Hostname:       &hostname,
		Status:         domain.ServerStatusRunning,
	}
	_ = serverRepo.Create(context.Background(), server)

	factory := storage.NewStorageFactory()
	mockAdapter := mock.NewMockStorageAdapter()
	factory.RegisterAdapter(domain.StorageProviderMinIO, mockAdapter)

	storageUc := storageUcPkg.NewStorageUsecase(bucketRepo, factory)
	backupUc := backupUcPkg.NewBackupUsecase(backupRepo, serverRepo, bucketRepo, factory)

	routerConfig := deliveryHttp.RouterConfig{
		JWTManager: jwtManager,
		Handlers: deliveryHttp.Handlers{
			StorageHandler: v1.NewStorageHandler(storageUc),
			BackupHandler:  v1.NewBackupHandler(backupUc),
		},
	}
	router := deliveryHttp.NewRouter(routerConfig)

	// 1. POST /api/v1/storage/buckets (Create Bucket)
	createBucketBody, _ := json.Marshal(map[string]any{
		"name":          "tenant-media-bucket",
		"provider_type": "minio",
		"region":        "us-east-1",
		"is_public":     false,
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/storage/buckets", bytes.NewReader(createBucketBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for bucket creation, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. GET /api/v1/storage/buckets (List Buckets)
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/storage/buckets", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list buckets, got %d", w.Code)
	}

	// 3. POST /api/v1/storage/buckets/tenant-media-bucket/objects/signed-url (Presigned URL)
	signedURLBody, _ := json.Marshal(map[string]any{
		"key":            "avatars/user1.png",
		"operation":      "upload",
		"expiry_minutes": 30,
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/storage/buckets/tenant-media-bucket/objects/signed-url", bytes.NewReader(signedURLBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for signed URL generation, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. POST /api/v1/backups/policies (Create Policy)
	createPolicyBody, _ := json.Marshal(map[string]any{
		"server_id":       server.ID,
		"name":            "Weekly Snapshot",
		"cron_expression": "0 1 * * 0",
		"retention_days":  14,
		"include_disks":   true,
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/backups/policies", bytes.NewReader(createPolicyBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for backup policy creation, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 5. POST /api/v1/backups/trigger/{server_id} (Trigger Backup)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/backups/trigger/%s", server.ID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for backup trigger, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 6. GET /api/v1/backups/records (List Records)
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/backups/records", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for backup records, got %d", w.Code)
	}
}
