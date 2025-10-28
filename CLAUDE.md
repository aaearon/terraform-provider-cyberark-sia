# terraform-provider-cyberark-sia Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-27 (Canonical Testing Reference Established)

## Active Technologies
- **Go**: 1.25.0 (confirmed in Phase 2)
- **ARK SDK**: github.com/cyberark/ark-sdk-golang v1.5.0
- **Terraform Plugin Framework**: v1.16.1 (Plugin Framework v6)
- **Terraform Plugin Log**: v0.9.0

## UAP Service Availability

**IMPORTANT**: The UAP (Unified Access Policy) service is not available on all CyberArk tenants. Tenants must have the UAP service specifically provisioned to use policy management features.

**Verification**: Check Platform Discovery API to confirm UAP service availability:
```bash
curl -s "https://platform-discovery.cyberark.cloud/api/v2/services/subdomain/{your-tenant}" | jq '.uap // .jit // .dpa'
```

- If UAP/JIT/DPA appears in the response, the tenant supports policy management
- If missing, the tenant does not have UAP service enabled
- Contact CyberArk support to provision UAP service if needed

**Note**: The `cyberiamtest` tenant used in early development did not have UAP provisioned, which caused DNS lookup failures when attempting to use policy resources.

## API Documentation and Source of Truth

**CRITICAL**: The CyberArk SIA API is not fully documented. The ARK SDKs serve as the **source of truth** for understanding the API's actual behavior and available fields.

### Cross-SDK Validation Requirements

**ALWAYS check BOTH SDKs** when determining if an attribute/field exists:
- **ark-sdk-golang** (Go): github.com/cyberark/ark-sdk-golang
- **ark-sdk-python** (Python): github.com/cyberark/ark-sdk-python

**Why Both?**: The SDKs do NOT have feature parity. A field may exist in one SDK but not the other, or be named differently. To determine the true API schema:

1. **Check ark-sdk-golang first** (since we use it directly)
2. **Verify against ark-sdk-python** (may expose additional fields or patterns)
3. **Cross-reference both** to confirm:
   - Field actually exists in the API (not SDK-fabricated)
   - Field name and type are correct
   - Whether field is required or optional
   - Valid values and constraints

### Validation Process

When adding or modifying a resource attribute:
```bash
# 1. Check Go SDK struct definition
go doc github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db.ArkSIADBAdd

# 2. Download and check Python SDK (if not already available)
# Clone: git clone https://github.com/cyberark/ark-sdk-python
# Search: rg "field_name" ark-sdk-python/

# 3. Compare both - if mismatch, investigate further or test against live API
```

### Common Pitfalls

- ❌ **Don't assume** a field exists just because it's logical
- ❌ **Don't trust** API documentation alone (often outdated)
- ✅ **DO verify** against both SDKs before adding to provider schema
- ✅ **DO document** which SDK version confirmed each field

**Example**: The `database_workspace_id` field in secrets was provider-fabricated and didn't exist in either SDK, causing user confusion. Always validate against SDK source code.

