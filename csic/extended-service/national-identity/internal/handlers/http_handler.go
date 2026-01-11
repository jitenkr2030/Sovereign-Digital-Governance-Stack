package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/domain"
	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/service"
)

// IdentityHandler handles HTTP requests for identity operations
type IdentityHandler struct {
	service service.IdentityService
}

// NewIdentityHandler creates a new identity handler
func NewIdentityHandler(svc service.IdentityService) *IdentityHandler {
	return &IdentityHandler{
		service: svc,
	}
}

// CreateIdentity creates a new citizen identity
func (h *IdentityHandler) CreateIdentity(c *gin.Context) {
	var identity domain.Identity
	if err := c.ShouldBindJSON(&identity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.CreateIdentity(c.Request.Context(), &identity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "identity created successfully",
		"identity_id": identity.ID,
		"national_id": identity.NationalID,
	})
}

// GetIdentity retrieves an identity by ID
func (h *IdentityHandler) GetIdentity(c *gin.Context) {
	id := c.Param("id")

	identity, err := h.service.GetIdentity(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "identity not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, identity)
}

// UpdateIdentity updates an existing identity
func (h *IdentityHandler) UpdateIdentity(c *gin.Context) {
	id := c.Param("id")

	var identity domain.Identity
	if err := c.ShouldBindJSON(&identity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	identity.ID = id

	if err := h.service.UpdateIdentity(c.Request.Context(), &identity); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "identity not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "identity updated successfully",
	})
}

// DeleteIdentity soft-deletes an identity
func (h *IdentityHandler) DeleteIdentity(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteIdentity(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "identity not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "deletion_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "identity deleted successfully",
	})
}

// GetIdentityHistory retrieves the history of an identity
func (h *IdentityHandler) GetIdentityHistory(c *gin.Context) {
	id := c.Param("id")
	
	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	history, err := h.service.GetIdentityHistory(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity_id": id,
		"history":     history,
		"limit":       limit,
		"offset":      offset,
	})
}

// Verification endpoints

// VerifyIdentity handles identity verification requests
func (h *IdentityHandler) VerifyIdentity(c *gin.Context) {
	var req service.VerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.VerifyIdentity(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "verification_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verification_id": result.VerificationID,
		"passed":          result.Passed,
		"score":           result.Score,
		"threshold":       result.Threshold,
		"failure_reason":  result.FailureReason,
		"details":         result.Details,
	})
}

// BiometricVerification handles biometric verification requests
func (h *IdentityHandler) BiometricVerification(c *gin.Context) {
	var req service.BiometricVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.BiometricVerification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "verification_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verification_id": result.VerificationID,
		"passed":          result.Passed,
		"score":           result.Score,
		"threshold":       result.Threshold,
		"failure_reason":  result.FailureReason,
		"details":         result.Details,
	})
}

// DocumentVerification handles document verification requests
func (h *IdentityHandler) DocumentVerification(c *gin.Context) {
	var req service.DocumentVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.DocumentVerification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "verification_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verification_id": result.VerificationID,
		"passed":          result.Passed,
		"score":           result.Score,
		"threshold":       result.Threshold,
		"failure_reason":  result.FailureReason,
	})
}

// LivenessCheck handles liveness check requests
func (h *IdentityHandler) LivenessCheck(c *gin.Context) {
	var req service.LivenessCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.LivenessCheck(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "liveness_check_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passed":           result.Passed,
		"score":            result.Score,
		"spoofing_detected": result.SpoofingDetected,
		"session_id":       result.SessionID,
	})
}

// Authentication endpoints

// LoginRequest represents a login request
type LoginRequest struct {
	NationalID   string                   `json:"national_id" binding:"required"`
	Password     string                   `json:"password" binding:"required"`
	DeviceInfo   *service.DeviceInfo      `json:"device_info"`
	BehavioralData *service.BehavioralDataRequest `json:"behavioral_data"`
}

