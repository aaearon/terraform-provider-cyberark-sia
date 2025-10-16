# SIA API Contract

**Version**: 1.0
**Date**: 2025-10-15
**Purpose**: Define expected API interactions between Terraform provider and CyberArk SIA REST API via ARK SDK

## Overview

This contract specifies the API operations required by the Terraform provider. All operations are performed via the CyberArk ARK SDK for Golang, which abstracts the underlying REST API calls.

---

## Authentication Operations

### 1. Acquire Bearer Token

**SDK Method**: `ispAuth.Authenticate()`

**Purpose**: Obtain bearer token from ISPSS platformtoken endpoint for SIA API authentication

**Request**:
```go
ispAuth := auth.NewArkISPAuth(false) // Disable internal caching
token, err := ispAuth.Authenticate(
    nil, // profile (use default)
    &authmodels.ArkAuthProfile{
        Username:           "{client_id}@cyberark.cloud.{tenant_id}",
        AuthMethod:         authmodels.Identity,
        AuthMethodSettings: &authmodels.IdentityArkAuthMethodSettings{
            IdentityURL:             "{identity_url}",
            IdentityTenantSubdomain: "{tenant_subdomain}",
        },
    },
    &authmodels.ArkSecret{
        Secret: "{client_secret}",
    },
    false, // force - do not force new auth if token exists
    false, // refreshAuth - do not attempt refresh
)
```

**Response Success**:
```go
type ArkAuthToken struct {
    Token     string        // Bearer token value
    ExpiresIn time.Duration // 15 minutes typically
    TokenType string        // "Bearer"
}
```

**Response Errors**:
- **401 Unauthorized**: Invalid client_id or client_secret
- **403 Forbidden**: Service account lacks required role memberships
- **5xx Server Error**: ISPSS service unavailable

**Provider Action on Error**:
- Return Terraform diagnostic with actionable guidance
- Do not retry authentication failures (user config issue)
- Retry 5xx errors up to `max_retries` times

---

### 2. Refresh Bearer Token

**SDK Method**: `ispAuth.Authenticate()` with `refreshAuth=true`

**Purpose**: Proactively refresh token before expiration

**Request**:
```go
newToken, err := ispAuth.Authenticate(
    nil,
    &authmodels.ArkAuthProfile{/* same as acquire */},
    &authmodels.ArkSecret{Secret: "{client_secret}"},
    false, // force
    true,  // refreshAuth - attempt to refresh existing token
)
```

**Response**: Same as Acquire Bearer Token

**Provider Behavior** (Updated based on research):
- **Enable SDK caching**: `NewArkISPAuth(true)` - SDK handles automatic refresh
- SDK manages token lifecycle internally when caching is enabled
- No custom refresh goroutine needed - SDK is thread-safe
- Log authentication events at TRACE level

---

## Database Target Operations

**Note**: Database operations use `sia.NewArkSIAAPI(ispAuth)` for access to WorkspacesDB() and SecretsDB() APIs.

### 3. Create Database Target

**Purpose**: Register existing database with SIA

**SDK Method**: `siaAPI.WorkspacesDB().AddDatabase()`

**Package Imports**:
```go
import (
    "github.com/cyberark/ark-sdk-golang/pkg/services/sia"
    dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
)
```

**Request Model** (actual ARK SDK structure):
```go
// Using ARK SDK models directly
database, err := siaAPI.WorkspacesDB().AddDatabase(&dbmodels.ArkSIADBAddDatabase{
    Name:              "MyDatabase",
    ProviderEngine:    dbmodels.EngineTypeAuroraMysql, // or other engine types
    ReadWriteEndpoint: "database.example.com:5432",
    SecretID:          secretID, // Reference to strong account secret
    // Additional fields as needed per SIA requirements
})
```

**Response Success**:
```go
// ARK SDK returns database model with ID and metadata
// Exact structure depends on ARK SDK dbmodels.ArkSIADatabase type
type DatabaseResponse struct {
    ID           string
    Name         string
    // ... additional fields from ARK SDK response
}
```

**Response Errors**:
- **400 Bad Request**: Invalid database type or version
- **409 Conflict**: Database target with same name already exists
- **422 Unprocessable**: Validation error (e.g., version below minimum)
- **5xx Server Error**: SIA service unavailable