## Project Structure
```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform provider implementation
│   ├── client/          # ARK SDK wrappers, retry, error handling
│   ├── models/          # Data models
│   └── validators/      # Custom Terraform validators (DatabaseEngine, etc.)
├── examples/            # Terraform HCL examples
│   ├── complete/        # Complete working examples
│   ├── provider/        # Provider configuration examples
│   ├── resources/       # Per-resource examples
│   └── testing/         # CRUD testing framework templates
├── docs/                # Documentation
│   ├── guides/          # User guides
│   ├── resources/       # Resource documentation
│   ├── sdk-integration.md      # ARK SDK reference
│   ├── development-history.md  # Development timeline
│   └── troubleshooting.md      # Common issues & solutions
├── specs/               # Feature specifications
│   └── 001-build-a-terraform/  # Spec artifacts (spec.md, plan.md, tasks.md, etc.)
├── specs-archive/       # Archived specifications
└── tests/               # Acceptance tests (empty - placeholder for future)
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
- **CRUD Testing Framework**: Use `examples/testing/TESTING-GUIDE.md` for systematic testing of all resources (canonical reference)

## CRUD Testing Standards

**CANONICAL REFERENCE**: `examples/testing/TESTING-GUIDE.md`

### Mandatory Testing Workflow

**ALL CRUD testing MUST follow** `examples/testing/TESTING-GUIDE.md`. This is the **single source of truth** for:
- Test configuration templates
- Testing workflow (CREATE → READ → UPDATE → DELETE)
- Validation checklists
- Resource dependency patterns
- Troubleshooting procedures

### When to Update TESTING-GUIDE.md

Update `examples/testing/TESTING-GUIDE.md` when:
1. ✅ Adding a new resource type
2. ✅ Changing resource schemas or dependencies
3. ✅ Discovering new validation requirements
4. ✅ Adding new troubleshooting scenarios
5. ✅ Updating provider configuration requirements

### Template Usage Rules

1. **Start from templates**: Always copy templates from `examples/testing/crud-test-*.tf`
2. **Working directory**: `/tmp/sia-crud-validation` (or timestamped variant)
3. **Never modify templates**: Copy to working directory, then customize
4. **Report issues**: If templates are outdated/broken, update TESTING-GUIDE.md first

### Testing Checklist (Before Committing Resource Changes)

- [ ] Run full CRUD cycle using TESTING-GUIDE.md workflow
- [ ] All validation checks pass (validation_summary outputs)
- [ ] Update TESTING-GUIDE.md if resource behavior changed
- [ ] Update template files if new dependencies added
- [ ] Document any new troubleshooting scenarios

**CRITICAL**: Do NOT create ad-hoc test configurations. Always use the canonical templates.

## ARK SDK Integration Patterns

### Authentication
```go
// Disable credential caching - fresh auth for each provider run
ispAuth := auth.NewArkISPAuth(false)

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

### Policy Database Assignment Bug Fix - Azure Database Support (2025-10-27)

**Critical Fix**: Fixed policy database assignment to correctly handle ALL cloud providers (Azure, AWS, GCP, on-premise).

**Root Cause**: The provider incorrectly assumed different cloud providers use different policy target sets (`"AWS"`, `"AZURE"`, `"GCP"`, etc.). The actual API behavior is that **ALL database workspaces use the `"FQDN/IP"` target set**, regardless of the `cloud_provider` attribute.

**Error Before Fix**:
```
Error: The only allowed key in the targets dictionary is "FQDN/IP".
```

**Evidence**: User-provided curl request from SIA UI showed Azure PostgreSQL database successfully assigned to policy using `"FQDN/IP"` target set:
```json
{
  "targets": {
    "FQDN/IP": {
      "instances": [{
        "instanceName": "azure-postgres-test",
        "instanceType": "Postgres",
        "authenticationMethod": "db_auth"
      }]
    }
  }
}
```

**Files Modified**:
- `internal/provider/policy_database_assignment_resource.go:1110` - Fixed `determineWorkspaceType()` to always return `"FQDN/IP"`
- `internal/provider/policy_database_assignment_resource.go:566-589` - Removed platform drift detection (no longer needed)
- `docs/resources/policy_database_assignment.md` - Updated documentation to reflect correct behavior
- `examples/testing/TESTING-GUIDE.md` - Added cloud provider testing section

**Before** (Incorrect):
```go
func determineWorkspaceType(platform string) string {
    switch strings.ToUpper(platform) {
    case "AWS": return "AWS"
    case "AZURE": return "AZURE"  // ← WRONG!
    case "GCP": return "GCP"
    // ...
    }
}
```

**After** (Correct):
```go
func determineWorkspaceType(platform string) string {
    // ALL database workspaces use "FQDN/IP" target set
    // regardless of cloud provider (AWS/Azure/GCP/on-premise)
    // The cloud_provider field is metadata only
    return "FQDN/IP"
}
```

