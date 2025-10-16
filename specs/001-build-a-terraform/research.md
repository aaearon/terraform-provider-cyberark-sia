# Research: Terraform Provider for CyberArk SIA

**Phase**: 0 - Outline & Research
**Date**: 2025-10-15
**Status**: Complete

## Executive Summary

Research validates technical feasibility of building a Terraform provider for CyberArk Secure Infrastructure Access using the official ARK SDK for Golang and HashiCorp's Terraform Plugin Framework. Key findings establish authentication patterns, API integration strategies, resource lifecycle management, and testing approaches.

## Research Areas

### 1. CyberArk ARK SDK for Golang

**Decision**: Use `github.com/cyberark/ark-sdk-golang` as the primary API client library

**Rationale**:
- Official CyberArk SDK with 168 code snippets and 9.5 trust score
- Provides pre-built authentication via `ArkISPAuth` supporting ISPSS OAuth2 flows
- Includes SIA-specific API clients (`uap.NewArkUAPAPI`) for database policy and target management
- Handles token caching internally via OS keystore
- Supports concurrent operations with thread-safe authentication

**Key Patterns Identified**:

1. **Authentication Flow** (from SDK examples):
```go
ispAuth := auth.NewArkISPAuth(false) // false = disable caching for provider control
token, err := ispAuth.Authenticate(
    nil,
    &authmodels.ArkAuthProfile{
        Username:   "user@cyberark.cloud.12345",
        AuthMethod: authmodels.Identity,
        AuthMethodSettings: &authmodels.IdentityArkAuthMethodSettings{},
    },
    &authmodels.ArkSecret{Secret: os.Getenv("ARK_SECRET")},
    false, // force new auth
    false, // don't attempt refresh
)
```

2. **SIA API Client Initialization**:
```go
uapAPI, err := uap.NewArkUAPAPI(ispAuth.(*auth.ArkISPAuth))
// uapAPI provides access to:
// - uapAPI.Db() for database policies
// - Database target management APIs (needs SDK exploration)
```

3. **Token Lifecycle**:
- SDK returns `token.ExpiresIn` for expiration tracking
- Automatic caching in OS keystore (can be disabled)
- Provider will need custom refresh logic for Terraform's long-running operations

**Alternatives Considered**:
- Direct REST API calls: Rejected due to authentication complexity, no official support, maintenance burden
- Python SDK: Rejected as Terraform providers require Go

**Integration Requirements**:
- Provider must wrap `ArkISPAuth` to manage token refresh cycle
- Need to disable SDK's internal caching (`NewArkISPAuth(false)`) to control refresh timing
- Must handle concurrent resource operations with shared auth state

---

### 2. Terraform Plugin Framework

**Decision**: Use `github.com/hashicorp/terraform-plugin-framework` (latest stable)

**Rationale**:
- Current recommended framework (replaces deprecated SDKv2)
- 827 code snippets, 9.8 trust score
- Type-safe schema definitions with compile-time validation
- Built-in support for plan modifiers, validators, state upgrades
- Better performance than SDKv2 for large resource counts

**Key Implementation Patterns**:

1. **Provider Structure**:
```go
type ExampleCloudProvider struct {
    Version string
    client  *SIAClient // Provider-owned API client
}

func (p *ExampleCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Initialize SIA client with credentials from config
    // Store on p.client for use by resources
}

func (p *ExampleCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewDatabaseTargetResource,
        NewStrongAccountResource,
    }
}
```

