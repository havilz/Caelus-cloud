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

func NewCredentialUsecase(credRepo domain.CredentialRepository, providerRepo domain.ProviderRepository, encryptionKey []byte) CredentialUsecase {
	return &credentialUsecase{
		credRepo:      credRepo,
		providerRepo:  providerRepo,
		encryptionKey: encryptionKey,
	}
}

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

func (u *credentialUsecase) ListCredentials(ctx context.Context, orgID uuid.UUID) ([]domain.Credential, error) {
	return u.credRepo.ListByOrg(ctx, orgID)
}

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

func (u *credentialUsecase) ListSupportedProviders(ctx context.Context) ([]domain.Provider, error) {
	return u.providerRepo.List(ctx)
}

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
