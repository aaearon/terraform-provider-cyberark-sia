# Feature Specification: Replace ARK SDK with Custom OAuth2 Implementation

**Feature Branch**: `002-we-need-to`
**Created**: 2025-10-25
**Status**: Draft
**Input**: User description: "We need to replace all `ark-sdk-golang` usage in our Terraform provider with our own custom code as we need to use Identity service users and the tokens `ark-sdk-golang` creates are not the right ones (id tokens vs access tokens.) We have no people using our Terraform provider so we do not need to care about backward compatibility."

## User Scenarios & Testing

### User Story 1 - OAuth2 Authentication Migration (Priority: P1)

Infrastructure operators need to authenticate the Terraform provider to CyberArk SIA using service account credentials and have all API operations succeed without authentication errors.

**Why this priority**: This is the foundation for all provider functionality. Without proper authentication, no Terraform resources can be managed. The current ARK SDK implementation creates ID tokens instead of access tokens, causing 401 errors for database workspace and secret resources.

**Independent Test**: Provider can be configured with service account credentials, successfully authenticate to CyberArk Identity, obtain a valid access token, and make an API call to any SIA endpoint (certificates, database workspaces, or secrets) without receiving 401 Unauthorized errors.

**Acceptance Scenarios**:

1. **Given** a service account with valid credentials (username and client_secret), **When** the provider is configured and initialized, **Then** an OAuth2 access token is obtained via the `/oauth2/platformtoken` endpoint and stored for subsequent API requests
2. **Given** a configured provider with a valid access token, **When** any Terraform resource operation is performed (create, read, update, delete), **Then** the request includes the access token in the Authorization header and succeeds without 401 errors
3. **Given** a service account with invalid credentials, **When** the provider attempts authentication, **Then** a clear error message indicates authentication failure with actionable guidance
4. **Given** a valid access token that is nearing expiration (within 5 minutes), **When** a resource operation is attempted, **Then** the provider automatically refreshes the token before making the API request

---

### User Story 2 - Database Workspace Resource Operations (Priority: P2)

Infrastructure operators need to create, update, read, and delete database workspaces in CyberArk SIA using Terraform configurations without authentication failures.

**Why this priority**: Database workspaces are a core SIA resource type. Currently broken due to ARK SDK ID token issue. This is the highest-priority resource after authentication is fixed.

**Independent Test**: Can execute a complete CRUD lifecycle for a database workspace: `terraform apply` creates a new workspace, `terraform plan` shows no drift, modifications to the configuration trigger updates, and `terraform destroy` removes the workspace - all without 401 errors.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration defining a database workspace, **When** `terraform apply` is executed, **Then** the workspace is created via POST to `/api/workspaces/db` using the access token and returns a valid workspace ID
2. **Given** an existing database workspace in Terraform state, **When** `terraform refresh` is executed, **Then** the workspace details are retrieved via GET from `/api/workspaces/db/{id}` and state is updated without drift
3. **Given** a database workspace configuration with modified attributes, **When** `terraform apply` is executed, **Then** the workspace is updated via PUT to `/api/workspaces/db/{id}` with the changed fields
4. **Given** an existing database workspace, **When** `terraform destroy` is executed, **Then** the workspace is deleted via DELETE to `/api/workspaces/db/{id}` and removed from state

---

### User Story 3 - Secret Resource Operations (Priority: P3)

Infrastructure operators need to create, update, read, and delete database secrets in CyberArk SIA using Terraform configurations without authentication failures.

**Why this priority**: Secrets are complementary to database workspaces and enable Zero Standing Privilege workflows. Lower priority than database workspaces as they depend on workspaces being functional.

**Independent Test**: Can execute a complete CRUD lifecycle for a secret: `terraform apply` creates a new secret, `terraform plan` shows no drift, secret rotation updates the credentials, and `terraform destroy` removes the secret - all without 401 errors.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration defining a database secret, **When** `terraform apply` is executed, **Then** the secret is created via POST to `/api/secrets/db` using the access token and returns a valid secret ID
2. **Given** an existing secret in Terraform state, **When** `terraform refresh` is executed, **Then** the secret metadata (not password) is retrieved via GET from `/api/secrets/db/{id}` and state is updated
3. **Given** a secret configuration with rotated credentials, **When** `terraform apply` is executed, **Then** the secret is updated via PUT to `/api/secrets/db/{id}` with the new password (marked sensitive)
4. **Given** an existing secret, **When** `terraform destroy` is executed, **Then** the secret is deleted via DELETE to `/api/secrets/db/{id}` and removed from state

---

### User Story 4 - ARK SDK Removal and Cleanup (Priority: P4)

Development team needs to completely remove all ARK SDK dependencies from the codebase to eliminate technical debt and simplify maintenance.

**Why this priority**: This is cleanup work after all resources are migrated to custom OAuth2 clients. Does not directly impact user functionality but improves long-term maintainability.

**Independent Test**: Can build the provider successfully without any ARK SDK packages in `go.mod`, run `go mod tidy` without adding ARK SDK dependencies, and verify all imports reference only custom OAuth2 client code.

**Acceptance Scenarios**:

1. **Given** all resources have been migrated to custom OAuth2 clients, **When** ARK SDK imports are removed from all Go files, **Then** the provider compiles without errors and all tests pass
2. **Given** ARK SDK packages have been removed from `go.mod`, **When** `go mod tidy` is executed, **Then** no ARK SDK dependencies are re-added
3. **Given** the provider codebase after ARK SDK removal, **When** searching for "ark-sdk-golang" across all files, **Then** zero occurrences are found (excluding historical documentation)
4. **Given** documentation referencing ARK SDK integration patterns, **When** documentation is updated, **Then** all references to ARK SDK are replaced with custom OAuth2 implementation details

---

### Edge Cases

- **Token Expiration During Long Operations**: What happens when an access token expires mid-operation (e.g., during a large bulk update)? The provider must detect token expiration (401 with specific error message), refresh the token, and retry the operation automatically.

- **Concurrent Resource Operations**: How does the provider handle multiple resources being created/updated simultaneously sharing the same access token? The OAuth2 client must be thread-safe with proper mutex locks around token refresh operations.

- **SIA URL Resolution from Different Token Claims**: What happens when JWT token claims have unexpected formats (e.g., missing "shell." prefix, non-standard platform domains)? The SIA URL resolution logic must handle variations gracefully and provide clear error messages for malformed tokens.

- **ARK SDK Model Types After Removal**: How are existing ARK SDK model types (e.g., `dbmodels.AddDatabaseRequest`, `secretsmodels.SecretRequest`) replaced? The provider must define custom model types that match the SIA API contract without depending on SDK types.

- **Migration from Existing Provider Installations**: What happens to users who have the current provider version with ARK SDK dependencies? Since we have no active users, this is not a concern - breaking changes are acceptable.

## Requirements

### Functional Requirements

- **FR-001**: Provider MUST authenticate to CyberArk Identity using OAuth2 client credentials grant flow (POST `/oauth2/platformtoken`)
- **FR-002**: Provider MUST use access tokens (NOT ID tokens) for all SIA API requests with `Authorization: Bearer {access_token}` header
- **FR-003**: Provider MUST automatically extract SIA API base URL from JWT access token claims (`subdomain` and `platform_domain`)
- **FR-004**: Provider MUST implement automatic token refresh when access token is expired or nearing expiration (within 5 minutes)
- **FR-005**: Provider MUST support thread-safe concurrent API operations with shared access token and proper synchronization
- **FR-006**: Database workspace resource MUST support all CRUD operations using custom OAuth2 HTTP client without ARK SDK dependencies
- **FR-007**: Secret resource MUST support all CRUD operations using custom OAuth2 HTTP client without ARK SDK dependencies
- **FR-008**: Certificate resource MUST continue working with existing OAuth2 implementation (already migrated)
- **FR-009**: Provider MUST handle all HTTP status codes appropriately (401 auth failure, 403 permission denied, 404 not found, 409 conflict, 429 rate limit, 500 server error)
- **FR-010**: Provider MUST implement exponential backoff retry logic for transient errors (network timeouts, 429 rate limits, 500 server errors)
- **FR-011**: Provider MUST remove all ARK SDK imports and dependencies from `go.mod` after migration is complete
- **FR-012**: Provider MUST define custom model types for database workspace and secret requests/responses matching SIA API contracts
- **FR-013**: Provider MUST preserve all existing Terraform schema attributes and behaviors (no user-facing breaking changes to resource configuration)
- **FR-014**: Provider MUST properly handle sensitive data (client_secret, access tokens, secret passwords) with no logging of sensitive values
- **FR-015**: Provider MUST support environment variable configuration (`CYBERARK_USERNAME`, `CYBERARK_CLIENT_SECRET`, `CYBERARK_IDENTITY_URL`)

### Key Entities

- **OAuth2 Client**: Manages authentication flow, token lifecycle, automatic refresh, and provides access tokens to resource clients
- **Access Token**: JWT bearer token containing API authorization claims (scope, audience, permissions) with 3600-second expiration
- **Database Workspace Client**: HTTP client for database workspace CRUD operations using access token authentication
- **Secret Client**: HTTP client for secret CRUD operations using access token authentication and sensitive data handling
- **SIA API Contract Models**: Custom Go structs representing database workspace and secret request/response payloads matching API JSON schema
- **Provider Data**: Central configuration object holding OAuth2 client and resource-specific HTTP clients (certificates, database workspaces, secrets)

## Success Criteria

### Measurable Outcomes

- **SC-001**: All database workspace operations (create, read, update, delete) complete successfully without 401 Unauthorized errors (100% success rate)
- **SC-002**: All secret operations (create, read, update, delete) complete successfully without 401 Unauthorized errors (100% success rate)
- **SC-003**: Provider initialization completes in under 3 seconds including OAuth2 authentication and token acquisition
- **SC-004**: Token refresh operations complete transparently without user intervention or resource operation failures (zero failed operations due to token expiration)
- **SC-005**: Provider codebase contains zero ARK SDK imports after migration (verified by searching for "ark-sdk-golang" returns no results)
- **SC-006**: Provider build completes successfully with only custom OAuth2 dependencies in dependencies file (no ARK SDK entries)
- **SC-007**: All existing Terraform configurations continue to work without modification (100% backward compatibility for resource schemas)
- **SC-008**: Concurrent resource operations (minimum 10 simultaneous database workspace creates) complete without race conditions or authentication failures
- **SC-009**: Provider handles transient network errors gracefully with automatic retry completing within 30 seconds (3 retries with exponential backoff)
- **SC-010**: Sensitive data (tokens, passwords) never appears in Terraform logs or error messages (verified by log audit)

