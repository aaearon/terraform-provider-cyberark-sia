# Implementation Plan: Terraform Provider for CyberArk Secure Infrastructure Access

**Branch**: `001-build-a-terraform` | **Date**: 2025-10-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-build-a-terraform/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a Terraform provider for CyberArk Secure Infrastructure Access (SIA) that enables Infrastructure as Code management of database targets and strong accounts. The provider will leverage the official CyberArk ARK SDK for Golang (`github.com/cyberark/ark-sdk-golang`) for API communication and the HashiCorp Terraform Plugin Framework for provider implementation. Users will define database targets (onboarding existing databases from AWS RDS, Azure SQL, or on-premise) and strong accounts (credentials stored directly in SIA for ephemeral access provisioning) using declarative Terraform resources. Authentication to SIA APIs occurs via CyberArk Identity Security Platform Shared Services (ISPSS) OAuth2 client credentials flow with automatic token caching and proactive refresh.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `github.com/cyberark/ark-sdk-golang` - Official CyberArk SDK for API interactions
- `github.com/hashicorp/terraform-plugin-framework` - HashiCorp's framework for building Terraform providers
- `github.com/hashicorp/terraform-plugin-testing` - Acceptance testing framework
- `github.com/hashicorp/terraform-plugin-log` - Structured logging

**Storage**: N/A (Terraform state managed by Terraform core; provider is stateless)
**Testing**:
- Unit tests via `go test` with `-race` flag
- Acceptance tests via `terraform-plugin-testing` framework
- Mock SIA API for unit tests; optional real API for acceptance tests

**Target Platform**: Linux, macOS, Windows (cross-platform Go binary)
**Project Type**: Single Go module (Terraform provider binary)
**Performance Goals**:
- Token acquisition and caching: <2s initial, <100ms cached
- Resource CRUD operations: <30s per operation (network-dependent)
- Parallel resource creation: support 10+ concurrent operations
- Token refresh: proactive at 80% lifetime (~12 min for 15 min tokens)

**Constraints**:
- Bearer token lifecycle: 15-minute expiration requiring proactive refresh
- No direct database connectivity validation (SIA validates on first access)
- Terraform Plugin Protocol version 6 (latest)
- Read-only access to cloud provider databases (AWS/Azure resources managed by their respective providers)

**Scale/Scope**:
- Support 7 database types (SQL Server, Db2, MariaDB, MongoDB, MySQL, Oracle, PostgreSQL)
- Manage 50+ database targets per Terraform workspace
- 3 authentication methods for strong accounts (local, AD, AWS IAM)
- API integration: SIA REST API via ARK SDK

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Code Quality and SOLID Principles
- [x] **Single Responsibility**: Each resource (database target, strong account) has one purpose; provider handles only SIA integration
- [x] **Open/Closed**: Schema extensible via Terraform Plugin Framework; new database types/auth methods added without modifying core
- [x] **Dependency Inversion**: Depends on ARK SDK abstractions (auth interfaces), not concrete API implementations
- [x] **Simplicity**: No custom validation logic; SIA API validates database compatibility (user responsibility per requirements)

### Maintainability First
- [x] **Clear Naming**: Resource names follow Terraform conventions (`cyberark_sia_database_target`, `cyberark_sia_strong_account`)
- [x] **Error Handling**: All API errors mapped to actionable Terraform diagnostics per FR-027
- [x] **Go Standards**: Will use `gofmt`, `golangci-lint`, follow Effective Go guidelines
- [x] **Provider Standards**: Schema attributes include descriptions; validators are reusable

### Testing Strategy (Terraform Provider Best Practices)
- [x] **Acceptance Test Focus**: Primary testing via terraform-plugin-testing framework per HashiCorp standards
- [x] **Resource Lifecycle Coverage**: Test Create, Read, Update, Delete, ImportState for each resource
- [x] **Minimal Unit Tests**: Only for complex helper functions; avoid over-testing framework behavior
- [x] **Real API Testing**: Acceptance tests against real SIA API (controlled by TF_ACC env var)

