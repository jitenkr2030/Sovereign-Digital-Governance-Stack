package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"neam-platform/audit-service/models"
)

// AccessControl provides zero-trust access control
type AccessControl interface {
	// CheckPermission checks if a user has permission to perform an action
	CheckPermission(ctx context.Context, resource, action string) error

	// CheckRole checks if a user has a specific role
	CheckRole(ctx context.Context, role string) error

	// CheckMultiplePermissions checks multiple permissions
	CheckMultiplePermissions(ctx context.Context, permissions []string) error

	// EvaluatePolicy evaluates an access control policy
	EvaluatePolicy(ctx context.Context, policy models.Policy, request models.AuthorizationRequest) bool

	// GetUserPermissions gets all permissions for a user
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)

	// HasPermission checks if a user has a specific permission
	HasPermission(ctx context.Context, userID, permission string) bool

	// IsResourceAllowed checks if an action on a resource is allowed
	IsResourceAllowed(ctx context.Context, userID, resource, action string) bool

	// AddPolicy adds a new policy
	AddPolicy(policy models.Policy)

	// RemovePolicy removes a policy
	RemovePolicy(policyID string)

	// GetApplicablePolicies gets all applicable policies for a request
	GetApplicablePolicies(resource, action string) []models.Policy
}

// accessControl implements AccessControl
type accessControl struct {
	policies    map[string][]models.Policy
	roles       map[string]models.Role
	permissions map[string][]string
	mu          sync.RWMutex
}

// NewAccessControlService creates a new AccessControl service
func NewAccessControlService() AccessControl {
	ac := &accessControl{
		policies:    make(map[string][]models.Policy),
		roles:       make(map[string]models.Role),
		permissions: make(map[string][]string),
	}

	// Initialize default roles and permissions
	ac.initializeDefaults()

	return ac
}

// initializeDefaults initializes default roles and permissions
func (ac *accessControl) initializeDefaults() {
	// Define roles
	roles := []models.Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system access",
			Permissions: []string{"*"},
			IsSystem:    true,
		},
		{
			ID:          "auditor",
			Name:        "Auditor",
			Description: "Read-only access to audit logs",
			Permissions: []string{"audit:read", "audit:verify", "report:read"},
			IsSystem:    true,
		},
		{
			ID:          "security-officer",
			Name:        "Security Officer",
			Description: "Manage security policies and view security events",
			Permissions: []string{"audit:*", "policy:*", "security:*", "forensic:*"},
			IsSystem:    true,
		},
		{
			ID:          "investigator",
			Name:        "Investigator",
			Description: "Conduct forensic investigations",
			Permissions: []string{"audit:read", "forensic:create", "forensic:read", "report:read"},
			IsSystem:    true,
		},
		{
			ID:          "operator",
			Name:        "Operator",
			Description: "Day-to-day system operations",
			Permissions: []string{"audit:read", "system:read", "system:configure"},
			IsSystem:    true,
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "Read-only access",
			Permissions: []string{"audit:read", "report:read"},
			IsSystem:    true,
		},
	}

	for _, role := range roles {
		ac.roles[role.ID] = role
		ac.permissions[role.ID] = role.Permissions
	}

	// Define default policies
	ac.policies["audit"] = []models.Policy{
		{
			ID:          "audit-read-own",
			Name:        "Read Own Audit Events",
			Description: "Users can read their own audit events",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin", "auditor", "security-officer", "investigator", "operator", "viewer"},
			},
			Resources:   []string{"audit"},
			Actions:     []string{"read"},
		},
		{
			ID:          "audit-verify",
			Name:        "Verify Audit Integrity",
			Description: "Only security officers and admins can verify integrity",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin", "security-officer"},
			},
			Resources:   []string{"audit"},
			Actions:     []string{"verify"},
		},
		{
			ID:          "audit-seal",
			Name:        "Seal Audit Records",
			Description: "Only admins can seal audit records",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin"},
			},
			Resources:   []string{"audit"},
			Actions:     []string{"seal"},
		},
	}

	ac.policies["forensic"] = []models.Policy{
		{
			ID:          "forensic-create",
			Name:        "Create Forensic Investigations",
			Description: "Investigators and security officers can create investigations",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin", "security-officer", "investigator"},
			},
			Resources:   []string{"forensic"},
			Actions:     []string{"create"},
		},
		{
			ID:          "forensic-approve",
			Name:        "Approve Forensic Investigations",
			Description: "Only security officers can approve forensic investigations",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin", "security-officer"},
			},
			Resources:   []string{"forensic"},
			Actions:     []string{"approve"},
		},
	}

	ac.policies["policy"] = []models.Policy{
		{
			ID:          "policy-manage",
			Name:        "Manage Policies",
			Description: "Only admins can manage policies",
			Effect:      "allow",
			Priority:    100,
			Active:      true,
			Subjects:    models.PolicySubjects{
				Roles: []string{"admin"},
			},
			Resources:   []string{"policy"},
			Actions:     []string{"*"},
		},
	}
}

