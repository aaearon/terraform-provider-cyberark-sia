# terraform-provider-cyberark-sia Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-15 (Phase 3 Cleanup)

## Active Technologies
- **Go**: 1.25.0 (confirmed in Phase 2)
- **ARK SDK**: github.com/cyberark/ark-sdk-golang v1.5.0
- **Terraform Plugin Framework**: v1.16.1 (Plugin Framework v6)
- **Terraform Plugin Log**: v0.9.0
- Go 1.25.0 (002-we-need-to)
- N/A (Terraform provider - state managed by Terraform) (002-we-need-to)

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
- 002-we-need-to: Added Go 1.25.0

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
