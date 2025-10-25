// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/auth"
	authmodels "github.com/cyberark/ark-sdk-golang/pkg/models/auth"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username     string // Service account username in full format (e.g., "user@cyberark.cloud.12345")
	ClientSecret string // Service account password/secret
	IdentityURL  string // Optional - SDK auto-resolves from username if empty
}

// NewISPAuth creates a new ARK SDK authentication client using IdentityServiceUser method
// Caching is enabled for automatic token refresh
// The SDK automatically extracts tenant information from the username and resolves the Identity URL
func NewISPAuth(ctx context.Context, config *AuthConfig) (*auth.ArkISPAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("auth config cannot be nil")
	}

	// Validate required fields
	if config.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	// Note: IdentityURL is optional - SDK will resolve it from username if empty

	// Initialize ARK SDK auth with caching enabled for automatic token refresh
	ispAuth := auth.NewArkISPAuth(true)

	// Create authentication profile using IdentityServiceUser method for service accounts
	// This uses OAuth 2.0 client credentials flow (not interactive user auth)
	profile := &authmodels.ArkAuthProfile{
		Username:   config.Username, // Full username - SDK extracts tenant from @suffix
		AuthMethod: authmodels.IdentityServiceUser,
		AuthMethodSettings: &authmodels.IdentityServiceUserArkAuthMethodSettings{
			IdentityURL:                      config.IdentityURL,          // Optional - SDK auto-resolves from username
			IdentityTenantSubdomain:          "",                          // Empty - SDK extracts from username
			IdentityAuthorizationApplication: "__idaptive_cybr_user_oidc", // SDK default OAuth app
		},
	}

	// Create secret
	secret := &authmodels.ArkSecret{
		Secret: config.ClientSecret,
	}

	// Authenticate and get initial token
	// Parameters: profile (optional, nil for default), authProfile, secret, force, refreshAuth
	// force=false: don't force new auth if token exists
	// refreshAuth=false: don't attempt refresh (getting initial token)
	// Note: ARK SDK v1.5.0 Authenticate() does not accept context.Context as first parameter
	// The first parameter is an optional *ArkProfile (nil uses default profile)
	_, err := ispAuth.Authenticate(nil, profile, secret, false, false)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return ispAuth.(*auth.ArkISPAuth), nil
}
