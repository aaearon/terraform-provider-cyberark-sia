# terraform-provider-cyberark-sia Development Guidelines

**Purpose**: Quick reference for LLM-assisted development of the CyberArk SIA Terraform Provider

**Last Updated**: 2025-10-30 (Complete rewrite with Environment Setup, Provider Overview, CRUD automation, Release guidelines)

## Quick Start

**Technology Stack**:
- **Go**: 1.25.0
- **ARK SDK**: github.com/cyberark/ark-sdk-golang v1.5.0 (has DELETE bug - see Critical Constraints)
- **Terraform Plugin Framework**: v1.16.1 (Plugin Framework v6)
- **Terraform Plugin Log**: v0.9.0

**Build & Run**:
```bash
go build -v                    # Build provider
go install                     # Install locally for testing
TF_ACC=1 go test ./... -v     # Run acceptance tests
```

**Key Files**:
- Authentication: `internal/client/auth.go`
- Profile Factory: `internal/provider/profile_factory.go` (centralized auth profile building)
- Error Handling: `internal/client/errors.go`, `internal/client/retry.go`
- DELETE Workaround: `internal/client/delete_workarounds.go` (ARK SDK bug fix)

## Environment Setup

### Prerequisites
- CyberArk Identity tenant with SIA enabled
- OAuth2 service account credentials (username format: `service-account@cyberark.cloud.XXXXX`)
- Go 1.25.0 installed
- Terraform CLI v1.5+ installed

### Required Environment Variables

For acceptance tests and local development, export these variables:

```bash
# CyberArk Identity Authentication
export CYBERARK_USERNAME="service-account@cyberark.cloud.12345"
export CYBERARK_CLIENT_SECRET="your-client-secret"

# Optional: CyberArk Identity URL (only needed for GovCloud or custom deployments)
# If not provided, URL is automatically resolved from username by ARK SDK
export CYBERARK_IDENTITY_URL="https://abc123.cyberark.cloud"

# Enable Terraform acceptance tests
export TF_ACC=1

# Optional: Terraform logging
export TF_LOG=DEBUG           # For verbose provider logs
export TF_LOG_PATH=./tf.log   # Save logs to file
```

### Terraform CLI Configuration (Local Development)

To use the locally-built provider, configure Terraform CLI dev overrides:

**Configuration File Location**:
- Linux/macOS: `~/.terraformrc`
- Windows: `%APPDATA%/terraform.rc`

**Configuration Content**:
```hcl
provider_installation {
  dev_overrides {
    "aaearon/cyberarksia" = "~/.terraform.d/plugins/local/aaearon/cyberark-sia/dev/linux_amd64"
  }
  direct {}
}
```

**Note**: Adjust the path to match your `make install` target. When dev overrides are active, Terraform will skip version constraints and use your local binary.

### Obtaining SIA Credentials

1. Log into CyberArk Identity admin console
2. Navigate to **Applications** → **Add Web Apps** → **Custom**
3. Create OAuth2 confidential client:
   - **Application ID**: `terraform-provider-sia` (or your preferred name)
   - **Grant Types**: `Client Credentials`
   - **Scopes**: `sia`, `identity`
4. Save the **username** (format: `app-name@cyberark.cloud.XXXXX`) and **client secret**

### Verifying Your Setup

Test that credentials work:

```bash
# Build and install provider
make build
make install

# Run provider configuration test
TF_ACC=1 go test ./internal/provider -v -run TestAccProvider_Configure
```

If successful, you're ready for development.

## Project Structure

```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform provider implementation
│   │   ├── profile_factory.go    # Authentication profile factory
│   │   └── helpers/               # Shared utilities
│   │       ├── id_conversion.go   # Database ID conversion
│   │       └── composite_ids.go   # Composite ID parsing/building
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
│   ├── development/     # Design decisions, implementation summaries
│   ├── sdk-integration.md      # ARK SDK reference
│   ├── development-history.md  # Development timeline
│   └── troubleshooting.md      # Common issues & solutions
├── specs/               # Feature specifications (active)
└── specs-archive/       # Archived specifications
```

## Provider Overview

### Available Resources & Data Sources