## Assumptions

- **OAuth2 Endpoint**: The CyberArk Identity `/oauth2/platformtoken` endpoint is available and supports client credentials grant with service account authentication
- **Token Claims**: All access tokens contain `subdomain` and `platform_domain` claims needed to construct SIA API URLs
- **API Stability**: SIA API endpoints (`/api/workspaces/db`, `/api/secrets/db`, `/api/certificates`) have stable contracts matching current ARK SDK behavior
- **Service Account Permissions**: Service accounts used with the provider have `DpaAdmin` role or equivalent permissions for all resource types
- **No Active Users**: There are zero active users of the provider, so breaking changes to internal implementation are acceptable
- **Single Tenant**: Each provider instance operates within a single CyberArk tenant (no cross-tenant operations)
- **Token Expiration**: Access tokens expire in 3600 seconds (1 hour) and must be refreshed before expiration
- **Environment**: Provider runs in Terraform CLI context with standard Go runtime (not serverless or embedded scenarios)
- **Certificate Resource**: The existing OAuth2 implementation for certificates is correct and should be preserved as a reference pattern
- **HTTP Timeouts**: Default HTTP client timeout of 30 seconds is sufficient for all SIA API operations (no long-running operations requiring custom timeouts)

## Dependencies

- **CyberArk Identity Service**: Provider depends on availability of CyberArk Identity platform for OAuth2 authentication
- **CyberArk SIA API**: Provider depends on availability of SIA API endpoints for resource management
- **JWT Token Library**: Provider requires JWT parsing library to extract claims from access tokens
- **Terraform Plugin Framework**: Provider is built on Terraform Plugin Framework v6 which dictates schema, resource, and provider patterns
- **Existing Certificate OAuth2 Code**: Database workspace and secret implementations will follow patterns from existing certificate OAuth2 implementation

## Out of Scope

- **Breaking Changes to Resource Schemas**: Terraform resource configurations (HCL) must remain unchanged - only internal implementation is modified
- **Token Caching to Disk**: Token persistence across provider invocations is not required - each Terraform command authenticates fresh
- **Multi-Tenant Support**: Provider does not support managing resources across multiple CyberArk tenants in a single configuration
- **Legacy ARK SDK Compatibility Layer**: No wrapper or compatibility layer for ARK SDK - full replacement only
- **Custom Identity URL Discovery**: Provider will not implement CyberArk Platform Discovery Service - users must provide identity URL or use environment variables
- **Performance Optimization Beyond Retry Logic**: Advanced features like request batching, response caching, or connection pooling are not included
- **Observability Enhancements**: No additional metrics, tracing, or monitoring beyond existing Terraform plugin logging
- **Automated SDK Upgrades**: No tooling or process for tracking ARK SDK updates since SDK will be completely removed
- **GovCloud or Special Regions**: No special handling for CyberArk GovCloud deployments beyond standard URL configuration
- **Role-Based Token Scoping**: Provider assumes service account has full permissions - no role-based token scope management

## Notes

**Critical Decision - Token Type**: The ARK SDK v1.5.0 obtains an OAuth2 access token via client credentials grant, then exchanges it for an OpenID Connect ID token via authorization code flow. The SIA API requires **access tokens** (containing API authorization claims like `scope`, `aud`, `permissions`) and rejects **ID tokens** (containing only identity claims like `sub`, `email`, `name`). This architectural flaw in the SDK is the root cause of all 401 Unauthorized errors.

**Proven Solution Pattern**: The certificate resource has already been migrated to use custom OAuth2 implementation calling `/oauth2/platformtoken` directly. This implementation is working in production with zero 401 errors. The migration for database workspaces and secrets will follow the exact same pattern.

**Breaking Change Acceptance**: Since there are no active users of this provider, we can make any breaking changes to internal implementation without concern for backward compatibility. This allows complete removal of ARK SDK rather than maintaining a hybrid approach.

**Migration Risk Assessment**: Low risk - certificate resource proves the OAuth2 pattern works. Database workspace and secret resources have well-documented API contracts from ARK SDK source code review. The main effort is mechanical code replacement following existing patterns.

**Future Token Refresh Enhancement**: Current implementation does not cache tokens or implement automatic refresh. Each provider invocation authenticates fresh. Future enhancement could add token refresh logic with mutex-protected token expiry checks to reduce authentication overhead during long Terraform operations.

**Model Type Replacement Strategy**: ARK SDK model types will be replaced with custom structs defined in new files under internal models directory. These will use JSON struct tags matching the SIA API contract to ensure correct serialization and deserialization.