**Provider Action on Error**:
- ARK SDK returns Go error with descriptive message
- Map SDK error to Terraform diagnostic with actionable guidance
- Retry logic for transient failures (provider implements retry wrapper)

**Note**: ARK SDK abstracts HTTP status codes - provider relies on SDK error messages for diagnostics.

---

### 4. Read Database Target

**Purpose**: Retrieve current state of database target from SIA

**SDK Method**: `siaAPI.WorkspacesDB().GetDatabase(id)`

**Request**:
```go
database, err := siaAPI.WorkspacesDB().GetDatabase(databaseID)
if err != nil {
    // SDK error handling - descriptive messages included
    resp.Diagnostics.AddError("Failed to Read Database Target", err.Error())
    return
}
```

**Response Success**: Same as `DatabaseTargetResponse` in Create operation

**Response Errors**:
- **404 Not Found**: Resource deleted outside Terraform
- **5xx Server Error**: SIA service unavailable

**Provider Action on Error**:
- 404: Remove from Terraform state (drift detection)
- 5xx: Retry, then return diagnostic

---

### 5. Update Database Target

**Purpose**: Modify existing database target configuration

**SDK Method**: `siaAPI.WorkspacesDB().UpdateDatabase(id, updates)`

**Request**:
```go
// Update using ARK SDK models
updatedDB, err := siaAPI.WorkspacesDB().UpdateDatabase(
    databaseID,
    &dbmodels.ArkSIADBUpdateDatabase{
        // Only changed fields needed
        Name:              updatedName,
        ReadWriteEndpoint: updatedEndpoint,
    },
)
```

**Response Success**: Same as `DatabaseTargetResponse`

**Response Errors**:
- **404 Not Found**: Resource deleted
- **422 Unprocessable**: Invalid configuration change
- **5xx Server Error**: SIA service unavailable

**Provider Action**:
- Handle partial updates (only send changed fields)
- 404: Remove from state
- 422: Validation diagnostic

---

### 6. Delete Database Target

**Purpose**: Remove database target from SIA

**SDK Method**: `siaAPI.WorkspacesDB().DeleteDatabase(id)`

**Request**:
```go
err := siaAPI.WorkspacesDB().DeleteDatabase(databaseID)
if err != nil {
    // Check if resource already deleted (treat as success)
    // ARK SDK error message will indicate if resource not found
    resp.Diagnostics.AddError("Failed to Delete Database Target", err.Error())
    return
}
```

**Provider Action**:
- Parse SDK error message to determine if resource already deleted
- Handle dependency conflicts gracefully with clear diagnostic
- Retry transient failures per provider retry logic

---

## Strong Account Operations

**Note**: Strong accounts are managed as secrets in SIA using the SecretsDB() API.

### 7. Create Strong Account

**Purpose**: Create credentials for SIA to provision ephemeral access

**SDK Method**: `siaAPI.SecretsDB().AddSecret()`

**Package Imports**:
```go
import (
    dbsecretsmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/secrets/db/models"
)
```

**Request Model** (actual ARK SDK structure):
```go
// Create secret using ARK SDK models
secret, err := siaAPI.SecretsDB().AddSecret(&dbsecretsmodels.ArkSIADBAddSecret{
    SecretType: "username_password", // or other types for aws_iam
    Username:   "admin_user",
    Password:   "strong_password",
    // Additional fields per authentication type
})
```

**Response Success**:
```go
// ARK SDK returns secret model with ID
// Note: Sensitive fields (password) NOT returned in response
type SecretResponse struct {
    ID        string
    SecretID  string // Reference ID for use with database targets
    // ... metadata fields, no sensitive data
}
```

**Response Errors**:
- **400 Bad Request**: Invalid authentication type for target
- **404 Not Found**: Database target ID doesn't exist
- **422 Unprocessable**: Missing required credential fields
- **5xx Server Error**: SIA service unavailable

**Provider Action**:
- Validate credential fields match authentication type before API call
- Never log password/secret key values
- 404: Return diagnostic - target must exist before strong account
- 422: Map to credential field validation error

