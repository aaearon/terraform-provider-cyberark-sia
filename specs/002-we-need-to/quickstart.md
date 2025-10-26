# Quickstart: ARK SDK Replacement Implementation

**Feature**: Replace ARK SDK with Custom OAuth2 Implementation
**Date**: 2025-10-25
**Audience**: Developers implementing this feature

## Overview

This quickstart guide provides step-by-step instructions for replacing CyberArk ARK SDK usage with custom OAuth2 implementation. The migration eliminates 401 Unauthorized errors caused by the SDK's ID token/access token issue.

## Prerequisites

Before starting implementation:

1. ✅ **Understand the Problem**: Read `docs/oauth2-authentication-fix.md` for root cause analysis
2. ✅ **Review Reference Implementation**: Study `internal/client/certificates_oauth2.go` (proven working pattern)
3. ✅ **Review Data Models**: Study `specs/002-we-need-to/data-model.md` for model structure
4. ✅ **Review API Contracts**: Study `specs/002-we-need-to/contracts/` for API expectations
5. ✅ **Development Environment**: Go 1.25.0, Terraform CLI, access to CyberArk SIA test environment

## Phase 1: Database Workspace OAuth2 Client

**Goal**: Create custom HTTP client for database workspace operations

**Estimated Time**: 2-3 hours

### Step 1.1: Create Model Types

**File**: `internal/models/database_workspace.go` (NEW)

```go
package models

// DatabaseWorkspaceCreateRequest represents the request to create a database workspace
type DatabaseWorkspaceCreateRequest struct {
    Name                                 string            `json:"name"`
    ProviderEngine                       string            `json:"provider_engine"`
    NetworkName                          string            `json:"network_name,omitempty"`
    Platform                             string            `json:"platform,omitempty"`
    AuthDatabase                         string            `json:"auth_database,omitempty"`
    Services                             []string          `json:"services,omitempty"`
    Account                              string            `json:"account,omitempty"`
    ReadWriteEndpoint                    string            `json:"read_write_endpoint,omitempty"`
    ReadOnlyEndpoint                     string            `json:"read_only_endpoint,omitempty"`
    Port                                 int               `json:"port,omitempty"`
    SecretID                             string            `json:"secret_id,omitempty"`
    Tags                                 map[string]string `json:"tags,omitempty"`
    ConfiguredAuthMethodType             string            `json:"configured_auth_method_type,omitempty"`
    Region                               string            `json:"region,omitempty"`
    EnableCertificateValidation          bool              `json:"enable_certificate_validation,omitempty"`
    Certificate                          string            `json:"certificate,omitempty"`
    Domain                               string            `json:"domain,omitempty"`
    DomainControllerName                 string            `json:"domain_controller_name,omitempty"`
    DomainControllerNetbios              string            `json:"domain_controller_netbios,omitempty"`
    DomainControllerUseLDAPS             bool              `json:"domain_controller_use_ldaps,omitempty"`
    DomainControllerEnableCertValidation bool              `json:"domain_controller_enable_certificate_validation,omitempty"`
    DomainControllerLDAPSCertificate     string            `json:"domain_controller_ldaps_certificate,omitempty"`
}

// DatabaseWorkspaceUpdateRequest (alias for create request - all fields optional)
type DatabaseWorkspaceUpdateRequest = DatabaseWorkspaceCreateRequest

// DatabaseWorkspace represents a database workspace response
type DatabaseWorkspace struct {
    ID                                   string            `json:"id"`
    TenantID                             string            `json:"tenant_id,omitempty"`
    Name                                 string            `json:"name"`
    ProviderEngine                       string            `json:"provider_engine"`
    NetworkName                          string            `json:"network_name,omitempty"`
    Platform                             string            `json:"platform,omitempty"`
    AuthDatabase                         string            `json:"auth_database,omitempty"`
    Services                             []string          `json:"services,omitempty"`
    Account                              string            `json:"account,omitempty"`
    ReadWriteEndpoint                    string            `json:"read_write_endpoint,omitempty"`
    ReadOnlyEndpoint                     string            `json:"read_only_endpoint,omitempty"`
    Port                                 int               `json:"port,omitempty"`
    SecretID                             string            `json:"secret_id,omitempty"`
    Tags                                 map[string]string `json:"tags,omitempty"`
    ConfiguredAuthMethodType             string            `json:"configured_auth_method_type,omitempty"`
    Region                               string            `json:"region,omitempty"`
    EnableCertificateValidation          bool              `json:"enable_certificate_validation,omitempty"`
    Certificate                          string            `json:"certificate,omitempty"`
    Domain                               string            `json:"domain,omitempty"`
    DomainControllerName                 string            `json:"domain_controller_name,omitempty"`
    DomainControllerNetbios              string            `json:"domain_controller_netbios,omitempty"`
    DomainControllerUseLDAPS             bool              `json:"domain_controller_use_ldaps,omitempty"`
    DomainControllerEnableCertValidation bool              `json:"domain_controller_enable_certificate_validation,omitempty"`
    DomainControllerLDAPSCertificate     string            `json:"domain_controller_ldaps_certificate,omitempty"`
    CreatedTime                          string            `json:"created_time,omitempty"`
    ModifiedTime                         string            `json:"modified_time,omitempty"`
}
```

