package provider

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

type CreateCredentialInput struct {
	OrganizationID uuid.UUID      `json:"organization_id"`
	ProviderID     uuid.UUID      `json:"provider_id"`
	Name           string         `json:"name"`
	APIKey         string         `json:"api_key,omitempty"`
	APISecret      string         `json:"api_secret,omitempty"`
	SSHKey         string         `json:"ssh_key,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type UpdateCredentialInput struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Name           string         `json:"name"`
	APIKey         string         `json:"api_key,omitempty"`
	APISecret      string         `json:"api_secret,omitempty"`
	SSHKey         string         `json:"ssh_key,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type CredentialUsecase interface {
	CreateCredential(ctx context.Context, input CreateCredentialInput) (*domain.Credential, error)
	GetCredential(ctx context.Context, orgID, credID uuid.UUID) (*domain.Credential, error)
	ListCredentials(ctx context.Context, orgID uuid.UUID) ([]domain.Credential, error)
	UpdateCredential(ctx context.Context, input UpdateCredentialInput) (*domain.Credential, error)
	DeleteCredential(ctx context.Context, orgID, credID uuid.UUID) error
	ListSupportedProviders(ctx context.Context) ([]domain.Provider, error)
}

type credentialUsecase struct {
	credRepo      domain.CredentialRepository
	providerRepo  domain.ProviderRepository
	encryptionKey []byte
}

// NewCredentialUsecase menginisialisasi use case manajemen kredensial provider dengan enkripsi data sensitif.
// Parameter credRepo merupakan implementasi domain.CredentialRepository.
// Parameter providerRepo merupakan implementasi domain.ProviderRepository.
// Parameter encryptionKey merupakan byte slice 32-byte kunci enkripsi AES-256.
// Mengembalikan instance interface CredentialUsecase.
func NewCredentialUsecase(credRepo domain.CredentialRepository, providerRepo domain.ProviderRepository, encryptionKey []byte) CredentialUsecase {
	return &credentialUsecase{
		credRepo:      credRepo,
		providerRepo:  providerRepo,
		encryptionKey: encryptionKey,
	}
}

// CreateCredential mengenkripsi kredensial API/SSH dan menyimpan entitas kredensial baru ke database.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter input memuat data kredensial provider baru.
// Mengembalikan pointer *domain.Credential yang berhasil disimpan atau error jika validasi/enkripsi gagal.
func (u *credentialUsecase) CreateCredential(ctx context.Context, input CreateCredentialInput) (*domain.Credential, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.OrganizationID == uuid.Nil || input.ProviderID == uuid.Nil {
		return nil, domain.ErrBadRequest
	}

	provider, err := u.providerRepo.GetByID(ctx, input.ProviderID)
	if err != nil {
		return nil, err
	}

	encryptedKey, encryptedSecret, encryptedSSH, err := u.encryptSecretFields(input.APIKey, input.APISecret, input.SSHKey)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	cred := &domain.Credential{
		ID:                 uuid.New(),
		OrganizationID:     input.OrganizationID,
		ProviderID:         input.ProviderID,
		Name:               input.Name,
		EncryptedAPIKey:    encryptedKey,
		EncryptedAPISecret: encryptedSecret,
		EncryptedSSHKey:    encryptedSSH,
		Metadata:           input.Metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
		Provider:           provider,
	}

	if err := u.credRepo.Create(ctx, cred); err != nil {
		return nil, err
	}

	return cred, nil
}

// GetCredential mengambil detail kredensial provider berdasarkan ID dan memastikan kepemilikan organisasi yang sah.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik kredensial.
// Parameter credID merupakan UUID kredensial yang diminta.
// Mengembalikan pointer *domain.Credential atau error jika kredensial tidak ditemukan atau tidak berhak diakses.
func (u *credentialUsecase) GetCredential(ctx context.Context, orgID, credID uuid.UUID) (*domain.Credential, error) {
	cred, err := u.credRepo.GetByID(ctx, credID)
	if err != nil {
		return nil, err
	}

	if cred.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	return cred, nil
}

// ListCredentials mengambil seluruh daftar kredensial provider yang terdaftar pada organisasi tertentu.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik kredensial.
// Mengembalikan slice []domain.Credential dan error jika query gagal.
func (u *credentialUsecase) ListCredentials(ctx context.Context, orgID uuid.UUID) ([]domain.Credential, error) {
	return u.credRepo.ListByOrg(ctx, orgID)
}