2. **Resource CRUD Pattern**:
```go
type databaseTargetResource struct {
    provider *Provider // Access to provider.client
}

func (r *databaseTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data databaseTargetResourceData
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    // Call SIA API via r.provider.client
    result, err := r.provider.client.CreateDatabaseTarget(...)

    // Map result to state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

3. **Schema Definition**:
```go
func (r *databaseTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "database_type": schema.StringAttribute{
                Required:    true,
                Description: "Type of database (postgresql, mysql, etc.)",
                Validators: []validator.String{
                    stringvalidator.OneOf("postgresql", "mysql", "mariadb", ...),
                },
            },
        },
    }
}
```

**Best Practices from Framework Docs**:
- Resources access provider-configured client via provider instance
- Use plan modifiers for computed attributes that shouldn't trigger replacements
- Custom validators for domain-specific validation (database types, versions, ports)
- Sensitive attributes automatically masked in logs when marked `Sensitive: true`
- Use `UseStateForUnknown()` plan modifier for ID attributes

**Alternatives Considered**:
- SDKv2: Deprecated, less type-safe, inferior performance
- terraform-plugin-go (low-level): Too much boilerplate, no abstraction benefits

---

### 3. Authentication & Token Management

**Decision**: Implement custom token refresh goroutine with proactive renewal at 80% token lifetime

**Rationale**:
- ISPSS bearer tokens expire in 15 minutes (per spec clarifications)
- Terraform operations can exceed 15 minutes (e.g., provisioning 10+ databases)
- FR-023a mandates proactive refresh to prevent mid-operation failures
- SDK's internal refresh may not align with Terraform's concurrency model

**Implementation Strategy**:

1. **Token Refresh Architecture**:
```go
type SIAClient struct {
    auth         *auth.ArkISPAuth
    uapClient    *uap.ArkUAPAPI
    tokenMutex   sync.RWMutex
    currentToken *authmodels.ArkAuthToken
    refreshTimer *time.Timer
}

func (c *SIAClient) startTokenRefresh(ctx context.Context) {
    expiresIn := c.currentToken.ExpiresIn
    refreshAt := time.Duration(float64(expiresIn) * 0.8) // 80% lifetime

    c.refreshTimer = time.AfterFunc(refreshAt, func() {
        c.tokenMutex.Lock()
        defer c.tokenMutex.Unlock()

        newToken, err := c.auth.Authenticate(..., false, true) // attempt refresh
        if err == nil {
            c.currentToken = newToken
            c.startTokenRefresh(ctx) // Schedule next refresh
        }
    })
}
```

2. **Concurrent Request Handling**:
- Use `sync.RWMutex` for token access (many readers, single writer during refresh)
- API calls acquire read lock to check token validity
- Refresh goroutine acquires write lock during token update

**Security Considerations**:
- Client ID and secret stored in provider config (Terraform will mark as sensitive)
- Never log token values (use `tflog.Trace` for auth flow debugging only)
- ARK SDK handles secure token storage if caching enabled

**Alternatives Considered**:
- On-demand refresh on 401: Reactive, causes mid-operation failures, race conditions
- Per-resource token: Inefficient, multiplies auth overhead, complexity
- Rely on SDK caching: Insufficient control over refresh timing for Terraform patterns

---

### 4. Database Type Support & Validation

**Decision**: NO custom validation of database types or versions - user responsibility per requirements

**Rationale**:
- **USER RESPONSIBILITY**: Users must ensure database compatibility with SIA before onboarding
- **SIA API Validation**: SIA REST API validates database type/version compatibility on resource creation
- **Simpler Provider**: Removes custom validation logic, complex cross-attribute checks, version parsing
- **Terraform Philosophy**: Provider accepts user input, lets API be source of truth for compatibility

**Provider Validation Scope** (limited to basic input sanitation):
- String attributes: Non-empty where required
- Numeric attributes: Valid ranges (e.g., port 1-65535)
- Format validation: URLs, hostnames (basic regex)
- **NO semantic validation**: Database type compatibility, version minimums, authentication method restrictions

**Implementation Approach**:
```go
// NO custom database validators needed
// Use only framework built-in validators:
schema.StringAttribute{
    Required:    true,
    Description: "Database type (e.g., postgresql, mysql). Must be supported by SIA.",
    Validators: []validator.String{
        stringvalidator.LengthAtLeast(1), // Basic non-empty check only
    },
}
```

**Error Handling**:
- If user provides unsupported database type/version, **SIA API returns 400/422**
- Provider maps API error to diagnostic: "SIA validation failed: {api_message}"
- User receives actionable feedback from SIA's own validation logic

**Alternatives Considered**:
- Custom validation logic: Rejected - duplicates SIA's validation, requires maintenance as SIA adds database support
- Schema enum constraints: Rejected - overly restrictive, prevents future SIA database type additions

---

### 5. Testing Strategy (Terraform Provider Best Practices)

**Decision**: Acceptance-test-heavy strategy per HashiCorp Terraform Provider Development guidelines

**Testing Approach** (aligned with Terraform provider conventions):

1. **Acceptance Tests** (Primary testing method):
   - Database target CRUD lifecycle (Create, Read, Update, Delete)
   - Strong account CRUD lifecycle
   - Terraform import functionality
   - State drift detection and refresh
   - Concurrent resource creation (parallel apply)
   - Plan-only operations (no-op updates)
   - ForceNew behavior (resource replacement scenarios)
   - **Run against real SIA API** when `TF_ACC=1` environment variable set

2. **Unit Tests** (Minimal, only for complex logic):
   - Error message mapping helpers (if logic is non-trivial)
   - Utility functions for data transformation
   - **NO unit tests for validators** (framework handles, acceptance tests validate behavior)
   - **NO mock API tests** (acceptance tests use real API)

**Rationale for Acceptance-Heavy Approach**:
- **HashiCorp Recommendation**: Terraform providers are tested primarily via acceptance tests
- **Real Behavior**: Tests exercise actual Terraform lifecycle with real provider binary
- **Integration Confidence**: Validates provider ↔ SIA API ↔ Terraform core interaction
- **State Management**: Tests real Terraform state handling, not mocked scenarios

**Testing Framework**:
```go
// Acceptance test example (Terraform provider standard)
func TestAccDatabaseTarget_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create resource
            {
                Config: testAccDatabaseTargetConfig_basic,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "database_type", "postgresql"),
                    resource.TestCheckResourceAttrSet("cyberark_sia_database_target.test", "id"),
                    testAccCheckDatabaseTargetExists("cyberark_sia_database_target.test"),
                ),
            },
            // Step 2: ImportState test
            {
                ResourceName:      "cyberark_sia_database_target.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Step 3: Update and verify
            {
                Config: testAccDatabaseTargetConfig_updated,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "description", "updated"),
                ),
            },
        },
    })
}
```

**Test Coverage Focus**:
- ✅ Each resource: Create, Read, Update, Delete tests
- ✅ ImportState: Verify state import works correctly
- ✅ Plan behavior: Verify computed attributes, ForceNew triggers
- ✅ Error scenarios: Invalid configurations, API errors
- ❌ NOT testing: Framework behavior, schema definitions, basic validation

**Test Environment Requirements**:
```bash
# Required environment variables for acceptance tests
export TF_ACC=1                           # Enable acceptance tests
export CYBERARK_CLIENT_ID="..."           # ISPSS credentials
export CYBERARK_CLIENT_SECRET="..."
export CYBERARK_IDENTITY_URL="..."
export CYBERARK_TENANT_SUBDOMAIN="..."

