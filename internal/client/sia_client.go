// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/services/sia"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
func NewSIAClient(ctx context.Context, authCtx *ISPAuthContext) (*sia.ArkSIAAPI, error) {
	if authCtx == nil || authCtx.ISPAuth == nil {
		return nil, fmt.Errorf("auth context cannot be nil")
	}

	tflog.Debug(ctx, "Initializing SIA API client", map[string]interface{}{
		"service": "sia",
	})

	// Initialize SIA API client with authenticated ISP Auth
	// This client provides WorkspacesDB() and SecretsDB() access for Phase 3+ resources
	siaAPI, err := sia.NewArkSIAAPI(authCtx.ISPAuth)
	if err != nil {
		tflog.Error(ctx, "Failed to initialize SIA API client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize SIA API client: %w", err)
	}

	tflog.Info(ctx, "SIA API client initialized successfully", map[string]interface{}{
		"services": "WorkspacesDB, SecretsDB",
	})

	return siaAPI, nil
}
