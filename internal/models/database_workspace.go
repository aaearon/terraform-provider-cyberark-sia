// Package models provides data structures for Terraform resources
package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatabaseWorkspaceModel represents a database workspace resource in Terraform state.
// This model maps to the cyberark_sia_database_workspace resource schema.
// Note: SIA API uses integer IDs internally, we store as string in Terraform for consistency
type DatabaseWorkspaceModel struct {
	// Core Identifiers
	ID   types.String `tfsdk:"id"` // Stored as string, converted to int for API calls
	Name types.String `tfsdk:"name"`

	// Database Configuration
	DatabaseType types.String `tfsdk:"database_type"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	AuthDatabase types.String `tfsdk:"auth_database"` // MongoDB authentication database (SDK: AuthDatabase)
	Services     types.List   `tfsdk:"services"`      // Oracle/SQL Server services (SDK: Services)
	Account      types.String `tfsdk:"account"`       // Snowflake/Atlas account name (SDK: Account)

	// Network & Endpoints
	NetworkName      types.String `tfsdk:"network_name"`       // Network segmentation (SDK: NetworkName)
	ReadOnlyEndpoint types.String `tfsdk:"read_only_endpoint"` // Read replica endpoint (SDK: ReadOnlyEndpoint)

	// Authentication
	AuthenticationMethod types.String `tfsdk:"authentication_method"`

	// Secret Integration
	SecretID types.String `tfsdk:"secret_id"` // Reference to secret (cyberark_sia_secret) for ZSP/JIT

	// TLS/Certificate Management
	EnableCertificateValidation types.Bool   `tfsdk:"enable_certificate_validation"` // Enforce TLS cert validation (SDK: EnableCertificateValidation, default: true)
	CertificateID               types.String `tfsdk:"certificate_id"`                // Certificate ID for TLS/mTLS (SDK: Certificate) - will reference cyberark_sia_certificate resource

	// Cloud Provider Metadata
	CloudProvider types.String `tfsdk:"cloud_provider"`
	Region        types.String `tfsdk:"region"` // Used for AWS RDS IAM authentication

	// Metadata
	LastModified types.String `tfsdk:"last_modified"`
}