# Run acceptance tests
go test -v ./internal/provider -run TestAcc
```

**Alternatives Considered**:
- Test pyramid (many unit, few acceptance): Rejected - not idiomatic for Terraform providers
- Mock API for acceptance tests: Rejected - doesn't test real API integration
- Manual testing only: Rejected - not repeatable, doesn't prevent regressions

---

### 6. Error Handling & Diagnostics

**Decision**: Map SIA API errors to actionable Terraform diagnostics per FR-027

**Error Mapping Strategy**:

| SIA API Error | Terraform Diagnostic | User Guidance |
|---------------|---------------------|---------------|
| 401 Unauthorized | Authentication failed | Check client_id and client_secret in provider config |
| 403 Forbidden | Insufficient permissions | Verify ISPSS service account has required role memberships |
| 404 Not Found (on read) | Resource not found | May have been deleted outside Terraform; run `terraform refresh` |
| 409 Conflict | Resource already exists | Use `terraform import` to manage existing resource |
| 422 Unprocessable (validation) | Invalid configuration | Specific field validation error with remediation |
| 5xx Server Error | SIA service unavailable | Retry operation; check SIA service status |

**Implementation Pattern**:
```go
func mapSIAError(err error, operation string) diag.Diagnostic {
    // Parse ARK SDK error
    // Return structured diagnostic with summary, detail, severity
    return diag.NewErrorDiagnostic(
        fmt.Sprintf("Failed to %s database target", operation),
        fmt.Sprintf("SIA API error: %s\n\nRecommended action: %s", err, getRecommendation(err)),
    )
}
```

**Partial State Handling** (per FR-027a):
- When database created by AWS/Azure provider but SIA onboarding fails:
  - Terraform apply fails with clear error
  - Diagnostic includes: which step failed, current state, manual remediation
  - Example: "Database 'prod-db' exists in AWS RDS but failed to onboard to SIA. To resolve: 1) Fix SIA connectivity/auth, 2) Run terraform apply again, OR 3) Manually remove database from AWS and run terraform destroy"

---

### 7. Cloud Provider Integration Patterns

**Decision**: Use Terraform data sources and resource references for cloud database connection details

**AWS RDS Integration Pattern**:
```hcl
# User's Terraform configuration
resource "aws_db_instance" "main" {
  identifier = "production-db"
  engine     = "postgres"
  # ... other AWS config
}

