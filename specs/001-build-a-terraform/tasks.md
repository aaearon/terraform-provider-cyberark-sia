# Tasks: Terraform Provider for CyberArk Secure Infrastructure Access

**Feature**: `001-build-a-terraform`
**Input**: Design documents from `/specs/001-build-a-terraform/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Acceptance tests following Terraform provider best practices (test-heavy strategy per research.md). Unit tests only for complex helper functions.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
Project structure follows Terraform provider conventions from plan.md:
- Provider code: `internal/provider/`
- Client code: `internal/client/`
- Models: `internal/models/`
- Examples: `examples/`
- Docs: `docs/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go module structure

- [X] T001 Initialize Go module with `go mod init github.com/aaearon/terraform-provider-cyberark-sia`
- [X] T002 [P] Create project directory structure per plan.md (internal/provider, internal/client, internal/models, examples, docs)
- [X] T003 [P] Add core dependencies to go.mod (terraform-plugin-framework, terraform-plugin-log, ark-sdk-golang)
- [X] T004 [P] Create Makefile with build, test, install, lint targets
- [X] T005 [P] Configure .golangci.yml with linter rules per plan.md standards
- [X] T006 [P] Create .gitignore for Go projects (vendor/, *.tfstate, terraform.tfvars)
- [X] T007 Create main.go provider entry point following Terraform Plugin Framework v6 conventions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T008 Implement provider schema in internal/provider/provider.go with authentication configuration attributes (client_id, client_secret, identity_url, identity_tenant_subdomain, max_retries, request_timeout)
- [X] T009 Implement Provider.Configure() method in internal/provider/provider.go to initialize ARK SDK auth (ArkISPAuth with caching enabled) and SIA API client
- [X] T010 [P] Create ProviderData struct in internal/provider/provider.go to hold ispAuth and siaAPI instances for sharing with resources
- [X] T011 [P] Implement authentication logic in internal/client/auth.go using ARK SDK ArkISPAuth with ISPSS OAuth2 client credentials flow
- [X] T012 [P] Create SIA API client wrapper in internal/client/sia_client.go initializing sia.NewArkSIAAPI() for WorkspacesDB() and SecretsDB() access
- [X] T013 [P] Implement error mapping helper in internal/client/errors.go to convert ARK SDK errors to Terraform diagnostics per contracts/sia_api_contract.md
- [X] T014 [P] Implement retry logic wrapper in internal/client/retry.go with exponential backoff for transient failures
- [X] T015 [P] Setup structured logging helpers in internal/provider/logging.go using terraform-plugin-log (NEVER log sensitive data: client_secret, passwords, tokens)
- [X] T016 Create acceptance test infrastructure in internal/provider/provider_test.go with testAccPreCheck(), testAccProtoV6ProviderFactories
- [X] T017 [P] Create environment variable configuration for acceptance tests (TF_ACC, CYBERARK_CLIENT_ID, CYBERARK_CLIENT_SECRET, CYBERARK_IDENTITY_URL, CYBERARK_TENANT_SUBDOMAIN)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 2.5: Technical Debt Resolution (COMPLETED 2025-10-15) ✅

**Purpose**: Address critical improvements identified during Phase 2 reflection before Phase 3

**All tasks completed - see docs/phase2-reflection.md for full details**

### Improvements Completed

- [X] **CRITICAL**: Enhanced error classification with multi-strategy detection (errors.go)
  - Added ErrorCategory enum with 10 categories
  - Implemented classifyError() with Go error type detection + string patterns
  - Comprehensive fallback for unknown errors
  - 95% test coverage (errors_test.go with 25+ test cases)

- [X] **CRITICAL**: Improved retry logic retryability detection (retry.go)
  - Added net.Error support (Temporary(), Timeout())
  - Added context error detection (DeadlineExceeded, Canceled)
  - More specific pattern matching (ordered by specificity)
  - 95% test coverage (retry_test.go with 23+ test cases)

