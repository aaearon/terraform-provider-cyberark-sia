// Package provider implements the CyberArk SIA Terraform provider
package provider

import (
	"context"
	"os"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/client"
	"github.com/cyberark/ark-sdk-golang/pkg/auth"
	"github.com/cyberark/ark-sdk-golang/pkg/services/sia"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var _ provider.Provider = &CyberArkSIAProvider{}

// CyberArkSIAProvider defines the provider implementation
type CyberArkSIAProvider struct {
	// version is set to the provider version on release
	version string
}

// CyberArkSIAProviderModel describes the provider data model
type CyberArkSIAProviderModel struct {
	ClientID                types.String `tfsdk:"client_id"`
	ClientSecret            types.String `tfsdk:"client_secret"`
	IdentityURL             types.String `tfsdk:"identity_url"`
	IdentityTenantSubdomain types.String `tfsdk:"identity_tenant_subdomain"`
}

// ProviderData holds the ARK SDK instances shared with resources
// This struct is passed to resources via resp.ResourceData in Configure()
type ProviderData struct {
	// ISPAuth handles authentication with CyberArk Identity Security Platform
	// Caching is enabled for automatic token refresh
	ISPAuth *auth.ArkISPAuth

	// SIAAPI provides access to SIA WorkspacesDB() and SecretsDB() APIs
	SIAAPI *sia.ArkSIAAPI
}

// New is a helper function to simplify provider server and testing implementation
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CyberArkSIAProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *CyberArkSIAProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cyberark_sia"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *CyberArkSIAProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for CyberArk Secure Infrastructure Access (SIA). " +
			"Manages database workspaces and secrets using the CyberArk ARK SDK.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "ISPSS service account client ID. Can also be set via CYBERARK_CLIENT_ID environment variable.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_secret": schema.StringAttribute{
				Description: "ISPSS service account client secret. Can also be set via CYBERARK_CLIENT_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"identity_url": schema.StringAttribute{
				Description: "CyberArk Identity tenant URL (e.g., https://abc123.cyberark.cloud). " +
					"Optional - only needed for GovCloud (https://abc123.cyberarkgov.cloud) or custom identity deployments. " +
					"If not provided, the URL is automatically resolved from identity_tenant_subdomain. " +
					"Can also be set via CYBERARK_IDENTITY_URL environment variable.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"identity_tenant_subdomain": schema.StringAttribute{
				Description: "CyberArk Identity tenant subdomain (e.g., 'abc123' from abc123.cyberark.cloud). " +
					"Required for constructing the service account username. " +
					"Can also be set via CYBERARK_TENANT_SUBDOMAIN environment variable.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

// Configure prepares a CyberArk SIA API client for data sources and resources
func (p *CyberArkSIAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CyberArkSIAProviderModel

	// Read configuration from Terraform
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from environment variables if not set in configuration
	clientID := getEnvOrConfig(config.ClientID.ValueString(), EnvClientID)
	clientSecret := getEnvOrConfig(config.ClientSecret.ValueString(), EnvClientSecret)
	identityURL := getEnvOrConfig(config.IdentityURL.ValueString(), EnvIdentityURL)
	tenantSubdomain := getEnvOrConfig(config.IdentityTenantSubdomain.ValueString(), EnvTenantSubdomain)

	// Validate required fields
	if clientID == "" {
		resp.Diagnostics.AddError(
			"Missing client_id",
			"client_id must be set in provider configuration or via CYBERARK_CLIENT_ID environment variable",
		)
	}
	if clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing client_secret",
			"client_secret must be set in provider configuration or via CYBERARK_CLIENT_SECRET environment variable",
		)
	}
	if tenantSubdomain == "" {
		resp.Diagnostics.AddError(
			"Missing identity_tenant_subdomain",
			"identity_tenant_subdomain must be set in provider configuration or via CYBERARK_TENANT_SUBDOMAIN environment variable",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Log configuration (without sensitive data)
	LogProviderConfig(ctx, &config)

	// Initialize authentication
	// Returns *auth.ArkISPAuth with caching enabled for automatic token refresh
	LogAuthStart(ctx)
	ispAuth, err := client.NewISPAuth(ctx, &client.AuthConfig{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		IdentityURL:             identityURL,
		IdentityTenantSubdomain: tenantSubdomain,
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "provider configuration"))
		return
	}
	LogAuthSuccess(ctx)

	// Initialize SIA API client
	// Returns *sia.ArkSIAAPI for WorkspacesDB() and SecretsDB() access
	LogSIAClientInit(ctx)
	siaAPI, err := client.NewSIAClient(ispAuth)
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "SIA client initialization"))
		return
	}
	LogSIAClientSuccess(ctx)

	// Create provider data for resource sharing
	providerData := &ProviderData{
		ISPAuth: ispAuth,
		SIAAPI:  siaAPI,
	}

	// Make provider data available to resources
	resp.ResourceData = providerData
}

// Resources defines the resources implemented in the provider
func (p *CyberArkSIAProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseWorkspaceResource,
		NewSecretResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *CyberArkSIAProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources not in scope for initial version
	}
}

// getEnvOrConfig returns config value if set, otherwise falls back to environment variable
func getEnvOrConfig(configValue string, envVar string) string {
	if configValue != "" {
		return configValue
	}
	return os.Getenv(envVar)
}
