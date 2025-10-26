// Package provider implements the CyberArk SIA Terraform provider
package provider

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	Username     types.String `tfsdk:"username"`
	ClientSecret types.String `tfsdk:"client_secret"`
	IdentityURL  types.String `tfsdk:"identity_url"`
}

// ProviderData holds the ARK SDK instances shared with resources
// This struct is passed to resources via resp.ResourceData in Configure()
type ProviderData struct {
	// ISPAuth handles authentication with CyberArk Identity Security Platform
	// Caching is enabled for automatic token refresh
	ISPAuth *auth.ArkISPAuth

	// SIAAPI provides access to SIA WorkspacesDB() and SecretsDB() APIs
	SIAAPI *sia.ArkSIAAPI

	// CertificatesClient handles certificate CRUD operations (Phase 6 - User Story 4)
	// Uses OAuth2 access token directly (not ID token) to avoid 401 errors
	CertificatesClient *client.CertificatesClientOAuth2

	// SecretsClient handles secret CRUD operations using custom OAuth2 client
	// Replaces ARK SDK SecretsDB() API
	SecretsClient *client.SecretsClient

	// DatabaseWorkspaceClient handles database workspace CRUD operations using custom OAuth2 client
	// Replaces ARK SDK WorkspacesDB() API
	DatabaseWorkspaceClient *client.DatabaseWorkspaceClient
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
	resp.TypeName = "cyberarksia"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *CyberArkSIAProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for CyberArk Secure Infrastructure Access (SIA). " +
			"Manages database workspaces, certificates, and secrets using the CyberArk ARK SDK.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Service account username in full format (e.g., 'my-service-account@cyberark.cloud.12345'). " +
					"The tenant information is automatically extracted from the username by the ARK SDK. " +
					"Can also be set via CYBERARK_USERNAME environment variable.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_secret": schema.StringAttribute{
				Description: "Service account password/secret. " +
					"Can also be set via CYBERARK_CLIENT_SECRET environment variable.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"identity_url": schema.StringAttribute{
				Description: "CyberArk Identity tenant URL (e.g., https://abc123.cyberark.cloud). " +
					"OPTIONAL - only needed for GovCloud (https://abc123.cyberarkgov.cloud) or custom identity deployments. " +
					"If not provided, the URL is automatically resolved from the username by the ARK SDK. " +
					"Can also be set via CYBERARK_IDENTITY_URL environment variable.",
				Optional: true,
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
	username := getEnvOrConfig(config.Username.ValueString(), EnvUsername)
	clientSecret := getEnvOrConfig(config.ClientSecret.ValueString(), EnvClientSecret)
	identityURL := getEnvOrConfig(config.IdentityURL.ValueString(), EnvIdentityURL)

	// Validate required fields
	if username == "" {
		resp.Diagnostics.AddError(
			"Missing username",
			"username must be set in provider configuration or via CYBERARK_USERNAME environment variable",
		)
	}
	if clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing client_secret",
			"client_secret must be set in provider configuration or via CYBERARK_CLIENT_SECRET environment variable",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Log configuration (without sensitive data)
	tflog.Info(ctx, "Configuring CyberArk SIA provider", map[string]interface{}{
		"identity_url_provided": identityURL != "",
	})

	// Initialize authentication
	// Returns *auth.ArkISPAuth with caching enabled for automatic token refresh
	// Uses IdentityServiceUser method - SDK auto-resolves Identity URL from username if not provided
	tflog.Info(ctx, "Starting authentication")
	ispAuth, err := client.NewISPAuth(ctx, &client.AuthConfig{
		Username:     username,
		ClientSecret: clientSecret,
		IdentityURL:  identityURL, // Optional - SDK auto-resolves from username
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "provider configuration"))
		return
	}
	tflog.Info(ctx, "Authentication successful")

	// Initialize SIA API client
	// Returns *sia.ArkSIAAPI for WorkspacesDB() and SecretsDB() access
	tflog.Info(ctx, "Initializing SIA client")
	siaAPI, err := client.NewSIAClient(ispAuth)
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "SIA client initialization"))
		return
	}
	tflog.Info(ctx, "SIA client initialized successfully")

	// Initialize OAuth2-based Certificates Client
	// This uses the access token directly instead of the ID token to avoid 401 errors
	tflog.Info(ctx, "Initializing OAuth2 certificates client")
	// Use the identity URL from ISPAuth token endpoint if not explicitly provided
	if identityURL == "" {
		identityURL = ispAuth.Token.Endpoint
	}
	certsClient, err := initCertificatesClient(ctx, username, clientSecret, identityURL)
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "certificates client initialization"))
		return
	}
	tflog.Info(ctx, "Certificates client initialized successfully with OAuth2 access token")

	// Initialize OAuth2-based Secrets Client
	tflog.Info(ctx, "Initializing OAuth2 secrets client")
	secretsClient, err := initSecretsClient(ctx, username, clientSecret, identityURL)
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "secrets client initialization"))
		return
	}
	tflog.Info(ctx, "Secrets client initialized successfully with OAuth2 access token")

	// Initialize OAuth2-based Database Workspace Client
	tflog.Info(ctx, "Initializing OAuth2 database workspace client")
	dbWorkspaceClient, err := initDatabaseWorkspaceClient(ctx, username, clientSecret, identityURL)
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "database workspace client initialization"))
		return
	}
	tflog.Info(ctx, "Database workspace client initialized successfully with OAuth2 access token")

	// Create provider data for resource sharing
	providerData := &ProviderData{
		ISPAuth:                 ispAuth,
		SIAAPI:                  siaAPI,
		CertificatesClient:      certsClient,
		SecretsClient:           secretsClient,
		DatabaseWorkspaceClient: dbWorkspaceClient,
	}

	// Make provider data available to resources
	resp.ResourceData = providerData
}