// CheckPermission checks if the current context has permission
func (ac *accessControl) CheckPermission(ctx context.Context, resource, action string) error {
	// Get user from context
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return fmt.Errorf("unauthorized: no user in context")
	}

	if !ac.HasPermission(ctx, userID, fmt.Sprintf("%s:%s", resource, action)) {
		return fmt.Errorf("forbidden: insufficient permissions for %s:%s", resource, action)
	}

	return nil
}

// CheckRole checks if the current context has a specific role
func (ac *accessControl) CheckRole(ctx context.Context, role string) error {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return fmt.Errorf("unauthorized: no user in context")
	}

	userRole := getRoleFromContext(ctx)
	if userRole == "" {
		return fmt.Errorf("forbidden: no role in context")
	}

	if userRole != role && userRole != "admin" {
		return fmt.Errorf("forbidden: requires role %s", role)
	}

	return nil
}

// CheckMultiplePermissions checks multiple permissions
func (ac *accessControl) CheckMultiplePermissions(ctx context.Context, permissions []string) error {
	for _, perm := range permissions {
		if err := ac.CheckPermission(ctx, parseResource(perm), parseAction(perm)); err != nil {
			return err
		}
	}
	return nil
}

// EvaluatePolicy evaluates an access control policy
func (ac *accessControl) EvaluatePolicy(ctx context.Context, policy models.Policy, request models.AuthorizationRequest) bool {
	// Check if policy is active
	if !policy.Active {
		return false
	}

	// Check if resource matches
	matchesResource := false
	for _, r := range policy.Resources {
		if matchWildcard(request.Resource, r) {
			matchesResource = true
			break
		}
	}
	if !matchesResource {
		return false
	}

	// Check if action matches
	matchesAction := false
	for _, a := range policy.Actions {
		if a == "*" || a == request.Action {
			matchesAction = true
			break
		}
	}
	if !matchesAction {
		return false
	}

	return policy.Effect == "allow"
}

// GetUserPermissions gets all permissions for a user
func (ac *accessControl) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	role := getRoleFromContext(ctx)
	if role == "" {
		// Default to empty permissions
		return []string{}, nil
	}

	perms, ok := ac.permissions[role]
	if !ok {
		return []string{}, nil
	}

	// Return a copy
	result := make([]string, len(perms))
	copy(result, perms)
	return result, nil
}

// HasPermission checks if a user has a specific permission
func (ac *accessControl) HasPermission(ctx context.Context, userID, permission string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	role := getRoleFromContext(ctx)
	if role == "" {
		return false
	}

	// Admin has all permissions
	if role == "admin" {
		return true
	}

	perms, ok := ac.permissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == "*" || p == permission {
			return true
		}
		// Check for wildcard resource
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, ":*")
			if strings.HasPrefix(permission, prefix+":") {
				return true
			}
		}
	}

	return false
}