**Validation**: Full integration test with Azure PostgreSQL Flexible Server confirmed fix:
- Created Azure PostgreSQL B1ms (~$0.01 for test)
- Created SIA certificate, secret, database workspace
- **Successfully assigned to "Terraform-Test-Policy"** ✅
- Policy assignment ID: `80b9f727-116d-4e6a-b682-f52fa8c25766:193512`
- Test results: `/tmp/sia-azure-test-20251027-185657/TEST-RESULTS.md`

**Key Learning**: The `cloud_provider` attribute on database workspaces is **metadata only** and does not affect policy target set selection. This applies to ALL cloud providers:
- `cloud_provider: "aws"` → uses `"FQDN/IP"` target set
- `cloud_provider: "azure"` → uses `"FQDN/IP"` target set
- `cloud_provider: "gcp"` → uses `"FQDN/IP"` target set
- `cloud_provider: "on_premise"` → uses `"FQDN/IP"` target set
- `cloud_provider: "atlas"` → uses `"FQDN/IP"` target set

### Policy Database Assignment Resource Complete (2025-10-27)

**New Resource**: `cyberarksia_policy_database_assignment` - Manages assignment of database workspaces to existing SIA access policies.

**Design Pattern**: Follows AWS Security Group Rule pattern (separate resource per database assignment rather than managing entire policy).

**Features Implemented**:
- ✅ Full CRUD operations with idempotency
- ✅ All 6 SIA authentication methods: `db_auth`, `ldap_auth`, `oracle_auth`, `mongo_auth`, `sqlserver_auth`, `rds_iam_user_auth`
- ✅ Read-modify-write pattern (preserves UI-managed and other Terraform-managed databases)
- ✅ Platform drift detection (force replacement when database workspace platform changes)
- ✅ Import support with composite ID format: `policy-id:database-id`
- ✅ Retry logic with exponential backoff (3 attempts, 500ms-30s delays)
- ✅ Comprehensive validation (authentication method, profile type matching)

**UAP Client Integration** (NEW):
- Added `UAPClient` field to `ProviderData` struct
- Created `internal/client/uap_client.go` wrapper following sia_client.go pattern
- Initialize UAP API client in provider `Configure()` method
- Use for policy management operations

**Key Implementation Patterns**:

1. **Composite ID Management**:
```go
// Building composite ID
id := fmt.Sprintf("%s:%s", policyID, databaseID)

// Parsing composite ID
parts := strings.SplitN(id, ":", 2)
policyID, databaseID := parts[0], parts[1]
```

2. **Read-Modify-Write Pattern** (preserves UI databases):
```go
// Fetch full policy
policy, err := r.providerData.UAPClient.Db().Policy(&req)

// Modify only our database in the fetched policy
workspaceType := determineWorkspaceType(database.Platform)
targets := policy.Targets[workspaceType]
targets.Instances = append(targets.Instances, *instanceTarget)
policy.Targets[workspaceType] = targets

// Write back with ONLY the modified workspace type (API constraint)
updatePolicy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
    ArkUAPSIACommonAccessPolicy: policy.ArkUAPSIACommonAccessPolicy,
    Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
        workspaceType: policy.Targets[workspaceType],  // Single workspace type only
    },
}
err = r.providerData.UAPClient.Db().UpdatePolicy(updatePolicy)
```

3. **Platform to Workspace Type Mapping**:
```go
func determineWorkspaceType(platform string) string {
    switch strings.ToUpper(platform) {
    case "AWS": return "AWS"
    case "AZURE": return "AZURE"
    case "GCP": return "GCP"
    case "ATLAS": return "ATLAS"
    case "ON-PREMISE", "": return "FQDN/IP"  // Note: slash, not underscore!
    default: return "FQDN/IP"
    }
}
```

4. **Type Conversion Pattern** (Database IDs):
```go
// Terraform string → SDK int (for API calls)
databaseIDInt, err := strconv.Atoi(data.DatabaseWorkspaceID.ValueString())
database, err := r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
    ID: databaseIDInt,  // int required
})

// SDK int → Terraform string (for state storage)
data.DatabaseWorkspaceID = types.StringValue(strconv.Itoa(database.ID))
```

