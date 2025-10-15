// Package provider implements the CyberArk SIA Terraform provider
package provider

// TestEnvVars documents the environment variables required for acceptance tests
// These variables must be set when running acceptance tests (TF_ACC=1)
const (
	// TF_ACC must be set to "1" to enable acceptance tests
	EnvTFAcc = "TF_ACC"

	// CYBERARK_CLIENT_ID is the ISPSS service account client ID
	EnvClientID = "CYBERARK_CLIENT_ID"

	// CYBERARK_CLIENT_SECRET is the ISPSS service account client secret
	EnvClientSecret = "CYBERARK_CLIENT_SECRET"

	// CYBERARK_IDENTITY_URL is the CyberArk Identity tenant URL
	// Example: https://example.cyberark.cloud
	EnvIdentityURL = "CYBERARK_IDENTITY_URL"

	// CYBERARK_TENANT_SUBDOMAIN is the CyberArk Identity tenant subdomain
	// Example: example (from example.cyberark.cloud)
	EnvTenantSubdomain = "CYBERARK_TENANT_SUBDOMAIN"
)

// TestAccPreCheckVars lists the required environment variables for acceptance tests
var TestAccPreCheckVars = []string{
	EnvClientID,
	EnvClientSecret,
	EnvIdentityURL,
	EnvTenantSubdomain,
}