// IsResourceAllowed checks if an action on a resource is allowed
func (ac *accessControl) IsResourceAllowed(ctx context.Context, userID, resource, action string) bool {
	// Get applicable policies
	policies := ac.GetApplicablePolicies(resource, action)

	// Evaluate policies
	request := models.AuthorizationRequest{
		UserID:     userID,
		Resource:   resource,
		Action:     action,
		Attributes: map[string]string{},
	}

	for _, policy := range policies {
		if ac.EvaluatePolicy(ctx, policy, request) {
			return policy.Effect == "allow"
		}
	}

	// Default deny
	return false
}

// AddPolicy adds a new policy
func (ac *accessControl) AddPolicy(policy models.Policy) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.policies[policy.ID] = append(ac.policies[policy.ID], policy)
}

// RemovePolicy removes a policy
func (ac *accessControl) RemovePolicy(policyID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	delete(ac.policies, policyID)
}

// GetApplicablePolicies gets all applicable policies for a request
func (ac *accessControl) GetApplicablePolicies(resource, action string) []models.Policy {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var allPolicies []models.Policy

	for _, policies := range ac.policies {
		for _, policy := range policies {
			if policy.Active {
				// Check if policy applies to this resource
				for _, r := range policy.Resources {
					if matchWildcard(resource, r) {
						// Check if policy applies to this action
						for _, a := range policy.Actions {
							if a == "*" || a == action {
								allPolicies = append(allPolicies, policy)
								break
							}
						}
						break
					}
				}
			}
		}
	}

	// Sort by priority (higher priority first)
	for i := 0; i < len(allPolicies)-1; i++ {
		for j := i + 1; j < len(allPolicies); j++ {
			if allPolicies[j].Priority > allPolicies[i].Priority {
				allPolicies[i], allPolicies[j] = allPolicies[j], allPolicies[i]
			}
		}
	}

	return allPolicies
}

// matchWildcard checks if a string matches a pattern with wildcards
func matchWildcard(s, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(s, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(s, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(s, pattern[:len(pattern)-1])
	}
	return s == pattern
}

// parseResource parses a permission string to extract the resource
func parseResource(perm string) string {
	parts := strings.SplitN(perm, ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return perm
}

// parseAction parses a permission string to extract the action
func parseAction(perm string) string {
	parts := strings.SplitN(perm, ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return "*"
}

// getUserIDFromContext extracts user ID from context
func getUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// getRoleFromContext extracts role from context
func getRoleFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if role, ok := ctx.Value("role").(string); ok {
		return role
	}
	return ""
}

// ContextWithUser creates a context with user information
func ContextWithUser(ctx context.Context, userID, role string) context.Context {
	return context.WithValue(ctx, "user_id", userID).
		WithValue(ctx, "role", role)
}

// ContextWithUserAndSession creates a context with user and session information
func ContextWithUserAndSession(ctx context.Context, userID, role, sessionID string) context.Context {
	return context.WithValue(ContextWithUser(ctx, userID, role), "session_id", sessionID)
}

// ZeroTrustPolicy represents a zero-trust access policy
type ZeroTrustPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Rules       []ZeroTrustRule        `json:"rules"`
	Priority    int                    `json:"priority"`
	Active      bool                   `json:"active"`
}

// ZeroTrustRule represents a single rule in a zero-trust policy
type ZeroTrustRule struct {
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Attributes  map[string]interface{} `json:"attributes"`
	RiskScore   int                    `json:"risk_score"`
}

// RiskAssessment represents a risk assessment for access
type RiskAssessment struct {
	Score       int                    `json:"score"`
	Factors     []RiskFactor           `json:"factors"`
	Recommendation string              `json:"recommendation"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RiskFactor represents a factor in risk assessment
type RiskFactor struct {
	Name        string `json:"name"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}