**Verification**:
```bash
go build ./internal/models/
```

### Step 1.2: Create OAuth2 Client

**File**: `internal/client/database_workspace_oauth2.go` (NEW)

**Pattern**: Follow `certificates_oauth2.go` structure

```go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "terraform-provider-cyberark-sia/internal/models"
)

// DatabaseWorkspaceClientOAuth2 manages database workspace operations using OAuth2 access tokens
type DatabaseWorkspaceClientOAuth2 struct {
    baseURL     string
    accessToken string
    httpClient  *http.Client
}

// NewDatabaseWorkspaceClientOAuth2 creates a new OAuth2-based database workspace client
func NewDatabaseWorkspaceClientOAuth2(baseURL, accessToken string) (*DatabaseWorkspaceClientOAuth2, error) {
    if baseURL == "" {
        return nil, fmt.Errorf("baseURL is required")
    }
    if accessToken == "" {
        return nil, fmt.Errorf("access token is required")
    }

    return &DatabaseWorkspaceClientOAuth2{
        baseURL:     strings.TrimSuffix(baseURL, "/"),
        accessToken: accessToken,
        httpClient:  &http.Client{Timeout: 30 * time.Second},
    }, nil
}

// CreateDatabaseWorkspace creates a new database workspace
func (c *DatabaseWorkspaceClientOAuth2) CreateDatabaseWorkspace(ctx context.Context, req *models.DatabaseWorkspaceCreateRequest) (*models.DatabaseWorkspace, error) {
    // 1. Serialize request to JSON
    // 2. Create HTTP POST request to /api/workspaces/db
    // 3. Set Authorization: Bearer {accessToken} header
    // 4. Execute request with retry logic
    // 5. Parse response JSON to DatabaseWorkspace
    // 6. Return workspace or error
}

// GetDatabaseWorkspace retrieves a database workspace by ID
func (c *DatabaseWorkspaceClientOAuth2) GetDatabaseWorkspace(ctx context.Context, id string) (*models.DatabaseWorkspace, error) {
    // Similar pattern to CreateDatabaseWorkspace
    // HTTP GET to /api/workspaces/db/{id}
}

// UpdateDatabaseWorkspace updates an existing database workspace
func (c *DatabaseWorkspaceClientOAuth2) UpdateDatabaseWorkspace(ctx context.Context, id string, req *models.DatabaseWorkspaceUpdateRequest) (*models.DatabaseWorkspace, error) {
    // HTTP PUT to /api/workspaces/db/{id}
}

// DeleteDatabaseWorkspace deletes a database workspace
func (c *DatabaseWorkspaceClientOAuth2) DeleteDatabaseWorkspace(ctx context.Context, id string) error {
    // HTTP DELETE to /api/workspaces/db/{id}
    // 204 No Content = success
}

// ListDatabaseWorkspaces lists all database workspaces
func (c *DatabaseWorkspaceClientOAuth2) ListDatabaseWorkspaces(ctx context.Context) ([]*models.DatabaseWorkspace, error) {
    // HTTP GET to /api/workspaces/db
}
```

