package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

type RegisterInput struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	FullName         string `json:"full_name"`
	OrganizationName string `json:"organization_name"`
}

type RegisterOutput struct {
	User         *domain.User         `json:"user"`
	Organization *domain.Organization `json:"organization"`
	Tokens       *jwt.TokenPair       `json:"tokens"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	User   *domain.User   `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

type Usecase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
}

type authUsecase struct {
	userRepo   domain.UserRepository
	orgRepo    domain.OrganizationRepository
	jwtManager jwt.Manager
}

// NewAuthUsecase menginisialisasi use case autentikasi dengan dependensi repository User, Organization, dan manajer JWT.
// Parameter userRepo merupakan implementasi domain.UserRepository.
// Parameter orgRepo merupakan implementasi domain.OrganizationRepository.
// Parameter jwtManager merupakan implementasi interface jwt.Manager.
// Mengembalikan instance interface Usecase.
func NewAuthUsecase(userRepo domain.UserRepository, orgRepo domain.OrganizationRepository, jwtManager jwt.Manager) Usecase {
	return &authUsecase{
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		jwtManager: jwtManager,
	}
}

// Register mengorkestrasi validasi input, pendaftaran akun pengguna baru, inisialisasi organisasi awal, dan penerbitan token sesi JWT.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter input memuat data pendaftaran (Email, Password, FullName, OrganizationName).
// Mengembalikan pointer *RegisterOutput yang memuat entitas User, Organization, dan TokenPair, atau error jika validasi/penyimpanan gagal.
func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := validateRegisterInput(&input); err != nil {
		return nil, err
	}

	if err := u.ensureEmailUnique(ctx, input.Email); err != nil {
		return nil, err
	}

	newUser, err := u.createUserEntity(ctx, input.Email, input.Password, input.FullName)
	if err != nil {
		return nil, err
	}

	newOrg, err := u.createOrganizationWithMember(ctx, newUser.ID, input.OrganizationName, input.FullName)
	if err != nil {
		return nil, err
	}

	tokens, err := u.jwtManager.GenerateTokenPair(newUser, &newOrg.ID)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		User:         newUser,
		Organization: newOrg,
		Tokens:       tokens,
	}, nil
}

// Login memvalidasi kredensial pengguna, memverifikasi hash Argon2id, memeriksa keaktifan akun, dan menghasilkan pasangan token JWT aktif.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter input memuat data login (Email dan Password).
// Mengembalikan pointer *LoginOutput yang memuat data profil User dan pasangan TokenPair, atau error jika kredensial tidak valid.
func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := u.authenticateUser(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	activeOrgID := u.resolveActiveOrganizationID(ctx, user.ID)

	tokens, err := u.jwtManager.GenerateTokenPair(user, activeOrgID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshToken memvalidasi Refresh Token yang dikirimkan dan menerbitkan pasangan Access Token serta Refresh Token yang baru.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter refreshToken merupakan string Refresh Token JWT yang valid.
// Mengembalikan pointer *jwt.TokenPair baru atau error jika token kedaluwarsa, tidak valid, atau akun pengguna nonaktif.
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := u.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	return u.jwtManager.GenerateTokenPair(user, claims.OrganizationID)
}

// validateRegisterInput memverifikasi kelayakan format data registrasi pengguna sebelum diproses.
// Parameter input merupakan pointer data input registrasi.
// Mengembalikan error validasi jika format email tidak sah, password kurang dari 8 karakter, atau nama kosong.
func validateRegisterInput(input *RegisterInput) error {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.OrganizationName = strings.TrimSpace(input.OrganizationName)

	if input.Email == "" {
		return domain.ErrEmailInvalid
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return domain.ErrEmailInvalid
	}
	if len(input.Password) < 8 {
		return domain.ErrPasswordTooShort
	}
	if input.FullName == "" {
		return domain.ErrBadRequest
	}
	return nil
}

// ensureEmailUnique memastikan alamat email yang didaftarkan belum digunakan oleh akun lain.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter email merupakan alamat email yang diperiksa.
// Mengembalikan domain.ErrEmailAlreadyInUse jika email sudah terdaftar.
func (u *authUsecase) ensureEmailUnique(ctx context.Context, email string) error {
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return domain.ErrEmailAlreadyInUse
	}
	return nil
}

