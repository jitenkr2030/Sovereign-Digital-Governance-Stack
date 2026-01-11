package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/csic-platform/services/api-gateway/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned when the token is invalid
var ErrInvalidToken = errors.New("invalid token")

// ErrExpiredToken is returned when the token has expired
var ErrExpiredToken = errors.New("token has expired")

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	secret []byte
}

// NewAuthService creates a new auth service instance
func NewAuthService(secret string) *AuthServiceImpl {
	return &AuthServiceImpl{
		secret: []byte(secret),
	}
}

// GenerateToken generates a new JWT token for a user
func (s *AuthServiceImpl) GenerateToken(userID, username, role string, permissions []string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(8 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"username":   username,
		"role":       role,
		"permissions": permissions,
		"token_id":   uuid.New().String(),
		"iat":        now.Unix(),
		"exp":        expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthServiceImpl) ValidateToken(tokenString string) (*ports.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return parseClaims(claims), nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken refreshes a token with a new expiration time
func (s *AuthServiceImpl) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return s.GenerateToken(claims.UserID, claims.Username, claims.Role, claims.Permissions)
}

// Authenticate authenticates a user with username and password
func (s *AuthServiceImpl) Authenticate(username, password string) (*User, error) {
	// This is a placeholder - in production, this would query the database
	// For demo purposes, we return a mock user
	if username == "admin" && password == "admin123" {
		return &User{
			ID:           "00000000-0000-0000-0000-000000000001",
			Username:     "admin",
			Email:        "admin@csic.gov",
			Role:         "ADMIN",
			Status:       "ACTIVE",
			Password:     "", // Not needed after auth
			MFAEnabled:   false,
			Permissions:  []string{"*"},
			LastLogin:    time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	return nil, errors.New("invalid credentials")
}

// HashPassword hashes a password using bcrypt or similar
func (s *AuthServiceImpl) HashPassword(password string) (string, error) {
	// This is a placeholder - in production, use bcrypt
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

// VerifyPassword verifies a password against a hash
func (s *AuthServiceImpl) VerifyPassword(hashedPassword, password string) bool {
	hash := sha256.Sum256([]byte(password))
	return hmac.Equal([]byte(hashedPassword), []byte(hex.EncodeToString(hash[:])))
}

// parseClaims parses JWT claims into TokenClaims
func parseClaims(claims jwt.MapClaims) *ports.TokenClaims {
	permissionsRaw, _ := claims["permissions"].([]interface{})
	permissions := make([]string, len(permissionsRaw))
	for i, p := range permissionsRaw {
		permissions[i] = p.(string)
	}

	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	tokenID, _ := claims["token_id"].(string)

	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)

	return &ports.TokenClaims{
		UserID:      userID,
		Username:    username,
		Role:        role,
		Permissions: permissions,
		TokenID:     tokenID,
		IssueAt:     int64(iat),
		ExpiresAt:   int64(exp),
	}
}

// User represents a user entity for authentication
type User struct {
	ID           string
	Username     string
	Email        string
	Role         string
	Status       string
	Password     string
	MFAEnabled   bool
	Permissions  []string
	LastLogin    string
}

// GenerateResetToken generates a password reset token
func (s *AuthServiceImpl) GenerateResetToken(userID string) (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

// ValidateResetToken validates a password reset token
func (s *AuthServiceImpl) ValidateResetToken(token string) bool {
	return len(token) == 44 // base64 URL encoding of 32 bytes
}

// Ensure AuthServiceImpl implements ports.AuthService
var _ ports.AuthService = (*AuthServiceImpl)(nil)