| Type | Name | Implementation | Status | Purpose |
|------|------|----------------|--------|---------|
| Resource | `cyberarksia_database_workspace` | `internal/provider/database_workspace_resource.go` | ✅ Stable | Database target configuration (60+ engines supported) |
| Resource | `cyberarksia_secret` | `internal/provider/secret_resource.go` | ✅ Stable | Strong account credentials (username/password, AWS IAM) |
| Resource | `cyberarksia_certificate` | `internal/provider/certificate_resource.go` | ✅ Stable | TLS/mTLS certificates for database connections |
| Resource | `cyberarksia_database_policy` | `internal/provider/database_policy_resource.go` | ✅ Stable | Access policies with time-based conditions |
| Resource | `cyberarksia_database_policy_principal_assignment` | `internal/provider/database_policy_principal_assignment_resource.go` | ✅ Stable | Assign users/groups/roles TO policies (WHO gets access) |
| Resource | `cyberarksia_policy_database_assignment` | `internal/provider/policy_database_assignment_resource.go` | ✅ Stable | Assign database workspaces TO policies (WHAT they access) |
| Data Source | `cyberarksia_principal` | `internal/provider/principal_data_source.go` | ✅ Stable | Lookup users/groups/roles by name (no manual UUID needed) |

### Resource Dependencies

Typical configuration flow:

```
1. cyberarksia_secret (credentials)
     ↓
2. cyberarksia_database_workspace (database target, references secret)
     ↓
3. cyberarksia_database_policy (access conditions)
     ↓
     ├→ 4a. cyberarksia_database_policy_principal_assignment (WHO: assign users/groups/roles)
     └→ 4b. cyberarksia_policy_database_assignment (WHAT: assign database workspaces)
```

**Note**: Principal and database assignments can be managed independently by different teams (security team manages WHO, app team manages WHAT).

## Architecture Patterns

### Profile Factory Pattern

**When to Use**: Creating/updating ANY resource with authentication profiles (db_auth, ldap_auth, oracle_auth, mongo_auth, sqlserver_auth, rds_iam_user_auth)

**Location**: `internal/provider/profile_factory.go`

**Usage**:
```go
// In Create() or Update() methods
profile := BuildAuthenticationProfile(ctx, authMethod, data, &diags)
if diags.HasError() {
    return
}
instanceTarget.Profile = profile
```

**Why**: Eliminates 410 LOC duplication, centralizes validation for all 6 authentication methods, prevents auth drift bugs

**Anti-Pattern**: ❌ Don't manually construct authentication profiles in resource CRUD methods

### Helper Utilities

**Composite ID Parsing** (`internal/provider/helpers/composite_ids.go`):
```go
// Policy-database assignments (2-part ID)
policyID, databaseID, err := helpers.ParsePolicyDatabaseID(id)

// Policy-principal assignments (3-part ID)
policyID, principalID, principalType, err := helpers.ParsePolicyPrincipalID(id)
```

**Database ID Conversion** (`internal/provider/helpers/id_conversion.go`):
```go
// API returns string IDs in JSON but accepts int64 in URLs
dbID, err := helpers.ConvertDatabaseID(stringID)
```

### Error Handling Pattern

**Always Use**:
```go
import "github.com/aaearon/terraform-provider-cyberark-sia/internal/client"

// Wrap SDK calls with retry logic
err := client.RetryWithBackoff(ctx, func() error {
    _, err := siaAPI.WorkspacesDB().AddDatabase(...)
    return err
})

// Convert to Terraform diagnostics
if err != nil {
    resp.Diagnostics.Append(client.MapError(err, "Failed to create database workspace")...)
    return
}
```

**Why**: Automatic exponential backoff (3 retries, 30s max delay), error classification, actionable user messages

### Read-Modify-Write for Policy Assignments

**Critical Pattern**: When updating policy assignments, ALWAYS fetch full policy first

```go
// CORRECT: Preserves other assignments
existingPolicy, err := siaAPI.AccessPolicies().GetAccessPolicy(policyID)
// Modify ONLY managed element
existingPolicy.Principals = append(existingPolicy.Principals, newPrincipal)
// Write back
updated, err := siaAPI.AccessPolicies().UpdatePolicy(policyID, existingPolicy)

// WRONG: Overwrites all other assignments
newPolicy := &models.Policy{Principals: []Principal{newPrincipal}}
updated, err := siaAPI.AccessPolicies().UpdatePolicy(policyID, newPolicy)
```

