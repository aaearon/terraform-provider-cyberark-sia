package client

import (
	"context"

	"github.com/cyberark/ark-sdk-golang/pkg/services/identity"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewIdentityClient initializes the ARK SDK Identity API client for principal lookups
// Returns *identity.ArkIdentityAPI for UsersService() and DirectoriesService() access
func NewIdentityClient(ctx context.Context, authCtx *ISPAuthContext) (*identity.ArkIdentityAPI, error) {
	tflog.Debug(ctx, "Initializing Identity API client")

	identityAPI, err := identity.NewArkIdentityAPI(authCtx.ISPAuth)
	if err != nil {
		tflog.Error(ctx, "Failed to initialize Identity API client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	tflog.Debug(ctx, "Identity API client initialized successfully")
	return identityAPI, nil
}
