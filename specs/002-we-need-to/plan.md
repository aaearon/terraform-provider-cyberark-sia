# Implementation Plan: Replace ARK SDK with Custom OAuth2 Implementation

**Branch**: `002-we-need-to` | **Date**: 2025-10-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-we-need-to/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Replace all ARK SDK (`github.com/cyberark/ark-sdk-golang`) usage with custom OAuth2 client implementation to fix authentication failures. The ARK SDK creates ID tokens instead of access tokens, causing 401 Unauthorized errors for database workspace and secret resources. The certificate resource has already been successfully migrated to custom OAuth2 implementation as a proven reference pattern.

**Technical Approach**:
1. Create custom OAuth2 HTTP clients for database workspaces and secrets following the existing `certificates_oauth2.go` pattern
2. Define custom model types matching SIA API contracts to replace ARK SDK model dependencies
3. Update provider initialization to use only custom OAuth2 clients
4. Remove all ARK SDK imports and dependencies from the codebase
5. Update documentation to reflect custom OAuth2 implementation

## Technical Context

**Language/Version**: Go 1.25.0
**Primary Dependencies**:
- Terraform Plugin Framework v1.16.1 (Plugin Framework v6)
- Terraform Plugin Log v0.9.0
- golang-jwt/jwt/v5 (JWT parsing for token claims)
- Standard library (net/http, encoding/json, context)

**Storage**: N/A (Terraform provider - state managed by Terraform)
**Testing**:
- Go standard testing (`go test`)
- Terraform acceptance testing framework (`terraform-plugin-testing`)
- Unit tests for OAuth2 client, error handling, retry logic
- Acceptance tests for resource CRUD operations

**Target Platform**: Linux/macOS/Windows (Terraform CLI context)
**Project Type**: Single project (Terraform provider plugin)
**Performance Goals**:
- Provider initialization <3 seconds (including OAuth2 authentication)
- API operations complete within 30 seconds (including retries)
- Support minimum 10 concurrent resource operations

**Constraints**:
- No breaking changes to Terraform resource schemas (HCL configurations must work unchanged)
- Access tokens expire in 3600 seconds (1 hour)
- Retry logic: 3 attempts with exponential backoff (500ms base, 30s max delay)
- No sensitive data logging (client_secret, tokens, passwords)

**Scale/Scope**:
- 3 resource types (certificates, database workspaces, secrets)
- ~15 ARK SDK import statements to remove
- ~5 Go files to create (OAuth2 clients + model types)
- ~10 Go files to modify (provider, resources, validators)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality and SOLID Principles ✅

- **Single Responsibility**: Each OAuth2 client (certificates, database workspaces, secrets) handles one resource type
- **Open/Closed**: New resource types can be added without modifying existing clients
- **Dependency Inversion**: Resources depend on OAuth2 client interfaces, not concrete ARK SDK types
- **Rationale**: Replacing ARK SDK with custom clients actually improves SOLID adherence by removing tight coupling to SDK internals

### II. Maintainability First ✅

- **Clear naming**: `DatabaseWorkspaceClientOAuth2`, `SecretsClientOAuth2` (descriptive, no abbreviations)
- **Explicit error handling**: All HTTP errors mapped to actionable Terraform diagnostics
- **Go standards**: Following Effective Go, using `gofmt` and `golangci-lint`
- **Terraform standards**: Following Plugin Framework patterns from certificate resource
- **Rationale**: Custom implementation is more maintainable than debugging SDK quirks (ID token vs access token issue)

### III. LLM-Driven Development ✅

- **Clear context**: Certificate OAuth2 implementation provides concrete reference pattern
- **Iterative refinement**: Implementation will follow proven pattern with incremental verification
- **Documentation**: oauth2-authentication-fix.md documents architecture decision and rationale
- **Rationale**: LLM can generate database workspace and secret clients by pattern-matching against certificates_oauth2.go

### IV. Light-Touch Testing ✅

- **Test pyramid compliance**:
  - Many: Unit tests for OAuth2 clients, error mapping, retry logic (fast, isolated)
  - Some: Integration tests for token refresh and concurrent operations (medium scope)
  - Few: Acceptance tests for complete CRUD workflows (end-to-end)