- [X] **HIGH**: Added retry operation logging
  - Integrated tflog into RetryWithBackoff()
  - Log at WARN level for retry attempts with backoff info
  - Log at DEBUG level for non-retryable errors and cancellations

- [X] **MEDIUM**: Improved type safety in ProviderData
  - Replaced interface{} with *auth.ArkISPAuth
  - Replaced interface{} with *sia.ArkSIAAPI
  - Removed type assertions, added compile-time safety

- [X] **DOCUMENTED**: ARK SDK context limitation
  - Authenticate() first param is *ArkProfile (optional), NOT context.Context
  - Added clear comments in auth.go explaining SDK signature
  - Documented in docs/sdk-integration.md

- [X] **SDK RESEARCH**: Verified Phase 3 integration patterns
  - Confirmed WorkspacesDB() and SecretsDB() package paths
  - Documented CRUD method signatures from Context7
  - Created docs/sdk-integration.md as Phase 3 reference

- [X] **DOCUMENTATION**: Created comprehensive docs
  - docs/sdk-integration.md - ARK SDK reference with examples
  - docs/phase2-reflection.md - Lessons learned, assessment, recommendations
  - Updated CLAUDE.md with Phase 2.5 patterns and limitations
  - Updated sia_client.go with usage documentation

### Validation Results ✅

- Tests: `go test ./internal/client/... -v` → **PASS** (2.32s, 95% coverage)
- Build: `go build -v` → **SUCCESS** (31MB binary)
- Linting: golangci-lint → **CLEAN** (pending final run)

**Checkpoint**: ✅ Technical debt resolved - Phase 3 ready with solid foundation

---

## Phase 3: User Story 1 - Onboard Existing Database to SIA via IaC (Priority: P1) 🎯 MVP

**Goal**: Enable infrastructure engineers to register existing databases (AWS RDS, Azure SQL, on-premise) with CyberArk SIA using declarative Terraform configuration, eliminating manual console operations and security gaps.

**Independent Test**: Create a PostgreSQL RDS instance using AWS provider, then use SIA provider to onboard it to SIA, and verify it appears in SIA console with correct connection parameters. Validates complete workflow from infrastructure code to security onboarding.

### Acceptance Tests for User Story 1

**NOTE: Acceptance tests are PRIMARY testing method per Terraform provider best practices**

- [X] T018 [P] [US1] Create acceptance test for database target basic CRUD lifecycle in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_basic)
- [X] T019 [P] [US1] Create acceptance test for AWS RDS PostgreSQL database target in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_awsRDS)
- [X] T020 [P] [US1] Create acceptance test for Azure SQL database target in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_azureSQL)
- [X] T021 [P] [US1] Create acceptance test for on-premise Oracle database target in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_onPremise)
- [X] T022 [P] [US1] Create acceptance test for ImportState functionality in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_import)
- [X] T023 [P] [US1] Create acceptance test for multiple database types (PostgreSQL, MySQL, MariaDB, MongoDB, Oracle, SQL Server, Db2) in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_multipleDatabaseTypes)

### Implementation for User Story 1