// Login handles user login
func (h *IdentityHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Get IP and UserAgent from request
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if req.DeviceInfo == nil {
		req.DeviceInfo = &service.DeviceInfo{
			IPAddress: ipAddress,
			UserAgent: userAgent,
		}
	} else {
		req.DeviceInfo.IPAddress = ipAddress
		req.DeviceInfo.UserAgent = userAgent
	}

	svcReq := &service.LoginRequest{
		NationalID:    req.NationalID,
		Password:      req.Password,
		DeviceInfo:    req.DeviceInfo,
		BehavioralData: req.BehavioralData,
	}

	result, err := h.service.Login(c.Request.Context(), svcReq)
	if err != nil {
		status := http.StatusUnauthorized
		c.JSON(status, gin.H{
			"error":   "login_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_token":    result.SessionToken,
		"refresh_token":    result.RefreshToken,
		"expires_at":       result.ExpiresAt.Format(time.RFC3339),
		"mfa_required":     result.MFARequired,
		"mfa_methods":      result.MFAMethods,
		"risk_score":       result.RiskScore,
		"behavioral_score": result.BehavioralScore,
	})
}

// Logout handles user logout
func (h *IdentityHandler) Logout(c *gin.Context) {
	sessionToken := c.GetHeader("Authorization")
	if sessionToken != "" {
		sessionToken = sessionToken[len("Bearer "):]
	}

	if err := h.service.Logout(c.Request.Context(), sessionToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "logout_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// RefreshToken handles token refresh
func (h *IdentityHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "refresh_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_token": result.SessionToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt.Format(time.RFC3339),
	})
}

// SetupMFA handles MFA setup requests
func (h *IdentityHandler) SetupMFA(c *gin.Context) {
	var req struct {
		IdentityID string `json:"identity_id" binding:"required"`
		TokenType  string `json:"token_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.SetupMFA(c.Request.Context(), req.IdentityID, req.TokenType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "mfa_setup_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":         result.Secret,
		"qr_code":        result.QRCode,
		"recovery_codes": result.RecoveryCodes,
	})
}

// VerifyMFA handles MFA verification requests
func (h *IdentityHandler) VerifyMFA(c *gin.Context) {
	var req service.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.VerifyMFA(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "mfa_verification_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA verified successfully",
	})
}

// PasswordReset handles password reset requests
func (h *IdentityHandler) PasswordReset(c *gin.Context) {
	var req service.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.PasswordReset(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "password_reset_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password reset email sent",
	})
}

// ChangePassword handles password change requests
func (h *IdentityHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "password_change_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}

// Behavioral biometrics endpoints

// SubmitKeystrokeData handles keystroke data submission
func (h *IdentityHandler) SubmitKeystrokeData(c *gin.Context) {
	var req struct {
		IdentityID string                 `json:"identity_id" binding:"required"`
		Data       *service.KeystrokeData `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SubmitKeystrokeData(c.Request.Context(), req.IdentityID, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "submission_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "keystroke data submitted successfully",
	})
}

// SubmitMouseData handles mouse data submission
func (h *IdentityHandler) SubmitMouseData(c *gin.Context) {
	var req struct {
		IdentityID string               `json:"identity_id" binding:"required"`
		Data       *service.MouseData   `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SubmitMouseData(c.Request.Context(), req.IdentityID, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "submission_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "mouse data submitted successfully",
	})
}

// AnalyzeBehavioralPattern handles behavioral pattern analysis requests
func (h *IdentityHandler) AnalyzeBehavioralPattern(c *gin.Context) {
	var req struct {
		IdentityID string `json:"identity_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.AnalyzeBehavioralPattern(c.Request.Context(), req.IdentityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "analysis_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"overall_score":     result.OverallScore,
		"anomaly_detected":  result.AnomalyDetected,
		"anomaly_score":     result.AnomalyScore,
		"recommendation":    result.Recommendation,
		"profile_data":      result.ProfileData,
	})
}

// GetBehavioralProfile retrieves the behavioral profile for an identity
func (h *IdentityHandler) GetBehavioralProfile(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.service.GetBehavioralProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "profile_not_found",
			"message": "No behavioral profile found for this identity",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// Fraud detection endpoints

// GetFraudScore retrieves the fraud score for an identity
func (h *IdentityHandler) GetFraudScore(c *gin.Context) {
	id := c.Param("id")

	result, err := h.service.GetFraudScore(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"score":       result.Score,
		"risk_level":  result.RiskLevel,
		"factors":     result.Factors,
		"last_updated": result.LastUpdated.Format(time.RFC3339),
	})
}

// ReportFraud handles fraud alert creation
func (h *IdentityHandler) ReportFraud(c *gin.Context) {
	var alert domain.FraudAlert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.ReportFraud(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "report_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "fraud alert created successfully",
		"alert_id":   alert.ID,
	})
}

// GetFraudAlerts retrieves fraud alerts for an identity
func (h *IdentityHandler) GetFraudAlerts(c *gin.Context) {
	id := c.Param("id")

	alerts, err := h.service.GetFraudAlerts(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity_id": id,
		"alerts":      alerts,
		"count":       len(alerts),
	})
}

// ResolveFraudAlert handles fraud alert resolution
func (h *IdentityHandler) ResolveFraudAlert(c *gin.Context) {
	alertID := c.Param("id")
	
	var req struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
		Resolution string `json:"resolution" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.ResolveFraudAlert(c.Request.Context(), alertID, req.ResolvedBy, req.Resolution); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "resolution_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "fraud alert resolved successfully",
	})
}

// Federated learning endpoints

// GetModelStatus retrieves the current federated model status
func (h *IdentityHandler) GetModelStatus(c *gin.Context) {
	status, err := h.service.GetModelStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SubmitModelUpdate handles federated model update submission
func (h *IdentityHandler) SubmitModelUpdate(c *gin.Context) {
	var update domain.FederatedModelUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SubmitModelUpdate(c.Request.Context(), &update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "submission_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "model update submitted successfully",
		"update_id":  update.ID,
	})
}

// GetClientContributions retrieves client contributions to the federated model
func (h *IdentityHandler) GetClientContributions(c *gin.Context) {
	clientID := c.Param("id")

	contributions, err := h.service.GetClientContributions(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":     clientID,
		"contributions": contributions,
	})
}

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Liveness handles liveness probe requests
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Readiness handles readiness probe requests
func (h *HealthHandler) Readiness(c *gin.Context) {
	// In a real implementation, check all dependencies (database, cache, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Health handles detailed health check requests
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "national-identity-service",
		"version":   "1.0.0",
	})
}