**Implementation Tips**:
- Copy request/response handling from `certificates_oauth2.go`
- Use `client.RetryWithBackoff()` for HTTP requests
- Map HTTP status codes to errors using `client.MapError()`
- Never log access tokens or sensitive data

**Verification**:
```bash
go build ./internal/client/
```

### Step 1.3: Update Provider to Initialize Database Workspace Client

**File**: `internal/provider/provider.go` (MODIFY)

**Changes**:
1. Add field to `ProviderData` struct:
   ```go
   type ProviderData struct {
       ISPAuth                 *auth.ArkISPAuth                        // Keep temporarily
       SIAAPI                  *sia.ArkSIAAPI                          // Keep temporarily
       CertificatesClient      *client.CertificatesClientOAuth2        // Existing
       DatabaseWorkspaceClient *client.DatabaseWorkspaceClientOAuth2  // NEW
   }
   ```

2. Initialize in `Configure()` method (copy pattern from `initCertificatesClient`):
   ```go
   // Initialize OAuth2-based Database Workspace Client
   tflog.Info(ctx, "Initializing OAuth2 database workspace client")
   dbWorkspaceClient, err := initDatabaseWorkspaceClient(ctx, username, clientSecret, identityURL)
   if err != nil {
       resp.Diagnostics.Append(client.MapError(err, "database workspace client initialization"))
       return
   }
   resp.ResourceData = &ProviderData{
       // ... existing fields ...
       DatabaseWorkspaceClient: dbWorkspaceClient,
   }
   ```

3. Add helper function:
   ```go
   func initDatabaseWorkspaceClient(ctx context.Context, username, clientSecret, identityURL string) (*client.DatabaseWorkspaceClientOAuth2, error) {
       // Same pattern as initCertificatesClient
       // 1. Get OAuth2 access token
       // 2. Resolve SIA URL from token
       // 3. Create client with access token
   }
   ```

**Verification**:
```bash
go build .
terraform init
terraform plan  # Should initialize provider without errors
```

### Step 1.4: Update Database Workspace Resource

**File**: `internal/provider/resource_database_workspace.go` (MODIFY)

**Changes**:
1. Update struct to use OAuth2 client:
   ```go
   type DatabaseWorkspaceResource struct {
       providerData           *ProviderData
       databaseWorkspaceAPI   *client.DatabaseWorkspaceClientOAuth2  // Changed from ARK SDK type
   }
   ```

2. Update `Configure()` method:
   ```go
   func (r *DatabaseWorkspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
       // ... existing validation ...
       r.databaseWorkspaceAPI = providerData.DatabaseWorkspaceClient  // Changed
   }
   ```

3. Update CRUD methods to use custom models:
   ```go
   func (r *DatabaseWorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
       // Replace: dbmodels.AddDatabaseRequest
       // With: models.DatabaseWorkspaceCreateRequest

       // Replace: siaAPI.WorkspacesDB().AddDatabase()
       // With: r.databaseWorkspaceAPI.CreateDatabaseWorkspace()
   }
   ```

4. Remove ARK SDK imports:
   ```go
   // DELETE: import dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
   // ADD: import "terraform-provider-cyberark-sia/internal/models"
   ```

**Verification**:
```bash
go build .
TF_ACC=1 go test ./internal/provider/ -v -run TestAccDatabaseWorkspaceResource
```

### Step 1.5: Test Database Workspace Operations

**Test Plan**:
1. Create database workspace
2. Read database workspace (verify no drift)
3. Update database workspace
4. Delete database workspace

**Test Configuration**:
```hcl
provider "cyberarksia" {
  username      = var.cyberark_username
  client_secret = var.cyberark_client_secret
}

resource "cyberarksia_database_workspace" "test" {
  name                          = "oauth2-test-db"
  database_type                 = "postgres"
  address                       = "postgres.example.com"
  port                          = 5432
  enable_certificate_validation = true
  tags = {
    test = "oauth2_migration"
  }
}
```