resource "cyberark_sia_database_target" "main" {
  name          = aws_db_instance.main.identifier
  database_type = "postgresql"
  address       = aws_db_instance.main.endpoint  # Reference RDS output
  port          = aws_db_instance.main.port

  # Cloud provider metadata for AWS IAM auth
  cloud_provider = "aws"
  aws_region     = "us-east-1"
  aws_account_id = data.aws_caller_identity.current.account_id
}
```

**Azure SQL Integration Pattern**:
```hcl
resource "azurerm_mssql_server" "main" {
  name = "production-sql"
  # ... Azure config
}

resource "cyberark_sia_database_target" "main" {
  name          = azurerm_mssql_server.main.name
  database_type = "sqlserver"
  address       = azurerm_mssql_server.main.fully_qualified_domain_name
  port          = 1433

  cloud_provider     = "azure"
  azure_tenant_id    = data.azurerm_client_config.current.tenant_id
  azure_subscription = data.azurerm_client_config.current.subscription_id
}
```

**Design Rationale**:
- Provider does NOT create databases (FR out of scope)
- Provider ONLY registers existing databases with SIA
- Cloud metadata enables AWS IAM and Azure AD authentication methods
- User controls dependency ordering via Terraform's implicit dependencies

**Alternatives Considered**:
- Auto-discovery: Out of scope per spec, requires excessive cloud API permissions
- Embedded cloud provider logic: Violates single responsibility, increases complexity
- Manual string input for all fields: Error-prone, doesn't leverage Terraform references

---

## Decision Summary Table

| Decision Area | Choice | Key Rationale |
|---------------|--------|---------------|
| API Client Library | CyberArk ARK SDK for Golang | Official support, pre-built auth, SIA-specific APIs |
| Provider Framework | Terraform Plugin Framework v6 | Current standard, type-safe, best performance |
| Authentication | ARK SDK internal token management | SDK handles refresh/caching; provider holds SDK instances per framework pattern |
| Validation | **NO custom validation** - SIA API validates | User ensures database compatibility; simpler provider; API is source of truth |
| Testing | **Acceptance-test-heavy per Terraform standards** | HashiCorp best practices; tests real provider behavior; minimal unit tests |
| Error Mapping | Structured diagnostics with guidance | FR-027 requirement, improved user experience |
| Cloud Integration | Terraform references to cloud resources | Leverages IaC workflow, no provider complexity |

---

## Open Questions & Follow-Up

### Resolved During Research
- ✅ Token refresh strategy: ARK SDK internal management (optional caching via `NewArkISPAuth(cachingEnabled)`)
- ✅ Provider data pattern: Hold ispAuth + uapClient instances per Terraform framework conventions
- ✅ Validation strategy: NO custom validation; SIA API is source of truth
- ✅ Testing approach: Acceptance-test-heavy per Terraform provider best practices

### Requires Phase 1 (Design)
- API endpoints for database targets (not found in SDK examples - need SIA API exploration)
- Strong account storage format in SIA (local vs. Privilege Cloud distinction)
- Target Set assignment APIs (if needed despite being out of scope)
- State upgrade scenarios (initial version, no migrations needed)

---

## References

1. CyberArk ARK SDK for Golang - Context7 Documentation
2. HashiCorp Terraform Plugin Framework - Context7 Documentation
3. Feature Specification - `/specs/001-build-a-terraform/spec.md`
4. Project Constitution - `/.specify/memory/constitution.md`
5. Terraform Plugin Best Practices - HashiCorp Developer Portal
6. Go Project Layout - https://github.com/golang-standards/project-layout

---

**Next Phase**: Design (Phase 1) - Generate data-model.md, API contracts, quickstart.md

---

## Additional Research (Phase 1 Enhancement)

### 8. ARK SDK SIA Database Target & Strong Account APIs

**Decision**: Use `sia.NewArkSIAAPI()` for database/secret operations, `uap.NewArkUAPAPI()` for policies

**Concrete API Methods Discovered**:

```go
// Initialize SIA API client
siaAPI, err := sia.NewArkSIAAPI(ispAuth.(*auth.ArkISPAuth))

