package handler

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/services/security/key-management/internal/config"
	"github.com/csic-platform/services/security/key-management/internal/core/domain"
	"github.com/csic-platform/services/security/key-management/internal/core/ports"
	"github.com/gin-gonic/gin"
)

// HTTPHandler contains all HTTP handlers for the Key Management Service
type HTTPHandler struct {
	service ports.KeyService
	cfg     *config.Config
}

// NewHTTPHandler creates a new HTTP handler instance
func NewHTTPHandler(service ports.KeyService, cfg *config.Config) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		cfg:     cfg,
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthCheck returns the health status of the service
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNHEALTHY",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   h.cfg.App.Version,
		},
	})
}

// CreateKey creates a new cryptographic key
func (h *HTTPHandler) CreateKey(c *gin.Context) {
	var req domain.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	key, err := h.service.CreateKey(c.Request.Context(), &req, actorID)
	if err != nil {
		if err.Error() == "key with this alias already exists" {
			c.JSON(http.StatusConflict, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "KEY_EXISTS",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "CREATE_FAILED",
				Message: "Failed to create key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    key,
	})
}

// ListKeys returns a paginated list of keys
func (h *HTTPHandler) ListKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	result, err := h.service.ListKeys(c.Request.Context(), page, pageSize, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "LIST_FAILED",
				Message: "Failed to list keys",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    result,
	})
}

// GetKey returns a specific key by ID
func (h *HTTPHandler) GetKey(c *gin.Context) {
	keyID := c.Param("id")

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	key, err := h.service.GetKey(c.Request.Context(), keyID, actorID)
	if err != nil {
		if err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Key not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "GET_FAILED",
				Message: "Failed to get key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    key,
	})
}

// RotateKey rotates a key to a new version
func (h *HTTPHandler) RotateKey(c *gin.Context) {
	keyID := c.Param("id")

	var req domain.RotateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	key, err := h.service.RotateKey(c.Request.Context(), keyID, &req, actorID)
	if err != nil {
		if err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Key not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "ROTATE_FAILED",
				Message: "Failed to rotate key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    key,
	})
}

// RevokeKey revokes a key
func (h *HTTPHandler) RevokeKey(c *gin.Context) {
	keyID := c.Param("id")

	var req domain.RevokeKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	if err := h.service.RevokeKey(c.Request.Context(), keyID, &req, actorID); err != nil {
		if err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Key not found",
				},
			})
			return
		}

		if err.Error() == "key is already revoked" {
			c.JSON(http.StatusConflict, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "ALREADY_REVOKED",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "REVOKE_FAILED",
				Message: "Failed to revoke key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":     keyID,
			"status": "REVOKED",
		},
	})
}

// GenerateDataKey generates a data key for envelope encryption
func (h *HTTPHandler) GenerateDataKey(c *gin.Context) {
	var req domain.GenerateDataKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	response, err := h.service.GenerateDataKey(c.Request.Context(), &req, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "GENERATE_FAILED",
				Message: "Failed to generate data key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// EncryptData encrypts data using a key
func (h *HTTPHandler) EncryptData(c *gin.Context) {
	var req domain.EncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	response, err := h.service.EncryptData(c.Request.Context(), &req, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "ENCRYPT_FAILED",
				Message: "Failed to encrypt data",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// DecryptData decrypts data using a key
func (h *HTTPHandler) DecryptData(c *gin.Context) {
	var req domain.DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	response, err := h.service.DecryptData(c.Request.Context(), &req, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "DECRYPT_FAILED",
				Message: "Failed to decrypt data",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// GetKeyAuditLog returns the audit log for a key
func (h *HTTPHandler) GetKeyAuditLog(c *gin.Context) {
	keyID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	logs, err := h.service.GetKeyAuditLog(c.Request.Context(), keyID, limit, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "AUDIT_FAILED",
				Message: "Failed to get audit log",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    logs,
	})
}

// GetKeyVersions returns all versions of a key
func (h *HTTPHandler) GetKeyVersions(c *gin.Context) {
	keyID := c.Param("id")

	actorID := c.GetString("user_id")
	if actorID == "" {
		actorID = "system"
	}

	versions, err := h.service.GetKeyVersions(c.Request.Context(), keyID, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "VERSIONS_FAILED",
				Message: "Failed to get key versions",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    versions,
	})
}

// ErrorHandler handles panics and returns proper error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Error: &ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: "An internal error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Writer.Header().Set("Pragma", "no-cache")

		c.Next()
	}
}

// decodeBase64 decodes a base64-encoded string
func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// encodeBase64 encodes bytes to a base64 string
func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
