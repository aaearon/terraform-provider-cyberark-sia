package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// SecretModel represents a secret resource in Terraform.
// Maps to cyberark_sia_secret resource.
// Secrets are standalone credentials that can be referenced by database workspaces, VM workspaces, etc.
type SecretModel struct {
	// Computed attributes
	ID           types.String `tfsdk:"id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	LastModified types.String `tfsdk:"last_modified"`

	// Required attributes
	Name                types.String `tfsdk:"name"`
	DatabaseWorkspaceID types.String `tfsdk:"database_workspace_id"`
	AuthenticationType  types.String `tfsdk:"authentication_type"`

	// Conditional attributes (based on authentication_type)
	// Local/Domain authentication
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"` // Sensitive

	// Domain authentication (Active Directory)
	Domain types.String `tfsdk:"domain"`

	// AWS IAM authentication
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`     // Sensitive
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"` // Sensitive

	// Rotation settings
	RotationEnabled      types.Bool  `tfsdk:"rotation_enabled"`
	RotationIntervalDays types.Int64 `tfsdk:"rotation_interval_days"`
}