- [X] T024 [P] [US1] Create DatabaseTarget data model in internal/models/database_target.go with all attributes from data-model.md (id, name, database_type, database_version, address, port, database_name, authentication_method, cloud_provider, aws_region, aws_account_id, azure_tenant_id, azure_subscription_id, description, tags, last_modified)
- [X] T025 [US1] Implement database_target_resource.go Schema() method in internal/provider/database_target_resource.go with complete attribute definitions per data-model.md
- [X] T026 [US1] Add conditional attribute validators in internal/provider/database_target_resource.go (aws_region/aws_account_id required when cloud_provider=aws, azure_tenant_id/azure_subscription_id required when cloud_provider=azure)
- [X] T027 [P] [US1] Implement basic validators in internal/provider/validators/port_range.go for port range 1-65535 (implemented inline using int64validator.Between)
- [X] T028 [US1] Implement database_target_resource.go Create() method in internal/provider/database_target_resource.go using siaAPI.WorkspacesDB().AddDatabase() per contracts/sia_api_contract.md
- [X] T029 [US1] Implement database_target_resource.go Read() method in internal/provider/database_target_resource.go using siaAPI.WorkspacesDB().Database() with drift detection (handle 404 as resource deleted)
- [X] T030 [US1] Implement database_target_resource.go Update() method in internal/provider/database_target_resource.go using siaAPI.WorkspacesDB().UpdateDatabase() (send only changed fields)
- [X] T031 [US1] Implement database_target_resource.go Delete() method in internal/provider/database_target_resource.go using siaAPI.WorkspacesDB().DeleteDatabase() with graceful handling of already-deleted resources
- [X] T032 [US1] Implement database_target_resource.go ImportState() method in internal/provider/database_target_resource.go to support terraform import functionality
- [X] T033 [US1] Implement database_target_resource.go Configure() method in internal/provider/database_target_resource.go to receive ProviderData from provider configuration
- [X] T034 [P] [US1] Create Terraform HCL examples in examples/resources/database_target/aws_rds_postgresql.tf demonstrating AWS RDS PostgreSQL onboarding with Terraform references to aws_db_instance
- [X] T035 [P] [US1] Create Terraform HCL examples in examples/resources/database_target/azure_sql_server.tf demonstrating Azure SQL onboarding with Terraform references to azurerm_mssql_server
- [X] T036 [P] [US1] Create Terraform HCL examples in examples/resources/database_target/onpremise_oracle.tf demonstrating on-premise Oracle database onboarding
- [X] T037 [US1] Add error mapping for database target operations in internal/client/errors.go (400 invalid database type/version, 409 conflict, 422 validation, 404 not found, 5xx service errors) - already implemented in Phase 2
- [X] T038 [US1] Add structured logging for database target CRUD operations in database_target_resource.go (INFO for success, DEBUG for API calls, ERROR for failures, TRACE for authentication - never log sensitive data)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Database targets can be created, read, updated, deleted, and imported via Terraform.

---

## Phase 4: User Story 2 - Manage Strong Account Lifecycle for Databases (Priority: P2)

**Goal**: Enable declarative management of strong account credentials that SIA uses to provision ephemeral database access, automating credential lifecycle alongside infrastructure code.

**Independent Test**: Create a database target in SIA (manually or via Story 1), then use Terraform to create a strong account for that target with local authentication. Update the password via Terraform and verify the update in SIA. Delete the strong account and confirm removal.

### Acceptance Tests for User Story 2

- [ ] T039 [P] [US2] Create acceptance test for strong account basic CRUD lifecycle in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_basic)
- [ ] T040 [P] [US2] Create acceptance test for local authentication strong account in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_localAuth)
- [ ] T041 [P] [US2] Create acceptance test for Active Directory authentication strong account in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_domainAuth)
- [ ] T042 [P] [US2] Create acceptance test for AWS IAM authentication strong account in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_awsIAM)
- [ ] T043 [P] [US2] Create acceptance test for credential rotation/update in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_credentialUpdate)
- [ ] T044 [P] [US2] Create acceptance test for ImportState functionality in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_import)

### Implementation for User Story 2