// Database operations
database, err := siaAPI.WorkspacesDB().AddDatabase(&dbmodels.ArkSIADBAddDatabase{
    Name:              "MyDatabase",
    ProviderEngine:    dbmodels.EngineTypeAuroraMysql, // or other types
    ReadWriteEndpoint: "myrds.com",
    SecretID:          secretID, // Reference to strong account secret
})

// Strong account (secret) operations  
secret, err := siaAPI.SecretsDB().AddSecret(&dbsecretsmodels.ArkSIADBAddSecret{
    SecretType: "username_password",
    Username:   "admin_user",
    Password:   "strong_password",
})
```

**Package Imports**:
```go
import (
    "github.com/cyberark/ark-sdk-golang/pkg/services/sia"
    dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
    dbsecretsmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/secrets/db/models"
)
```

**Provider Implementation Pattern**:
```go
type ProviderData struct {
    ispAuth   *auth.ArkISPAuth
    siaAPI    *sia.ArkSIAAPI  // For database targets & secrets
    uapAPI    *uap.ArkUAPAPI  // For policies (out of scope for initial version)
}
```

---

### 9. Provider→Resource Data Sharing (Terraform Plugin Framework)

**Decision**: Use `resp.ResourceData` in Provider.Configure(), implement `resource.ResourceWithConfigure` interface

**Implementation Pattern**:

```go
// 1. Provider Configure method
func (p *CyberArkSIAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Initialize clients
    ispAuth := auth.NewArkISPAuth(true) // Enable caching
    // ... authenticate ...
    
    siaAPI, err := sia.NewArkSIAAPI(ispAuth.(*auth.ArkISPAuth))
    // ... error handling ...
    
    // Make available to resources
    resp.ResourceData = &ProviderData{
        ispAuth: ispAuth,
        siaAPI:  siaAPI,
    }
}

// 2. Resource struct with Configure method
type databaseTargetResource struct {
    providerData *ProviderData
}

func (r *databaseTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Nil check (Terraform sets ProviderData after ConfigureProvider RPC)
    if req.ProviderData == nil {
        return
    }
    
    // Type assertion with error handling
    providerData, ok := req.ProviderData.(*ProviderData)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *ProviderData, got: %T", req.ProviderData),
        )
        return
    }
    
    r.providerData = providerData
}

// 3. Use in CRUD methods
func (r *databaseTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.providerData == nil {
        resp.Diagnostics.AddError("Unconfigured API Client", "...")
        return
    }
    
    result, err := r.providerData.siaAPI.WorkspacesDB().AddDatabase(...)
}
```

---

### 10. Conditional Required Attributes (Terraform Plugin Framework)

**Decision**: Use attribute-level validators from `terraform-plugin-framework-validators`

**Implementation Patterns**:

```go
import (
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/path"
)

// Pattern 1: AlsoRequires (RequiredWith equivalent)
"aws_region": schema.StringAttribute{
    Optional: true,
    Validators: []validator.String{
        stringvalidator.AlsoRequires(
            path.MatchRelative().AtParent().AtName("aws_account_id"),
        ),
    },
},

// Pattern 2: ConflictsWith
"local_auth": schema.StringAttribute{
    Optional: true,
    Validators: []validator.String{
        stringvalidator.ConflictsWith(
            path.MatchRelative().AtParent().AtName("aws_iam_auth"),
        ),
    },
},