5. **Profile Type Matching Validation**:
```go
// internal/validators/profile_validator.go
type AuthenticationMethodValidator struct{}

func (v AuthenticationMethodValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    validMethods := []string{"db_auth", "ldap_auth", "oracle_auth", "mongo_auth", "sqlserver_auth", "rds_iam_user_auth"}
    // Validation logic...
}
```

**Important Limitations** (Documented, Not Bugs):

1. **Multi-workspace conflicts**: Managing the same policy from multiple Terraform workspaces causes race conditions
   - This is the same limitation as `aws_security_group_rule`, `google_project_iam_member`, etc.
   - Mitigation: Manage all assignments for a policy in a single workspace, use modules for organization

2. **Policy locationType constraint**: Policies are locked to a single location type (set at policy creation)
   - `locationType: "FQDN_IP"` → Only on-premise databases (`cloud_provider: "on_premise"`)
   - `locationType: "AWS"` → Only AWS databases (`cloud_provider: "aws"`)
   - `locationType: "AZURE"` → Only Azure databases (`cloud_provider: "azure"`)
   - `locationType: "GCP"` → Only GCP databases (`cloud_provider: "gcp"`)
   - `locationType: "ATLAS"` → Only Atlas databases (`cloud_provider: "atlas"`)
   - **Cannot mix location types** in the same policy (API enforced)
   - API error: `"The only allowed key in the targets dictionary is \"FQDN/IP\"."`
   - Mitigation: Create separate policies for each location type

**Usage Example**:
```hcl
# Lookup policy by name
data "cyberarksia_access_policy" "db_admins" {
  name = "Database Administrators"
}

# Assign database to policy
resource "cyberarksia_policy_database_assignment" "postgres" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.prod.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["db_reader", "db_writer"]
  }
}
```

**Files Created**:
- `internal/provider/policy_database_assignment_resource.go` - Main resource (1200+ lines)
- `internal/models/policy_database_assignment.go` - State models for all 6 auth profiles
- `internal/validators/profile_validator.go` - Authentication method validator
- `internal/client/uap_client.go` - UAP API wrapper
- `docs/resources/policy_database_assignment.md` - Comprehensive documentation
- `examples/resources/cyberarksia_policy_database_assignment/*.tf` - Examples for all 6 auth methods

**Files Modified**:
- `internal/provider/provider.go` - UAP client initialization, resource/data source registration
- `internal/provider/logging.go` - UAP logging functions

**Test Results** (2025-10-27):
- ✅ Build: SUCCESS (`go build -v`)
- ✅ Unit tests: 39/39 passing (validators + helpers)
- ⏳ Acceptance tests: Pending (requires real SIA API environment)

**Progress**: 40/63 tasks complete (63.5%) - Phases 0, 1, 2 complete. Documentation complete. Testing phase pending.

---

### Schema Correction: Secret-Database Workspace Relationship (2025-10-27)

**Critical Fix**: Corrected the relationship between secret and database_workspace resources to match the actual CyberArk SIA API architecture.

**Changes Made**:
1. **Secret Resource** - Removed `database_workspace_id` field:
   - ❌ Field did not exist in ARK SDK `ArkSIADBAddSecret` struct
   - ❌ Was never used in API calls (provider-fabricated field)
   - ✅ Secrets are now correctly modeled as **standalone resources**

2. **Database Workspace Resource** - Made `secret_id` required:
   - Changed from `Optional: true` → `Required: true`
   - Aligns with functional requirement: database workspaces MUST have a secret for ZSP/JIT access
   - SDK's `omitempty` tag is serialization-only, not business logic

**Correct Usage Pattern**:
```hcl
# Create secret first (standalone)
resource "cyberarksia_secret" "db_admin" {
  name                = "postgres-admin"
  authentication_type = "local"
  username            = "admin"
  password            = var.db_password
}

# Database workspace references secret (required)
resource "cyberarksia_database_workspace" "production" {
  name          = "prod-postgres"
  database_type = "postgres"
  secret_id     = cyberarksia_secret.db_admin.id  # Required!
  address       = "postgres.example.com"
  port          = 5432
}
```

