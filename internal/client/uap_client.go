// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/services/uap"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewUAPClient creates a new UAP API client using authenticated ARK SDK instance
//
// The returned *uap.ArkUAPAPI provides access to UAP services:
//   - Db() - Database access policy management (CRUD operations for SIA policies)
//   - Vm() - VM access policy management
//
// Usage in resources:
//
//	policy, err := uapAPI.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{PolicyID: id})
//	err = uapAPI.Db().UpdatePolicy(policy)
//
// See docs/policy-implementation-plan.md for UAP integration patterns.
func NewUAPClient(ctx context.Context, authCtx *ISPAuthContext) (*uap.ArkUAPAPI, error) {
	if authCtx == nil || authCtx.ISPAuth == nil {
		return nil, fmt.Errorf("auth context cannot be nil")
	}

	tflog.Debug(ctx, "Initializing UAP API client", map[string]interface{}{
		"service": "uap",
	})

	// Initialize UAP API client with authenticated ISP Auth
	// This client provides Db() access for policy database assignment operations
	uapAPI, err := uap.NewArkUAPAPI(authCtx.ISPAuth)
	if err != nil {
		tflog.Error(ctx, "Failed to initialize UAP API client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize UAP API client: %w", err)
	}

	tflog.Info(ctx, "UAP API client initialized successfully", map[string]interface{}{
		"services": "Db (policy management)",
	})

	return uapAPI, nil
}
