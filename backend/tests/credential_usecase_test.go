package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

type mockCredRepo struct {
	creds map[uuid.UUID]*domain.Credential
}

func newMockCredRepo() *mockCredRepo {
	return &mockCredRepo{creds: make(map[uuid.UUID]*domain.Credential)}
}

func (m *mockCredRepo) Create(ctx context.Context, cred *domain.Credential) error {
	m.creds[cred.ID] = cred
	return nil
}

func (m *mockCredRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Credential, error) {
	if c, exists := m.creds[id]; exists {
		return c, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockCredRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Credential, error) {
	var list []domain.Credential
	for _, c := range m.creds {
		if c.OrganizationID == orgID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *mockCredRepo) Update(ctx context.Context, cred *domain.Credential) error {
	if _, exists := m.creds[cred.ID]; !exists {
		return domain.ErrNotFound
	}
	m.creds[cred.ID] = cred
	return nil
}

func (m *mockCredRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.creds[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.creds, id)
	return nil
}

type mockProviderRepo struct {
	providers map[uuid.UUID]*domain.Provider
}

func newMockProviderRepo() *mockProviderRepo {
	return &mockProviderRepo{providers: make(map[uuid.UUID]*domain.Provider)}
}

func (m *mockProviderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Provider, error) {
	if p, exists := m.providers[id]; exists {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockProviderRepo) GetBySlug(ctx context.Context, slug string) (*domain.Provider, error) {
	for _, p := range m.providers {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockProviderRepo) List(ctx context.Context) ([]domain.Provider, error) {
	var list []domain.Provider
	for _, p := range m.providers {
		list = append(list, *p)
	}
	return list, nil
}

func TestEncryptor_AES256GCM(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plainText := "secret_api_key_value_98765"

	cipherText, err := encryptor.Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("enkripsi gagal: %v", err)
	}

	decrypted, err := encryptor.Decrypt(cipherText, key)
	if err != nil {
		t.Fatalf("dekripsi gagal: %v", err)
	}

	if decrypted != plainText {
		t.Errorf("hasil dekripsi tidak sama dengan plaintext asli")
	}

	wrongKey := []byte("wrong_key_32_bytes_long_12345678")
	_, err = encryptor.Decrypt(cipherText, wrongKey)
	if err == nil {
		t.Error("dekripsi dengan kunci yang salah harus mengembalikan error")
	}
}

func TestCredentialUsecase_Create_Success(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	uc := provider.NewCredentialUsecase(credRepo, provRepo, key)
	ctx := context.Background()

	providerID := uuid.New()
	provRepo.providers[providerID] = &domain.Provider{
		ID:       providerID,
		Name:     "Hetzner Cloud",
		Slug:     "hetzner",
		IsActive: true,
	}

	orgID := uuid.New()
	input := provider.CreateCredentialInput{
		OrganizationID: orgID,
		ProviderID:     providerID,
		Name:           "Hetzner Production Token",
		APIKey:         "hetzner_token_abcdef123456",
		APISecret:      "hetzner_secret_789",
	}

	created, err := uc.CreateCredential(ctx, input)
	if err != nil {
		t.Fatalf("gagal membuat kredensial: %v", err)
	}

	if created.Name != "Hetzner Production Token" {
		t.Errorf("nama kredensial tidak sesuai")
	}
	if created.EncryptedAPIKey == nil || *created.EncryptedAPIKey == "hetzner_token_abcdef123456" {
		t.Errorf("API Key harus tersimpan dalam bentuk terenkripsi")
	}

	decryptedKey, err := encryptor.Decrypt(*created.EncryptedAPIKey, key)
	if err != nil || decryptedKey != "hetzner_token_abcdef123456" {
		t.Errorf("dekripsi API Key yang disimpan tidak sesuai dengan aslinya")
	}
}

func TestCredentialUsecase_GetAndList_Success(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	uc := provider.NewCredentialUsecase(credRepo, provRepo, key)
	ctx := context.Background()

	providerID := uuid.New()
	provRepo.providers[providerID] = &domain.Provider{ID: providerID, Name: "AWS", Slug: "aws", IsActive: true}
	orgID := uuid.New()

	created, _ := uc.CreateCredential(ctx, provider.CreateCredentialInput{
		OrganizationID: orgID,
		ProviderID:     providerID,
		Name:           "AWS Key",
		APIKey:         "AKIAEXAMPLE123",
	})

	fetched, err := uc.GetCredential(ctx, orgID, created.ID)
	if err != nil || fetched.ID != created.ID {
		t.Fatalf("gagal mengambil kredensial: %v", err)
	}

	list, err := uc.ListCredentials(ctx, orgID)
	if err != nil || len(list) != 1 {
		t.Fatalf("gagal mengambil daftar kredensial: %v", err)
	}
}

func TestCredentialUsecase_Get_ForbiddenOrg(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	uc := provider.NewCredentialUsecase(credRepo, provRepo, key)
	ctx := context.Background()

	providerID := uuid.New()
	provRepo.providers[providerID] = &domain.Provider{ID: providerID, Name: "Mock", Slug: "mock", IsActive: true}
	ownerOrgID := uuid.New()
	otherOrgID := uuid.New()

	created, _ := uc.CreateCredential(ctx, provider.CreateCredentialInput{
		OrganizationID: ownerOrgID,
		ProviderID:     providerID,
		Name:           "My Token",
		APIKey:         "token123",
	})

	_, err := uc.GetCredential(ctx, otherOrgID, created.ID)
	if err != domain.ErrForbidden {
		t.Errorf("harus mengembalikan ErrForbidden, didapat: %v", err)
	}
}

func TestCredentialUsecase_Update_Success(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	uc := provider.NewCredentialUsecase(credRepo, provRepo, key)
	ctx := context.Background()

	providerID := uuid.New()
	provRepo.providers[providerID] = &domain.Provider{ID: providerID, Name: "Mock", Slug: "mock", IsActive: true}
	orgID := uuid.New()

	created, _ := uc.CreateCredential(ctx, provider.CreateCredentialInput{
		OrganizationID: orgID,
		ProviderID:     providerID,
		Name:           "Old Name",
		APIKey:         "old_key",
	})

	updated, err := uc.UpdateCredential(ctx, provider.UpdateCredentialInput{
		ID:             created.ID,
		OrganizationID: orgID,
		Name:           "Updated Name",
		APIKey:         "new_secret_key",
	})
	if err != nil {
		t.Fatalf("gagal memperbarui kredensial: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("nama tidak terperbarui")
	}

	decrypted, _ := encryptor.Decrypt(*updated.EncryptedAPIKey, key)
	if decrypted != "new_secret_key" {
		t.Errorf("hasil dekripsi API key baru tidak cocok")
	}
}

func TestCredentialUsecase_Delete_Success(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	uc := provider.NewCredentialUsecase(credRepo, provRepo, key)
	ctx := context.Background()

	providerID := uuid.New()
	provRepo.providers[providerID] = &domain.Provider{ID: providerID, Name: "Mock", Slug: "mock", IsActive: true}
	orgID := uuid.New()

	created, _ := uc.CreateCredential(ctx, provider.CreateCredentialInput{
		OrganizationID: orgID,
		ProviderID:     providerID,
		Name:           "To Delete",
		APIKey:         "delete_me",
	})

	if err := uc.DeleteCredential(ctx, orgID, created.ID); err != nil {
		t.Fatalf("gagal menghapus kredensial: %v", err)
	}

	_, err := uc.GetCredential(ctx, orgID, created.ID)
	if err != domain.ErrNotFound {
		t.Errorf("kredensial yang dihapus harus mengembalikan ErrNotFound, didapat: %v", err)
	}
}