**Gemini Expert Analysis Confirmed**:
- Secrets are API-level standalone resources
- `database_workspace_id` was a provider-only construct causing user issues
- `secret_id` is functionally essential for ZSP/JIT (core functionality)
- Removal is breaking but acceptable (no production users yet)

**Files Updated**:
- `internal/provider/secret_resource.go` - Removed database_workspace_id schema
- `internal/models/secret.go` - Removed DatabaseWorkspaceID field
- `internal/provider/database_workspace_resource.go` - Made secret_id required
- `examples/resources/secret/*.tf` - Updated all secret examples
- `examples/resources/database_workspace/*.tf` - Added required secret_id

### SDK Cross-Validation Complete (2025-10-27)

**Comprehensive Validation**: All Terraform provider resource attributes validated against BOTH ark-sdk-golang v1.5.0 and ark-sdk-python to ensure accuracy.

**Validation Results**:
- ✅ **Database Workspace**: All 17 exposed attributes match SDK fields in both Go and Python SDKs
- ✅ **Secret Resource**: All 11 exposed attributes match SDK fields (domain field documented as convenience-only)
- ✅ **Certificate Resource**: All 14 attributes correctly match API reality (6 Python SDK fabricated fields already removed 2025-10-25)

**Key Finding**: Python SDK's `ArkSIACertificate` class exposes 6 fields that do NOT exist in actual API responses (`created_by`, `last_updated_by`, `version`, `checksum`, `updated_time`, `cert_password`). Our Go SDK wrapper correctly omits these, and our provider correctly follows the Go SDK.

**Methodology**:
- Cross-referenced every Terraform attribute with Go SDK structs via `go doc`
- Validated against Python SDK structs (cloned from GitHub)
- Confirmed certificate fields through direct API testing
- Documented all intentionally unexposed fields (Active Directory domain controller, PAM secrets, Atlas secrets)

**Conclusion**: Provider schema is 100% accurate. All exposed attributes exist in ark-sdk-golang v1.5.0. No code changes required.

### Removed Provider-Level Retry Configuration (2025-10-15 - Phase 3.5)
- **max_retries Removed**: Removed provider-level retry configuration - now hard-coded constant (3 retries)
- **request_timeout Removed**: Removed unused parameter (never referenced in code)

### Removed Unused sia_api_url Parameter (2025-10-15)

### Provider Configuration Simplification (2025-10-15)

### Phase 3 Cleanup (2025-10-15) - Schema Validation & SDK Constraints

### Phase 3 (2025-10-15) - Database Workspace Resource (User Story 1)

### Phase 2.5 (2025-10-15) - Technical Debt Resolution

### Phase 2 (2025-10-15) - Foundation Complete

### Phase 1 (2025-10-15) - Project Initialization

## Known ARK SDK Limitations (v1.5.0)

1. **No Context Support**: `Authenticate()` doesn't accept `context.Context`
2. **No Structured Errors**: Returns generic `error` interface
3. **No HTTP Status Codes**: Status codes embedded in error strings
4. **Token Expiration**: 15-minute bearer tokens (SDK handles refresh)
5. **DELETE Panic Bug (CRITICAL)**: `DeleteDatabase()` and `DeleteSecret()` cause nil pointer panic (WORKAROUND IMPLEMENTED)

See `docs/sdk-integration.md` for detailed SDK integration patterns.

### DELETE Panic Bug Workaround (2025-10-27)

**Bug**: ARK SDK v1.5.0's `DeleteDatabase()` and `DeleteSecret()` methods pass `nil` body to HTTP DELETE requests, causing panic in `doRequest()` when `http.NewRequestWithContext()` calls `.Len()` on nil `*bytes.Buffer`.

**Root Cause** (`pkg/common/ark_client.go:556-576`):
```go
var bodyBytes *bytes.Buffer  // Defaults to nil
if body != nil {
    bodyBytes = bytes.NewBuffer(json.Marshal(body))
}
req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyBytes)  // Panic if bodyBytes is nil!
```

