package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
)

var (
	ErrInvalidTokenSignature = errors.New("tanda tangan token JWT tidak valid")
	ErrTokenExpired          = errors.New("token JWT telah kedaluwarsa")
	ErrInvalidTokenType      = errors.New("tipe token JWT tidak sesuai")
	ErrMalformedToken        = errors.New("format token JWT cacat")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type UserClaims struct {
	UserID         uuid.UUID  `json:"user_id"`
	Email          string     `json:"email"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	TokenType      TokenType  `json:"token_type"`
	jwtlib.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Manager interface {
	GenerateTokenPair(user *domain.User, orgID *uuid.UUID) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*UserClaims, error)
	ValidateRefreshToken(tokenString string) (*UserClaims, error)
}

type jwtManager struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
	issuer            string
}

// NewJWTManager menginisialisasi manajer token JWT berdasarkan konfigurasi keamanan aplikasi.
// Parameter cfg memuat konfigurasi kunci rahasia dan masa berlaku token JWT.
// Parameter appName merupakan nama aplikasi yang disematkan sebagai issuer token.
// Mengembalikan implementasi interface Manager untuk pembuatan dan validasi token.
func NewJWTManager(cfg *config.JWTConfig, appName string) Manager {
	return &jwtManager{
		secret:            []byte(cfg.Secret),
		accessExpiration:  cfg.AccessExpiration,
		refreshExpiration: cfg.RefreshExpiration,
		issuer:            appName,
	}
}

// GenerateTokenPair membuat pasangan Access Token dan Refresh Token berformat JWT yang ditandatangani dengan algoritma HMAC-SHA256.
// Parameter user merupakan pointer entitas *domain.User pemilik token.
// Parameter orgID merupakan pointer UUID organisasi aktif yang dapat bernilai nil.
// Mengembalikan pointer *TokenPair dan error jika proses penandatanganan token gagal.
func (m *jwtManager) GenerateTokenPair(user *domain.User, orgID *uuid.UUID) (*TokenPair, error) {
	accessToken, err := m.generateToken(user, orgID, TokenTypeAccess, m.accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani access token: %w", err)
	}

	refreshToken, err := m.generateToken(user, orgID, TokenTypeRefresh, m.refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessExpiration.Seconds()),
	}, nil
}

// ValidateAccessToken memvalidasi integritas tanda tangan kriptografi dan masa berlaku Access Token.
// Parameter tokenString merupakan string token JWT yang dikirimkan klien.
// Mengembalikan pointer *UserClaims jika token valid, atau error jika tanda tangan salah/kedaluwarsa.
func (m *jwtManager) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	return m.validateTokenWithType(tokenString, TokenTypeAccess)
}

// ValidateRefreshToken memvalidasi integritas tanda tangan kriptografi dan masa berlaku Refresh Token.
// Parameter tokenString merupakan string refresh token JWT yang dikirimkan klien.
// Mengembalikan pointer *UserClaims jika token valid, atau error jika tanda tangan salah/kedaluwarsa.
func (m *jwtManager) ValidateRefreshToken(tokenString string) (*UserClaims, error) {
	return m.validateTokenWithType(tokenString, TokenTypeRefresh)
}

// generateToken membuat string token JWT individual berdasarkan tipe dan durasi kedaluwarsa.
// Parameter user merupakan entitas pengguna pemilik token.
// Parameter orgID merupakan identifier organisasi aktif.
// Parameter tokenType menentukan tipe token (access atau refresh).
// Parameter duration menentukan masa berlaku token.
// Mengembalikan string token yang telah ditandatangani dan error jika terjadi kegagalan.
func (m *jwtManager) generateToken(user *domain.User, orgID *uuid.UUID, tokenType TokenType, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := UserClaims{
		UserID:         user.ID,
		Email:          user.Email,
		OrganizationID: orgID,
		TokenType:      tokenType,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
	}

	tokenObj := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return tokenObj.SignedString(m.secret)
}

// validateTokenWithType memverifikasi integritas token dan memastikan tipe token sesuai dengan yang diharapkan.
// Parameter tokenString merupakan teks token mentah.
// Parameter expectedType merupakan tipe token yang diharapkan (access atau refresh).
// Mengembalikan pointer *UserClaims jika token valid dan sesuai tipe, atau error jika tidak valid.
func (m *jwtManager) validateTokenWithType(tokenString string, expectedType TokenType) (*UserClaims, error) {
	claims, err := m.parseAndValidate(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// parseAndValidate mem-parsing string token JWT, memverifikasi algoritma HMAC, dan mengekstrak klaim pengguna.
// Parameter tokenString merupakan teks token JWT mentah.
// Mengembalikan pointer *UserClaims hasil ekstraksi dan error jika parsing atau verifikasi gagal.
func (m *jwtManager) parseAndValidate(tokenString string) (*UserClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, ErrInvalidTokenSignature
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrMalformedToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