**Expected Result**: ✅ All operations succeed with HTTP 200/201 responses (no 401 errors)

---

## Phase 2: Secret OAuth2 Client

**Goal**: Create custom HTTP client for secret operations

**Estimated Time**: 2-3 hours

### Step 2.1: Create Secret Model Types

**File**: `internal/models/secret.go` (NEW)

```go
package models

// SecretCreateRequest represents the request to create a database secret
type SecretCreateRequest struct {
    Name        string            `json:"name"`
    DatabaseID  string            `json:"database_id"`
    Username    string            `json:"username"`
    Password    string            `json:"password"`
    Description string            `json:"description,omitempty"`
    Tags        map[string]string `json:"tags,omitempty"`
}

// SecretUpdateRequest (alias for create request - all fields optional)
type SecretUpdateRequest = SecretCreateRequest

// Secret represents a database secret response
type Secret struct {
    ID              string            `json:"id"`
    TenantID        string            `json:"tenant_id,omitempty"`
    Name            string            `json:"name"`
    DatabaseID      string            `json:"database_id"`
    Username        string            `json:"username"`
    Password        string            `json:"password,omitempty"` // Write-only (not returned by GET)
    Description     string            `json:"description,omitempty"`
    Tags            map[string]string `json:"tags,omitempty"`
    CreatedTime     string            `json:"created_time,omitempty"`
    ModifiedTime    string            `json:"modified_time,omitempty"`
    LastRotatedTime string            `json:"last_rotated_time,omitempty"`
}
```

### Step 2.2: Create Secrets OAuth2 Client

**File**: `internal/client/secrets_oauth2.go` (NEW)

**Pattern**: Copy from `database_workspace_oauth2.go` and adjust for secrets

```go
package client

// SecretsClientOAuth2 manages secret operations using OAuth2 access tokens
type SecretsClientOAuth2 struct {
    baseURL     string
    accessToken string
    httpClient  *http.Client
}

// Methods: CreateSecret, GetSecret, UpdateSecret, DeleteSecret, ListSecrets
```

### Step 2.3: Update Provider to Initialize Secrets Client

**File**: `internal/provider/provider.go` (MODIFY)

**Changes**:
1. Add `SecretsClient *client.SecretsClientOAuth2` to `ProviderData`
2. Initialize in `Configure()` method
3. Add `initSecretsClient()` helper function

### Step 2.4: Update Secret Resource

**File**: `internal/provider/resource_secret.go` (MODIFY)

**Changes**:
1. Replace ARK SDK client with `SecretsClientOAuth2`
2. Replace ARK SDK models with custom models
3. Remove ARK SDK imports

### Step 2.5: Test Secret Operations

**Test Configuration**:
```hcl
resource "cyberarksia_secret" "test" {
  name         = "oauth2-test-secret"
  database_id  = cyberarksia_database_workspace.test.id
  username     = "test_user"
  password     = "Test@1234"
  description  = "Test secret for OAuth2 migration"
}
```

---

## Phase 3: ARK SDK Removal

**Goal**: Completely remove ARK SDK dependencies

**Estimated Time**: 1-2 hours

### Step 3.1: Update Database Engine Validator

**File**: `internal/validators/database_engine_validator.go` (MODIFY)

**Changes**:
1. Remove ARK SDK import
2. Define engine types locally or use string constants

### Step 3.2: Delete ARK SDK Wrapper Files

**Files to Delete**:
- `internal/client/auth.go`
- `internal/client/certificates.go`
- `internal/client/sia_client.go`

**Verification**:
```bash
rg "ark-sdk-golang" internal/  # Should find zero imports
```

### Step 3.3: Remove ARK SDK from Provider Data

**File**: `internal/provider/provider.go` (MODIFY)

**Changes**:
1. Remove `ISPAuth` and `SIAAPI` fields from `ProviderData`
2. Remove ARK SDK initialization code
3. Remove ARK SDK imports