**Affected SDK Methods**:
- `pkg/services/sia/workspaces/db/ark_sia_workspaces_db_service.go:188` - `DeleteDatabase()` passes `nil`
- `pkg/services/sia/secrets/db/ark_sia_secrets_db_service.go:343` - `DeleteSecret()` passes `nil`

**Workaround Implemented**:
- **File**: `internal/client/delete_workarounds.go`
- **Functions**: `DeleteDatabaseWorkspaceDirect()`, `DeleteSecretDirect()`
- **Pattern**: Create temporary ISP client, call `client.Delete()` with `map[string]string{}` instead of `nil`
- **Why It Works**: Empty map JSON-marshals to `"{}"`, creating valid `bytes.Buffer` → no panic
- **Same Pattern**: Already used successfully in `certificates.go:570`

**Provider Changes**:
- `database_workspace_resource.go:726-736` - Uses `DeleteDatabaseWorkspaceDirect()` instead of SDK method
- `secret_resource.go:496-506` - Uses `DeleteSecretDirect()` instead of SDK method
- Certificate DELETE already working (uses direct HTTP calls with workaround)

**CRUD Test Results** (2025-10-27):
- ✅ **CREATE**: All resources work correctly
- ✅ **READ**: All resources work correctly
- ✅ **UPDATE**: All resources work correctly (certificate fix applied 2025-10-27)
- ✅ **DELETE**: All resources work correctly with workaround

**TODO**: Remove workaround when ARK SDK v1.6.0+ fixes nil body handling in `doRequest()`.

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
| `secret_id` | `SecretID` | ✅ Required | Links to cyberarksia_secret resource for ZSP/JIT access provisioning |
| `enable_certificate_validation` | `EnableCertificateValidation` | Optional | Enforce TLS cert validation (default: true) |
| `certificate_id` | `Certificate` | Optional | TLS/mTLS certificate reference |
| `cloud_provider` | `Platform` | Optional | aws, azure, gcp, on_premise, atlas |
| `region` | `Region` | Optional | **Required for RDS IAM auth** |
| `read_only_endpoint` | `ReadOnlyEndpoint` | Optional | Read replica endpoint |
| `tags` | `Tags` | Optional | Key-value metadata |

**Removed** (Phase 3 Cleanup): database_version, aws_account_id, azure_tenant_id, azure_subscription_id

**Not Exposed Yet**: Active Directory domain controller fields (6 fields)

See `docs/sdk-integration.md` for complete field mapping table and unexposed SDK fields.

## Implementation Status

**Completed Resources**:
- ✅ Certificate Resource - Full CRUD with TLS/mTLS support
- ✅ Database Workspace Resource - Full CRUD with all major database types
- ✅ Secret Resource - Full CRUD with local, domain, and AWS IAM authentication
- ✅ Access Policy Data Source - Lookup policies by ID or name
- ✅ Policy Database Assignment Resource - Manage database assignments to policies (6 authentication methods)

**Future Enhancements** (pending user demand):
- Active Directory domain controller integration (6 fields available in SDK)
- CyberArk PAM secret integration
- MongoDB Atlas secret type
- Enhanced lifecycle management (ignore_changes, prevent_destroy patterns)
- Additional database engines as they become available

<!-- MANUAL ADDITIONS START -->

## Certificate Resource Changes (Breaking - 2025-10-25)

**Removed Fabricated Fields**: The following attributes were removed as they don't exist in the CyberArk SIA Certificates API:
- `created_by` - User who created certificate (not returned by API)
- `last_updated_by` - User who last updated (not returned by API)
- `version` - Version number (not returned by API)
- `checksum` - SHA256 hash for drift detection (not returned by API)
- `updated_time` - Last modification timestamp (not returned by API)
- `cert_password` - Password for encrypted certificates (API only supports public keys)

**Actual Certificate Attributes** (per API documentation):
- Core: `id`, `certificate_id`, `tenant_id`
- Input: `cert_name`, `cert_body`, `cert_description`, `cert_type`, `domain_name`, `labels`
- Computed: `expiration_date`, `metadata` (issuer, subject, valid_from, valid_to, serial_number, subject_alternative_name)

