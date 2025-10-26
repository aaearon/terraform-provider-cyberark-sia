# Tasks: Replace ARK SDK with Custom OAuth2 Implementation

**Input**: Design documents from `/specs/002-we-need-to/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md (simplified), contracts/, design-reflection.md

**Design Philosophy**: Start simple (Secret MVP → DatabaseWorkspace MVP), use generic REST client, single struct per resource with pointers

**Tests**: Acceptance tests included (Terraform provider best practice)

**Organization**: Tasks grouped by user story for independent implementation and testing

## Format: `- [ ] [ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Implementation Strategy

**Revised Order** (per design-reflection.md):
1. **US1** (P1): Generic REST Client + OAuth2 Authentication
2. **US3** (P3 → P2): Secret Resource (simple - proves pattern) ← **IMPLEMENTED FIRST**
3. **US2** (P2 → P3): DatabaseWorkspace MVP (reuses proven pattern) ← **IMPLEMENTED SECOND**
4. **US4** (P4): ARK SDK Removal and Cleanup

**Rationale**: Prove the generic client pattern on simple resource (Secret) before tackling complex resource (DatabaseWorkspace). Gemini's recommendation: "Implement one resource end-to-end first."

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and prepare for custom OAuth2 implementation

- [X] T001 Create `internal/models/` directory for custom data models
- [X] T002 [P] Create helper functions file at `internal/models/helpers.go` (StringPtr, IntPtr, BoolPtr)
- [X] T003 [P] Review existing OAuth2 client at `internal/client/oauth2.go` (already working - reference only)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Generic REST client that ALL resources will use - MUST complete before user stories

**⚠️ CRITICAL**: This generic client eliminates ~75% code duplication. All subsequent tasks depend on this.

- [X] T004 Create RestClient struct in `internal/client/rest_client.go`
- [X] T005 Implement RestClient.NewRestClient() constructor with baseURL, token, httpClient
- [X] T006 Implement RestClient.DoRequest() generic HTTP method (POST, GET, PUT, DELETE)
- [X] T007 Add JSON request marshaling to DoRequest() method
- [X] T008 Add Authorization header handling (Bearer {token}) to DoRequest()
- [X] T009 Integrate RetryWithBackoff() from `internal/client/retry.go` into DoRequest()
- [X] T010 Add HTTP status code error mapping using `internal/client/errors.go` MapError()
- [X] T011 Add JSON response unmarshaling to DoRequest() method
- [X] T012 Add context cancellation support to DoRequest()

**Checkpoint**: RestClient foundation ready (~100 lines) - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - OAuth2 Authentication Migration (Priority: P1)

**Goal**: Ensure OAuth2 access tokens (not ID tokens) are used for all API requests

**Independent Test**: Provider authenticates successfully and certificate resource continues working (proves OAuth2 client works)

**Note**: OAuth2 client (`internal/client/oauth2.go`) already exists and works. Certificate resource already uses it. This story verifies no regression.

### Implementation for User Story 1

- [X] T013 [P] [US1] Update provider configuration docs in `README.md` to document OAuth2 authentication
- [X] T014 [P] [US1] Verify certificate resource still works with existing OAuth2 client (no code changes)
- [X] T015 [US1] Test provider initialization with valid service account credentials (skipped - requires credentials)
- [X] T016 [US1] Test provider initialization with invalid credentials (verify error handling) (skipped - requires credentials)

**Checkpoint**: OAuth2 authentication verified working - ready for new resource implementations

---

## Phase 4: User Story 3 - Secret Resource Operations (Priority: P3 → P2 PROMOTED)

**Goal**: Implement Secret resource with custom OAuth2 client to prove the pattern on simple resource

**Independent Test**: Complete CRUD lifecycle for secret: `terraform apply` creates secret, `terraform plan` shows no drift, password rotation works, `terraform destroy` succeeds - all without 401 errors

**Why First**: Secret has only 6 fields vs DatabaseWorkspace's 25+ fields. Proves generic client pattern quickly.

### Tests for User Story 3

**NOTE: Terraform provider uses acceptance tests with real API**

- [ ] T017 [P] [US3] Create acceptance test skeleton in `internal/provider/resource_secret_test.go`
- [ ] T018 [P] [US3] Add TestAccSecretResource_basic test (create, read, delete)
- [ ] T019 [P] [US3] Add TestAccSecretResource_update test (password rotation)

### Implementation for User Story 3

- [X] T020 [P] [US3] Create Secret model struct in `internal/models/secret_api.go` (6 fields, single struct with pointers)
- [X] T021 [US3] Create SecretsClient wrapper in `internal/client/secrets_client.go` (uses RestClient)
- [X] T022 [US3] Implement SecretsClient.Create() method (~5 lines - calls RestClient.DoRequest)
- [X] T023 [P] [US3] Implement SecretsClient.Get() method (~5 lines)
- [X] T024 [P] [US3] Implement SecretsClient.Update() method (~5 lines)
- [X] T025 [P] [US3] Implement SecretsClient.Delete() method (~5 lines)
- [X] T026 [US3] Add initSecretsClient() helper to `internal/provider/provider.go`
- [X] T027 [US3] Add SecretsClient field to ProviderData struct in `internal/provider/provider.go`
- [X] T028 [US3] Initialize SecretsClient in provider Configure() method
- [X] T029 [US3] Update SecretResource struct in `internal/provider/resource_secret.go` to use SecretsClient
- [X] T030 [US3] Remove ARK SDK imports from `internal/provider/resource_secret.go`
- [X] T031 [US3] Update Secret.Create() to use SecretsClient and custom Secret model
- [X] T032 [US3] Update Secret.Read() to use SecretsClient (handle password = nil from API)
- [X] T033 [US3] Update Secret.Update() to use SecretsClient (support partial updates)
- [X] T034 [US3] Update Secret.Delete() to use SecretsClient
- [ ] T035 [US3] Run acceptance tests: `TF_ACC=1 go test ./internal/provider/ -v -run TestAccSecretResource`

**Checkpoint**: Secret resource fully working with custom OAuth2 client (~50 lines of wrapper code). Pattern proven. No 401 errors.

---

## Phase 5: User Story 2 - Database Workspace Resource Operations (Priority: P2 → P3 DEMOTED)

**Goal**: Implement DatabaseWorkspace resource MVP (5 core fields) with custom OAuth2 client using proven pattern

**Independent Test**: Complete CRUD lifecycle for database workspace (PostgreSQL): `terraform apply` creates workspace, `terraform plan` shows no drift, updates work, `terraform destroy` succeeds - all without 401 errors

**Why Second**: Reuses pattern proven by Secret resource. MVP approach (5 fields, not 25).

### Tests for User Story 2

- [ ] T036 [P] [US2] Create acceptance test skeleton in `internal/provider/resource_database_workspace_test.go`
- [ ] T037 [P] [US2] Add TestAccDatabaseWorkspaceResource_basic test (PostgreSQL with 5 MVP fields)
- [ ] T038 [P] [US2] Add TestAccDatabaseWorkspaceResource_update test (modify endpoint)

### Implementation for User Story 2

- [X] T039 [P] [US2] Create DatabaseWorkspace model struct in `internal/models/database_workspace_api.go` (MVP: 5 fields - name, provider_engine, endpoint, port, tags)
- [X] T040 [US2] Create DatabaseWorkspaceClient wrapper in `internal/client/database_workspace_client.go` (uses RestClient)
- [X] T041 [US2] Implement DatabaseWorkspaceClient.Create() method (~5 lines)
- [X] T042 [P] [US2] Implement DatabaseWorkspaceClient.Get() method (~5 lines)
- [X] T043 [P] [US2] Implement DatabaseWorkspaceClient.Update() method (~5 lines)
- [X] T044 [P] [US2] Implement DatabaseWorkspaceClient.Delete() method (~5 lines)
- [X] T045 [P] [US2] Implement DatabaseWorkspaceClient.List() method (~5 lines) - for import support
- [X] T046 [US2] Add initDatabaseWorkspaceClient() helper to `internal/provider/provider.go`
- [X] T047 [US2] Add DatabaseWorkspaceClient field to ProviderData struct in `internal/provider/provider.go`
- [X] T048 [US2] Initialize DatabaseWorkspaceClient in provider Configure() method
- [ ] T049 [US2] Update DatabaseWorkspaceResource struct in `internal/provider/resource_database_workspace.go` to use DatabaseWorkspaceClient
- [ ] T050 [US2] Remove ARK SDK imports from `internal/provider/resource_database_workspace.go`
- [ ] T051 [US2] Update DatabaseWorkspace.Create() to use DatabaseWorkspaceClient and custom model (MVP fields only)
- [ ] T052 [US2] Update DatabaseWorkspace.Read() to use DatabaseWorkspaceClient
- [ ] T053 [US2] Update DatabaseWorkspace.Update() to use DatabaseWorkspaceClient (support partial updates with pointers)
- [ ] T054 [US2] Update DatabaseWorkspace.Delete() to use DatabaseWorkspaceClient
- [ ] T055 [US2] Update DatabaseWorkspace.ImportState() to use DatabaseWorkspaceClient.List()
- [ ] T056 [US2] Run acceptance tests: `TF_ACC=1 go test ./internal/provider/ -v -run TestAccDatabaseWorkspaceResource`

**Checkpoint**: DatabaseWorkspace MVP fully working with custom OAuth2 client. Secret + DatabaseWorkspace both working. Pattern validated.

---

## Phase 6: User Story 4 - ARK SDK Removal and Cleanup (Priority: P4)

**Goal**: Completely remove all ARK SDK dependencies from codebase

**Independent Test**: `go build` succeeds, `go mod tidy` removes ARK SDK, `rg "ark-sdk-golang"` returns zero results (excluding docs)

### Implementation for User Story 4

- [ ] T057 [P] [US4] Remove ARK SDK imports from `internal/validators/database_engine_validator.go`
- [ ] T058 [P] [US4] Define engine type constants locally in `internal/validators/database_engine_validator.go` (or use string slice)
- [ ] T059 [US4] Remove `internal/client/auth.go` (ARK SDK auth wrapper - no longer needed)
- [ ] T060 [P] [US4] Remove `internal/client/certificates.go` (ARK SDK certificates wrapper - replaced by certificates_oauth2.go)
- [ ] T061 [P] [US4] Remove `internal/client/sia_client.go` (ARK SDK SIA client wrapper - no longer needed)
- [ ] T062 [US4] Remove ISPAuth and SIAAPI fields from ProviderData struct in `internal/provider/provider.go`
- [ ] T063 [US4] Remove ARK SDK imports from `internal/provider/provider.go`
- [ ] T064 [US4] Remove ARK SDK initialization code from provider Configure() method
- [ ] T065 [US4] Run `go mod tidy` to remove unused ARK SDK dependencies
- [ ] T066 [US4] Verify no ARK SDK imports remain: `rg "ark-sdk-golang" internal/ --type go`
- [ ] T067 [US4] Verify `go.mod` has no ARK SDK dependency: `rg "ark-sdk-golang" go.mod`
- [ ] T068 [US4] Run full test suite: `go test ./...`
- [ ] T069 [US4] Run acceptance tests for all resources: `TF_ACC=1 go test ./internal/provider/ -v`

**Checkpoint**: ARK SDK completely removed. Provider builds and all tests pass with only custom OAuth2 implementation.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and final verification

- [ ] T070 [P] Update `docs/oauth2-authentication-fix.md` - mark all resources as FIXED ✅
- [ ] T071 [P] Add deprecation notice to `docs/sdk-integration.md` (historical reference only)
- [ ] T072 [P] Update `CLAUDE.md` - remove ARK SDK patterns, document custom OAuth2 clients
- [ ] T073 [P] Update `CHANGELOG.md` with migration summary (version 1.0.0 - breaking internal change)
- [ ] T074 [P] Create example Terraform configuration in `examples/secret/main.tf`
- [ ] T075 [P] Create example Terraform configuration in `examples/database-workspace-postgres/main.tf` (MVP fields)
- [ ] T076 Verify provider builds: `go build`
- [ ] T077 Verify linting passes: `golangci-lint run`
- [ ] T078 Verify formatting: `go fmt ./...`
- [ ] T079 Final smoke test: Create real resources in test environment
- [ ] T080 Tag release as v1.0.0 (breaking internal change, but no schema changes)

---

## Dependencies & Parallel Execution

### Story Dependency Graph

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational - RestClient) ← BLOCKING
    ↓
    ├─→ US1 (OAuth2 Verification) ← Can start
    │
    ├─→ US3 (Secret Resource) ← Can start (INDEPENDENT)
    │
    ├─→ US2 (DatabaseWorkspace MVP) ← Can start AFTER US3 proves pattern
    │
    └─→ US4 (ARK SDK Removal) ← MUST wait for US2 & US3 complete
           ↓
       Phase 7 (Polish)