---

### 8. Read Strong Account

**Purpose**: Retrieve strong account metadata (not credentials)

**SDK Method**: `siaAPI.SecretsDB().GetSecret(id)`

**Request**:
```go
secret, err := siaAPI.SecretsDB().GetSecret(secretID)
if err != nil {
    resp.Diagnostics.AddError("Failed to Read Strong Account", err.Error())
    return
}
// Response contains metadata only, no sensitive credentials
```

**Drift Detection**:
- Terraform detects drift in non-sensitive metadata
- Sensitive fields (password, keys) NOT returned by SDK - handled via ignore_changes strategy

---

### 9. Update Strong Account

**Purpose**: Modify strong account configuration or rotate credentials

**SDK Method**: `siaAPI.SecretsDB().UpdateSecret(id, updates)`

**Request**:
```go
// Update secret using ARK SDK models
updatedSecret, err := siaAPI.SecretsDB().UpdateSecret(
    secretID,
    &dbsecretsmodels.ArkSIADBUpdateSecret{
        // Metadata updates
        // Credential rotation (password/keys if changed)
        Password: newPassword, // Sensitive - only if rotating
    },
)
```

**Special Behavior**:
- SIA updates credentials immediately per FR-015a
- New sessions use new credentials
- ARK SDK handles secure transmission of sensitive updates

---

### 10. Delete Strong Account

**Purpose**: Remove strong account from SIA

**SDK Method**: `siaAPI.SecretsDB().DeleteSecret(id)`

**Request**:
```go
err := siaAPI.SecretsDB().DeleteSecret(secretID)
if err != nil {
    // Parse error to determine if already deleted
    resp.Diagnostics.AddError("Failed to Delete Strong Account", err.Error())
    return
}
```

**Provider Action**:
- Parse SDK error for resource-not-found scenarios (treat as success)
- Handle dependency conflicts with clear diagnostics

---

## API Client Requirements

### Base Client Configuration

**Updated based on research findings**:

```go
// Provider data structure (shared with resources)
type ProviderData struct {
    ispAuth   *auth.ArkISPAuth  // Initialized with caching enabled
    siaAPI    *sia.ArkSIAAPI    // For WorkspacesDB() and SecretsDB() operations
    // Optional: uapAPI for policies (future scope)
}

// Provider Configure method
func (p *CyberArkSIAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Enable SDK caching - handles token refresh automatically
    ispAuth := auth.NewArkISPAuth(true)

    // Authenticate
    _, err := ispAuth.Authenticate(nil, profile, secret, false, false)
    if err != nil {
        resp.Diagnostics.AddError("Authentication Failed", err.Error())
        return
    }

    // Initialize SIA API client
    siaAPI, err := sia.NewArkSIAAPI(ispAuth.(*auth.ArkISPAuth))
    if err != nil {
        resp.Diagnostics.AddError("Failed to Initialize SIA Client", err.Error())
        return
    }

    // Make available to resources
    resp.ResourceData = &ProviderData{
        ispAuth: ispAuth,
        siaAPI:  siaAPI,
    }
}
```

### Retry Logic

**Retryable Errors**:
- HTTP 5xx (server errors)
- Network timeout errors
- Connection refused

**Non-Retryable Errors**:
- HTTP 4xx (client errors, except 429 rate limit)
- Authentication failures (401, 403)

**Retry Strategy**:
- Exponential backoff: `delay = base_delay * 2^attempt` (max 30s)
- Max retries from provider config (`max_retries`, default 3)
- Log each retry attempt at DEBUG level

### Concurrent Request Handling

**Updated based on SDK caching research**:

- Multiple resources may call API simultaneously during `terraform apply`
- ARK SDK handles token refresh internally when caching enabled - thread-safe
- No custom mutex needed - SDK manages concurrent access
- Provider creates one `ispAuth` and `siaAPI` instance shared across all resources via `resp.ResourceData`

### Logging Requirements

**Log Levels**:
- TRACE: Authentication flows, token refresh events, request/response bodies (no secrets)
- DEBUG: API operation start/completion, retry attempts
- INFO: Resource lifecycle events (create, update, delete success)
- WARN: Validation warnings, rate limit encountered
- ERROR: API failures, authentication errors

