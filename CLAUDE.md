# terraform-provider-cyberark-sia Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-15 (Phase 3 Cleanup)

## Active Technologies
- **Go**: 1.25.0 (confirmed in Phase 2)
- **ARK SDK**: github.com/cyberark/ark-sdk-golang v1.5.0
- **Terraform Plugin Framework**: v1.16.1 (Plugin Framework v6)
- **Terraform Plugin Log**: v0.9.0

## Project Structure
```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform provider implementation
│   ├── client/          # ARK SDK wrappers, retry, error handling
│   └── models/          # Data models (Phase 3+)
├── examples/            # Terraform HCL examples (Phase 3+)
├── docs/                # Documentation
│   ├── sdk-integration.md     # ARK SDK reference
│   ├── phase2-reflection.md   # Phase 2 lessons learned
│   └── phase3-reflection.md   # Phase 3 schema validation findings
├── specs/               # Feature specifications
└── tests/               # Acceptance tests (Phase 3+)
```

## Commands

### Build & Test
```bash
# Build provider
go build -v

# Run tests (unit)
go test ./...

# Run tests (acceptance - requires TF_ACC=1)
TF_ACC=1 go test ./... -v

# Run specific package tests
go test ./internal/client/... -v
```

### Code Quality
```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...
gofmt -w .
```

### Development Workflow
```bash
# Install locally for testing
go install

# Clean build artifacts
go clean

# Update dependencies
go mod tidy
go mod download
```

## Code Style

### Go Standards
- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before commits
- Write godoc comments for exported functions

### Terraform Provider Patterns
- Use Terraform Plugin Framework v6
- Mark sensitive attributes with `Sensitive: true`
- Use `terraform-plugin-log/tflog` for structured logging
- **NEVER log sensitive data** (passwords, tokens, secrets)

### Error Handling (Phase 2.5 Improvements)
- Use `internal/client.MapError()` for Terraform diagnostics
- Wrap operations with `internal/client.RetryWithBackoff()`
- Classify errors by type (auth, permission, network, etc.)
- Provide actionable error messages with guidance

### Testing Strategy
- **Primary**: Acceptance tests (test against real SIA API)
- **Selective**: Unit tests for complex validators and helpers
- Use `TF_ACC=1` environment variable for acceptance tests
- Mock only when necessary (prefer real integration tests)

## ARK SDK Integration Patterns

### Authentication
```go
// Enable token caching for auto-refresh
ispAuth := auth.NewArkISPAuth(true)

// Authenticate (note: first param is *ArkProfile, NOT context.Context)
_, err := ispAuth.Authenticate(nil, profile, secret, false, false)
```

### Error Handling
```go
// ARK SDK v1.5.0 returns standard error interface (no structured types)
// Use multi-layer detection:
// 1. Standard Go errors (net.Error, context errors)
// 2. Pattern matching (case-insensitive, ordered by specificity)
// 3. Fallback for unknown errors
```

### Retry Logic
```go
// Wrap SDK calls with exponential backoff (uses hardcoded constants)
err := client.RetryWithBackoff(ctx, &client.RetryConfig{
    MaxRetries: client.DefaultMaxRetries,  // 3 retries
    BaseDelay:  client.BaseDelay,          // 500ms
    MaxDelay:   client.MaxDelay,           // 30s
}, func() error {
    return siaAPI.WorkspacesDB().AddDatabase(...)
})
```

### Logging
```go
// Use structured logging with context
tflog.Info(ctx, "Operation succeeded", map[string]interface{}{
    "operation": "create",
    "resource_id": id,
})

// NEVER log: password, client_secret, aws_secret_access_key, tokens
```

## Recent Changes

### Removed Provider-Level Retry Configuration (2025-10-15 - Phase 3.5)
- **max_retries Removed**: Removed provider-level retry configuration - now hard-coded constant (3 retries)
- **request_timeout Removed**: Removed unused parameter (never referenced in code)
- **Breaking Change**: Users with these parameters in provider config will get schema validation errors
- **Rationale**: Modern Terraform best practice (2025) - providers handle transient errors internally, users control operation timeouts via resource-level `timeouts` blocks
- **Provider Attributes**: Now 4 attributes (3 required: client_id, client_secret, identity_tenant_subdomain; 1 optional: identity_url)
- **Internal Constants**: Retry logic uses `client.DefaultMaxRetries=3`, `client.BaseDelay=500ms`, `client.MaxDelay=30s`
- **Future Enhancement**: Resource-level `timeouts` blocks for long-running operations (Phase 4+)
- **See**: `docs/phase3.5-reflection.md` for architectural decision rationale

### Removed Unused sia_api_url Parameter (2025-10-15)
- **sia_api_url Removed**: Deleted unused configuration parameter that provided no functionality
- **SDK Auto-Construction**: ARK SDK automatically constructs SIA API URL as `https://{subdomain}.dpa.{domain}` from authenticated JWT token
- **No Override Mechanism**: SDK doesn't support custom SIA API URLs - URL is always derived from token
- **Breaking Change**: Users with `sia_api_url` in configs will get schema validation error (parameter was non-functional)

### Provider Configuration Simplification (2025-10-15)
- **identity_url Now Optional**: Reduced required configuration fields from 5 to 4
- **Automatic URL Resolution**: SDK resolves identity URL via discovery service when not provided
- **identity_tenant_subdomain Required**: Explicitly required in schema (was implicitly required)
- **GovCloud Support**: Users can override URL for GovCloud or set `DEPLOY_ENV=gov-prod` environment variable
- **Discovery Service Dependency**: Adds ~100-300ms latency at provider init when URL not provided

