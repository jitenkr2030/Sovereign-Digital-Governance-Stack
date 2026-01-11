// Auth Service - JWT Authentication for CSIC Platform
// Provides authentication and authorization utilities

package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AuthService handles JWT token generation and validation
type AuthService struct {
	secret []byte
}

// NewAuthService creates a new authentication service
func NewAuthService(secret string) *AuthService {
	return &AuthService{
		secret: []byte(secret),
	}
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	Subject      string   `json:"sub"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions"`
	SessionID    string   `json:"session_id"`
	IssuedAt     int64    `json:"iat"`
	ExpiresAt    int64    `json:"exp"`
	NotBefore    int64    `json:"nbf"`
	Issuer       string   `json:"iss"`
}

// GenerateToken creates a new JWT token for a user
func (s *AuthService) GenerateToken(
	subject string,
	email string,
	name string,
	role string,
	permissions []string,
	sessionID string,
	expiryHours int,
) (string, error) {
	// Generate session ID if not provided
	if sessionID == "" {
		var err error
		sessionID, err = generateSessionID()
		if err != nil {
			return "", fmt.Errorf("failed to generate session ID: %w", err)
		}
	}

	now := time.Now()
	claims := TokenClaims{
		Subject:     subject,
		Email:       email,
		Name:        name,
		Role:        role,
		Permissions: permissions,
		SessionID:   sessionID,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(time.Duration(expiryHours) * time.Hour).Unix(),
		NotBefore:   now.Unix(),
		Issuer:      "csic-platform",
	}

	// Encode header
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	signatureInput := fmt.Sprintf("%s.%s", header, claimsB64)
	signature := s.sign(signatureInput)

	return fmt.Sprintf("%s.%s.%s", header, claimsB64, signature), nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Verify signature
	signatureInput := fmt.Sprintf("%s.%s", parts[0], parts[1])
	expectedSignature := s.sign(signatureInput)
	if parts[2] != expectedSignature {
		return nil, errors.New("invalid token signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	// Check not before
	if time.Now().Unix() < claims.NotBefore {
		return nil, errors.New("token is not yet valid")
	}

	return &claims, nil
}

// RefreshToken creates a new token from an existing valid token
func (s *AuthService) RefreshToken(token string, expiryHours int) (string, error) {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return "", err
	}

	return s.GenerateToken(
		claims.Subject,
		claims.Email,
		claims.Name,
		claims.Role,
		claims.Permissions,
		claims.SessionID,
		expiryHours,
	)
}

// sign creates an HMAC-SHA256 signature
func (s *AuthService) sign(data string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// generateSessionID creates a unique session identifier
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HasPermission checks if the claims include a specific permission
func (c *TokenClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission || p == "write:all" {
			return true
		}
	}
	return false
}

// HasRole checks if the claims include a specific role
func (c *TokenClaims) HasRole(roles ...string) bool {
	for _, role := range roles {
		if c.Role == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims include any of the specified roles
func (c *TokenClaims) HasAnyRole(roles ...string) bool {
	return c.HasRole(roles...)
}