## LLM Testing Guide: Provider CRUD Operations

This section provides a structured plan for LLMs to test CRUD operations of the Terraform provider.

### Prerequisites

Before testing, ensure:
1. Valid SIA credentials (username and client_secret)
2. Provider built and installed: `make install`
3. Test certificates available (can generate with `openssl`)

### Test Environment Setup

**Step 1: Create Test Directory**
```bash
mkdir -p /tmp/sia-crud-test
cd /tmp/sia-crud-test
```

**Step 2: Generate Test Certificate**
```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=test-cert/O=Testing/C=US"
```

**Step 3: Create Base Terraform Configuration**
```hcl
# main.tf
terraform {
  required_providers {
    cyberarksia = {
      source  = "terraform.local/local/cyberark-sia"
      version = "0.1.0"
    }
  }
}

provider "cyberarksia" {
  username      = "your-service-account@cyberark.cloud.XXXX"
  client_secret = "your-secret"
}
```

### CRUD Testing Plan

#### Test 1: CREATE - Certificate Resource

**Objective**: Validate certificate creation and schema correctness

**Configuration**:
```hcl
resource "cyberarksia_certificate" "test_create" {
  cert_name        = "crud-test-create-${timestamp()}"
  cert_description = "CRUD test - Create operation"
  cert_body        = file("${path.module}/cert.pem")
  cert_type        = "PEM"

  labels = {
    test      = "crud_create"
    timestamp = formatdate("YYYY-MM-DD", timestamp())
  }
}

output "certificate_id" {
  value = cyberarksia_certificate.test_create.id
}

output "expiration_date" {
  value = cyberarksia_certificate.test_create.expiration_date
}

output "metadata_issuer" {
  value = cyberarksia_certificate.test_create.metadata.issuer
}
```

**Steps**:
1. Run `terraform init`
2. Run `terraform plan` - verify no warnings
3. Run `terraform apply -auto-approve`
4. **Validate**:
   - Certificate created successfully
   - `id` and `certificate_id` are populated
   - `expiration_date` is populated with ISO 8601 timestamp
   - `metadata` object is populated with certificate details
   - **NO WARNINGS** about unknown attributes

**Expected Output**:
```
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:
certificate_id   = "1234567890123456"
expiration_date  = "2026-10-24T12:00:00+00:00"
metadata_issuer  = "CN=test-cert,O=Testing,C=US"
```

#### Test 2: READ - State Refresh

**Objective**: Validate drift detection and state refresh

**Steps**:
1. Run `terraform plan` (after CREATE)
2. Run `terraform refresh`
3. **Validate**:
   - No changes detected
   - All computed fields match state
   - Sensitive fields (cert_body) remain in state

**Expected Output**:
```
No changes. Your infrastructure matches the configuration.
```

#### Test 3: UPDATE - Modify Certificate Attributes

**Objective**: Validate update operations and field persistence

**Configuration Update**:
```hcl
resource "cyberarksia_certificate" "test_create" {
  cert_name        = "crud-test-updated-${timestamp()}"  # Changed
  cert_description = "CRUD test - Update operation"      # Changed
  cert_body        = file("${path.module}/cert.pem")     # Required for updates
  cert_type        = "PEM"

  labels = {
    test      = "crud_update"  # Changed
    timestamp = formatdate("YYYY-MM-DD", timestamp())
    updated   = "true"         # Added
  }
}
```

**Steps**:
1. Modify configuration (change description, labels, cert_name)
2. Run `terraform plan`
3. Run `terraform apply -auto-approve`
4. **Validate**:
   - Update successful
   - Changes reflected in state
   - `cert_body` persisted correctly (required for ALL updates)
   - **NO WARNINGS** about unknown attributes

**Expected Output**:
```
Apply complete! Resources: 0 added, 1 changed, 0 destroyed
```

#### Test 4: IMPORT - State Import

**Objective**: Validate import functionality