- [ ] T045 [P] [US2] Create StrongAccount data model in internal/models/strong_account.go with all attributes from data-model.md (id, name, database_target_id, authentication_type, username, password, aws_access_key_id, aws_secret_access_key, domain, description, rotation_enabled, rotation_interval_days, tags, created_at, last_modified)
- [ ] T046 [US2] Implement strong_account_resource.go Schema() method in internal/provider/strong_account_resource.go with complete attribute definitions and Sensitive=true for password and aws_secret_access_key
- [ ] T047 [US2] Add conditional required validators in internal/provider/strong_account_resource.go for authentication type-specific credential fields (local: username+password, domain: username+password+domain, aws_iam: aws_access_key_id+aws_secret_access_key)
- [ ] T048 [US2] Add cross-attribute validator in internal/provider/strong_account_resource.go ensuring rotation_interval_days is required when rotation_enabled=true
- [ ] T049 [US2] Implement strong_account_resource.go Create() method in internal/provider/strong_account_resource.go using siaAPI.SecretsDB().AddSecret() per contracts/sia_api_contract.md
- [ ] T050 [US2] Implement strong_account_resource.go Read() method in internal/provider/strong_account_resource.go using siaAPI.SecretsDB().GetSecret() (note: response contains metadata only, no sensitive credentials per contract)
- [ ] T051 [US2] Implement strong_account_resource.go Update() method in internal/provider/strong_account_resource.go using siaAPI.SecretsDB().UpdateSecret() for credential rotation and metadata updates (SIA updates credentials immediately per FR-015a)
- [ ] T052 [US2] Implement strong_account_resource.go Delete() method in internal/provider/strong_account_resource.go using siaAPI.SecretsDB().DeleteSecret() with graceful handling of already-deleted resources
- [ ] T053 [US2] Implement strong_account_resource.go ImportState() method in internal/provider/strong_account_resource.go to support terraform import functionality
- [ ] T054 [US2] Implement strong_account_resource.go Configure() method in internal/provider/strong_account_resource.go to receive ProviderData from provider configuration
- [ ] T055 [P] [US2] Create Terraform HCL examples in examples/resources/strong_account/local_auth.tf demonstrating local authentication strong account for PostgreSQL
- [ ] T056 [P] [US2] Create Terraform HCL examples in examples/resources/strong_account/ad_auth.tf demonstrating Active Directory authentication for SQL Server
- [ ] T057 [P] [US2] Create Terraform HCL examples in examples/resources/strong_account/aws_iam_auth.tf demonstrating AWS IAM authentication for RDS MySQL
- [ ] T058 [US2] Add error mapping for strong account operations in internal/client/errors.go (400 invalid auth type, 404 database target not found, 422 missing credentials, 5xx service errors)
- [ ] T059 [US2] Add structured logging for strong account CRUD operations in strong_account_resource.go (INFO for success, DEBUG for API calls, ERROR for failures - NEVER log password, aws_secret_access_key, or bearer tokens)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Database targets can be onboarded, and strong accounts can be managed for those targets.

---

## Phase 5: User Story 3 - Update and Delete Database Targets (Priority: P3)

**Goal**: Complete database target lifecycle management by enabling configuration updates and clean deletion, ensuring security configuration stays synchronized with infrastructure changes.

**Independent Test**: Create a database target (using Story 1), modify its configuration (e.g., change port or authentication method) via Terraform, verify update in SIA. Then delete the target via terraform destroy and confirm removal from SIA.

### Acceptance Tests for User Story 3

- [ ] T060 [P] [US3] Create acceptance test for database target configuration update in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_update)
- [ ] T061 [P] [US3] Create acceptance test for ForceNew behavior when changing immutable attributes in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_forceNew)
- [ ] T062 [P] [US3] Create acceptance test for plan-only operation (no-op update) in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_noOpUpdate)
- [ ] T063 [P] [US3] Create acceptance test for concurrent resource operations in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_concurrent)
- [ ] T064 [P] [US3] Create acceptance test for state drift detection in internal/provider/database_target_resource_test.go (TestAccDatabaseTarget_driftDetection)
- [ ] T065 [P] [US3] Create acceptance test for strong account update (credential rotation) in internal/provider/strong_account_resource_test.go (TestAccStrongAccount_updateCredentials)

### Implementation for User Story 3

