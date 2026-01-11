package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"neam-platform/audit-service/crypto"
	"neam-platform/audit-service/models"
	"neam-platform/audit-service/repository"
	"neam-platform/audit-service/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepo         repository.UserRepository
	accessControl    services.AccessControl
	keycloakClient   interface{}
	signatureService crypto.SignatureService
	jwtSecret        []byte
	jwtIssuer        string
	jwtAudience      string
	tokenExpiry      time.Duration
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	userRepo repository.UserRepository,
	accessControl services.AccessControl,
	keycloakClient interface{},
	signatureService crypto.SignatureService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		accessControl:    accessControl,
		keycloakClient:   keycloakClient,
		signatureService: signatureService,
		jwtSecret:        []byte("your-jwt-secret-key"),
		jwtIssuer:        "neam-platform",
		jwtAudience:      "neam-audit",
		tokenExpiry:      time.Hour,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/verify", h.VerifyToken)
		auth.GET("/me", h.GetCurrentUser)
		auth.POST("/mfa/verify", h.VerifyMFA)
	}
}

// Login handles POST /api/v1/auth/login
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Find user
	user, err := h.userRepo.FindByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid credentials",
		})
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "account_disabled",
			Message: "Account is disabled",
		})
		return
	}

	// Verify password (in production, use proper password hashing comparison)
	if !verifyPassword(req.Password, user.PasswordHash) {
		// Increment failed login count
		h.userRepo.IncrementFailedLogins(c.Request.Context(), user.ID)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid credentials",
		})
		return
	}

	// Check if MFA is required
	if user.MFEnabled {
		// Generate MFA token
		mfaToken, err := generateMFAToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "mfa_generation_failed",
				Message: "Failed to generate MFA token",
			})
			return
		}

		c.JSON(http.StatusOK, MFARequiredResponse{
			MFAToken: mfaToken,
			Message:  "MFA verification required",
		})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate tokens",
		})
		return
	}

	// Update last login
	now := time.Now()
	h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID, &now)

	// Create session
	session := models.Session{
		ID:           generateSessionID(),
		UserID:       user.ID,
		Token:        accessToken,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		ExpiresAt:    time.Now().Add(h.tokenExpiry),
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		Revoked:      false,
	}

	// Store session (implementation depends on session store)
	_ = session

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.tokenExpiry.Seconds()),
		User: UserSummary{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        user.Role,
			Permissions: user.Permissions,
			Groups:      user.Groups,
			MFEnabled:   user.MFEnabled,
		},
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(req.RefreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid refresh token",
		})
		return
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid refresh token claims",
		})
		return
	}

	// Check if token is expired
	if claims.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "token_expired",
			Message: "Refresh token has expired",
		})
		return
	}

	// Get user
	user, err := h.userRepo.FindByID(c.Request.Context(), claims.UserID)
	if err != nil || !user.Active {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found or inactive",
		})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := h.generateTokens(user, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.tokenExpiry.Seconds()),
	})
}

// Logout handles POST /api/v1/auth/logout
// @Summary User logout
// @Description Invalidate current session
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} SuccessResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from context
	sessionID, _ := c.Get("session_id")

	// Revoke session (implementation depends on session store)
	if sid, ok := sessionID.(string); ok {
		_ = h.revokeSession(sid)
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// VerifyToken handles POST /api/v1/auth/verify
// @Summary Verify token
// @Description Verify if a token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Param token body TokenVerifyRequest true "Token to verify"
// @Success 200 {object} TokenVerifyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/verify [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req TokenVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	token, err := jwt.ParseWithClaims(req.Token, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: err.Error(),
		})
		return
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid token claims",
		})
		return
	}

	c.JSON(http.StatusOK, TokenVerifyResponse{
		Valid:     true,
		ExpiresAt: claims.ExpiresAt,
		Subject:   claims.Subject,
	})
}

// GetCurrentUser handles GET /api/v1/auth/me
// @Summary Get current user
// @Description Get information about the currently authenticated user
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200 {object} UserSummary
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "not_authenticated",
			Message: "User not authenticated",
		})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, UserSummary{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
		Groups:      user.Groups,
		MFEnabled:   user.MFEnabled,
	})
}

// VerifyMFA handles POST /api/v1/auth/mfa/verify
// @Summary Verify MFA code
// @Description Verify MFA code and complete authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param mfa_code body MFAVerifyRequest true "MFA verification code"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/mfa/verify [post]
func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Verify MFA token and get user
	mfaToken, err := validateMFAToken(req.MFAToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_mfa_token",
			Message: "Invalid or expired MFA token",
		})
		return
	}

	// Verify MFA code (implementation depends on MFA method)
	if !verifyMFACode(mfaToken.UserID, req.MFACode) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_mfa_code",
			Message: "Invalid MFA code",
		})
		return
	}

	// Get user
	user, err := h.userRepo.FindByID(c.Request.Context(), mfaToken.UserID)
	if err != nil || !user.Active {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found or inactive",
		})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.tokenExpiry.Seconds()),
		User: UserSummary{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        user.Role,
			Permissions: user.Permissions,
			Groups:      user.Groups,
			MFEnabled:   user.MFEnabled,
		},
	})
}

// Token generation helpers

func (h *AuthHandler) generateTokens(user *models.User, ipAddress string) (string, string, error) {
	now := time.Now()

	// Access token
	accessClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.jwtIssuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{h.jwtAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(h.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
		SessionID:   generateSessionID(),
		IPAddress:   ipAddress,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(h.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh token (longer expiry)
	refreshClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.jwtIssuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{h.jwtAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		SessionID: accessClaims.SessionID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(h.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *AuthHandler) revokeSession(sessionID string) error {
	// Implementation depends on session store
	return nil
}

// Helper functions

func verifyPassword(password, hash string) bool {
	// In production, use proper password hashing (bcrypt, argon2)
	return password == hash // Simplified for demo
}

func generateSessionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func generateMFAToken(userID string) (string, error) {
	// Simplified - in production use proper MFA token generation
	return userID + "|" + strconv.FormatInt(time.Now().Add(5*time.Minute).Unix(), 36), nil
}

func validateMFAToken(token string) (*MFAToken, error) {
	parts := splitToken(token)
	if len(parts) != 2 {
		return nil, nil
	}
	return &MFAToken{
		UserID: parts[0],
		Expiry: parts[1],
	}, nil
}

func verifyMFACode(userID, code string) bool {
	// Simplified - in production use proper TOTP verification
	return code == "123456"
}

func splitToken(token string) []string {
	// Simple split - in production use proper parsing
	var parts []string
	start := 0
	for i, c := range token {
		if c == '|' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

// Claims types

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	IPAddress   string   `json:"ip_address"`
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	SessionID string `json:"session_id"`
}

type MFAToken struct {
	UserID string
	Expiry string
}

// Request/Response types

type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	User         UserSummary `json:"user"`
	SessionID    string      `json:"session_id"`
}

type MFARequiredResponse struct {
	MFAToken string `json:"mfa_token"`
	Message  string `json:"message"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenVerifyRequest struct {
	Token string `json:"token" binding:"required"`
}

type TokenVerifyResponse struct {
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Subject   string    `json:"subject,omitempty"`
}

type MFAVerifyRequest struct {
	MFAToken string `json:"mfa_token" binding:"required"`
	MFACode  string `json:"mfa_code" binding:"required"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