### Step 3.4: Clean Dependencies

**Commands**:
```bash
go mod tidy
rg "ark-sdk-golang" go.mod  # Should return no results
```

### Step 3.5: Full Test Suite

**Commands**:
```bash
go test ./...
TF_ACC=1 go test ./internal/provider/ -v
```

**Expected Result**: ✅ All tests pass with zero ARK SDK dependencies

---

## Phase 4: Documentation Updates

**Goal**: Update documentation to reflect custom OAuth2 implementation

**Estimated Time**: 1 hour

### Step 4.1: Update oauth2-authentication-fix.md

**File**: `docs/oauth2-authentication-fix.md` (MODIFY)

**Changes**:
- Update "Current State" section: Mark database workspace and secret resources as FIXED ✅
- Update "Migration Plan" section: Mark all phases as COMPLETE
- Add "Completion Summary" section documenting final state

### Step 4.2: Deprecate sdk-integration.md

**File**: `docs/sdk-integration.md` (MODIFY)

**Changes**:
- Add deprecation notice at top
- Note that custom OAuth2 implementation replaced ARK SDK
- Keep for historical reference only

### Step 4.3: Update CLAUDE.md

**File**: `CLAUDE.md` (MODIFY)

**Changes**:
- Remove ARK SDK integration patterns
- Document custom OAuth2 client patterns
- Update dependency list (remove ARK SDK)

---

## Common Issues and Solutions

### Issue 1: 401 Unauthorized Errors

**Symptom**: API requests fail with 401 after migration

**Solution**:
- Verify access token (not ID token) is being used
- Check Authorization header: `Authorization: Bearer {access_token}`
- Confirm token hasn't expired (3600s lifetime)

### Issue 2: JSON Serialization Errors

**Symptom**: Request body doesn't match API expectations

**Solution**:
- Verify JSON struct tags match API field names (snake_case)
- Check for missing `omitempty` on optional fields
- Compare serialized JSON with API contract in `contracts/` directory

### Issue 3: Import Cycle Errors

**Symptom**: Go build fails with "import cycle not allowed"

**Solution**:
- Ensure `internal/models/` doesn't import `internal/provider/`
- Keep models independent (only JSON serialization, no business logic)

### Issue 4: Provider Initialization Fails

**Symptom**: Terraform init/plan fails to configure provider

**Solution**:
- Check OAuth2 token acquisition logs
- Verify service account credentials are valid
- Confirm identity URL is correct

---

## Rollback Plan

If critical issues discovered during implementation:

1. **Phase 1 Rollback**: Revert `resource_database_workspace.go` to use ARK SDK
2. **Phase 2 Rollback**: Revert `resource_secret.go` to use ARK SDK
3. **Phase 3 Rollback**: Keep ARK SDK files and dependencies

**Note**: Certificate resource remains on OAuth2 (already proven working)

---

## Success Criteria Checklist

- [ ] Database workspace CRUD operations work without 401 errors
- [ ] Secret CRUD operations work without 401 errors
- [ ] Certificate operations continue working (no regression)
- [ ] `rg "ark-sdk-golang"` returns zero results in code (excluding docs)
- [ ] `go mod tidy` removes all ARK SDK dependencies
- [ ] All unit tests pass
- [ ] All acceptance tests pass
- [ ] Documentation updated

---

## Next Steps After Completion

1. Tag release as `v1.0.0` (breaking internal change)
2. Update `CHANGELOG.md` with migration details
3. Update provider documentation with OAuth2 authentication details
4. Monitor production usage for any unexpected errors

---

## Resources

- **OAuth2 Fix Documentation**: `docs/oauth2-authentication-fix.md`
- **Reference Implementation**: `internal/client/certificates_oauth2.go`
- **Data Models**: `specs/002-we-need-to/data-model.md`
- **API Contracts**: `specs/002-we-need-to/contracts/`
- **OAuth2 RFC**: https://datatracker.ietf.org/doc/html/rfc6749#section-4.4