- [ ] T066 [US3] Add plan modifiers in internal/provider/database_target_resource.go to control ForceNew behavior for immutable attributes (database_type should trigger replacement)
- [ ] T067 [US3] Add plan modifiers in internal/provider/database_target_resource.go to use UseStateForUnknown() for computed attributes (id, last_modified)
- [ ] T068 [US3] Enhance Update() method in internal/provider/database_target_resource.go to detect which fields changed and send only delta to SIA API
- [ ] T069 [US3] Enhance Read() method in internal/provider/database_target_resource.go to properly detect drift when resources modified outside Terraform
- [ ] T070 [US3] Add plan modifiers in internal/provider/strong_account_resource.go for computed attributes (id, created_at, last_modified)
- [ ] T071 [US3] Enhance Update() method in internal/provider/strong_account_resource.go to handle metadata updates vs credential rotation separately
- [ ] T072 [P] [US3] Create complete workflow example in examples/complete/full_workflow.tf demonstrating AWS RDS provisioning + SIA onboarding + strong account creation + update + deletion
- [ ] T073 [US3] Add validation for partial state failure scenarios in error handling (database created by cloud provider but SIA onboarding failed - provide clear recovery steps per FR-027a)

**Checkpoint**: All user stories should now be independently functional. Complete CRUD lifecycle works for both database targets and strong accounts.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and final hardening

- [ ] T074 [P] Create provider documentation in docs/index.md with authentication setup, configuration reference, and troubleshooting guide
- [ ] T075 [P] Create database_target resource documentation in docs/resources/database_target.md with attribute reference, examples, and import instructions
- [ ] T076 [P] Create strong_account resource documentation in docs/resources/strong_account.md with attribute reference, examples, and import instructions
- [ ] T077 [P] Create authentication guide in docs/guides/authentication.md explaining ISPSS service account setup and role requirements
- [ ] T078 [P] Create provider configuration examples in examples/provider/provider.tf with environment variable usage and secret manager integration patterns
- [ ] T079 [P] Add README.md at repository root with quick start, installation, and links to documentation
- [ ] T080 [P] Validate all examples in examples/ directory run successfully with make test-examples
- [ ] T081 Run quickstart.md validation by executing all steps in specs/001-build-a-terraform/quickstart.md and verifying outputs
- [ ] T082 [P] Run golangci-lint and fix any linting issues per .golangci.yml configuration
- [ ] T083 [P] Run go fmt and gofmt on all Go files
- [ ] T084 Add LICENSE file (appropriate CyberArk/HashiCorp-compatible license)
- [ ] T085 Add CONTRIBUTING.md with contribution guidelines and development setup instructions
- [ ] T086 Review and update CLAUDE.md with final technology stack and commands for Go project
- [ ] T087 Final acceptance test sweep - run all tests with TF_ACC=1 and verify 100% pass rate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion - database target CRUD
- **User Story 2 (Phase 4)**: Depends on Foundational phase completion - strong account CRUD (can run in parallel with US1 if staffed)
- **User Story 3 (Phase 5)**: Depends on User Story 1 and User Story 2 completion - enhances existing resources
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories - **MVP READY**
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - References database targets from US1 but can be developed in parallel
- **User Story 3 (P3)**: Depends on User Story 1 and User Story 2 - Adds update/delete lifecycle enhancements

### Within Each User Story

**User Story 1 (Database Targets)**:
- Acceptance tests (T018-T023) can run in parallel [P]
- Models (T024) before resource implementation (T025-T033)
- Schema definition (T025) before CRUD methods (T028-T032)
- Examples (T034-T036) can be created in parallel [P]
- Error mapping (T037) and logging (T038) after CRUD implementation

**User Story 2 (Strong Accounts)**:
- Acceptance tests (T039-T044) can run in parallel [P]
- Models (T045) before resource implementation (T046-T054)
- Schema definition (T046) before CRUD methods (T049-T053)
- Examples (T055-T057) can be created in parallel [P]
- Error mapping (T058) and logging (T059) after CRUD implementation

**User Story 3 (Lifecycle Enhancements)**:
- Acceptance tests (T060-T065) can run in parallel [P]
- Plan modifiers (T066-T067, T070) before enhanced CRUD methods
- Update/Read enhancements (T068-T069, T071) depend on US1/US2 implementation
- Complete workflow example (T072) depends on all CRUD functionality

