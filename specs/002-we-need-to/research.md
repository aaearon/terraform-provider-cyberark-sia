# Research: ARK SDK Replacement with Custom OAuth2

**Feature**: Replace ARK SDK with Custom OAuth2 Implementation
**Date**: 2025-10-25
**Status**: Complete

## Overview

This document consolidates research findings for replacing the CyberArk ARK SDK (`github.com/cyberark/ark-sdk-golang`) with custom OAuth2 implementation. The research leverages existing working implementations and documented API patterns.

## Research Tasks

### 1. OAuth2 Authentication Pattern (RESOLVED)

**Research Question**: What OAuth2 flow is required for CyberArk SIA API authentication using service accounts?

**Decision**: Use OAuth2 Client Credentials Grant Flow with `/oauth2/platformtoken` endpoint

**Rationale**:
- Certificate resource already implements this pattern successfully (zero 401 errors)
- ARK SDK's ID token exchange is the root cause of authentication failures
- Access tokens contain required API authorization claims (scope, aud, permissions)
- Standard OAuth2 pattern widely documented and supported

**Reference Implementation**: `internal/client/oauth2.go` (existing)

**Key Implementation Details**:
```go
// POST https://{identity_tenant_id}.id.cyberark.cloud/oauth2/platformtoken
// Authorization: Basic base64(username:password)
// Content-Type: application/x-www-form-urlencoded
// Body: grant_type=client_credentials

// Response:
{
  "access_token": "eyJ...",  // JWT with 3600s expiration
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Alternatives Considered**:
- ❌ **Patching ARK SDK**: Requires maintaining fork, unsustainable long-term
- ❌ **Wrapper around ARK SDK**: Fragile, still depends on SDK internals
- ✅ **Direct OAuth2 implementation**: Clean separation, full control, proven pattern

**Sources**:
- `docs/oauth2-authentication-fix.md` (comprehensive analysis of ARK SDK flaw)
- `internal/client/oauth2.go` (working implementation)
- OAuth2 RFC 6749 Section 4.4 (Client Credentials Grant)

---

### 2. SIA API Endpoints and Contracts (RESOLVED)

**Research Question**: What are the exact API contracts for database workspace and secret operations?

**Decision**: Use REST API endpoints with JSON payloads matching ARK SDK model structures

**Database Workspaces API**:
- **Create**: `POST /api/workspaces/db` → Returns workspace ID
- **Read**: `GET /api/workspaces/db/{id}` → Returns full workspace details
- **Update**: `PUT /api/workspaces/db/{id}` → Accepts partial updates
- **Delete**: `DELETE /api/workspaces/db/{id}` → Returns 204 No Content
- **List**: `GET /api/workspaces/db` → Returns array of workspaces

**Secrets API**:
- **Create**: `POST /api/secrets/db` → Returns secret ID
- **Read**: `GET /api/secrets/db/{id}` → Returns secret metadata (not password)
- **Update**: `PUT /api/secrets/db/{id}` → Accepts password rotation
- **Delete**: `DELETE /api/secrets/db/{id}` → Returns 204 No Content

**Request/Response Patterns**:
- All requests require `Authorization: Bearer {access_token}` header
- Content-Type: `application/json`
- Error responses include HTTP status codes + JSON error body
- Snake_case JSON field naming (following Go struct tags)

**ARK SDK Model Mapping**:
- `dbmodels.AddDatabaseRequest` → Custom `DatabaseWorkspaceCreateRequest`
- `dbmodels.Database` → Custom `DatabaseWorkspace`
- `secretsmodels.SecretRequest` → Custom `SecretCreateRequest`
- `secretsmodels.Secret` → Custom `Secret`

**Rationale**:
- ARK SDK source code provides definitive API contract reference
- Certificate resource proves HTTP client pattern works for SIA API
- Custom models eliminate ARK SDK dependency while preserving API contract

**Sources**:
- ARK SDK source: `/pkg/services/sia/workspaces/db/` (API implementation reference)
- ARK SDK source: `/pkg/services/sia/secrets/db/` (API implementation reference)
- `docs/oauth2-authentication-fix.md` (endpoint documentation)
- `CLAUDE.md` (Database Workspace Field Mappings table)

---

### 3. HTTP Client Best Practices for Terraform Providers (RESOLVED)

**Research Question**: What are HashiCorp's recommended patterns for custom HTTP clients in Terraform providers?

**Decision**: Follow Terraform Plugin Framework HTTP client patterns with context, retry logic, and error mapping

**Best Practices**:

1. **Context Propagation**:
   - All HTTP requests MUST accept `context.Context` parameter
   - Use `http.NewRequestWithContext(ctx, ...)` for cancellation support
   - Respect Terraform timeouts and user interrupts

2. **Error Handling**:
   - Map HTTP status codes to Terraform diagnostics
   - Provide actionable error messages with guidance
   - Distinguish between retryable (network, 429, 500) and non-retryable (401, 404) errors

3. **Retry Logic**:
   - Implement exponential backoff for transient errors
   - Limit retry attempts (3 attempts recommended)
   - Use existing `client.RetryWithBackoff()` helper

4. **Sensitive Data**:
   - Never log access tokens, passwords, or client secrets
   - Mark sensitive fields in Terraform schema with `Sensitive: true`
   - Sanitize error messages to exclude credentials

5. **Connection Management**:
   - Reuse `http.Client` instances (connection pooling)
   - Set reasonable timeouts (30 seconds default)
   - Handle concurrent requests safely (OAuth2 client thread-safety)

**Reference Implementation**: `internal/client/certificates_oauth2.go`

**Rationale**:
- Certificate OAuth2 client follows these patterns successfully
- Terraform Plugin Framework documentation recommends context-aware operations
- Industry-standard HTTP client patterns (connection pooling, timeouts, retries)

**Sources**:
- HashiCorp Terraform Plugin Framework Documentation
- `internal/client/certificates_oauth2.go` (reference implementation)
- `internal/client/retry.go` (retry logic implementation)
- `internal/client/errors.go` (error mapping implementation)

---

### 4. Model Type Design Patterns (RESOLVED)

**Research Question**: How should custom model types be structured to replace ARK SDK models?

**Decision**: Define Go structs with JSON tags matching SIA API field names (snake_case)

**Model Structure Pattern**:

```go
// Request models for API input
type DatabaseWorkspaceCreateRequest struct {
    Name                         string            `json:"name"`
    ProviderEngine               string            `json:"provider_engine"`
    ReadWriteEndpoint            string            `json:"read_write_endpoint,omitempty"`
    Port                         int               `json:"port,omitempty"`
    NetworkName                  string            `json:"network_name,omitempty"`
    // ... additional fields with omitempty for optional values
}

