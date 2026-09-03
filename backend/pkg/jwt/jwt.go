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

func NewJWTManager(cfg *config.JWTConfig, appName string) Manager {
	return &jwtManager{
		secret:            []byte(cfg.Secret),
		accessExpiration:  cfg.AccessExpiration,
		refreshExpiration: cfg.RefreshExpiration,
		issuer:            appName,
	}
}

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

func (m *jwtManager) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	return m.validateTokenWithType(tokenString, TokenTypeAccess)
}

func (m *jwtManager) ValidateRefreshToken(tokenString string) (*UserClaims, error) {
	return m.validateTokenWithType(tokenString, TokenTypeRefresh)
}

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