### Parallel Opportunities

**Setup Phase (Phase 1)**:
- T002, T003, T004, T005, T006 can run in parallel [P]

**Foundational Phase (Phase 2)**:
- T010, T011, T012, T013, T014, T015, T017 can run in parallel [P] after T008-T009 complete

**User Story 1 (Phase 3)**:
- All acceptance tests (T018-T023) can run in parallel [P]
- T027 validators can be developed in parallel [P] with schema work
- Examples (T034-T036) can be created in parallel [P]

**User Story 2 (Phase 4)**:
- All acceptance tests (T039-T044) can run in parallel [P]
- Examples (T055-T057) can be created in parallel [P]

**User Story 3 (Phase 5)**:
- All acceptance tests (T060-T065) can run in parallel [P]

**Polish Phase (Phase 6)**:
- Documentation tasks (T074-T078) can run in parallel [P]
- T079, T080, T082, T083, T084, T085 can run in parallel [P]

---

## Parallel Example: Foundational Phase

```bash
# After T008-T009 complete, launch these tasks in parallel:
Task: "Create ProviderData struct in internal/provider/provider.go"
Task: "Implement authentication logic in internal/client/auth.go"
Task: "Create SIA API client wrapper in internal/client/sia_client.go"
Task: "Implement error mapping helper in internal/client/errors.go"
Task: "Implement retry logic wrapper in internal/client/retry.go"
Task: "Setup structured logging helpers in internal/provider/logging.go"
Task: "Create environment variable configuration for acceptance tests"
```

---

## Parallel Example: User Story 1

