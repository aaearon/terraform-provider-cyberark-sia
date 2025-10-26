# Data Model: Custom SIA API Models (Simplified)

**Feature**: Replace ARK SDK with Custom OAuth2 Implementation
**Date**: 2025-10-25
**Status**: Design Revised (see design-reflection.md)

## Design Philosophy

**Key Principle**: Use **ONE struct per resource** with pointers for optional fields (AWS SDK for Go pattern).

**Why This Approach**:
- Eliminates separate CreateRequest/UpdateRequest/Response types
- Simplifies maintenance (single struct to update)
- Uses `omitempty` for optional JSON fields
- Uses **pointers** (`*string`, `*int`, `*bool`) to distinguish "not set" vs "set to zero value"
- Standard Go SDK pattern (not Java-style DTOs)

## Implementation Strategy

### MVP (Minimum Viable Product) First
1. **Phase 1**: Implement core fields only (name, provider_engine, endpoint, port, tags)
2. **Phase 2**: Add certificate support fields
3. **Phase 3**: Add cloud provider fields
4. **Phase 4+**: Add specialized DB fields incrementally (MongoDB, Oracle, Snowflake, AD)

**Rationale**: Don't try to achieve 1:1 API coverage from day one. Start simple, iterate based on actual usage.

---

## Secret Model

### Secret (Single Struct)

**Purpose**: Used for Create, Update, and Read operations

**Location**: `internal/models/secret.go`

**Fields**:

```go
type Secret struct {
    // Computed fields (read-only from API)
    ID              *string `json:"id,omitempty"`
    TenantID        *string `json:"tenant_id,omitempty"`
    CreatedTime     *string `json:"created_time,omitempty"`
    ModifiedTime    *string `json:"modified_time,omitempty"`
    LastRotatedTime *string `json:"last_rotated_time,omitempty"`

    // Required fields (no pointers)
    Name       string `json:"name"`
    DatabaseID string `json:"database_id"`
    Username   string `json:"username"`

    // Sensitive field (write-only - not returned by GET)
    Password *string `json:"password,omitempty"`

    // Optional fields
    Description *string            `json:"description,omitempty"`
    Tags        map[string]string  `json:"tags,omitempty"`
}
```

**Usage Patterns**:

```go
// Create secret
secret := &Secret{
    Name:       "app-user-secret",
    DatabaseID: "1234567890",
    Username:   "app_user",
    Password:   StringPtr("s3cr3t"),
    Description: StringPtr("Application user credentials"),
}

// Update password only (partial update)
update := &Secret{
    Password: StringPtr("newS3cr3t"),
}

// Read response (password is nil - never returned)
// secret.Password == nil
```

**Field Notes**:
- **Password**: Write-only field. API never returns it for security. Use pointer to detect "not provided" vs "empty string".
- **Computed fields** (ID, timestamps): Always nil on Create, populated by API on response
- **Required fields**: No pointers (always must be set for Create)

**Helper Function** (add to package):
```go
func StringPtr(s string) *string { return &s }
func IntPtr(i int) *int { return &i }
func BoolPtr(b bool) *bool { return &b }
```

---

## DatabaseWorkspace Model

### DatabaseWorkspace (Single Struct - MVP)

**Purpose**: Used for Create, Update, and Read operations

**Location**: `internal/models/database_workspace.go`

**Phase 1: MVP Fields** (Core functionality):

```go
type DatabaseWorkspace struct {
    // Computed fields (read-only from API)
    ID           *string `json:"id,omitempty"`
    TenantID     *string `json:"tenant_id,omitempty"`
    CreatedTime  *string `json:"created_time,omitempty"`
    ModifiedTime *string `json:"modified_time,omitempty"`

    // Required fields (no pointers)
    Name           string `json:"name"`
    ProviderEngine string `json:"provider_engine"`

    // MVP Optional fields (core database connection)
    ReadWriteEndpoint *string           `json:"read_write_endpoint,omitempty"`
    Port              *int              `json:"port,omitempty"`
    Tags              map[string]string `json:"tags,omitempty"`
}
```

**Phase 2: Certificate Fields** (Add after MVP working):

```go
    // Certificate support
    EnableCertificateValidation *bool   `json:"enable_certificate_validation,omitempty"`
    Certificate                 *string `json:"certificate,omitempty"`
```

**Phase 3: Cloud Provider Fields** (Add when needed):