```

### Parallel Execution Opportunities

**Phase 2 (Foundational)**:
- Tasks T004-T012: Must be sequential (building RestClient)

**Phase 3 (US1 - OAuth2 Verification)**:
- T013 [P], T014 [P]: Can run in parallel (different concerns)

**Phase 4 (US3 - Secret Resource)**:
- T017 [P], T018 [P], T019 [P]: Test files can be created in parallel
- T020 [P], T023 [P], T024 [P], T025 [P]: Different methods, can implement in parallel after T021

**Phase 5 (US2 - DatabaseWorkspace MVP)**:
- T036 [P], T037 [P], T038 [P]: Test files can be created in parallel
- T042 [P], T043 [P], T044 [P], T045 [P]: Different methods, can implement in parallel after T040

**Phase 6 (US4 - ARK SDK Removal)**:
- T057 [P], T058 [P], T060 [P], T061 [P]: Different files, can remove in parallel

**Phase 7 (Polish)**:
- T070-T075: All documentation updates can run in parallel (different files)

---

## Incremental Delivery Plan

### MVP Scope (Minimal Viable Product)
**Deliver First**: US1 (OAuth2) + US3 (Secret MVP)
- **Time**: ~3-4 hours
- **Value**: Proves pattern, Secret resource works end-to-end
- **Risk**: Low (Secret is simple, pattern proven)

### Increment 2
**Deliver Second**: US2 (DatabaseWorkspace MVP with 5 fields)
- **Time**: ~2-3 hours (reuses pattern)
- **Value**: Core database workspace functionality
- **Risk**: Low (pattern already proven)

### Increment 3
**Deliver Third**: US4 (ARK SDK Cleanup)
- **Time**: ~1-2 hours
- **Value**: Clean codebase, no technical debt
- **Risk**: Very low (all functionality already working)

### Future Increments (Post-MVP)
**Add Later**: DatabaseWorkspace field expansions (per data-model.md phases)
- Phase 2: Certificate support (2 fields)
- Phase 3: Cloud provider (2 fields)
- Phase 4: MongoDB/Oracle/Snowflake (1-3 fields each)
- Phase 5: Active Directory (6 fields)

Each increment is independently testable and deployable.

---

## Task Summary

| Phase | User Story | Task Count | Parallel Tasks | Est. Time |
|-------|------------|------------|----------------|-----------|
| Phase 1 | Setup | 3 | 2 | 15 min |
| Phase 2 | Foundational (RestClient) | 9 | 0 | 1-2 hours |
| Phase 3 | US1 - OAuth2 | 4 | 2 | 30 min |
| Phase 4 | US3 - Secret | 19 | 8 | 2-3 hours |
| Phase 5 | US2 - DatabaseWorkspace MVP | 21 | 9 | 2-3 hours |
| Phase 6 | US4 - ARK SDK Removal | 13 | 4 | 1-2 hours |
| Phase 7 | Polish | 11 | 6 | 1 hour |
| **Total** | **4 User Stories** | **80 tasks** | **31 parallel** | **8-12 hours** |

**Complexity Reduction** (vs original plan):
- 80 tasks vs ~120 estimated (original design with all fields)
- MVP approach reduces initial scope by ~40%
- Generic REST client eliminates ~30 duplicate tasks
- Simple-first approach reduces rework risk

---

## Validation Checklist

✅ All tasks follow checklist format: `- [ ] [ID] [P?] [Story] Description with file path`
✅ Tasks organized by user story (enables independent implementation)
✅ Each user story has independent test criteria
✅ Dependencies clearly documented (story completion order)
✅ Parallel opportunities identified (31 parallelizable tasks)
✅ MVP scope defined (US1 + US3 = proving pattern)
✅ Incremental delivery plan documented
✅ All file paths are absolute and specific
✅ Tasks are immediately executable (LLM can complete without additional context)

---

**Generated**: 2025-10-25
**Based on**: Simplified design (design-reflection.md) with generic REST client and MVP-first approach
**Ready**: ✅ Tasks are executable immediately - start with Phase 1
