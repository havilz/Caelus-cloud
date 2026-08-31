package settings

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
)

type UpdateProfileInput struct {
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UpdateOrganizationInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type InviteMemberInput struct {
	Email string                  `json:"email"`
	Role  domain.OrganizationRole `json:"role"`
}

type Usecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error

	GetOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, input UpdateOrganizationInput) (*domain.Organization, error)

	ListMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error)
	InviteMember(ctx context.Context, orgID, invitedBy uuid.UUID, input InviteMemberInput) (*domain.OrganizationInvitation, error)
	UpdateMemberRole(ctx context.Context, orgID, targetUserID uuid.UUID, role domain.OrganizationRole) error
	RemoveMember(ctx context.Context, orgID, targetUserID uuid.UUID) error
	ListInvitations(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationInvitation, error)
	DeleteInvitation(ctx context.Context, id uuid.UUID) error

	ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]domain.APIKey, error)
	CreateAPIKey(ctx context.Context, orgID, userID uuid.UUID, req domain.CreateAPIKeyRequest) (*domain.APIKey, error)
	DeleteAPIKey(ctx context.Context, orgID, keyID uuid.UUID) error

	ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]domain.Webhook, error)
	CreateWebhook(ctx context.Context, orgID uuid.UUID, req domain.CreateWebhookRequest) (*domain.Webhook, error)
	UpdateWebhook(ctx context.Context, orgID, webhookID uuid.UUID, req domain.UpdateWebhookRequest) (*domain.Webhook, error)
	TestWebhook(ctx context.Context, orgID, webhookID uuid.UUID) (int, error)
	DeleteWebhook(ctx context.Context, orgID, webhookID uuid.UUID) error

	ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
}

type settingsUsecase struct {
	userRepo    domain.UserRepository
	orgRepo     domain.OrganizationRepository
	apiKeyRepo  domain.APIKeyRepository
	webhookRepo domain.WebhookRepository
	auditRepo   domain.AuditLogRepository
}

func NewSettingsUsecase(
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
	apiKeyRepo domain.APIKeyRepository,
	webhookRepo domain.WebhookRepository,
	auditRepo domain.AuditLogRepository,
) Usecase {
	return &settingsUsecase{
		userRepo:    userRepo,
		orgRepo:     orgRepo,
		apiKeyRepo:  apiKeyRepo,
		webhookRepo: webhookRepo,
		auditRepo:   auditRepo,
	}
}

func (u *settingsUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}

func (u *settingsUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.FullName) == "" {
		return nil, errors.New("nama lengkap tidak boleh kosong")
	}

	user.FullName = strings.TrimSpace(input.FullName)
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *settingsUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := hasher.Compare(input.OldPassword, user.PasswordHash)
	if err != nil || !match {
		return errors.New("kata sandi lama tidak sesuai")
	}

	if len(input.NewPassword) < 8 {
		return errors.New("kata sandi baru minimal 8 karakter")
	}

	newHash, err := hasher.Hash(input.NewPassword, nil)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi kata sandi baru: %w", err)
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *settingsUsecase) GetOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	return u.orgRepo.GetByID(ctx, orgID)
}

func (u *settingsUsecase) UpdateOrganization(ctx context.Context, orgID uuid.UUID, input UpdateOrganizationInput) (*domain.Organization, error) {
	org, err := u.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Name) != "" {
		org.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Slug) != "" {
		org.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	}
	org.UpdatedAt = time.Now()

	if err := u.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

func (u *settingsUsecase) ListMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	return u.orgRepo.ListMembers(ctx, orgID)
}

func (u *settingsUsecase) InviteMember(ctx context.Context, orgID, invitedBy uuid.UUID, input InviteMemberInput) (*domain.OrganizationInvitation, error) {
	if strings.TrimSpace(input.Email) == "" {
		return nil, errors.New("email undangan wajib diisi")
	}

	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("gagal menghasilkan token undangan: %w", err)
	}
	token := hex.EncodeToString(randomBytes)

	inv := &domain.OrganizationInvitation{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          strings.ToLower(strings.TrimSpace(input.Email)),
		Role:           input.Role,
		Token:          token,
		InvitedBy:      &invitedBy,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	if inv.Role == "" {
		inv.Role = domain.RoleMember
	}

	if err := u.orgRepo.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (u *settingsUsecase) UpdateMemberRole(ctx context.Context, orgID, targetUserID uuid.UUID, role domain.OrganizationRole) error {
	return u.orgRepo.UpdateMemberRole(ctx, orgID, targetUserID, role)
}

func (u *settingsUsecase) RemoveMember(ctx context.Context, orgID, targetUserID uuid.UUID) error {
	return u.orgRepo.RemoveMember(ctx, orgID, targetUserID)
}

func (u *settingsUsecase) ListInvitations(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationInvitation, error) {
	return u.orgRepo.ListInvitations(ctx, orgID)
}

func (u *settingsUsecase) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	return u.orgRepo.DeleteInvitation(ctx, id)
}