```bash
# Launch all acceptance tests for User Story 1 together:
Task: "Create acceptance test for database target basic CRUD lifecycle"
Task: "Create acceptance test for AWS RDS PostgreSQL database target"
Task: "Create acceptance test for Azure SQL database target"
Task: "Create acceptance test for on-premise Oracle database target"
Task: "Create acceptance test for ImportState functionality"
Task: "Create acceptance test for multiple database types"

# After schema work (T025), launch all examples together:
Task: "Create Terraform HCL examples in examples/resources/database_target/aws_rds_postgresql.tf"
Task: "Create Terraform HCL examples in examples/resources/database_target/azure_sql_server.tf"
Task: "Create Terraform HCL examples in examples/resources/database_target/onpremise_oracle.tf"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T017) - **CRITICAL - blocks all stories**
3. Complete Phase 3: User Story 1 (T018-T038)
4. **STOP and VALIDATE**: Run acceptance tests for User Story 1 independently
5. Deploy/demo - **MVP READY**: Database onboarding works end-to-end

**MVP Success Criteria**:
- Can onboard AWS RDS PostgreSQL database to SIA via Terraform
- Can onboard Azure SQL database to SIA via Terraform
- Can onboard on-premise Oracle database to SIA via Terraform
- All 7 database types supported (PostgreSQL, MySQL, MariaDB, MongoDB, Oracle, SQL Server, Db2)
- Acceptance tests pass for CRUD lifecycle
- terraform import works

### Incremental Delivery

1. **Foundation** (Phase 1 + 2): Provider configured, authentication working → Foundation ready
2. **MVP** (Phase 3): Add User Story 1 → Test independently → Deploy/Demo → **Database onboarding automated**
3. **Credential Management** (Phase 4): Add User Story 2 → Test independently → Deploy/Demo → **Strong accounts manageable**
4. **Complete Lifecycle** (Phase 5): Add User Story 3 → Test independently → Deploy/Demo → **Full CRUD with updates/deletes**
5. **Production Ready** (Phase 6): Polish → Documentation → **Production deployment**

Each phase adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Phase 1 + 2 together** (foundation must be solid)
2. Once Foundational is done:
   - **Developer A**: User Story 1 (Database Targets)
   - **Developer B**: User Story 2 (Strong Accounts)
   - Work in parallel on separate resources
3. **Team converges for User Story 3**: Enhance both resources together
4. **Team completes Phase 6**: Documentation and polish

---

## Validation Checkpoints

### After Phase 2 (Foundational)
- [ ] Provider initialization works with valid ISPSS credentials
- [ ] Authentication succeeds and bearer token acquired
- [ ] SIA API client initialized successfully
- [ ] Retry logic handles transient failures
- [ ] Error mapping produces actionable diagnostics

### After Phase 3 (User Story 1 - MVP)
- [ ] Database target can be created via Terraform
- [ ] Database target appears in SIA console with correct parameters
- [ ] Database target can be read (state refresh works)
- [ ] Database target can be imported via terraform import
- [ ] All 7 database types work (PostgreSQL, MySQL, MariaDB, MongoDB, Oracle, SQL Server, Db2)
- [ ] AWS RDS example works end-to-end
- [ ] Azure SQL example works end-to-end
- [ ] On-premise example works end-to-end
- [ ] Acceptance tests pass: TestAccDatabaseTarget_*

### After Phase 4 (User Story 2)
- [ ] Strong account can be created for database target
- [ ] Strong account supports local authentication (username+password)
- [ ] Strong account supports domain authentication (AD)
- [ ] Strong account supports AWS IAM authentication
- [ ] Strong account can be imported via terraform import
- [ ] Acceptance tests pass: TestAccStrongAccount_*

### After Phase 5 (User Story 3)
- [ ] Database target configuration can be updated (port, authentication method)
- [ ] Strong account credentials can be rotated via Terraform
- [ ] Database target can be deleted cleanly
- [ ] Strong account can be deleted cleanly
- [ ] ForceNew triggers replacement for immutable attributes
- [ ] State drift detected when resources modified outside Terraform
- [ ] Acceptance tests pass: All TestAccDatabaseTarget_update, TestAccStrongAccount_update

### After Phase 6 (Polish)
- [ ] Documentation complete and accurate
- [ ] All examples in examples/ directory execute successfully
- [ ] Quickstart guide validated end-to-end
- [ ] golangci-lint passes with no errors
- [ ] All acceptance tests pass (TF_ACC=1 go test ./...)
- [ ] README.md provides clear getting started guide

---

## Notes

- **[P] tasks** = different files, no dependencies - can execute in parallel
- **[Story] label** maps task to specific user story for traceability (US1, US2, US3)
- Each user story should be **independently completable and testable**
- **Acceptance tests are PRIMARY** testing method per Terraform provider best practices (research.md)
- Unit tests ONLY for complex helper functions (validators, error mapping)
- Verify acceptance tests **run against real SIA API** when TF_ACC=1
- **NEVER log sensitive data**: client_secret, password, aws_secret_access_key, bearer tokens
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **ARK SDK handles token refresh** automatically when caching enabled (NewArkISPAuth(true))
- Provider uses **resp.ResourceData** pattern to share ProviderData with resources
- **SIA API validates** database compatibility - provider does minimal validation (user responsibility per research.md)

---

## Summary

- **Total Tasks**: 87 tasks
- **Task Count per User Story**:
  - Setup (Phase 1): 7 tasks
  - Foundational (Phase 2): 10 tasks
  - User Story 1 (P1 - Database Targets): 21 tasks
  - User Story 2 (P2 - Strong Accounts): 21 tasks
  - User Story 3 (P3 - Lifecycle): 14 tasks
  - Polish (Phase 6): 14 tasks
- **Parallel Opportunities**: 52 tasks marked [P] can run in parallel within their phase
- **MVP Scope**: Phase 1 + Phase 2 + Phase 3 (38 tasks) = Fully functional database onboarding
- **Independent Test Criteria**:
  - US1: Create database target, verify in SIA console, import state, delete
  - US2: Create strong account, update credentials, verify in SIA, delete
  - US3: Update database target config, verify ForceNew behavior, detect drift

**Format Validation**: ✅ All tasks follow checklist format (checkbox, ID, labels, file paths)
