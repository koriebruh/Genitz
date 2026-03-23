package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	cfg := JWTConfig{
		SecretKey:     "test-secret-key-1234567890abcdef",
		TokenDuration: 1 * time.Hour,
		Issuer:        "test-issuer",
	}
	mgr := NewJWTManager(cfg)

	token, err := mgr.Generate("user-123", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if token == "" {
		t.Fatal("Expected non-empty token string")
	}

	claims, err := mgr.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("Expected UserID=user-123, got %s", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected Role=admin, got %s", claims.Role)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	cfg := JWTConfig{SecretKey: "secret", TokenDuration: time.Hour, Issuer: "app"}
	mgr := NewJWTManager(cfg)

	_, err := mgr.Validate("this.is.invalid")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	cfg1 := JWTConfig{SecretKey: "secret-a", TokenDuration: time.Hour, Issuer: "app"}
	mgr1 := NewJWTManager(cfg1)
	token, _ := mgr1.Generate("u1", "a@b.com", "user")

	cfg2 := JWTConfig{SecretKey: "secret-b", TokenDuration: time.Hour, Issuer: "app"}
	mgr2 := NewJWTManager(cfg2)
	_, err := mgr2.Validate(token)
	if err == nil {
		t.Error("Expected validation to fail with different secret key")
	}
}
