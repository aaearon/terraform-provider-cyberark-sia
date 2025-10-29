# Contributing to terraform-provider-cyberark-sia

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.25.0+
- Make
- ARK SDK golang v1.5.0 (automatically managed via go.mod)
- CyberArk SIA tenant access (for acceptance testing)

### Building

```bash
# Clone repository
git clone https://github.com/aaearon/terraform-provider-cyberark-sia
cd terraform-provider-cyberark-sia

# Build provider
go build -v

# Install locally for testing
go install
```

### Testing

```bash
# Unit tests
go test ./... -v

# Acceptance tests (requires CyberArk SIA tenant)
export TF_ACC=1
export CYBERARK_USERNAME="your-username@cyberark.cloud.XXXX"
export CYBERARK_CLIENT_SECRET="your-secret"
go test ./... -v
```

See [TESTING.md](TESTING.md) for comprehensive testing guide.

## Code Style

### Go Standards

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write godoc comments for exported functions
- **NEVER log sensitive data** (passwords, tokens, secrets)

### Formatting

```bash
# Format code
go fmt ./...
gofmt -w .

# Run linter
golangci-lint run
```

## Project Structure

```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform resource implementations
│   │   ├── profile_factory.go    # Authentication profile factory
│   │   └── helpers/               # Shared utilities
│   │       ├── id_conversion.go   # Database ID conversion
│   │       └── composite_ids.go   # Composite ID parsing/building
│   ├── client/          # ARK SDK wrappers, retry, error handling
│   ├── models/          # Terraform state models
│   └── validators/      # Custom Terraform validators
├── examples/            # Terraform HCL examples
│   ├── complete/        # Complete working examples
│   ├── provider/        # Provider configuration examples
│   ├── resources/       # Per-resource examples
│   └── testing/         # CRUD testing framework templates
├── docs/                # Documentation
│   ├── guides/          # User guides
│   ├── resources/       # Resource documentation
│   └── development/     # Design decisions, implementation summaries
└── specs/               # Feature specifications
```

## Adding a New Resource

### Step 1: Create Resource File

Create `internal/provider/<resource_name>_resource.go`:

```go
package provider

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure interface compliance
var _ resource.Resource = &ResourceName{}
var _ resource.ResourceWithConfigure = &ResourceName{}
var _ resource.ResourceWithImportState = &ResourceName{}

type ResourceName struct {
    providerData *ProviderData
}

func NewResourceName() resource.Resource {
    return &ResourceName{}
}

func (r *ResourceName) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_resource_name"
}

func (r *ResourceName) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // Define schema
}

func (r *ResourceName) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Configure provider data
}

func (r *ResourceName) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Implement create
}

func (r *ResourceName) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implement read
}

func (r *ResourceName) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Implement update
}

func (r *ResourceName) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Implement delete
}

func (r *ResourceName) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Implement import
}
```

### Step 2: Register Resource

Add to `internal/provider/provider.go`:

```go
func (p *CyberArkSIAProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewResourceName,
        // ... other resources
    }
}
```

### Step 3: Add Tests

Create `internal/provider/<resource_name>_resource_test.go`:

```go
func TestResourceName_CRUD(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccResourceConfig_basic(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("cyberarksia_resource_name.test", "id"),
                ),
            },
            // Import
            {
                ResourceName:      "cyberarksia_resource_name.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

### Step 4: Add Documentation

- `docs/resources/<resource_name>.md` - Resource documentation
- `examples/resources/cyberarksia_<resource_name>/` - HCL examples

### Step 5: Update Testing Guide

Add resource patterns to `examples/testing/TESTING-GUIDE.md`.

## Coding Conventions

### Error Handling

Use `internal/client.MapError()` for Terraform diagnostics:

```go
if err != nil {
    resp.Diagnostics.Append(client.MapError(err, "create certificate")...)
    return
}
```

Wrap API operations with `internal/client.RetryWithBackoff()`:

```go
err := client.RetryWithBackoff(ctx, &client.RetryConfig{
    MaxRetries: client.DefaultMaxRetries,  // 3 retries
    BaseDelay:  client.BaseDelay,          // 500ms
    MaxDelay:   client.MaxDelay,           // 30s
}, func() error {
    return siaAPI.WorkspacesDB().AddDatabase(...)
})
```

### Logging

Use `terraform-plugin-log/tflog` for structured logging:

```go
tflog.Info(ctx, "Operation succeeded", map[string]interface{}{
    "operation": "create",
    "resource_id": id,
})
```

**NEVER log sensitive data**:
- ❌ password, client_secret, aws_secret_access_key, tokens
- ✅ Log resource IDs, operation names, non-sensitive metadata

### Helper Usage

Use shared helpers from `internal/provider/helpers/`:

```go
import "github.com/aaearon/terraform-provider-cyberark-sia/internal/provider/helpers"

// ID conversion
databaseIDInt, ok := helpers.ConvertDatabaseIDToInt(databaseID, &resp.Diagnostics, path.Root("database_workspace_id"))
if !ok {
    return
}

// Composite IDs
id := helpers.BuildCompositeID(policyID, databaseID)
policyID, databaseID, err := helpers.ParsePolicyDatabaseID(id)
```

Use `internal/provider/profile_factory` for authentication profiles:

```go
// Build profile from Terraform plan
profile := BuildAuthenticationProfile(ctx, authMethod, &data, &resp.Diagnostics)
if resp.Diagnostics.HasError() {
    return
}

// Set profile on SDK object
SetProfileOnInstanceTarget(instanceTarget, authMethod, profile)
```

### Keep Resources Focused

Resource files should focus on CRUD orchestration. Extract complex logic to:
- `internal/provider/helpers/` - Shared utilities
- `internal/provider/profile_factory.go` - Authentication profiles
- `internal/client/` - SDK wrappers and retry logic

## Testing Philosophy

- **Primary**: Acceptance tests (test against real SIA API)
- **Selective**: Unit tests for complex validators and helpers
- **No Mocks**: Prefer real integration tests over mocks
- **CRUD Testing Framework**: Use `examples/testing/TESTING-GUIDE.md` for systematic testing

### Running Tests

```bash
# Unit tests (fast)
go test ./internal/provider/helpers/... -v

# Acceptance tests (requires real SIA API)
TF_ACC=1 go test ./internal/provider -run TestPolicyDatabaseAssignment -v
```

## Pull Request Process

### Before Submitting

1. ✅ Create feature branch from `main`
2. ✅ Write clear commit messages
3. ✅ Run tests: `go test ./... -v`
4. ✅ Run linter: `golangci-lint run`
5. ✅ Update documentation
6. ✅ Test CRUD operations using `examples/testing/TESTING-GUIDE.md`

### PR Checklist

- [ ] All tests pass (unit + acceptance)
- [ ] Code follows Go conventions and passes linter
- [ ] No sensitive data logged (passwords, tokens, secrets)
- [ ] Documentation updated (`docs/resources/`, `examples/`)
- [ ] `TESTING.md` or `examples/testing/TESTING-GUIDE.md` updated (if resource added)
- [ ] Commit messages are clear and descriptive
- [ ] PR description explains changes and rationale

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

**Example**:
```
feat: add policy database assignment resource

Implement cyberarksia_policy_database_assignment resource to manage
database assignments to existing SIA access policies. Follows AWS
Security Group Rule pattern (separate resource per database).

Features:
- Full CRUD operations with idempotency
- All 6 SIA authentication methods
- Read-modify-write pattern (preserves UI-managed databases)
- Import support with composite ID format

Closes #123
```

## Questions?

- Check [`docs/development/design-decisions.md`](docs/development/design-decisions.md) for technical decisions
- See [`docs/troubleshooting.md`](docs/troubleshooting.md) for common issues
- Review [`TESTING.md`](TESTING.md) for test patterns
- Read [`CLAUDE.md`](CLAUDE.md) for development guidelines