// Response models for API output
type DatabaseWorkspace struct {
    ID                           string            `json:"id"`
    Name                         string            `json:"name"`
    ProviderEngine               string            `json:"provider_engine"`
    // ... all fields from API response
    CreatedTime                  string            `json:"created_time,omitempty"`
    ModifiedTime                 string            `json:"modified_time,omitempty"`
}
```

**Key Design Decisions**:

1. **Separate Request/Response Models**:
   - Create: `*CreateRequest` → `*Resource`
   - Update: `*UpdateRequest` → `*Resource`
   - Read: No request body → `*Resource`
   - Delete: No request/response bodies

2. **Field Naming**:
   - JSON tags use snake_case (API convention)
   - Go field names use PascalCase (Go convention)
   - Use `omitempty` for optional fields (reduce payload size)

3. **Type Mapping**:
   - Strings for IDs, names, enums
   - `int` for numeric fields (port, count)
   - `map[string]string` for key-value pairs (tags, labels)
   - `[]string` for arrays (services, ACLs)

4. **Validation**:
   - Required fields: no `omitempty` tag
   - Optional fields: `omitempty` tag
   - Enums: Use string type + Terraform validators

**Rationale**:
- Matches ARK SDK model structure (proven API compatibility)
- Standard Go JSON serialization patterns
- Clear separation of concerns (request vs response models)
- Eliminates ARK SDK dependency

**Sources**:
- ARK SDK models: `pkg/services/sia/workspaces/db/models/`
- ARK SDK models: `pkg/services/sia/secrets/db/models/`
- Go JSON encoding best practices
- `CLAUDE.md` Database Workspace Field Mappings

---

### 5. Dependency Management Strategy (RESOLVED)

**Research Question**: What is the safest approach to removing ARK SDK dependencies from `go.mod`?

**Decision**: Incremental removal with verification at each step

**Removal Strategy**:

1. **Phase 1: Create Replacements**
   - Implement custom OAuth2 clients for database workspaces and secrets
   - Define custom model types in `internal/models/`
   - Verify new implementations work independently

2. **Phase 2: Update References**
   - Modify `provider.go` to initialize only custom OAuth2 clients
   - Update `resource_database_workspace.go` to use `DatabaseWorkspaceClientOAuth2`
   - Update `resource_secret.go` to use `SecretsClientOAuth2`
   - Update validators to use custom model types (remove ARK SDK imports)

3. **Phase 3: Remove ARK SDK Code**
   - Delete `internal/client/auth.go` (ARK SDK auth wrapper)
   - Delete `internal/client/certificates.go` (ARK SDK certificates wrapper - superseded by certificates_oauth2.go)
   - Delete `internal/client/sia_client.go` (ARK SDK SIA client wrapper)
   - Remove all ARK SDK imports from remaining files

4. **Phase 4: Clean Dependencies**
   - Run `go mod tidy` to remove unused ARK SDK packages
   - Verify no ARK SDK dependencies remain: `rg "ark-sdk-golang" go.mod`
   - Run full test suite to ensure no regressions

**Verification Gates**:
- ✅ After each phase: `go build` succeeds
- ✅ After each phase: `go test ./...` passes
- ✅ After Phase 4: `rg "ark-sdk-golang"` returns zero results (excluding docs)

**Rationale**:
- Incremental approach reduces risk of breaking changes
- Each phase is independently verifiable
- Allows rollback if issues discovered
- Ensures no orphaned imports remain

**Sources**:
- Go Modules documentation
- `docs/oauth2-authentication-fix.md` (migration plan)
- Terraform provider development best practices

---

## Research Summary

All research tasks completed successfully. Key findings:

1. ✅ **OAuth2 Pattern**: Proven working implementation exists (`oauth2.go`, `certificates_oauth2.go`)
2. ✅ **API Contracts**: Well-documented via ARK SDK source code and field mapping tables
3. ✅ **HTTP Client Patterns**: HashiCorp best practices already implemented in certificate resource
4. ✅ **Model Design**: Clear pattern established by ARK SDK models (just need custom structs)
5. ✅ **Dependency Removal**: Incremental strategy minimizes risk

**Zero NEEDS CLARIFICATION items remain** - All unknowns resolved through existing implementations and documentation.

**Ready for Phase 1**: Design artifacts (data models, contracts, quickstart guide)