**Why**: API constraint - UpdatePolicy() accepts only ONE workspace type in Targets map per call. Must preserve unmanaged elements.

## Common Workflows

### Adding a New Resource

1. **Create Schema**: `internal/provider/<name>_resource.go`
   - Define schema with `schema.Schema{}`
   - Mark sensitive attributes: `Sensitive: true`
   - Use profile factory for authentication profiles

2. **Implement CRUD Methods**:
   - Create(): Use profile factory, wrap with RetryWithBackoff, convert errors with MapError
   - Read(): Handle 404 as deleted (drift detection)
   - Update(): Use Read-Modify-Write pattern if modifying shared resources
   - Delete(): Use `delete_workarounds.go` functions (see Critical Constraints)

3. **Add Tests**: `internal/provider/<name>_resource_test.go`
   - Acceptance tests with `TF_ACC=1`
   - Test CRUD lifecycle, ImportState, ForceNew behavior

4. **Create Examples**: `examples/resources/<name>/`
   - Basic usage example
   - Complete configuration example

5. **Generate Documentation**:
   ```bash
   tfplugindocs generate
   ```

6. **CRUD Validation**:
   - **Automated**: `make test-crud DESC=<resource-description>`
   - **Manual**: Follow `examples/testing/TESTING-GUIDE.md` for detailed workflow
   - Verify all validation checks pass (CREATE → READ → UPDATE → DELETE cycle)

### Fixing a Resource Bug

1. **Identify Affected Method**: Create/Read/Update/Delete
2. **Check Patterns**:
   - Using profile factory? (if auth-related)
   - Using delete workarounds? (if Delete method)
   - Using RetryWithBackoff? (if API calls)
   - Using Read-Modify-Write? (if policy updates)
3. **Verify SDK Mappings**: `docs/sdk-integration.md`
4. **Add/Update Test**: Reproduce bug in acceptance test
5. **CRUD Validation**: Run TESTING-GUIDE.md workflow

### Debugging Test Failures

1. **Enable Verbose Logging**:
   ```bash
   TF_LOG=DEBUG TF_ACC=1 go test ./... -v -run TestAccResourceName
   ```

2. **Common Issues**:
   - 401 Unauthorized → Check token refresh (see `docs/troubleshooting.md`)
   - 404 Not Found → Resource deleted externally (drift)
   - Nil pointer panic on Delete → Using SDK methods directly (use delete_workarounds.go)
   - Perpetual drift → Check profile pointer clearing in Read() method

3. **Consult References**:
   - API errors: `docs/troubleshooting.md`
   - SDK limitations: `docs/development/design-decisions.md`
   - Field mappings: `docs/sdk-integration.md`

## Critical Constraints

### ARK SDK v1.5.0 Limitations

1. **DELETE Panic Bug** ⚠️ **CRITICAL**
   - **Problem**: `DeleteDatabase()`, `DeleteSecret()`, `DeletePolicy()` pass nil body → panic in doRequest()
   - **Root Cause**: `pkg/common/ark_client.go:556-576` doesn't handle nil body pointer
   - **Solution**: Use `internal/client/delete_workarounds.go` functions:
     ```go
     // ✅ CORRECT
     err := client.DeleteDatabaseWorkspaceDirect(ctx, providerData.AuthContext, databaseID)
     err := client.DeleteSecretDirect(ctx, providerData.AuthContext, secretID)
     err := client.DeletePolicyDirect(ctx, providerData.AuthContext, policyID)

     // ❌ WRONG - Will panic
     err := siaAPI.WorkspacesDB().DeleteDatabase(databaseID)
     ```
   - **TODO**: Remove workaround when ARK SDK v1.6.0+ fixes nil body handling

2. **No Context Support in Authenticate()**
   - Cannot cancel authentication mid-flight via context
   - First parameter is `*ArkProfile` (optional), NOT `context.Context`

3. **No Structured Errors**
   - SDK returns generic `error` interface with string messages
   - Use `internal/client.MapError()` for error classification