**Security**:
- NEVER log: `client_secret`, `password`, `aws_secret_access_key`, bearer tokens
- Mask sensitive values if logging request payloads

---

## Error Response Format

**Updated based on research findings**:

ARK SDK abstracts HTTP status codes and returns standard Go `error` types with descriptive messages. Provider should rely on SDK error messages rather than parsing HTTP status codes.

**Provider Error Handling Pattern**:

```go
result, err := siaAPI.WorkspacesDB().AddDatabase(...)
if err != nil {
    // ARK SDK error messages are already descriptive
    // Parse error message for common patterns
    errorMsg := err.Error()

    if strings.Contains(errorMsg, "authentication failed") {
        resp.Diagnostics.AddError(
            "Authentication Failed",
            "Invalid client_id or client_secret. Verify provider configuration.",
        )
        return
    }

    if strings.Contains(errorMsg, "already exists") {
        resp.Diagnostics.AddError(
            "Database Target Already Exists",
            fmt.Sprintf("Use `terraform import` to manage existing resource: %s", errorMsg),
        )
        return
    }

    // Generic error with SDK message
    resp.Diagnostics.AddError(
        "Failed to Create Database Target",
        fmt.Sprintf("SIA API error: %s", errorMsg),
    )
    return
}
```

**Error Categories** (based on SDK message patterns):
- Authentication errors: "authentication failed", "invalid credentials"
- Permission errors: "insufficient permissions", "forbidden"
- Validation errors: Include field-specific details in message
- Resource not found: "not found", "does not exist"
- Conflict errors: "already exists", "duplicate"
- Service errors: Connection failures, timeouts

---

## API Versioning

**Current API Version**: TBD (determined by SIA instance)

**Version Handling**:
- Provider will query SIA API version during Configure()
- Warn if unsupported version detected
- Future: Schema version upgrades via Terraform state migration

---

## Testing Mocks

**Mock Strategy for Unit Tests**:

```go
type MockSIAClient struct {
    CreateDatabaseTargetFunc func(req *CreateDatabaseTargetRequest) (*DatabaseTargetResponse, error)
    GetDatabaseTargetFunc    func(id string) (*DatabaseTargetResponse, error)
    // ... other operations
}

func (m *MockSIAClient) CreateDatabaseTarget(req *CreateDatabaseTargetRequest) (*DatabaseTargetResponse, error) {
    if m.CreateDatabaseTargetFunc != nil {
        return m.CreateDatabaseTargetFunc(req)
    }
    // Default mock behavior
    return &DatabaseTargetResponse{ID: "mock-id", Name: req.Name}, nil
}
```

**Mock Scenarios**:
- Success responses for happy path tests
- Specific error codes for error handling tests
- Concurrent request scenarios for token refresh tests

---

## Conclusion

This contract defines the interface between the Terraform provider and SIA REST API via ARK SDK. Key requirements (updated with concrete SDK methods):

1. **Authentication**: ISPSS OAuth2 via ARK SDK with **automatic token refresh** when caching enabled (`NewArkISPAuth(true)`)
2. **CRUD Operations**:
   - Database targets via `siaAPI.WorkspacesDB()` (AddDatabase, GetDatabase, UpdateDatabase, DeleteDatabase)
   - Strong accounts via `siaAPI.SecretsDB()` (AddSecret, GetSecret, UpdateSecret, DeleteSecret)
3. **Concurrency Safety**: ARK SDK handles thread-safe token management - no custom mutex needed
4. **Error Mapping**: Trust SDK error messages - parse for common patterns, provide actionable diagnostics
5. **Logging**: Structured logging without sensitive data exposure

**Key Changes from Initial Contract**:
- **Simplified token management**: SDK internal caching instead of custom refresh goroutine
- **Concrete API methods**: Replaced placeholders with actual `sia.NewArkSIAAPI()` methods
- **Error handling**: Parse SDK error messages instead of HTTP status codes
- **Provider data pattern**: Use `resp.ResourceData` to share `ispAuth` and `siaAPI` with resources

**Next Steps**: Implement `internal/provider/` resources following this updated contract specification.
