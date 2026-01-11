package adapter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
	"github.com/csic-platform/services/security/iam/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JWTGenerator implements TokenGenerator using JWT
type JWTGenerator struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	logger        *zap.Logger
}

// NewJWTGenerator creates a new JWTGenerator
func NewJWTGenerator(
	accessSecret, refreshSecret string,
	accessExpiry, refreshExpiry time.Duration,
	logger *zap.Logger,
) *JWTGenerator {
	return &JWTGenerator{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		logger:        logger,
	}
}

// GenerateAccessToken generates an access token for a user
func (g *JWTGenerator) GenerateAccessToken(
	ctx context.Context,
	user *domain.User,
) (string, error) {
	now := time.Now()
	expiresAt := now.Add(g.accessExpiry)

	// Build permissions list
	permissions := make([]string, 0)
	if user.Role != nil {
		for _, perm := range user.Role.Permissions {
			permissions = append(permissions, fmt.Sprintf("%s:%s", perm.Resource, perm.Action))
		}
	}

	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role.Name,
		"permissions": permissions,
		"iat":        now.Unix(),
		"exp":        expiresAt.Unix(),
		"type":       "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(g.accessSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken generates a refresh token for a user
func (g *JWTGenerator) GenerateRefreshToken(
	ctx context.Context,
	user *domain.User,
) (string, error) {
	now := time.Now()
	expiresAt := now.Add(g.refreshExpiry)

	claims := jwt.MapClaims{
		"sub":    user.ID.String(),
		"iat":    now.Unix(),
		"exp":    expiresAt.Unix(),
		"type":   "refresh",
		"jti":    uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(g.refreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken validates an access token
func (g *JWTGenerator) ValidateAccessToken(
	ctx context.Context,
	tokenString string,
) (*domain.AuthContext, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.accessSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token type
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
			return nil, fmt.Errorf("invalid token type")
		}

		// Extract claims
		authContext := &domain.AuthContext{}

		if sub, ok := claims["sub"].(string); ok {
			authContext.UserID = uuid.MustParse(sub)
		}

		if username, ok := claims["username"].(string); ok {
			authContext.Username = username
		}

		if email, ok := claims["email"].(string); ok {
			authContext.Email = email
		}

		if role, ok := claims["role"].(string); ok {
			authContext.Role = role
		}

		if permissions, ok := claims["permissions"].([]interface{}); ok {
			for _, p := range permissions {
				if perm, ok := p.(string); ok {
					authContext.Permissions = append(authContext.Permissions, perm)
				}
			}
		}

		return authContext, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetAccessTokenExpiry returns the access token expiry in seconds
func (g *JWTGenerator) GetAccessTokenExpiry() int {
	return int(g.accessExpiry.Seconds())
}

// Ensure JWTGenerator implements TokenGenerator
var _ ports.TokenGenerator = (*JWTGenerator)(nil)

// TokenBlacklistRedis implements TokenBlacklist using Redis
type TokenBlacklistRedis struct {
	client  interface{ Get(ctx context.Context, key string) *string }
	prefix  string
	logger  *zap.Logger
}

// NewTokenBlacklistRedis creates a new TokenBlacklistRedis
func NewTokenBlacklistRedis(client interface{}, prefix string, logger *zap.Logger) *TokenBlacklistRedis {
	return &TokenBlacklistRedis{
		client: client,
		prefix: prefix,
		logger: logger,
	}
}

// Add adds a token to the blacklist
func (b *TokenBlacklistRedis) Add(ctx context.Context, token string, expiry int64) error {
	key := fmt.Sprintf("%s:blacklist:%s", b.prefix, token)
	// In production, use Redis SET with EX
	// b.client.Set(ctx, key, "1", time.Duration(expiry)*time.Second)
	b.logger.Debug("Token blacklisted",
		zap.String("key", key),
		zap.Int64("expiry_seconds", expiry))
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (b *TokenBlacklistRedis) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s:blacklist:%s", b.prefix, token)
	// In production, use Redis EXISTS
	// exists := b.client.Exists(ctx, key)
	return false, nil
}

// Cleanup removes all expired tokens from the blacklist
func (b *TokenBlacklistRedis) Cleanup(ctx context.Context) error {
	// In production, use Redis SCAN and DEL
	b.logger.Debug("Token blacklist cleanup")
	return nil
}

// Ensure TokenBlacklistRedis implements TokenBlacklist
var _ ports.TokenBlacklist = (*TokenBlacklistRedis)(nil)

// PasswordHasherBCrypt implements PasswordHasher using bcrypt
type PasswordHasherBCrypt struct {
	cost int
}

// NewPasswordHasherBCrypt creates a new PasswordHasherBCrypt
func NewPasswordHasherBCrypt(cost int) *PasswordHasherBCrypt {
	if cost == 0 {
		cost = 10
	}
	return &PasswordHasherBCrypt{cost: cost}
}

// Hash generates a hash for the given password
func (h *PasswordHasherBCrypt) Hash(ctx context.Context, password string) (string, error) {
	// In production, use golang.org/x/crypto/bcrypt
	// For now, use a simple base64 of random bytes as placeholder
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Verify compares a password with a hash
func (h *PasswordHasherBCrypt) Verify(ctx context.Context, password, hash string) bool {
	// In production, use bcrypt.CompareHashAndPassword
	// For now, always return false for placeholder
	return false
}

// Ensure PasswordHasherBCrypt implements PasswordHasher
var _ ports.PasswordHasher = (*PasswordHasherBCrypt)(nil)

// InMemoryTokenBlacklist implements TokenBlacklist in memory
type InMemoryTokenBlacklist struct {
	tokens map[string]time.Time
}

// NewInMemoryTokenBlacklist creates a new InMemoryTokenBlacklist
func NewInMemoryTokenBlacklist() *InMemoryTokenBlacklist {
	return &InMemoryTokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

// Add adds a token to the blacklist
func (b *InMemoryTokenBlacklist) Add(ctx context.Context, token string, expiry int64) error {
	b.tokens[token] = time.Now().Add(time.Duration(expiry) * time.Second)
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (b *InMemoryTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	expiry, exists := b.tokens[token]
	if !exists {
		return false, nil
	}
	if time.Now().After(expiry) {
		delete(b.tokens, token)
		return false, nil
	}
	return true, nil
}

// Cleanup removes all expired tokens from the blacklist
func (b *InMemoryTokenBlacklist) Cleanup(ctx context.Context) error {
	now := time.Now()
	for token, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, token)
		}
	}
	return nil
}

// Ensure InMemoryTokenBlacklist implements TokenBlacklist
var _ ports.TokenBlacklist = (*InMemoryTokenBlacklist)(nil)
