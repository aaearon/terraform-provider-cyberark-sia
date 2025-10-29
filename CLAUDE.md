# terraform-provider-cyberark-sia Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-29 (Refactoring Complete + Critical Bug Fix)

## Project Structure
```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform provider implementation
│   │   ├── profile_factory.go    # Authentication profile factory (NEW 2025-10-29)
│   │   └── helpers/               # Shared utilities (NEW 2025-10-29)
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

### Error Handling
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

## Recent Changes

### Phase 1 & 2 Refactoring Complete (2025-10-29)

**Profile Factory Refactoring** ✅:
- Created `internal/provider/profile_factory.go` (443 lines)
- Eliminated 410 LOC from `policy_database_assignment_resource.go` (1,177 → 767 lines / 35% reduction)
- Zero duplicated profile handling code across Create(), Read(), Update()
- All 6 authentication methods centralized

**Helper Extraction** ✅:
- Created `internal/provider/helpers/id_conversion.go` (30 lines)
- Created `internal/provider/helpers/composite_ids.go` (50 lines)
- Shared utilities ready for use across all resources

**Critical Bug Fix** (Discovered by Codex Code Review):
- Fixed perpetual Terraform drift when switching authentication methods
- Root cause: `ParseAuthenticationProfile()` didn't clear stale profile pointers before repopulating state
- Added explicit profile pointer clearing in `profile_factory.go:270-278`
- Added type safety panic checks in `SetProfileOnInstanceTarget()`

**Final Metrics**:
- Main resource file: 1,177 → 767 lines (-410 lines / 35%)
- Net change: -297 lines saved
- Critical bugs: 1 → 0 (fixed)
- All tests passing ✅

### Principal Lookup Data Source (2025-10-29)

**Feature**: `cyberarksia_principal` Terraform data source for looking up users, groups, and roles by name.

**Implementation**:
- Universal support for Cloud Directory, Federated Directory, and Active Directory
- Hybrid lookup strategy: Fast path for users (< 1s), fallback for all types (< 2s)
- Handles all principal types (USER, GROUP, ROLE) with single implementation
- Zero manual UUID lookups required

**Files**:
- `internal/provider/principal_data_source.go` - Data source implementation (388 lines)
- `internal/client/identity_client.go` - Identity API client wrapper (26 lines)
- `docs/data-sources/principal.md` - User documentation (225 lines)

**Key Design Decision**: Hybrid lookup strategy driven by ARK SDK API limitations (UserByName doesn't return directory info, ListDirectoriesEntities only searches DisplayName).

### Policy Database Assignment Bug Fix - Azure Database Support (2025-10-27)

**Critical Fix**: Fixed policy database assignment to correctly handle ALL cloud providers (Azure, AWS, GCP, on-premise).

**Root Cause**: The provider incorrectly assumed different cloud providers use different policy target sets. The actual API behavior is that **ALL database workspaces use the `"FQDN/IP"` target set**, regardless of the `cloud_provider` attribute.

**Key Learning**: The `cloud_provider` attribute on database workspaces is **metadata only** and does not affect policy target set selection.

### Schema Correction: Secret-Database Workspace Relationship (2025-10-27)

**Changes Made**:
1. **Secret Resource** - Removed `database_workspace_id` field (was provider-fabricated, didn't exist in SDK)
2. **Database Workspace Resource** - Made `secret_id` required (functionally essential for ZSP/JIT access)

## Technical Reference

For detailed technical information, see:
- **Design Decisions**: `docs/development/design-decisions.md` - Active technologies, SDK limitations, breaking changes
- **SDK Integration**: `docs/sdk-integration.md` - ARK SDK patterns and field mappings
- **Troubleshooting**: `docs/troubleshooting.md` - Common issues and solutions
- **Development History**: `docs/development-history.md` - Complete development timeline