**Steps**:
1. Remove resource from state: `terraform state rm cyberarksia_certificate.test_create`
2. Import using certificate ID: `terraform import cyberarksia_certificate.test_create <certificate_id>`
3. Run `terraform plan`
4. **Validate**:
   - Import successful
   - All computed fields populated
   - No unexpected changes

**Expected Output**:
```
Import successful!

The imported object is now in your Terraform state.
```

#### Test 5: DELETE - Resource Cleanup

**Objective**: Validate delete operation and error handling

**Steps**:
1. Run `terraform destroy -auto-approve`
2. **Validate**:
   - Certificate deleted successfully
   - State is empty
3. **Error Test** (if certificate is in use):
   - Create database_workspace referencing certificate
   - Try to delete certificate
   - **Validate**: Error message about CERTIFICATE_IN_USE

**Expected Output**:
```
Destroy complete! Resources: 1 destroyed.
```

**Error Case Output**:
```
Error: Certificate In Use - delete certificate

The certificate is currently referenced by the following database workspaces:
- workspace-id-1
- workspace-id-2

Remove the certificate from these workspaces before deleting.
```

### Database Workspace CRUD Testing

#### Test 6: CREATE - Database Workspace with Certificate

**Configuration**:
```hcl
resource "cyberarksia_certificate" "db_cert" {
  cert_name   = "db-test-cert"
  cert_body   = file("${path.module}/cert.pem")
  cert_type   = "PEM"
}

resource "cyberarksia_database_workspace" "test_db" {
  name                          = "crud-test-postgres"
  database_type                 = "postgres"
  address                       = "postgres.example.com"
  port                          = 5432
  certificate_id                = cyberarksia_certificate.db_cert.id
  enable_certificate_validation = true

  tags = {
    environment = "test"
    purpose     = "crud-validation"
  }
}
```

**Validate**:
- Workspace created with certificate reference
- Certificate ID correctly linked
- All computed fields populated

### Validation Checklist

For each CRUD operation, verify:

- [ ] **No Warnings**: No warnings about unknown attributes
- [ ] **Computed Fields**: All computed fields are populated (`expiration_date`, `tenant_id`, `metadata`)
- [ ] **Sensitive Fields**: `cert_body` marked as `(sensitive value)` in output
- [ ] **State Consistency**: State matches remote API
- [ ] **Error Messages**: Actionable guidance provided
- [ ] **Retry Logic**: Transient errors auto-retry (check logs)

### Common Testing Patterns

**Pattern 1: Quick Validation Test**
```bash
# One-liner to test create/destroy cycle
terraform init && \
terraform apply -auto-approve && \
terraform show && \
terraform destroy -auto-approve
```

**Pattern 2: Field Validation**
```bash
# Check specific field values
terraform show -json | jq '.values.root_module.resources[] |
  select(.type=="cyberarksia_certificate") |
  {id, certificate_id, expiration_date, tenant_id}'
```

**Pattern 3: Drift Detection**
```bash
# Manually modify resource in SIA UI, then:
terraform plan  # Should detect drift
terraform refresh
terraform plan  # Should show changes needed
```

### Cleanup

Always clean up test resources:
```bash
cd /tmp/sia-crud-test
terraform destroy -auto-approve
cd ~ && rm -rf /tmp/sia-crud-test
```

### LLM Testing Automation

When testing as an LLM, follow this sequence:

1. **Setup Phase**: Create test directory, generate certificates, initialize Terraform
2. **CREATE Test**: Apply configuration, capture outputs, validate schema
3. **READ Test**: Refresh state, verify no changes
4. **UPDATE Test**: Modify config, apply, verify changes reflected in state
5. **DELETE Test**: Destroy resources, verify cleanup
6. **Validation**: Check for warnings, errors, and unexpected behavior
7. **Cleanup**: Remove test directory and resources

**Key Success Criteria**:
- ✅ No warnings about unknown attributes during any operation
- ✅ All CRUD operations complete successfully
- ✅ Computed fields properly populated (expiration_date, tenant_id, metadata)
- ✅ State matches remote API
- ✅ Error messages are actionable

<!-- MANUAL ADDITIONS END -->