### Documentation as Code
- [x] **Examples**: Terraform HCL examples for each resource type (AWS RDS, Azure SQL, on-premise)
- [x] **Schema Descriptions**: All attributes documented inline per Terraform conventions
- [x] **Architecture Decisions**: Token caching strategy, API integration patterns documented in `research.md`

**GATE STATUS**: ✅ **PASS** (with justified complexities - see Complexity Tracking section)

---

### Post-Design Re-Evaluation (Phase 1 Complete)

**Date**: 2025-10-15

✅ **Code Quality and SOLID Principles**:
- Single Responsibility maintained: Provider/resources/client separation clear in project structure
- Open/Closed validated: Schema extensible via framework attributes/validators
- Dependency Inversion confirmed: `internal/client/` abstracts ARK SDK, resources depend on client interface

✅ **Maintainability First**:
- Clear naming conventions established in data-model.md
- Error mapping contract defined with actionable diagnostics
- Go standards enforced via Makefile with golangci-lint
- All schema attributes have descriptions per contract

✅ **Light-Touch Testing**:
- Test pyramid validated: Unit tests for validators, acceptance tests for CRUD
- Mock strategy defined in API contract
- No framework behavior testing

✅ **Documentation as Code**:
- research.md documents all architecture decisions
- data-model.md provides complete entity specifications
- quickstart.md offers working examples
- Contract-first design in contracts/ directory

**Additional Complexity Identified**:
- Cross-attribute validation complexity (e.g., auth method compatibility) - justified by FR requirements and better UX than API-side errors alone

**FINAL GATE STATUS**: ✅ **PASS** - No constitution violations, all complexities justified

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider implementation
│   │   ├── provider_test.go         # Provider tests
│   │   ├── database_target_resource.go      # Database target resource
│   │   ├── database_target_resource_test.go
│   │   ├── strong_account_resource.go       # Strong account resource
│   │   ├── strong_account_resource_test.go
│   │   └── validators/              # Custom validators
│   │       ├── database_type.go
│   │       ├── port_range.go
│   │       └── version.go
│   ├── client/
│   │   ├── sia_client.go            # SIA API client wrapper
│   │   ├── auth.go                  # Authentication/token management
│   │   ├── database_targets.go      # Database target API operations
│   │   └── strong_accounts.go       # Strong account API operations
│   └── models/
│       ├── database_target.go       # Database target data models
│       └── strong_account.go        # Strong account data models
├── examples/
│   ├── provider/
│   │   └── provider.tf              # Provider configuration examples
│   ├── resources/
│   │   ├── database_target/
│   │   │   ├── aws_rds_postgresql.tf
│   │   │   ├── azure_sql_server.tf
│   │   │   └── onpremise_oracle.tf
│   │   └── strong_account/
│   │       ├── local_auth.tf
│   │       ├── ad_auth.tf
│   │       └── aws_iam_auth.tf
│   └── complete/
│       └── full_workflow.tf         # End-to-end example
├── docs/
│   ├── index.md                     # Provider documentation
│   ├── resources/
│   │   ├── database_target.md
│   │   └── strong_account.md
│   └── guides/
│       ├── authentication.md
│       └── migration.md
├── main.go                          # Provider entry point
├── go.mod
├── go.sum
├── .golangci.yml                    # Linter configuration
└── Makefile                         # Build/test automation
```

**Structure Decision**: Single Go module following Terraform provider conventions. The `internal/` directory contains provider-specific code (not importable by other modules). Resources are in `internal/provider/`, client logic in `internal/client/`. Examples follow HashiCorp's structure for provider documentation generation. This aligns with Terraform Plugin Framework best practices and Go project layout standards.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Provider data struct holding ARK SDK instances | **NOT ACTUAL COMPLEXITY** - This is standard Terraform Plugin Framework pattern where Provider.Configure() initializes client instances (ispAuth, uapClient) that resources access via provider reference. Required by framework architecture for sharing authenticated session across resources. | N/A - This is idiomatic Terraform provider design, not additional complexity |

**Note**: Previous "custom SIA client wrapper" concern withdrawn - the provider simply holds ARK SDK instances per framework conventions. ARK SDK handles all auth/refresh internally. Token refresh is managed by SDK's built-in mechanisms (can enable/disable caching via `NewArkISPAuth(cachingEnabled)`).
