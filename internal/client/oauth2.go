// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OAuth2TokenResponse represents the response from the OAuth2 token endpoint
type OAuth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// OAuth2Config holds configuration for OAuth2 client credentials authentication
type OAuth2Config struct {
	IdentityURL  string // CyberArk Identity URL (e.g., https://abc123.cyberark.cloud)
	ClientID     string // Service account username (full format: user@cyberark.cloud.12345)
	ClientSecret string // Service account password/secret
	Scope        string // OAuth2 scope (default: "api")
}

// GetPlatformAccessToken obtains an OAuth2 access token using client credentials flow.
// This function performs the CORRECT authentication for ISPSS API access:
//
// 1. Uses OAuth2 client credentials grant (grant_type=client_credentials)
// 2. Returns the ACCESS TOKEN (not ID token)
// 3. Access token contains authorization claims for API access
//
// Unlike the ARK SDK's IdentityServiceUser method which exchanges the access token
// for an ID token, this function returns the access token directly.
//
// Authentication Flow:
//   POST https://{identity_tenant_id}.id.cyberark.cloud/oauth2/platformtoken
//   Authorization: Basic base64(username:password)
//   Content-Type: application/x-www-form-urlencoded
//   Body: grant_type=client_credentials
//
// Response:
//   {
//     "access_token": "eyJ...",  // ← THIS is what we need for SIA API
//     "token_type": "Bearer",
//     "expires_in": 3600
//   }
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - config: OAuth2 configuration with identity URL, client ID, and secret
//
// Returns:
//   - *OAuth2TokenResponse: Token response with access_token
//   - error: Authentication failure or network error
func GetPlatformAccessToken(ctx context.Context, config *OAuth2Config) (*OAuth2TokenResponse, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth2 config cannot be nil")
	}
	if config.IdentityURL == "" {
		return nil, fmt.Errorf("identity_url is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}

	// Construct token endpoint URL
	// Using the platformtoken endpoint for ISPSS API access tokens
	// Format: https://{identity_tenant_id}.id.cyberark.cloud/oauth2/platformtoken
	tokenURL := fmt.Sprintf("%s/oauth2/platformtoken", strings.TrimSuffix(config.IdentityURL, "/"))

	// Prepare request body (application/x-www-form-urlencoded)
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	// Note: scope is not included - platformtoken endpoint doesn't require it

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 token request: %w", err)
	}

	// Set Basic Authentication header (username:password)
	req.SetBasicAuth(config.ClientID, config.ClientSecret)

	// Set required headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OAuth2 response: %w", err)
	}

	// Check for non-200 status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth2 authentication failed - [%d] - [%s]", resp.StatusCode, string(bodyBytes))
	}

	// Parse JSON response
	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth2 token response: %w", err)
	}

	// Validate access_token is present
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("OAuth2 response missing access_token")
	}

	return &tokenResp, nil
}

// ResolveSIAURLFromToken extracts the SIA API base URL from a JWT access token.
// The token contains claims like "subdomain" and "platform_domain" that are used
// to construct the SIA API URL in the format: https://{subdomain}.dpa.{domain}
//
// Parameters:
//   - accessToken: JWT access token from GetPlatformAccessToken()
//
// Returns:
//   - string: SIA API base URL (e.g., https://abc123.dpa.cyberark.cloud)
//   - error: If token parsing fails or required claims are missing
func ResolveSIAURLFromToken(accessToken string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("access token cannot be empty")
	}

	// Parse JWT token without verification (we trust it came from CyberArk)
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to extract JWT claims")
	}

	// Extract subdomain from token claims
	subdomain, ok := claims["subdomain"].(string)
	if !ok || subdomain == "" {
		return "", fmt.Errorf("JWT token missing 'subdomain' claim")
	}

	// Extract platform_domain from token claims
	platformDomain, ok := claims["platform_domain"].(string)
	if !ok || platformDomain == "" {
		return "", fmt.Errorf("JWT token missing 'platform_domain' claim")
	}

	// Remove "shell." prefix if present (SDK pattern from ark_isp_service_client.go:162-164)
	if strings.HasPrefix(platformDomain, "shell.") {
		platformDomain = strings.TrimPrefix(platformDomain, "shell.")
	}

	// Construct SIA API URL: https://{subdomain}.dpa.{platform_domain}
	// "dpa" is the service name for SIA (Dynamic Privileged Access)
	siaURL := fmt.Sprintf("https://%s.dpa.%s", subdomain, platformDomain)

	return siaURL, nil
}
