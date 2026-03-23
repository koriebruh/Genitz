package auth

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds configuration for JWT token generation and validation.
type JWTConfig struct {
	SecretKey     string        `env:"JWT_SECRET"      envDefault:"change-me-in-production"`
	TokenDuration time.Duration `env:"JWT_DURATION"    envDefault:"24h"`
	Issuer        string        `env:"JWT_ISSUER"      envDefault:"genitz-app"`
}

// Claims defines the JWT payload structure.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles token generation and validation.
type JWTManager struct {
	cfg JWTConfig
}

// NewJWTManager creates a new JWTManager from config.
func NewJWTManager(cfg JWTConfig) *JWTManager {
	if cfg.SecretKey == "" || cfg.SecretKey == "change-me-in-production" {
		log.Println("⚠️  WARNING: JWT secret key is not set — use a strong random secret in production!")
	}
	return &JWTManager{cfg: cfg}
}

// Generate creates a signed JWT token for the given user.
func (m *JWTManager) Generate(userID, email, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.SecretKey))
}

// Validate parses and validates a JWT token string.
func (m *JWTManager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.cfg.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
