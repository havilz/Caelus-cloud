package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}

	manager := jwt.NewJWTManager(jwtCfg, "caelus-cloud-test")

	userID := uuid.New()
	orgID := uuid.New()
	user := &domain.User{
		ID:        userID,
		Email:     "user@example.com",
		FullName:  "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tokenPair, err := manager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("gagal menghasilkan token pair: %v", err)
	}

	if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
		t.Fatal("access token dan refresh token tidak boleh kosong")
	}

	accessClaims, err := manager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("validasi access token gagal: %v", err)
	}

	if accessClaims.UserID != userID {
		t.Errorf("UserID klaim tidak sesuai: didapat %v, diharapkan %v", accessClaims.UserID, userID)
	}
	if accessClaims.Email != "user@example.com" {
		t.Errorf("Email klaim tidak sesuai: didapat %v, diharapkan %v", accessClaims.Email, "user@example.com")
	}
	if accessClaims.OrganizationID == nil || *accessClaims.OrganizationID != orgID {
		t.Errorf("OrganizationID klaim tidak sesuai")
	}
	if accessClaims.TokenType != jwt.TokenTypeAccess {
		t.Errorf("TokenType klaim harus access")
	}

	refreshClaims, err := manager.ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("validasi refresh token gagal: %v", err)
	}
	if refreshClaims.UserID != userID {
		t.Errorf("UserID klaim refresh token tidak sesuai")
	}
	if refreshClaims.TokenType != jwt.TokenTypeRefresh {
		t.Errorf("TokenType klaim harus refresh")
	}

	_, err = manager.ValidateAccessToken(tokenPair.RefreshToken)
	if err != jwt.ErrInvalidTokenType {
		t.Errorf("validasi refresh token sebagai access token harus menghasilkan ErrInvalidTokenType, didapat: %v", err)
	}

	_, err = manager.ValidateAccessToken("invalid.jwt.token")
	if err == nil {
		t.Error("token cacat harus menghasilkan error")
	}
}