// Resources defines the resources implemented in the provider
func (p *CyberArkSIAProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseWorkspaceResource,
		NewSecretResource,
		NewCertificateResource,
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

// initCertificatesClient initializes the OAuth2-based certificates client.
// This function obtains an OAuth2 access token and creates a client that uses
// the access token directly for API calls (not the ID token).
//
// This resolves the 401 Unauthorized errors caused by the ARK SDK's
// IdentityServiceUser method which exchanges the access token for an ID token.
func initCertificatesClient(ctx context.Context, username, clientSecret, identityURL string) (*client.CertificatesClientOAuth2, error) {
	// Step 1: Obtain OAuth2 access token using client credentials flow
	// POST /oauth2/platformtoken with client credentials
	tokenResp, err := client.GetPlatformAccessToken(ctx, &client.OAuth2Config{
		IdentityURL:  identityURL,
		ClientID:     username,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth2 access token: %w", err)
	}

	// Step 2: Resolve SIA API URL from access token JWT claims
	siaURL, err := client.ResolveSIAURLFromToken(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SIA URL from token: %w", err)
	}

	// Step 3: Create certificates client with access token
	certsClient, err := client.NewCertificatesClientOAuth2(siaURL, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificates client: %w", err)
	}

	return certsClient, nil
}

// initSecretsClient initializes the OAuth2-based secrets client using the generic RestClient.
func initSecretsClient(ctx context.Context, username, clientSecret, identityURL string) (*client.SecretsClient, error) {
	// Step 1: Obtain OAuth2 access token
	tokenResp, err := client.GetPlatformAccessToken(ctx, &client.OAuth2Config{
		IdentityURL:  identityURL,
		ClientID:     username,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth2 access token: %w", err)
	}

	// Step 2: Resolve SIA API URL from access token JWT claims
	siaURL, err := client.ResolveSIAURLFromToken(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SIA URL from token: %w", err)
	}

	// Step 3: Create generic RestClient
	restClient, err := client.NewRestClient(siaURL, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create RestClient: %w", err)
	}

	// Step 4: Create secrets client wrapper
	secretsClient := client.NewSecretsClient(restClient)

	return secretsClient, nil
}

// initDatabaseWorkspaceClient initializes the OAuth2-based database workspace client using the generic RestClient.
func initDatabaseWorkspaceClient(ctx context.Context, username, clientSecret, identityURL string) (*client.DatabaseWorkspaceClient, error) {
	// Step 1: Obtain OAuth2 access token
	tokenResp, err := client.GetPlatformAccessToken(ctx, &client.OAuth2Config{
		IdentityURL:  identityURL,
		ClientID:     username,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth2 access token: %w", err)
	}

	// Step 2: Resolve SIA API URL from access token JWT claims
	siaURL, err := client.ResolveSIAURLFromToken(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SIA URL from token: %w", err)
	}

	// Step 3: Create generic RestClient
	restClient, err := client.NewRestClient(siaURL, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create RestClient: %w", err)
	}

	// Step 4: Create database workspace client wrapper
	dbWorkspaceClient := client.NewDatabaseWorkspaceClient(restClient)

	return dbWorkspaceClient, nil
}
