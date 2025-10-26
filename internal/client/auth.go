// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/auth"
	authmodels "github.com/cyberark/ark-sdk-golang/pkg/models/auth"
	"github.com/cyberark/ark-sdk-golang/pkg/models"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username     string // Service account username in full format (e.g., "user@cyberark.cloud.12345")
	ClientSecret string // Service account password/secret
	IdentityURL  string // Optional - SDK auto-resolves from username if empty
}

// ISPAuthContext holds authentication state for re-use across operations
// This prevents filesystem profile loading and keyring caching
type ISPAuthContext struct {
	ISPAuth     *auth.ArkISPAuth
	Profile     *models.ArkProfile
	AuthProfile *authmodels.ArkAuthProfile
	Secret      *authmodels.ArkSecret
}

// NewISPAuth creates a new ARK SDK authentication client using IdentityServiceUser method
// CRITICAL: Creates an in-memory profile to bypass filesystem profile loading and keyring caching
// This prevents stale token issues that cause 401 errors on subsequent operations
func NewISPAuth(ctx context.Context, config *AuthConfig) (*ISPAuthContext, error) {
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

	// Initialize ARK SDK auth with caching DISABLED
	ispAuth := auth.NewArkISPAuth(false)

	// Create authentication profile using IdentityServiceUser method
	authProfile := &authmodels.ArkAuthProfile{
		Username:   config.Username,
		AuthMethod: authmodels.IdentityServiceUser,
		AuthMethodSettings: &authmodels.IdentityServiceUserArkAuthMethodSettings{
			IdentityURL:                      config.IdentityURL,
			IdentityTenantSubdomain:          "",
			IdentityAuthorizationApplication: "__idaptive_cybr_user_oidc",
		},
	}

	// Create secret
	secret := &authmodels.ArkSecret{
		Secret: config.ClientSecret,
	}

	// CRITICAL: Create in-memory ArkProfile to bypass filesystem profile loading
	// Passing this explicit profile prevents the SDK from loading ~/.ark/profiles/ and ~/.ark_cache
	inMemoryProfile := &models.ArkProfile{
		ProfileName:  "terraform-ephemeral", // Non-persisted name
		AuthProfiles: map[string]*authmodels.ArkAuthProfile{
			"isp": authProfile,
		},
	}

	// Authenticate with explicit in-memory profile (NOT nil)
	// This bypasses the default profile loading mechanism
	// force=true: Always get a fresh token (no cache lookups)
	_, err := ispAuth.Authenticate(inMemoryProfile, authProfile, secret, true, false)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Return context with all auth state for re-use in refresh callbacks
	return &ISPAuthContext{
		ISPAuth:     ispAuth.(*auth.ArkISPAuth),
		Profile:     inMemoryProfile,
		AuthProfile: authProfile,
		Secret:      secret,
	}, nil
}