### Phase 3 Cleanup (2025-10-15) - Schema Validation & SDK Constraints
- **Removed Non-Existent Fields**: Deleted database_version, aws_account_id, azure_tenant_id, azure_subscription_id (no SDK equivalents)
- **Fixed Required Constraints**: Changed address, port, authentication_method from Required → Optional (SDK uses defaults)
- **database_type Now Required**: SDK v1.5.0 has unconditional validation that rejects empty strings, making this field mandatory
- **Renamed Fields**: aws_region → region (generic, used for RDS IAM authentication)
- **Validated Mappings**: Confirmed `name` and `database_type` are required; all other fields optional
- **Documentation**: Created `phase3-reflection.md` documenting SDK field audit
- **SDK Field Mappings**: Comprehensive Terraform ↔ SDK field mapping table in `sdk-integration.md`

### Phase 3 (2025-10-15) - Database Workspace Resource (User Story 1)
- **Resource Implementation**: Full CRUD for cyberark_sia_database_workspace
- **Schema Design**: 12 attributes (name, database_type, address, port, etc.)
- **SDK Integration**: WorkspacesDB().AddDatabase/Database/UpdateDatabase/DeleteDatabase
- **Error Handling**: 404 detection for drift, retry with backoff
- **Import Support**: State import via resource ID
- **Validation**: OneOf validators for enums, length/range validators

### Phase 2.5 (2025-10-15) - Technical Debt Resolution
- **Enhanced Error Handling**: Robust error classification with fallback
- **Improved Retry Logic**: Better error detection with `net.Error` support
- **Type Safety**: Removed `interface{}` from ProviderData
- **Comprehensive Tests**: 95% coverage for error/retry logic
- **SDK Research**: Documented ARK SDK v1.5.0 packages and limitations
- **Logging**: Added retry attempt logging at WARN level
- **Documentation**: Created `sdk-integration.md` and `phase2-reflection.md`

### Phase 2 (2025-10-15) - Foundation Complete
- **Provider Setup**: Schema with 4 attributes (3 required: client_id, client_secret, identity_tenant_subdomain; 1 optional: identity_url) + environment fallback
- **Authentication**: ISPSS OAuth2 via ARK SDK with caching
- **SIA Client**: Wrapper for WorkspacesDB() and SecretsDB() access
- **Error Mapping**: ARK SDK errors → actionable Terraform diagnostics
- **Retry Logic**: Exponential backoff with 30s max delay
- **Logging**: Structured logging with sensitive data protection
- **Testing**: Acceptance test scaffolding

### Phase 1 (2025-10-15) - Project Initialization
- Go module initialized
- Dependency management configured
- Project structure created

## Known ARK SDK Limitations (v1.5.0)

1. **No Context Support**: `Authenticate()` doesn't accept `context.Context`
2. **No Structured Errors**: Returns generic `error` interface
3. **No HTTP Status Codes**: Status codes embedded in error strings
4. **Token Expiration**: 15-minute bearer tokens (SDK handles refresh)

See `docs/sdk-integration.md` for detailed SDK integration patterns.

## Database Workspace Field Mappings (Phase 3 - VALIDATED + Extended)

| Terraform Attribute | SDK Field | Required? | Notes |
|---------------------|-----------|-----------|-------|
| `name` | `Name` | ✅ Required | Database name on server (e.g., "customers", "myapp") - actual DB that SIA connects to |
| `database_type` | `ProviderEngine` | ✅ Required | SDK v1.5.0 rejects empty strings. 60+ engine types: postgres, mysql, postgres-aws-rds, etc. |
| `network_name` | `NetworkName` | Optional | Network segmentation (default: "ON-PREMISE") |
| `address` | `ReadWriteEndpoint` | Optional | Hostname/IP/FQDN |
| `port` | `Port` | Optional | SDK uses family defaults |
| `auth_database` | `AuthDatabase` | Optional | MongoDB auth database (default: "admin") |
| `services` | `Services` | Optional | Oracle/SQL Server services ([]string) |
| `account` | `Account` | Optional | Snowflake/Atlas account name |
| `authentication_method` | `ConfiguredAuthMethodType` | Optional | ad_ephemeral_user, local_ephemeral_user, rds_iam_authentication, atlas_ephemeral_user |
| `secret_id` | `SecretID` | Optional | **Required for ZSP/JIT**. Links to secret |
| `enable_certificate_validation` | `EnableCertificateValidation` | Optional | Enforce TLS cert validation (default: true) |
| `certificate_id` | `Certificate` | Optional | TLS/mTLS certificate reference |
| `cloud_provider` | `Platform` | Optional | aws, azure, gcp, on_premise, atlas |
| `region` | `Region` | Optional | **Required for RDS IAM auth** |
| `read_only_endpoint` | `ReadOnlyEndpoint` | Optional | Read replica endpoint |
| `tags` | `Tags` | Optional | Key-value metadata |

**Removed** (Phase 3 Cleanup): database_version, aws_account_id, azure_tenant_id, azure_subscription_id

**Not Exposed Yet**: Active Directory domain controller fields (6 fields)

See `docs/sdk-integration.md` for complete field mapping table and unexposed SDK fields.

## Next Steps

- **Phase 4**: Implement secret resource (User Story 2)
- **Phase 5**: Lifecycle enhancements (User Story 3)
- **Phase 6**: Documentation and polish

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->