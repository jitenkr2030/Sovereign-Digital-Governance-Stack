package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// KeycloakClient provides Keycloak integration for authentication
type KeycloakClient interface {
	// Authenticate authenticates a user
	Authenticate(ctx context.Context, username, password string) (*TokenResponse, error)

	// ValidateToken validates a token
	ValidateToken(ctx context.Context, token string) (*TokenInfo, error)

	// RefreshToken refreshes an access token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// GetUserInfo returns user information
	GetUserInfo(ctx context.Context, token string) (*UserInfo, error)

	// Logout logs out a user
	Logout(ctx context.Context, refreshToken string) error

	// GetUserRoles returns user roles
	GetUserRoles(ctx context.Context, token, userID string) ([]string, error)

	// HasRole checks if user has a role
	HasRole(ctx context.Context, token, userID, role string) (bool, error)
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	ServerURL      string `json:"server_url"`
	Realm          string `json:"realm"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	RedirectURI    string `json:"redirect_uri"`
	Scopes         string `json:"scopes"`
}

// TokenResponse represents Keycloak token response
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	SessionState     string `json:"session_state"`
}

// TokenInfo represents token information
type TokenInfo struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope"`
	ClientID  string   `json:"client_id"`
	Username  string   `json:"username"`
	UserID    string   `json:"sub"`
	Exp       int64    `json:"exp"`
	Iat       int64    `json:"iat"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud"`
}

// UserInfo represents user information from Keycloak
type UserInfo struct {
	ID        string   `json:"sub"`
	Username  string   `json:"preferred_username"`
	Email     string   `json:"email"`
	FirstName string   `json:"given_name"`
	LastName  string   `json:"family_name"`
	Roles     []string `json:"roles"`
}

// keycloakClient implements KeycloakClient
type keycloakClient struct {
	config     KeycloakConfig
	httpClient *http.Client
}

// NewKeycloakClient creates a new KeycloakClient
func NewKeycloakClient(serverURL, realm, clientID, clientSecret string) (*keycloakClient, error) {
	config := KeycloakConfig{
		ServerURL:    strings.TrimSuffix(serverURL, "/"),
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       "openid profile email",
	}

	client := &keycloakClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.healthCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Keycloak: %w", err)
	}

	return client, nil
}

// NewKeycloakClientWithConfig creates a KeycloakClient with full config
func NewKeycloakClientWithConfig(config KeycloakConfig) (*keycloakClient, error) {
	client := &keycloakClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.healthCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Keycloak: %w", err)
	}

	return client, nil
}

// Authenticate authenticates a user with username and password
func (c *keycloakClient) Authenticate(ctx context.Context, username, password string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", c.config.Scopes)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// ValidateToken validates an access token using Keycloak introspection
func (c *keycloakClient) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)

	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("token validation failed: %s", string(body))
	}

	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenInfo, nil
}

// RefreshToken refreshes an access token
func (c *keycloakClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo returns user information
func (c *keycloakClient) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}

// Logout logs out a user
func (c *keycloakClient) Logout(ctx context.Context, refreshToken string) error {
	data := url.Values{}
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("logout failed: %s", string(body))
	}

	return nil
}

// GetUserRoles returns user roles
func (c *keycloakClient) GetUserRoles(ctx context.Context, token, userID string) ([]string, error) {
	rolesURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings",
		c.config.ServerURL, c.config.Realm, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rolesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get roles: %s", string(body))
	}

	var roles []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	return roleNames, nil
}

// HasRole checks if user has a specific role
func (c *keycloakClient) HasRole(ctx context.Context, token, userID, role string) (bool, error) {
	roles, err := c.GetUserRoles(ctx, token, userID)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil
}

// healthCheck checks Keycloak health
func (c *keycloakClient) healthCheck(ctx context.Context) error {
	healthURL := fmt.Sprintf("%s/realms/%s",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Keycloak not healthy: status %d", resp.StatusCode)
	}

	return nil
}

// GetAuthCodeURL returns the OAuth2 authorization code URL
func (c *keycloakClient) GetAuthCodeURL(state, redirectURI string) string {
	params := url.Values{}
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", c.config.Scopes)
	params.Set("state", state)

	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?%s",
		c.config.ServerURL, c.config.Realm, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (c *keycloakClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("code exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}
