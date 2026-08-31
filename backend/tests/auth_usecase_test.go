package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/auth"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

type mockUserRepo struct {
	usersByID    map[uuid.UUID]*domain.User
	usersByEmail map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:    make(map[uuid.UUID]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.usersByEmail[strings.ToLower(user.Email)]; exists {
		return domain.ErrEmailAlreadyInUse
	}
	m.usersByID[user.ID] = user
	m.usersByEmail[strings.ToLower(user.Email)] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, exists := m.usersByID[id]; exists {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if user, exists := m.usersByEmail[strings.ToLower(email)]; exists {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if _, exists := m.usersByID[user.ID]; !exists {
		return domain.ErrNotFound
	}
	m.usersByID[user.ID] = user
	m.usersByEmail[strings.ToLower(user.Email)] = user
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if user, exists := m.usersByID[id]; exists {
		delete(m.usersByID, id)
		delete(m.usersByEmail, strings.ToLower(user.Email))
		return nil
	}
	return domain.ErrNotFound
}

type mockOrgRepo struct {
	orgs    map[uuid.UUID]*domain.Organization
	members map[string]*domain.OrganizationMember
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{
		orgs:    make(map[uuid.UUID]*domain.Organization),
		members: make(map[string]*domain.OrganizationMember),
	}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *domain.Organization) error {
	if _, exists := m.orgs[org.ID]; exists {
		return domain.ErrConflict
	}
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if org, exists := m.orgs[id]; exists {
		return org, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	for _, o := range m.orgs {
		if o.Slug == slug {
			return o, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockOrgRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	var result []domain.Organization
	for _, member := range m.members {
		if member.UserID == userID {
			if o, exists := m.orgs[member.OrganizationID]; exists {
				result = append(result, *o)
			}
		}
	}
	return result, nil
}

func (m *mockOrgRepo) Update(ctx context.Context, org *domain.Organization) error {
	if _, exists := m.orgs[org.ID]; !exists {
		return domain.ErrNotFound
	}
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.orgs[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.orgs, id)
	return nil
}

func (m *mockOrgRepo) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	key := member.OrganizationID.String() + ":" + member.UserID.String()
	m.members[key] = member
	return nil
}

func (m *mockOrgRepo) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	key := orgID.String() + ":" + userID.String()
	if member, exists := m.members[key]; exists {
		return member, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockOrgRepo) ListMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	var list []domain.OrganizationMember
	for _, member := range m.members {
		if member.OrganizationID == orgID {
			list = append(list, *member)
		}
	}
	return list, nil
}

func (m *mockOrgRepo) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role domain.OrganizationRole) error {
	key := orgID.String() + ":" + userID.String()
	if member, exists := m.members[key]; exists {
		member.Role = role
		return nil
	}
	return domain.ErrNotFound
}

func (m *mockOrgRepo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	key := orgID.String() + ":" + userID.String()
	if _, exists := m.members[key]; !exists {
		return domain.ErrNotFound
	}
	delete(m.members, key)
	return nil
}

func (m *mockOrgRepo) CreateInvitation(ctx context.Context, inv *domain.OrganizationInvitation) error {
	return nil
}

func (m *mockOrgRepo) GetInvitationByToken(ctx context.Context, token string) (*domain.OrganizationInvitation, error) {
	return nil, domain.ErrNotFound
}

func (m *mockOrgRepo) ListInvitations(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationInvitation, error) {
	return nil, nil
}

func (m *mockOrgRepo) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	return nil
}

func setupTestAuthUsecase() (auth.Usecase, *mockUserRepo, *mockOrgRepo, jwt.Manager) {
	userRepo := newMockUserRepo()
	orgRepo := newMockOrgRepo()
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	uc := auth.NewAuthUsecase(userRepo, orgRepo, jwtManager)
	return uc, userRepo, orgRepo, jwtManager
}

// TestRegister_Success menguji alur registrasi pengguna baru yang berhasil dengan pembuatan organisasi dan token JWT.
func TestRegister_Success(t *testing.T) {
	uc, _, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	input := auth.RegisterInput{
		Email:            "alex@example.com",
		Password:         "SecurePassword123!",
		FullName:         "Alex Johnson",
		OrganizationName: "Alex Cloud Inc",
	}

	output, err := uc.Register(ctx, input)
	if err != nil {
		t.Fatalf("registrasi harus berhasil, didapat error: %v", err)
	}

	if output.User == nil || output.Organization == nil || output.Tokens == nil {
		t.Fatal("output User, Organization, dan Tokens tidak boleh bernilai nil")
	}

	if output.Tokens.AccessToken == "" || output.Tokens.RefreshToken == "" {
		t.Error("token pair JWT harus terisi lengkap")
	}

	if output.User.Email != "alex@example.com" {
		t.Errorf("email pengguna tidak sesuai: %s", output.User.Email)
	}

	match, err := hasher.Compare("SecurePassword123!", output.User.PasswordHash)
	if err != nil || !match {
		t.Error("hash password yang disimpan harus cocok dengan password input")
	}

	if output.Organization.Name != "Alex Cloud Inc" {
		t.Errorf("nama organisasi tidak sesuai: %s", output.Organization.Name)
	}
}

// TestRegister_DuplicateEmail menguji penolakan registrasi apabila alamat email sudah terdaftar.
func TestRegister_DuplicateEmail(t *testing.T) {
	uc, _, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	initialInput := auth.RegisterInput{
		Email:    "duplicate@example.com",
		Password: "Password123!",
		FullName: "First User",
	}
	_, _ = uc.Register(ctx, initialInput)

	duplicateInput := auth.RegisterInput{
		Email:    "duplicate@example.com",
		Password: "AnotherPassword123!",
		FullName: "Duplicate User",
	}

	_, err := uc.Register(ctx, duplicateInput)
	if err != domain.ErrEmailAlreadyInUse {
		t.Errorf("harus mengembalikan ErrEmailAlreadyInUse, didapat: %v", err)
	}
}

// TestRegister_ShortPassword menguji penolakan registrasi apabila panjang kata sandi kurang dari 8 karakter.
func TestRegister_ShortPassword(t *testing.T) {
	uc, _, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	input := auth.RegisterInput{
		Email:    "bob@example.com",
		Password: "short",
		FullName: "Bob Smith",
	}

	_, err := uc.Register(ctx, input)
	if err != domain.ErrPasswordTooShort {
		t.Errorf("harus mengembalikan ErrPasswordTooShort, didapat: %v", err)
	}
}

// TestLogin_Success menguji alur login berhasil untuk akun aktif dengan kata sandi yang valid.
func TestLogin_Success(t *testing.T) {
	uc, userRepo, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	hash, _ := hasher.Hash("ValidPassword123!", nil)
	activeUser := &domain.User{
		ID:           uuid.New(),
		Email:        "active@example.com",
		PasswordHash: hash,
		FullName:     "Active User",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(ctx, activeUser)

	output, err := uc.Login(ctx, auth.LoginInput{
		Email:    "active@example.com",
		Password: "ValidPassword123!",
	})
	if err != nil {
		t.Fatalf("login harus berhasil, didapat error: %v", err)
	}

	if output.User.ID != activeUser.ID {
		t.Errorf("ID pengguna tidak cocok: didapat %v, diharapkan %v", output.User.ID, activeUser.ID)
	}

	if output.Tokens == nil || output.Tokens.AccessToken == "" || output.Tokens.RefreshToken == "" {
		t.Error("tokens pada login output harus terisi lengkap")
	}
}

// TestLogin_RefreshToken_Success menguji perpanjangan sesi melalui Refresh Token yang sah.
func TestLogin_RefreshToken_Success(t *testing.T) {
	uc, userRepo, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	hash, _ := hasher.Hash("ValidPassword123!", nil)
	activeUser := &domain.User{
		ID:           uuid.New(),
		Email:        "active@example.com",
		PasswordHash: hash,
		FullName:     "Active User",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(ctx, activeUser)

	loginOutput, err := uc.Login(ctx, auth.LoginInput{
		Email:    "active@example.com",
		Password: "ValidPassword123!",
	})
	if err != nil {
		t.Fatalf("login gagal: %v", err)
	}

	newTokens, err := uc.RefreshToken(ctx, loginOutput.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh token gagal: %v", err)
	}

	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Error("pasangan token baru tidak boleh kosong")
	}
}

// TestLogin_WrongPassword menguji penolakan login jika kata sandi salah.
func TestLogin_WrongPassword(t *testing.T) {
	uc, userRepo, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	hash, _ := hasher.Hash("ValidPassword123!", nil)
	activeUser := &domain.User{
		ID:           uuid.New(),
		Email:        "active@example.com",
		PasswordHash: hash,
		FullName:     "Active User",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(ctx, activeUser)

	_, err := uc.Login(ctx, auth.LoginInput{
		Email:    "active@example.com",
		Password: "WrongPassword!",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("harus mengembalikan ErrInvalidCredentials, didapat: %v", err)
	}
}

// TestLogin_UserNotFound menguji penolakan login jika akun pengguna tidak terdaftar.
func TestLogin_UserNotFound(t *testing.T) {
	uc, _, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	_, err := uc.Login(ctx, auth.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "AnyPassword123!",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("harus mengembalikan ErrInvalidCredentials, didapat: %v", err)
	}
}

// TestLogin_InactiveUser menguji penolakan login jika akun pengguna dalam status nonaktif.
func TestLogin_InactiveUser(t *testing.T) {
	uc, userRepo, _, _ := setupTestAuthUsecase()
	ctx := context.Background()

	hash, _ := hasher.Hash("ValidPassword123!", nil)
	inactiveUser := &domain.User{
		ID:           uuid.New(),
		Email:        "inactive@example.com",
		PasswordHash: hash,
		FullName:     "Inactive User",
		IsActive:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(ctx, inactiveUser)

	_, err := uc.Login(ctx, auth.LoginInput{
		Email:    "inactive@example.com",
		Password: "ValidPassword123!",
	})
	if err != domain.ErrUserInactive {
		t.Errorf("harus mengembalikan ErrUserInactive, didapat: %v", err)
	}
}
