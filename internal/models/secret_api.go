package models

// SecretAPI represents a database secret for SIA API operations
// Used for Create, Update, and Read operations (single struct with pointers for optional fields)
//
// Design Philosophy:
//   - ONE struct per resource (AWS SDK for Go pattern, not Java-style DTOs)
//   - Pointers for optional fields (distinguish "not set" vs "set to zero value")
//   - omitempty for JSON serialization (reduces payload size)
//
// Field Notes:
//   - ID, TenantID, CreatedTime, ModifiedTime, LastRotatedTime: Computed fields (read-only from API)
//   - Name, DatabaseID, Username: Required fields for Create (no pointers)
//   - Password: Write-only field (never returned by API for security - use pointer to detect "not provided")
//   - Description, Tags: Optional fields
type SecretAPI struct {
	// Computed fields (read-only from API)
	ID              *string `json:"id,omitempty"`
	TenantID        *string `json:"tenant_id,omitempty"`
	CreatedTime     *string `json:"created_time,omitempty"`
	ModifiedTime    *string `json:"modified_time,omitempty"`
	LastRotatedTime *string `json:"last_rotated_time,omitempty"`

	// Required fields (no pointers - must be set for Create)
	Name       string `json:"name"`
	DatabaseID string `json:"database_id"`
	Username   string `json:"username"`

	// Sensitive field (write-only - not returned by GET)
	Password *string `json:"password,omitempty"`

	// Optional fields
	Description *string           `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}
