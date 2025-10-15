// Package models provides data structures for Terraform resources
package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatabaseTargetModel represents a database target resource in Terraform state.
// This model maps to the cyberark_sia_database_target resource schema.
// Note: SIA API uses integer IDs internally, we store as string in Terraform for consistency
type DatabaseTargetModel struct {
	// Core Identifiers
	ID   types.String `tfsdk:"id"` // Stored as string, converted to int for API calls
	Name types.String `tfsdk:"name"`

	// Database Configuration
	DatabaseType    types.String `tfsdk:"database_type"`
	DatabaseVersion types.String `tfsdk:"database_version"`
	Address         types.String `tfsdk:"address"`
	Port            types.Int64  `tfsdk:"port"`
	DatabaseName    types.String `tfsdk:"database_name"`

	// Authentication
	AuthenticationMethod types.String `tfsdk:"authentication_method"`

	// Cloud Provider Metadata
	CloudProvider        types.String `tfsdk:"cloud_provider"`
	AWSRegion            types.String `tfsdk:"aws_region"`
	AWSAccountID         types.String `tfsdk:"aws_account_id"`
	AzureTenantID        types.String `tfsdk:"azure_tenant_id"`
	AzureSubscriptionID  types.String `tfsdk:"azure_subscription_id"`

	// Metadata
	Description  types.String `tfsdk:"description"`
	Tags         types.Map    `tfsdk:"tags"`
	LastModified types.String `tfsdk:"last_modified"`
}