```go
    // Cloud provider support
    Platform *string `json:"platform,omitempty"` // aws, azure, gcp, on_premise, atlas
    Region   *string `json:"region,omitempty"`   // Required for RDS IAM auth
```

**Phase 4: Specialized DB Fields** (Add incrementally):

```go
    // MongoDB-specific
    AuthDatabase *string `json:"auth_database,omitempty"` // Default: "admin"

    // Oracle/SQL Server-specific
    Services []string `json:"services,omitempty"`

    // Snowflake/Atlas-specific
    Account *string `json:"account,omitempty"`

    // Network segmentation
    NetworkName *string `json:"network_name,omitempty"` // Default: "ON-PREMISE"

    // Read replicas
    ReadOnlyEndpoint *string `json:"read_only_endpoint,omitempty"`

    // Authentication methods
    ConfiguredAuthMethodType *string `json:"configured_auth_method_type,omitempty"`
    SecretID                 *string `json:"secret_id,omitempty"` // For ZSP/JIT
```

**Phase 5: Active Directory Fields** (Add if requested):

```go
    // Active Directory integration
    Domain                               *string `json:"domain,omitempty"`
    DomainControllerName                 *string `json:"domain_controller_name,omitempty"`
    DomainControllerNetbios              *string `json:"domain_controller_netbios,omitempty"`
    DomainControllerUseLDAPS             *bool   `json:"domain_controller_use_ldaps,omitempty"`
    DomainControllerEnableCertValidation *bool   `json:"domain_controller_enable_certificate_validation,omitempty"`
    DomainControllerLDAPSCertificate     *string `json:"domain_controller_ldaps_certificate,omitempty"`
```

**Usage Patterns**:

```go
// Create minimal PostgreSQL workspace (MVP)
workspace := &DatabaseWorkspace{
    Name:              "customers-db",
    ProviderEngine:    "postgres",
    ReadWriteEndpoint: StringPtr("postgres.example.com"),
    Port:              IntPtr(5432),
}

// Update endpoint only (partial update)
update := &DatabaseWorkspace{
    ReadWriteEndpoint: StringPtr("postgres-new.example.com"),
}

// Add certificate support (Phase 2)
workspace.Certificate = StringPtr("cert-id-123")
workspace.EnableCertificateValidation = BoolPtr(true)
```

---

## Incremental Field Addition Strategy

### When to Add Fields

Add new fields **only when**:
1. **User requests** the feature (real need, not speculation)
2. **Testing requires** the field (e.g., testing MongoDB needs `auth_database`)
3. **Critical functionality** blocked without it

### How to Add Fields

1. **One PR per feature area**:
   - PR #1: Certificate support (2 fields)
   - PR #2: Cloud provider (2 fields)
   - PR #3: MongoDB support (1 field)
   - PR #4: Oracle support (1 field)
   - etc.

2. **Always add**:
   - Field to struct with pointer type
   - Tests for the new field
   - Example in `examples/` directory
   - Documentation update

3. **Never add**:
   - "Just in case" fields
   - Unused SDK fields without clear use case

---

## Generic REST Client (Shared HTTP Logic)

### RestClient (Eliminates Duplication)

**Purpose**: Shared HTTP client handling requests, responses, errors, retry logic

**Location**: `internal/client/rest_client.go` (NEW)

**Design**:

```go
type RestClient struct {
    HTTPClient *http.Client
    BaseURL    string
    Token      string
}

func NewRestClient(baseURL, token string) *RestClient {
    return &RestClient{
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
        BaseURL:    strings.TrimSuffix(baseURL, "/"),
        Token:      token,
    }
}

// DoRequest handles all HTTP boilerplate
func (c *RestClient) DoRequest(
    ctx context.Context,
    method string,    // "POST", "GET", "PUT", "DELETE"
    path string,      // "/api/workspaces/db", "/api/secrets/db/{id}"
    body interface{}, // Request body (marshaled to JSON)
    responseData interface{}, // Response body (unmarshaled from JSON)
) error {
    // 1. Create HTTP request
    // 2. Marshal body to JSON if not nil
    // 3. Set headers (Authorization: Bearer {token}, Content-Type: application/json)
    // 4. Execute with RetryWithBackoff() for transient errors
    // 5. Handle non-2xx status codes (map to Terraform diagnostics)
    // 6. Unmarshal response JSON to responseData
    return nil
}
```

**Benefits**:
- ~90% of HTTP code shared between all resources
- Consistent error handling
- Centralized retry logic
- Easy to add logging/metrics in one place