- **Focus on value**: Testing authentication flow and API contracts, not framework behavior
- **Rationale**: Existing certificate tests prove pattern works; database workspace/secret tests will follow same structure

### V. Documentation as Code ✅

- **Continuous**: Each user story (P1-P4) includes documentation updates
- **Version-controlled**: All docs in Git alongside code changes
- **Actionable**: oauth2-authentication-fix.md provides working authentication flow
- **Current**: Documentation explicitly lists "REMAINING WORK: Database Workspace and Secret resources need migration"
- **Rationale**: Migration plan documented before implementation; will be updated as work progresses

### Quality Gates Summary

**Status**: ✅ **PASSED** - All constitution principles satisfied

| Principle | Compliance | Notes |
|-----------|------------|-------|
| SOLID Principles | ✅ Pass | Improves separation of concerns vs ARK SDK |
| Maintainability | ✅ Pass | Custom code more maintainable than SDK debugging |
| LLM-Driven Development | ✅ Pass | Proven reference pattern available |
| Light-Touch Testing | ✅ Pass | Test pyramid structure planned |
| Documentation | ✅ Pass | Architecture decision documented |

**No complexity violations** - No justifications required in Complexity Tracking section

## Project Structure

### Documentation (this feature)

```
specs/002-we-need-to/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── database-workspaces-api.yaml  # OpenAPI spec for WorkspacesDB API
│   └── secrets-api.yaml              # OpenAPI spec for SecretsDB API
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
terraform-provider-cyberark-sia/
├── internal/
│   ├── client/                         # API client implementations
│   │   ├── oauth2.go                   # ✅ OAuth2 authentication (existing)
│   │   ├── certificates_oauth2.go      # ✅ Certificates client (existing - reference pattern)
│   │   ├── database_workspace_oauth2.go # 🆕 Database workspace client (to create)
│   │   ├── secrets_oauth2.go           # 🆕 Secrets client (to create)
│   │   ├── errors.go                   # ✅ Error mapping (existing)
│   │   ├── retry.go                    # ✅ Retry logic (existing)
│   │   ├── auth.go                     # ❌ ARK SDK auth wrapper (to remove)
│   │   ├── certificates.go             # ❌ ARK SDK certificates wrapper (to remove)
│   │   └── sia_client.go               # ❌ ARK SDK SIA client wrapper (to remove)
│   │
│   ├── models/                         # Data models
│   │   ├── database_workspace.go       # 🆕 Database workspace models (to create)
│   │   └── secret.go                   # 🆕 Secret models (to create)
│   │
│   ├── provider/                       # Terraform provider implementation
│   │   ├── provider.go                 # 🔄 Provider configuration (to modify - remove ARK SDK init)
│   │   ├── resource_certificate.go     # ✅ Certificate resource (existing - uses OAuth2)
│   │   ├── resource_database_workspace.go # 🔄 Database workspace resource (to modify)
│   │   └── resource_secret.go          # 🔄 Secret resource (to modify)
│   │
│   └── validators/                     # Schema validators
│       └── database_engine_validator.go # 🔄 Engine validator (to modify - remove ARK SDK types)
│
├── docs/
│   ├── oauth2-authentication-fix.md    # ✅ Architecture decision record (existing)
│   ├── sdk-integration.md              # 🔄 ARK SDK integration (to update/deprecate)
│   └── CERTIFICATES-API.md             # ✅ Certificates API reference (existing)
│
├── examples/                           # Terraform configuration examples
│   ├── certificate/                    # ✅ Certificate examples (existing)
│   ├── database_workspace/             # 🔄 Database workspace examples (to verify/update)
│   └── secret/                         # 🔄 Secret examples (to verify/update)
│
└── go.mod                              # 🔄 Go dependencies (to update - remove ARK SDK)
```

**Structure Decision**: Single project (Terraform provider plugin) following standard Terraform Plugin Framework structure. The `internal/` directory contains provider implementation with three main subdirectories:

1. **client/** - API client implementations (OAuth2 authentication + resource-specific HTTP clients)
2. **models/** - Data models for API request/response serialization
3. **provider/** - Terraform resource and provider implementations

**Key Changes**:
- **NEW**: `internal/models/` directory for custom model types (replacing ARK SDK models)
- **NEW**: OAuth2 client files for database workspaces and secrets
- **REMOVE**: ARK SDK wrapper files (`auth.go`, `certificates.go`, `sia_client.go`)
- **MODIFY**: Provider and resource files to use custom OAuth2 clients

## Complexity Tracking

*No constitution violations - this section intentionally left empty*

**Rationale**: The ARK SDK replacement actually **reduces** complexity by:
1. Eliminating dependency on third-party SDK with known authentication issues
2. Providing full control over HTTP requests and error handling
3. Following established patterns (certificate OAuth2 implementation)
4. Improving code maintainability (explicit HTTP client vs SDK abstraction)

No additional complexity justifications required.

---

## Implementation Phases Summary

### ✅ Phase 0: Research (COMPLETE)

**Output**: `research.md`

**Key Findings**:
- OAuth2 authentication pattern proven by certificate resource
- SIA API contracts documented via ARK SDK source code
- HTTP client best practices from HashiCorp documentation
- Model design patterns established
- Incremental dependency removal strategy defined

**Status**: All NEEDS CLARIFICATION items resolved

---

### ✅ Phase 1: Design & Contracts (COMPLETE)

**Outputs**:
- `data-model.md` - Complete model type specifications
- `contracts/database-workspaces-api.yaml` - OpenAPI 3.0 spec
- `contracts/secrets-api.yaml` - OpenAPI 3.0 spec
- `quickstart.md` - Step-by-step implementation guide
- `CLAUDE.md` - Updated agent context (language, database, project type)

**Key Deliverables**:
1. **Data Models**: 6 model types defined (DatabaseWorkspaceCreateRequest, DatabaseWorkspaceUpdateRequest, DatabaseWorkspace, SecretCreateRequest, SecretUpdateRequest, Secret)
2. **API Contracts**: 10 endpoints documented with request/response schemas
3. **Quickstart Guide**: 4 implementation phases with code snippets and verification steps

**Status**: All design artifacts complete and ready for implementation

---

### ⏭️ Phase 2: Task Generation (NEXT)

**Command**: `/speckit.tasks`

**Prerequisites**: ✅ All met (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md complete)

**Expected Output**: `tasks.md` with dependency-ordered implementation tasks

**Task Categories**:
1. Create custom model types (`internal/models/`)
2. Create OAuth2 clients (`internal/client/*_oauth2.go`)
3. Update provider initialization (`internal/provider/provider.go`)
4. Update resource implementations (`internal/provider/resource_*.go`)
5. Update validators (`internal/validators/`)
6. Remove ARK SDK files and dependencies
7. Update documentation
8. Testing and verification

---

## Planning Phase Completion Report

**Branch**: `002-we-need-to`
**Spec**: `/home/tim/terraform-provider-cyberark-sia/specs/002-we-need-to/spec.md`
**Plan**: `/home/tim/terraform-provider-cyberark-sia/specs/002-we-need-to/plan.md`

### Artifacts Generated

| Artifact | Path | Status | Lines |
|----------|------|--------|-------|
| Research | `specs/002-we-need-to/research.md` | ✅ Complete | ~350 |
| Data Model | `specs/002-we-need-to/data-model.md` | ✅ Complete | ~450 |
| DB Workspaces API | `specs/002-we-need-to/contracts/database-workspaces-api.yaml` | ✅ Complete | ~450 |
| Secrets API | `specs/002-we-need-to/contracts/secrets-api.yaml` | ✅ Complete | ~350 |
| Quickstart Guide | `specs/002-we-need-to/quickstart.md` | ✅ Complete | ~650 |
| Implementation Plan | `specs/002-we-need-to/plan.md` | ✅ Complete | ~200 |

**Total**: 6 design artifacts, ~2450 lines of documentation

### Constitution Check Re-Validation

**Status**: ✅ **PASSED** (post-design)

All constitution principles remain satisfied after design phase:
- ✅ SOLID Principles: Model/client separation follows SRP and DIP
- ✅ Maintainability: Clear naming, explicit error handling, Go standards
- ✅ LLM-Driven: Reference patterns documented, iterative refinement planned
- ✅ Light-Touch Testing: Test pyramid structure defined in quickstart
- ✅ Documentation: All design artifacts version-controlled and actionable

**No new complexity violations introduced**

### Risk Assessment

| Risk | Mitigation | Residual Risk |
|------|------------|---------------|
| API contract mismatch | ARK SDK source code provides definitive reference | ⚠️ Low |
| Token refresh timing | Existing OAuth2 client handles expiration | ⚠️ Low |
| Model serialization errors | JSON struct tags match API exactly | ⚠️ Low |
| Test coverage gaps | Quickstart includes comprehensive test plan | ⚠️ Low |
| Breaking changes impact | No active users (documented assumption) | ✅ None |

**Overall Risk**: ⚠️ **LOW** - Proven pattern with clear implementation path

### Ready for Implementation

**Checklist**:
- [x] Feature specification complete (spec.md)
- [x] Constitution Check passed (no violations)
- [x] Research complete (all unknowns resolved)
- [x] Data models designed (complete field mappings)
- [x] API contracts documented (OpenAPI 3.0 specs)
- [x] Quickstart guide created (step-by-step implementation)
- [x] Agent context updated (CLAUDE.md)

**Next Command**: `/speckit.tasks` to generate dependency-ordered task list

---

**Plan Generated**: 2025-10-25
**Author**: Claude Code (Sonnet 4.5)
**Status**: ✅ Design Phase Complete - Ready for Task Generation

---

## Design Revision (Post-Gemini Review)

**Date**: 2025-10-25
**Reviewer**: Gemini 2.5 Pro (via clink)
**Status**: ✅ Simplified design approved

### Critical Simplifications Applied

Based on Gemini's review (see `design-reflection.md`), the following simplifications were applied:

1. **Model Structure** (-67% types):
   - Original: 6 types (CreateRequest, UpdateRequest, Response per resource)
   - Revised: 2 types (ONE struct per resource with pointers)
   - Pattern: AWS SDK for Go (not Java-style DTOs)

2. **DatabaseWorkspace Fields** (-80% initial complexity):
   - Original: 25 fields from day 1 (1:1 API mapping)
   - Revised: 5 MVP fields (name, provider_engine, endpoint, port, tags)
   - Strategy: Incremental field addition based on actual need

3. **Client Architecture** (-75% duplicated code):
   - Original: Separate OAuth2 clients with duplicated CRUD methods
   - Revised: Generic `RestClient` + thin wrappers (~50 lines each)
   - Benefit: ~90% of HTTP code shared

4. **Implementation Order** (Risk reduction):
   - Original: DatabaseWorkspace (complex) → Secret (simple)
   - Revised: Secret (simple) → DatabaseWorkspace MVP
   - Rationale: Prove pattern on smaller scope first

5. **Quickstart Guide** (-77% documentation):
   - Original: 650 lines covering all scenarios
   - Revised: ~150 lines with single use case
   - Move advanced examples to `examples/` directory

### Updated Architecture

```
internal/
├── models/
│   ├── database_workspace.go    # Single struct with pointers (MVP: 5 fields)
│   └── secret.go                # Single struct with pointers (6 fields)
│
└── client/
    ├── rest_client.go           # NEW: Generic HTTP client (~100 lines)
    ├── oauth2.go                # Existing: Token acquisition
    ├── secrets_client.go        # NEW: Thin wrapper (~50 lines)
    ├── database_workspace_client.go  # NEW: Thin wrapper (~50 lines)
    ├── errors.go                # Existing: Error mapping
    └── retry.go                 # Existing: Retry logic
```

### Complexity Reduction Summary

| Metric | Original | Revised | Reduction |
|--------|----------|---------|-----------|
| Model types | 6 | 2 | **-67%** |
| Initial fields (DatabaseWorkspace) | 25 | 5 | **-80%** |
| Client code per resource | ~200 lines | ~50 lines | **-75%** |
| Quickstart guide | 650 lines | 150 lines | **-77%** |
| **Overall Complexity** | High | **Low** | **~75%** |

### Key Insight from Gemini

> "The goal is to replace legacy SDK logic, not to build a new, all-encompassing SDK within the provider."

The revised design focuses on **simplicity and incrementalism** rather than comprehensive API coverage upfront.

---

**Planning Revision Complete**: 2025-10-25
**Next Step**: `/speckit.tasks` with simplified design