// UpdateCredential memperbarui informasi kredensial dan mengenkripsi ulang field rahasia jika terdapat pembaruan data.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter input memuat data pembaruan kredensial.
// Mengembalikan pointer *domain.Credential yang diperbarui atau error jika operasi gagal.
func (u *credentialUsecase) UpdateCredential(ctx context.Context, input UpdateCredentialInput) (*domain.Credential, error) {
	existing, err := u.credRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if existing.OrganizationID != input.OrganizationID {
		return nil, domain.ErrForbidden
	}

	if input.Name != "" {
		existing.Name = strings.TrimSpace(input.Name)
	}

	if err := u.applyEncryptedUpdates(existing, input.APIKey, input.APISecret, input.SSHKey); err != nil {
		return nil, err
	}

	if input.Metadata != nil {
		existing.Metadata = input.Metadata
	}
	existing.UpdatedAt = time.Now()

	if err := u.credRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteCredential menghapus kredensial provider dan memverifikasi kepemilikan organisasi sebelum penghapusan.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik kredensial.
// Parameter credID merupakan UUID kredensial yang akan dihapus.
// Mengembalikan error jika kredensial tidak ditemukan atau kepemilikan tidak sesuai.
func (u *credentialUsecase) DeleteCredential(ctx context.Context, orgID, credID uuid.UUID) error {
	existing, err := u.credRepo.GetByID(ctx, credID)
	if err != nil {
		return err
	}

	if existing.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	return u.credRepo.Delete(ctx, credID)
}

// ListSupportedProviders mengambil seluruh daftar provider cloud yang terdaftar dan didukung sistem.
// Parameter ctx merupakan konteks eksekusi use case.
// Mengembalikan slice []domain.Provider dan error jika terjadi kegagalan query.
func (u *credentialUsecase) ListSupportedProviders(ctx context.Context) ([]domain.Provider, error) {
	return u.providerRepo.List(ctx)
}

// encryptSecretFields mengenkripsi field APIKey, APISecret, dan SSHKey dengan kunci AES-256.
// Parameter apiKey merupakan teks API key mentah.
// Parameter apiSecret merupakan teks API secret mentah.
// Parameter sshKey merupakan teks private SSH key mentah.
// Mengembalikan pointer string terenkripsi untuk masing-masing field.
func (u *credentialUsecase) encryptSecretFields(apiKey, apiSecret, sshKey string) (*string, *string, *string, error) {
	var encKey, encSecret, encSSH *string

	if apiKey != "" {
		encrypted, err := encryptor.Encrypt(apiKey, u.encryptionKey)
		if err != nil {
			return nil, nil, nil, err
		}
		encKey = &encrypted
	}

	if apiSecret != "" {
		encrypted, err := encryptor.Encrypt(apiSecret, u.encryptionKey)
		if err != nil {
			return nil, nil, nil, err
		}
		encSecret = &encrypted
	}

	if sshKey != "" {
		encrypted, err := encryptor.Encrypt(sshKey, u.encryptionKey)
		if err != nil {
			return nil, nil, nil, err
		}
		encSSH = &encrypted
	}

	return encKey, encSecret, encSSH, nil
}

// applyEncryptedUpdates memperbarui field enkripsi pada entitas kredensial yang ada jika input teks rahasia baru diberikan.
// Parameter existing merupakan entitas kredensial yang sedang dimodifikasi.
// Parameter apiKey merupakan teks API key baru (opsional).
// Parameter apiSecret merupakan teks API secret baru (opsional).
// Parameter sshKey merupakan teks SSH key baru (opsional).
// Mengembalikan error jika proses enkripsi gagal.
func (u *credentialUsecase) applyEncryptedUpdates(existing *domain.Credential, apiKey, apiSecret, sshKey string) error {
	if apiKey != "" {
		encrypted, err := encryptor.Encrypt(apiKey, u.encryptionKey)
		if err != nil {
			return err
		}
		existing.EncryptedAPIKey = &encrypted
	}

	if apiSecret != "" {
		encrypted, err := encryptor.Encrypt(apiSecret, u.encryptionKey)
		if err != nil {
			return err
		}
		existing.EncryptedAPISecret = &encrypted
	}

	if sshKey != "" {
		encrypted, err := encryptor.Encrypt(sshKey, u.encryptionKey)
		if err != nil {
			return err
		}
		existing.EncryptedSSHKey = &encrypted
	}

	return nil
}
