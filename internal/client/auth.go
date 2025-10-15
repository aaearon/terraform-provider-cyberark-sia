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
	ClientID                string
	ClientSecret            string
	IdentityURL             string
	IdentityTenantSubdomain string
}

// NewISPAuth creates a new ARK SDK authentication client
// Caching is enabled for automatic token refresh
func NewISPAuth(ctx context.Context, config *AuthConfig) (*auth.ArkISPAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("auth config cannot be nil")
	}

	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	if config.IdentityURL == "" {
		return nil, fmt.Errorf("identity_url is required")
	}
	if config.IdentityTenantSubdomain == "" {
		return nil, fmt.Errorf("identity_tenant_subdomain is required")
	}

	// Initialize ARK SDK auth with caching enabled for automatic token refresh
	ispAuth := auth.NewArkISPAuth(true)

	// Create authentication profile
	profile := &authmodels.ArkAuthProfile{
		Username:   fmt.Sprintf("%s@cyberark.cloud.%s", config.ClientID, config.IdentityTenantSubdomain),
		AuthMethod: authmodels.Identity,
		AuthMethodSettings: &authmodels.IdentityArkAuthMethodSettings{
			IdentityURL:             config.IdentityURL,
			IdentityTenantSubdomain: config.IdentityTenantSubdomain,
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