// Pattern 3: Resource-level validators (for complex logic)
func (r *databaseTargetResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
    return []resource.ConfigValidator{
        resourcevalidator.RequiredTogether(
            path.MatchRoot("cloud_provider"),
            path.MatchRoot("aws_region"),
        ),
    }
}
```

---

### 11. ARK SDK Token Caching Behavior

**Decision**: Enable caching (`NewArkISPAuth(true)`) - SDK handles refresh automatically

**Caching Mechanics**:
- `NewArkISPAuth(true)`: Caches token to OS keystore, auto-refresh supported
- `NewArkISPAuth(false)`: No caching, provider manages token lifecycle
- Refresh: `Authenticate(nil, profile, secret, false, true)` - last param triggers refresh

**Recommendation for Terraform Provider**:
```go
// Enable caching - SDK manages token refresh internally
ispAuth := auth.NewArkISPAuth(true)

// Initial authentication
token, err := ispAuth.Authenticate(nil, profile, secret, false, false)

// SDK handles refresh automatically on subsequent API calls
// No custom token refresh goroutine needed!
```

**Rationale**: SDK's internal token management is thread-safe and handles the 15-minute expiration automatically when caching is enabled. Provider benefits from SDK's built-in refresh logic without custom complexity.

---

### 12. SIA API Error Response Structure

**Findings**: Specific error JSON structure not publicly documented. ARK SDK abstracts error handling.

**ARK SDK Error Handling Pattern**:
```go
result, err := siaAPI.WorkspacesDB().AddDatabase(...)
if err != nil {
    // SDK returns Go error - inspect error message
    // Common patterns from ARK SDK:
    // - Authentication errors: "authentication failed"
    // - Permission errors: "insufficient permissions"
    // - Validation errors: include field details in message
    
    resp.Diagnostics.AddError(
        "Failed to Create Database Target",
        fmt.Sprintf("SIA API error: %s", err.Error()),
    )
    return
}
```

**Error Mapping Strategy** (revised from original contract):
- Rely on ARK SDK error messages (already descriptive)
- Map common error patterns to diagnostics
- No need for custom HTTP status code handling (SDK abstracts this)

---

### 13. Terraform Acceptance Test Patterns

**Decision**: Multi-step tests with ConfigStateChecks, ImportState, plan validation

**Comprehensive Test Pattern**:

```go
func TestAccDatabaseTarget_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create
            {
                Config: testAccDatabaseTargetConfig_basic,
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "cyberark_sia_database_target.test",
                        tfjsonpath.New("database_type"),
                        knownvalue.StringExact("postgresql"),
                    ),
                },
            },
            // Step 2: ImportState
            {
                ResourceName:      "cyberark_sia_database_target.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Step 3: Update
            {
                Config: testAccDatabaseTargetConfig_updated,
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectResourceAction(
                            "cyberark_sia_database_target.test",
                            plancheck.ResourceActionUpdate,
                        ),
                    },
                },
            },
        },
    })
}

// ForceNew test (resource replacement)
func TestAccDatabaseTarget_forceNew(t *testing.T) {
    resource.Test(t, resource.TestCase{
        // ... setup ...
        Steps: []resource.TestStep{
            {Config: configOriginal},
            {
                Config: configWithDifferentType,
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectResourceAction(
                            "cyberark_sia_database_target.test",
                            plancheck.ResourceActionDestroyBeforeCreate,
                        ),
                    },
                },
            },
        },
    })
}
```

**Test Coverage**:
- ✅ Basic CRUD lifecycle
- ✅ Import functionality  
- ✅ Update in-place validation
- ✅ ForceNew behavior (resource replacement)
- ✅ Concurrent operations (multiple resources in single config)

---

## Updated Decision Summary

| Decision Area | Updated Choice | Key Change |
|---------------|----------------|------------|
| SIA Database API | `sia.NewArkSIAAPI()` with WorkspacesDB() and SecretsDB() | Concrete methods identified |
| Provider Data Sharing | `resp.ResourceData` + `resource.ResourceWithConfigure` | Implementation pattern confirmed |
| Conditional Attributes | `stringvalidator.AlsoRequires()` with path matching | Framework-specific validators |
| Token Caching | **Enable caching** (`NewArkISPAuth(true)`) | SDK handles refresh - simpler than custom goroutine |
| Error Handling | Trust ARK SDK error messages | SDK abstracts HTTP details |
| Acceptance Testing | Multi-step with ConfigStateChecks + plan validation | Terraform testing best practices |

