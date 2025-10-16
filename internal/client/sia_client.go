// Package client provides CyberArk SIA API client wrappers
package client

import (
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/auth"
	"github.com/cyberark/ark-sdk-golang/pkg/services/sia"
)

// NewSIAClient creates a new SIA API client using authenticated ARK SDK instance
//
// The returned *sia.ArkSIAAPI provides access to SIA services:
//   - WorkspacesDB() - Database target management (CRUD operations)
//   - SecretsDB() - Database secrets/strong accounts management (CRUD operations)
//
// Usage in resources:
//
//	database, err := siaAPI.WorkspacesDB().AddDatabase(&dbmodels.ArkSIADBAddDatabase{...})
//	secret, err := siaAPI.SecretsDB().AddSecret(&secretsmodels.ArkSIADBAddSecret{...})
//
// See docs/sdk-integration.md for detailed SDK integration patterns and examples.
func NewSIAClient(ispAuth *auth.ArkISPAuth) (*sia.ArkSIAAPI, error) {
	if ispAuth == nil {
		return nil, fmt.Errorf("ispAuth cannot be nil")
	}

	// Initialize SIA API client with authenticated ISP Auth
	// This client provides WorkspacesDB() and SecretsDB() access for Phase 3+ resources
	siaAPI, err := sia.NewArkSIAAPI(ispAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SIA API client: %w", err)
	}

	return siaAPI, nil
}
