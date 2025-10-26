package models

// DatabaseWorkspaceAPI represents a database workspace for SIA API operations
// Used for Create, Update, and Read operations (single struct with pointers for optional fields)
//
// Design Philosophy:
//   - ONE struct per resource (AWS SDK for Go pattern, not Java-style DTOs)
//   - Pointers for optional fields (distinguish "not set" vs "set to zero value")
//   - omitempty for JSON serialization (reduces payload size)
//   - MVP FIRST: Start with 5 core fields, add more incrementally based on need
//
// MVP Phase (5 fields):
//   - Name, ProviderEngine: Required fields
//   - ReadWriteEndpoint, Port, Tags: Optional fields
//
// Future Phases: Add certificate, cloud provider, specialized DB fields incrementally
type DatabaseWorkspaceAPI struct {
	// Computed fields (read-only from API)
	ID           *string `json:"id,omitempty"`
	TenantID     *string `json:"tenant_id,omitempty"`
	CreatedTime  *string `json:"created_time,omitempty"`
	ModifiedTime *string `json:"modified_time,omitempty"`

	// Required fields (no pointers - must be set for Create)
	Name           string `json:"name"`
	ProviderEngine string `json:"provider_engine"`

	// MVP Optional fields (core database connection)
	ReadWriteEndpoint *string           `json:"read_write_endpoint,omitempty"`
	Port              *int              `json:"port,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`

	// Phase 2: Certificate support (add when needed)
	// EnableCertificateValidation *bool   `json:"enable_certificate_validation,omitempty"`
	// Certificate                 *string `json:"certificate,omitempty"`

	// Phase 3: Cloud provider support (add when needed)
	// Platform *string `json:"platform,omitempty"` // aws, azure, gcp, on_premise, atlas
	// Region   *string `json:"region,omitempty"`   // Required for RDS IAM auth

	// Phase 4: Specialized DB fields (add incrementally)
	// AuthDatabase *string `json:"auth_database,omitempty"` // MongoDB
	// Services []string `json:"services,omitempty"` // Oracle/SQL Server
	// Account *string `json:"account,omitempty"` // Snowflake/Atlas
	// NetworkName *string `json:"network_name,omitempty"`
	// ReadOnlyEndpoint *string `json:"read_only_endpoint,omitempty"`
	// ConfiguredAuthMethodType *string `json:"configured_auth_method_type,omitempty"`
	// SecretID *string `json:"secret_id,omitempty"`

	// Phase 5: Active Directory integration (add if requested)
	// Domain, DomainController fields...
}
