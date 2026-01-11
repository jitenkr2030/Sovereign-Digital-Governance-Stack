package handler

import (
	"encoding/json"
	"net/http"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
	"github.com/csic-platform/services/security/iam/internal/core/ports"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// HTTPHandler handles HTTP requests for the IAM service
type HTTPHandler struct {
	authService   ports.AuthenticationService
	userService   ports.UserService
	roleService   ports.RoleService
	sessionService ports.SessionService
	logger        *zap.Logger
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(
	authService ports.AuthenticationService,
	userService ports.UserService,
	roleService ports.RoleService,
	sessionService ports.SessionService,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		authService:    authService,
		userService:    userService,
		roleService:    roleService,
		sessionService: sessionService,
		logger:         logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(r *chi.Mux) {
	// Authentication routes
	r.Post("/api/v1/auth/login", h.Login)
	r.Post("/api/v1/auth/register", h.Register)
	r.Post("/api/v1/auth/refresh", h.Refresh)
	r.Post("/api/v1/auth/logout", h.Logout)
	r.Post("/api/v1/auth/logout-all", h.LogoutAll)

	// User routes
	r.Get("/api/v1/users/me", h.GetCurrentUser)
	r.Put("/api/v1/users/me/password", h.ChangePassword)
	r.Post("/api/v1/users", h.CreateUser)
	r.Get("/api/v1/users", h.ListUsers)
	r.Get("/api/v1/users/{id}", h.GetUser)
	r.Put("/api/v1/users/{id}", h.UpdateUser)
	r.Delete("/api/v1/users/{id}", h.DeleteUser)
	r.Post("/api/v1/users/{id}/enable", h.EnableUser)
	r.Post("/api/v1/users/{id}/disable", h.DisableUser)
	r.Post("/api/v1/users/{id}/reset-password", h.ResetPassword)

	// Role routes
	r.Get("/api/v1/roles", h.ListRoles)
	r.Post("/api/v1/roles", h.CreateRole)
	r.Get("/api/v1/roles/{id}", h.GetRole)
	r.Put("/api/v1/roles/{id}", h.UpdateRole)
	r.Delete("/api/v1/roles/{id}", h.DeleteRole)

	// Session routes
	r.Get("/api/v1/sessions", h.GetSessions)
	r.Delete("/api/v1/sessions/{id}", h.RevokeSession)
	r.Delete("/api/v1/sessions", h.RevokeAllSessions)

	// Health check
	r.Get("/health", h.HealthCheck)
}

// Authentication handlers

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	resp, err := h.authService.Authenticate(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Registration failed", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})
}

func (h *HTTPHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	if err := h.authService.Logout(r.Context(), token); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *HTTPHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	// Extract user ID from token
	authContext, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	if err := h.authService.LogoutAll(r.Context(), authContext.UserID.String()); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "All sessions logged out"})
}

// User handlers

func (h *HTTPHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	authContext, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	user, err := h.userService.GetUser(r.Context(), authContext.UserID.String())
	if err != nil {
		h.writeError(w, http.StatusNotFound, "User not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

func (h *HTTPHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	authContext, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.userService.ChangePassword(r.Context(), authContext.UserID.String(), req.OldPassword, req.NewPassword); err != nil {
		h.writeError(w, http.StatusBadRequest, "Password change failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

func (h *HTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "User creation failed", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var req ports.ListUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.userService.ListUsers(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "User not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

func (h *HTTPHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateUserRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.userService.UpdateUser(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "User update failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

func (h *HTTPHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, "User deletion failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *HTTPHandler) EnableUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.userService.EnableUser(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, "User enable failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "User enabled successfully"})
}

func (h *HTTPHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.userService.DisableUser(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, "User disable failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "User disabled successfully"})
}

func (h *HTTPHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resp, err := h.userService.ResetPassword(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Password reset failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Role handlers

func (h *HTTPHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleService.ListRoles(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list roles", err)
		return
	}

	h.writeJSON(w, http.StatusOK, roles)
}

func (h *HTTPHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.roleService.CreateRole(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Role creation failed", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	role, err := h.roleService.GetRole(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Role not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, role)
}

func (h *HTTPHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateRoleRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.roleService.UpdateRole(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Role update failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Role updated successfully"})
}

func (h *HTTPHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.roleService.DeleteRole(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, "Role deletion failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

// Session handlers

func (h *HTTPHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	authContext, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	sessions, err := h.sessionService.GetSessions(r.Context(), authContext.UserID.String())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get sessions", err)
		return
	}

	h.writeJSON(w, http.StatusOK, sessions)
}

func (h *HTTPHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.sessionService.RevokeSession(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, "Session revocation failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Session revoked successfully"})
}

func (h *HTTPHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "Missing authorization token", nil)
		return
	}

	authContext, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	if err := h.sessionService.RevokeAllSessions(r.Context(), authContext.UserID.String()); err != nil {
		h.writeError(w, http.StatusBadRequest, "Session revocation failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "All sessions revoked successfully"})
}

// Health check

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "iam",
	})
}

// Helper methods

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		h.logger.Error(message, zap.Error(err))
	}
	h.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	})
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