4. **15-Minute Token Expiration**
   - SDK handles automatic token refresh
   - In-memory profile pattern (stateless, container-friendly)

### Database Workspace Constraints

1. **All Cloud Providers Use "FQDN/IP" Target Set**
   - Don't create cloud-specific policy target logic
   - `cloud_provider` attribute is metadata only
   - Validated in ARK SDK: `choices:"FQDN/IP"` annotation

2. **secret_id is Functionally Required**
   - Schema: Optional (SDK allows it)
   - Reality: Required for ZSP/JIT access provisioning
   - Document this requirement in examples

### Policy Management Constraints

1. **UpdatePolicy() Accepts ONE Workspace Type Only**
   - Can't update database targets and VM targets in same call
   - Use Read-Modify-Write pattern to preserve unmanaged assignments

2. **Composite ID Formats**
   - Principal assignments: `policy-id:principal-id:principal-type` (3-part)
   - Database assignments: `policy-id:database-id` (2-part)
   - Parsing: Use `helpers/composite_ids.go` utilities

## Anti-Patterns (What NOT to Do)

❌ **Don't bypass profile factory** - Creates validation inconsistencies and 410 LOC duplication
❌ **Don't log sensitive data** - Passwords, tokens, client_secret, aws_secret_access_key
❌ **Don't use SDK Delete methods directly** - Use `delete_workarounds.go` (prevents panics)
❌ **Don't assume cloud providers need different target sets** - All use "FQDN/IP"
❌ **Don't create ad-hoc test configs** - Use `examples/testing/TESTING-GUIDE.md` templates
❌ **Don't modify template files directly** - Copy to `/tmp` first, then customize
❌ **Don't skip Read-Modify-Write for policies** - Causes assignment overwrites
❌ **Don't assume SDK behavior** - Always verify in SDK source code

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

### Error Handling
- Use `internal/client.MapError()` for Terraform diagnostics
- Wrap operations with `internal/client.RetryWithBackoff()`
- Classify errors by type (auth, permission, network, etc.)
- Provide actionable error messages with guidance

## Testing Strategy

### Primary: Acceptance Tests
- Test against real SIA API when `TF_ACC=1`
- Verify CRUD operations end-to-end
- Test ImportState functionality
- Test ForceNew behavior and drift detection
- Mock only when necessary (prefer real integration tests)

### Selective: Unit Tests
- Complex validators only
- Error classification logic
- Retry logic
- Helper utilities

### Acceptance Test Prerequisites

**Quick Prerequisites Check** before running `TF_ACC=1 go test ./...`:
- [ ] Environment variables: `CYBERARK_USERNAME`, `CYBERARK_CLIENT_SECRET`, `TF_ACC=1`
- [ ] Service account scopes: `sia`, `identity`
- [ ] CyberArk tenant with SIA enabled

**For complete prerequisites** (test data, cloud providers, troubleshooting), see `examples/testing/TESTING-GUIDE.md`

### CRUD Testing Standards

**CANONICAL REFERENCE**: `examples/testing/TESTING-GUIDE.md`

**ALL CRUD testing MUST follow** `examples/testing/TESTING-GUIDE.md`. This is the **single source of truth** for:
- Test configuration templates
- Testing workflow (CREATE → READ → UPDATE → DELETE)
- Validation checklists
- Resource dependency patterns
- Troubleshooting procedures

**Template Usage**:
1. Start from templates in `examples/testing/crud-test-*.tf`
2. Copy to `/tmp/sia-crud-validation-<timestamp>/`
3. Never modify templates directly
4. Update TESTING-GUIDE.md if resource behavior changes

**Testing Checklist (Before Committing)**:
- [ ] Run full CRUD cycle using TESTING-GUIDE.md workflow
- [ ] All validation checks pass (validation_summary outputs)
- [ ] Update TESTING-GUIDE.md if resource behavior changed
- [ ] Update template files if new dependencies added
- [ ] Document any new troubleshooting scenarios

## Commands

**Quick Reference**: Run `make help` to see all available commands