func (u *settingsUsecase) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]domain.APIKey, error) {
	return u.apiKeyRepo.ListByOrg(ctx, orgID)
}

func (u *settingsUsecase) CreateAPIKey(ctx context.Context, orgID, userID uuid.UUID, req domain.CreateAPIKeyRequest) (*domain.APIKey, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("nama API key wajib diisi")
	}

	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("gagal menghasilkan secret token: %w", err)
	}
	rawSecret := hex.EncodeToString(randomBytes)
	prefix := rawSecret[:8]
	fullToken := fmt.Sprintf("caelus_pat_%s", rawSecret)

	hash := sha256.Sum256([]byte(fullToken))
	keyHash := hex.EncodeToString(hash[:])

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &exp
	}

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read", "write"}
	}

	apiKey := &domain.APIKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Name:           strings.TrimSpace(req.Name),
		KeyPrefix:      prefix,
		KeyHash:        keyHash,
		Scopes:         scopes,
		ExpiresAt:      expiresAt,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		RawToken:       fullToken,
	}

	if err := u.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (u *settingsUsecase) DeleteAPIKey(ctx context.Context, orgID, keyID uuid.UUID) error {
	key, err := u.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}
	if key.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	return u.apiKeyRepo.Delete(ctx, keyID)
}

func (u *settingsUsecase) ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]domain.Webhook, error) {
	return u.webhookRepo.ListByOrg(ctx, orgID)
}

func (u *settingsUsecase) CreateWebhook(ctx context.Context, orgID uuid.UUID, req domain.CreateWebhookRequest) (*domain.Webhook, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
		return nil, errors.New("nama dan URL webhook wajib diisi")
	}

	events := req.Events
	if len(events) == 0 {
		events = []string{"server.down", "alert.triggered", "backup.failed"}
	}

	webhook := &domain.Webhook{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		URL:            strings.TrimSpace(req.URL),
		Secret:         req.Secret,
		Events:         events,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (u *settingsUsecase) UpdateWebhook(ctx context.Context, orgID, webhookID uuid.UUID, req domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	webhook, err := u.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if webhook.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	if strings.TrimSpace(req.Name) != "" {
		webhook.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.URL) != "" {
		webhook.URL = strings.TrimSpace(req.URL)
	}
	if req.Secret != nil {
		webhook.Secret = req.Secret
	}
	if len(req.Events) > 0 {
		webhook.Events = req.Events
	}
	webhook.IsActive = req.IsActive
	webhook.UpdatedAt = time.Now()

	if err := u.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (u *settingsUsecase) TestWebhook(ctx context.Context, orgID, webhookID uuid.UUID) (int, error) {
	webhook, err := u.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return 0, err
	}
	if webhook.OrganizationID != orgID {
		return 0, domain.ErrForbidden
	}

	payload := map[string]any{
		"event":     "ping",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   "Test ping dari Caelus Cloud Webhook Dispatcher",
		"org_id":    orgID.String(),
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, fmt.Errorf("gagal membuat HTTP request test webhook: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Caelus-Cloud-Webhook/1.0")

	if webhook.Secret != nil && *webhook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(*webhook.Secret))
		mac.Write(payloadBytes)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Caelus-Signature", signature)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	now := time.Now()
	if err != nil {
		_ = u.webhookRepo.UpdateStatus(ctx, webhookID, 500, now)
		return 500, fmt.Errorf("koneksi ke endpoint webhook gagal: %w", err)
	}
	defer resp.Body.Close()

	_ = u.webhookRepo.UpdateStatus(ctx, webhookID, resp.StatusCode, now)
	return resp.StatusCode, nil
}

func (u *settingsUsecase) DeleteWebhook(ctx context.Context, orgID, webhookID uuid.UUID) error {
	webhook, err := u.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return err
	}
	if webhook.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	return u.webhookRepo.Delete(ctx, webhookID)
}

func (u *settingsUsecase) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.auditRepo.ListByOrg(ctx, orgID, page, limit)
}