---

## Resource-Specific Clients (Thin Wrappers)

### SecretsClient (Uses RestClient)

**Location**: `internal/client/secrets_client.go` (NEW)

**Design**:

```go
type SecretsClient struct {
    RestClient *RestClient
}

func NewSecretsClient(restClient *RestClient) *SecretsClient {
    return &SecretsClient{RestClient: restClient}
}

func (c *SecretsClient) Create(ctx context.Context, secret *Secret) (*Secret, error) {
    var response Secret
    err := c.RestClient.DoRequest(ctx, "POST", "/api/secrets/db", secret, &response)
    return &response, err
}

func (c *SecretsClient) Get(ctx context.Context, id string) (*Secret, error) {
    var response Secret
    path := fmt.Sprintf("/api/secrets/db/%s", id)
    err := c.RestClient.DoRequest(ctx, "GET", path, nil, &response)
    return &response, err
}

func (c *SecretsClient) Update(ctx context.Context, id string, secret *Secret) (*Secret, error) {
    var response Secret
    path := fmt.Sprintf("/api/secrets/db/%s", id)
    err := c.RestClient.DoRequest(ctx, "PUT", path, secret, &response)
    return &response, err
}

func (c *SecretsClient) Delete(ctx context.Context, id string) error {
    path := fmt.Sprintf("/api/secrets/db/%s", id)
    return c.RestClient.DoRequest(ctx, "DELETE", path, nil, nil)
}
```

**Code Size**: ~50 lines (vs ~200 lines without generic client)

### DatabaseWorkspaceClient (Uses RestClient)

**Location**: `internal/client/database_workspace_client.go` (NEW)

**Design**: Identical structure to SecretsClient, different paths

```go
type DatabaseWorkspaceClient struct {
    RestClient *RestClient
}

// Create, Get, Update, Delete - same pattern as SecretsClient
```

**Code Size**: ~50 lines (vs ~200 lines without generic client)

---

## Migration from ARK SDK Models

| ARK SDK Type | Custom Model Type | Simplification |
|---------------|-------------------|----------------|
| `dbmodels.ArkSIADBAddDatabase` | `DatabaseWorkspace` (same struct) | 3 types → 1 type |
| `dbmodels.ArkSIADBUpdateDatabase` | `DatabaseWorkspace` (same struct) | |
| `dbmodels.ArkSIADBDatabase` | `DatabaseWorkspace` (same struct) | |
| `secretsmodels.ArkSIADBSecretRequest` | `Secret` (same struct) | 3 types → 1 type |
| `secretsmodels.ArkSIADBSecret` | `Secret` (same struct) | |

**Total Reduction**: 6 ARK SDK types → 2 custom types (-67%)

---

## Validation Rules

### Secret

- **Required**: `name`, `database_id`, `username`, `password` (on Create)
- **Optional**: All other fields

### DatabaseWorkspace (MVP)

- **Required**: `name`, `provider_engine`
- **Optional**: All other fields
- **Constraints** (enforced by Terraform validators, not model):
  - `provider_engine`: Must be valid engine type
  - `port`: 1-65535 if provided
  - `region`: Required if `configured_auth_method_type` = "rds_iam_authentication"

---

## Testing Strategy

### Unit Tests
- JSON serialization round-trips
- Pointer handling (nil vs zero value)
- Partial updates (only changed fields)

### Integration Tests
- RestClient.DoRequest() with mock HTTP server
- Error mapping and retry logic

### Acceptance Tests
- Full CRUD lifecycle with real API
- MVP fields only initially
- Add tests as fields are added incrementally

---

## Summary of Simplifications

| Aspect | Original Design | Simplified Design | Reduction |
|--------|-----------------|-------------------|-----------|
| **Model Types** | 6 (3 per resource) | 2 (1 per resource) | **-67%** |
| **DatabaseWorkspace Fields (MVP)** | 25 fields | 5 fields | **-80%** |
| **Client Code per Resource** | ~200 lines (duplicated) | ~50 lines (wrapper) | **-75%** |
| **Shared Client Code** | 0 lines | ~100 lines | **Reused** |

**Key Insight**: "The goal is to replace legacy SDK logic, not to build a new, all-encompassing SDK within the provider." - Gemini

---

**Design Revised**: 2025-10-25
**Reviewer**: Gemini 2.5 Pro
**Status**: ✅ Simplified design approved - Ready for implementation