// createUserEntity mengenkripsi kata sandi dengan Argon2id dan menyimpan entitas User baru ke database.
// Parameter ctx merupakan konteks eksekusi database.
// Parameter email merupakan email akun pengguna.
// Parameter password merupakan teks sandi mentah.
// Parameter fullName merupakan nama lengkap pengguna.
// Mengembalikan pointer *domain.User yang berhasil dibuat atau error jika penyimpanan gagal.
func (u *authUsecase) createUserEntity(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	passwordHash, err := hasher.Hash(password, nil)
	if err != nil {
		return nil, domain.ErrInternal
	}

	now := time.Now()
	newUser := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// createOrganizationWithMember membuat entitas organisasi awal dan menetapkan pengguna sebagai pemilik organisasi (RoleOwner).
// Parameter ctx merupakan konteks eksekusi database.
// Parameter userID merupakan identifier pemilik organisasi.
// Parameter orgName merupakan nama organisasi yang diinginkan.
// Parameter fallbackName merupakan nama cadangan untuk nama organisasi jika kosong.
// Mengembalikan pointer *domain.Organization yang berhasil dibuat atau error jika operasi gagal.
func (u *authUsecase) createOrganizationWithMember(ctx context.Context, userID uuid.UUID, orgName, fallbackName string) (*domain.Organization, error) {
	if orgName == "" {
		orgName = fallbackName + "'s Workspace"
	}

	now := time.Now()
	newOrg := &domain.Organization{
		ID:        uuid.New(),
		Name:      orgName,
		Slug:      generateSlug(orgName),
		Tier:      "free",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.orgRepo.Create(ctx, newOrg); err != nil {
		return nil, err
	}

	orgMember := &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: newOrg.ID,
		UserID:         userID,
		Role:           domain.RoleOwner,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := u.orgRepo.AddMember(ctx, orgMember); err != nil {
		return nil, err
	}

	return newOrg, nil
}

// authenticateUser mencari pengguna berdasarkan email, menguji kata sandi dengan hash Argon2id, dan memeriksa keaktifan akun.
// Parameter ctx merupakan konteks eksekusi query.
// Parameter email merupakan email login pengguna.
// Parameter password merupakan kata sandi mentah pengujian.
// Mengembalikan pointer *domain.User jika autentikasi berhasil atau error jika kredensial tidak sesuai.
func (u *authUsecase) authenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	match, err := hasher.Compare(password, user.PasswordHash)
	if err != nil || !match {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	return user, nil
}

// resolveActiveOrganizationID mengambil identifier organisasi pertama yang diikuti oleh pengguna sebagai organisasi default.
// Parameter ctx merupakan konteks eksekusi query.
// Parameter userID merupakan identifier pengguna.
// Mengembalikan pointer *uuid.UUID organisasi aktif atau nil jika pengguna belum tergabung dalam organisasi.
func (u *authUsecase) resolveActiveOrganizationID(ctx context.Context, userID uuid.UUID) *uuid.UUID {
	orgs, err := u.orgRepo.ListByUser(ctx, userID)
	if err == nil && len(orgs) > 0 {
		return &orgs[0].ID
	}
	return nil
}

// generateSlug menghasilkan slug URL yang bersih dan unik berdasarkan nama organisasi.
// Parameter name merupakan teks nama organisasi yang akan dikonversi.
// Mengembalikan string slug yang telah dibersihkan dan disisipkan suffix acak 4 byte hex.
func generateSlug(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	slug := strings.ToLower(reg.ReplaceAllString(name, "-"))
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "workspace"
	}

	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	suffix := hex.EncodeToString(randomBytes)

	return slug + "-" + suffix
}
