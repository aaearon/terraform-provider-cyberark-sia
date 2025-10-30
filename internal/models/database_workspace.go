// Package models provides data structures for Terraform resources
package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatabaseWorkspaceModel represents a database workspace resource in Terraform state.
// This model maps to the cyberark_sia_database_workspace resource schema.
// Note: SIA API uses integer IDs internally, we store as string in Terraform for consistency
type DatabaseWorkspaceModel struct {
	// Oracle/SQL Server services (SDK: Services)
	Services types.List `tfsdk:"services"` // 24 bytes (slice)

	// Optional metadata - Key-value tags for organization
	Tags types.Map `tfsdk:"tags"` // 8 bytes (map)

	// Port configuration
	Port types.Int64 `tfsdk:"port"` // 8 bytes

	// TLS/Certificate Management - Enforce TLS cert validation (SDK: EnableCertificateValidation, default: true)
	EnableCertificateValidation types.Bool `tfsdk:"enable_certificate_validation"` // 1 byte

	// Core Identifiers - Stored as string, converted to int for API calls
	ID   types.String `tfsdk:"id"`   // types.String
	Name types.String `tfsdk:"name"` // types.String

	// Database Configuration
	DatabaseType types.String `tfsdk:"database_type"` // types.String
	Address      types.String `tfsdk:"address"`       // types.String
	AuthDatabase types.String `tfsdk:"auth_database"` // types.String - MongoDB authentication database (SDK: AuthDatabase)
	Account      types.String `tfsdk:"account"`       // types.String - Snowflake/Atlas account name (SDK: Account)

	// Network & Endpoints
	NetworkName      types.String `tfsdk:"network_name"`       // types.String - Network segmentation (SDK: NetworkName)
	ReadOnlyEndpoint types.String `tfsdk:"read_only_endpoint"` // types.String - Read replica endpoint (SDK: ReadOnlyEndpoint)

	// Authentication
	AuthenticationMethod types.String `tfsdk:"authentication_method"` // types.String

	// Secret Integration - Reference to secret (cyberark_sia_secret) for ZSP/JIT
	SecretID types.String `tfsdk:"secret_id"` // types.String

	// TLS/Certificate Management - Certificate ID for TLS/mTLS (SDK: Certificate) - will reference cyberark_sia_certificate resource
	CertificateID types.String `tfsdk:"certificate_id"` // types.String

	// Cloud Provider Metadata
	CloudProvider types.String `tfsdk:"cloud_provider"` // types.String
	Region        types.String `tfsdk:"region"`         // types.String - Used for AWS RDS IAM authentication

	// Computed attributes
	LastModified types.String `tfsdk:"last_modified"` // types.String
}