### Build & Test
```bash
make build                              # Build provider binary
make install                            # Install locally for Terraform development
make test                               # Run unit tests
make testacc                            # Run acceptance tests (requires TF_ACC=1)
make test-crud DESC=policy-assignment   # Run automated CRUD validation
make check-env                          # Verify environment variables are set
```

### Code Quality
```bash
make fmt                                # Format Go code
make lint                               # Run golangci-lint
```

### Development Workflow
```bash
make deps                               # Download and tidy Go dependencies
make generate                           # Generate provider documentation
make clean                              # Clean build artifacts
```

### Manual Commands (Advanced)

If you need more control, use Go commands directly:

```bash
# Build
go build -v

# Run specific tests
go test ./internal/client/... -v
go test ./internal/provider -v -run TestAccResourceName

# Acceptance tests with verbose logging
TF_LOG=DEBUG TF_ACC=1 go test ./internal/provider -v -run TestAccResourceName

# Install to custom location
go install

# Dependencies
go mod tidy
go mod download
```

## Release & Distribution

### Version Management

**Versioning**: Follow [Semantic Versioning 2.0.0](https://semver.org/)
- **Major** (v1.0.0): Breaking changes (incompatible API changes)
- **Minor** (v0.X.0): New features, backward compatible
- **Patch** (v0.0.X): Bug fixes, backward compatible

**Pre-1.0 Status**: Currently v0.1.0 - breaking changes are acceptable before 1.0 release

### Release Checklist

Before creating a new release:

- [ ] All tests passing (`make test && make testacc`)
- [ ] CRUD validation complete for affected resources (`make test-crud DESC=...`)
- [ ] Code formatted and linted (`make fmt && make lint`)
- [ ] Documentation generated (`make generate` or `tfplugindocs generate`)
- [ ] CHANGELOG.md updated with release notes
- [ ] Version numbers bumped (if applicable)
- [ ] Git tag created: `git tag v0.X.X && git push origin v0.X.X`
- [ ] GitHub release created with release notes

### CHANGELOG.md Format

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```
[0.X.X] - YYYY-MM-DD

Added:
- New features

Changed:
- Changes in existing functionality

Fixed:
- Bug fixes

Breaking Changes:
- Incompatible changes (pre-1.0 only)
```

### Future: Terraform Registry Publication

When ready for public distribution:

1. **Prerequisites**:
   - GitHub repository public
   - GPG key for signing releases
   - Terraform Registry account

2. **CI/CD Setup**:
   - GitHub Actions for automated builds
   - Automated testing on PRs
   - Release automation

3. **Registry Publishing**:
   - Follow [Terraform Registry publishing guide](https://www.terraform.io/docs/registry/providers/publishing.html)
   - Sign releases with GPG
   - Follow provider naming conventions

## Technical References

For detailed technical information, see:
- **Design Decisions**: `docs/development/design-decisions.md` - Active technologies, SDK limitations, breaking changes
- **SDK Integration**: `docs/sdk-integration.md` - ARK SDK patterns and field mappings
- **Troubleshooting**: `docs/troubleshooting.md` - Common issues and solutions
- **Development History**: `docs/development-history.md` - Complete development timeline and architectural decisions

## Known TODOs in Codebase

**Quick Scan**: `rg "TODO|FIXME" --glob "*.go"`

**Current Count**: 8 TODOs across 6 files (as of 2025-10-30)

### High-Priority TODOs

Track critical items as GitHub Issues for better visibility and prioritization:

| Priority | TODO | File | Blocked By | Notes |
|----------|------|------|------------|-------|
| **P1** | Remove delete_workarounds.go | `internal/client/delete_workarounds.go` | ARK SDK v1.6.0+ release | Critical workaround for nil body panic bug |
| **P2** | Add conditional validators for secret auth types | `internal/provider/secret_resource.go` | SDK field verification | Enforce required fields per auth type |
| **P3** | Complete profile_factory test coverage | `internal/provider/profile_factory_test.go` | - | Increase test coverage for all 6 auth methods |

**Recommendation**: Create GitHub Issues for P1 and P2 TODOs to track SDK dependency and coordinate with upstream ARK SDK team.

**Detailed TODO Report** (with context):
```bash
rg "TODO|FIXME" --glob "*.go" -A 2 -B 1
```
