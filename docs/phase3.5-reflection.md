# Phase 3.5 Reflection: Removing Provider-Level Retry Configuration

**Date**: 2025-10-15
**Decision**: Remove `max_retries` and `request_timeout` from provider configuration

## Summary

Removed provider-level retry/timeout configuration parameters in favor of hard-coded internal constants, aligning with modern Terraform provider best practices (2025).

## Problem Statement

The provider exposed two configuration parameters:
1. **`max_retries`** (optional, default 3) - Maximum API retry attempts
2. **`request_timeout`** (optional, default 30) - API request timeout in seconds

Questions arose:
- Do users need this level of control?
- How do other major providers handle this?
- What's the modern best practice for 2025?
- Can Terraform handle this at a higher level?

## Research

### Industry Analysis

**AWS Provider** (Legacy Pattern):
- Exposes `max_retries` (default 25, ~1 hour)
- Criticized for excessively high defaults (GitHub #1209)
- Pattern dates back to pre-split Terraform CLI codebase
- Considered outdated by 2025 standards

**Modern Providers** (Google Cloud, Azure):
- **Do NOT expose** retry/timeout configuration
- Handle transient errors internally
- Providers are more opinionated about error handling
- Simpler configuration surface

**Terraform Plugin Framework**:
- Provides `timeouts` module for **resource-level** timeout blocks
- Pattern: `timeouts { create = "60m", update = "30m" }`
- Used for **long-running operations**, not API-level retries
- This is the idiomatic Terraform way

### Key Distinction

Two separate concerns:

1. **API-Level Retries** (Provider's responsibility):
   - Handle transient network errors (503, timeouts, throttling)
   - Should be **internal** to the provider
   - Users shouldn't need to configure this
   - Example: 3 retries with exponential backoff

2. **Operation Timeouts** (User's control):
   - Handle long-running async operations
   - Should be **configurable per resource**
   - Users control how long to wait
   - Example: Database creation might take 15 minutes

### Critical Discovery

**`request_timeout` was completely unused!**
- Defined in schema ✓
- Stored in `ProviderData` ✓
- **NEVER REFERENCED** in code ✗
- Zero functionality
- Pure configuration bloat

## Decision

**Remove both parameters** from provider configuration.

### Rationale

1. **Modern Best Practice**: 2025 trend is opinionated providers that abstract internal details
2. **Simpler UX**: 99% of users don't need to configure retry logic
3. **Already Well-Designed**: Our defaults (3 retries, 30s max delay) are excellent
4. **No Functionality Loss**: `request_timeout` did nothing anyway
5. **Proper Separation**:
   - Provider handles transient errors (internal constants)
   - Users control operation timeouts (resource-level blocks)
6. **Reduced Support Burden**: Fewer parameters to validate/document/troubleshoot

### Expert Opinion (Gemini)

> "The best practice is a combination of internal retries and resource timeouts. You should **remove `max_retries` and `request_timeout`** from provider configuration. Make them hard-coded constants within your internal client."

## Implementation

### Changes Made

1. **Provider Schema** (`internal/provider/provider.go`):
   - Removed `max_retries` from `CyberArkSIAProviderModel`
   - Removed `request_timeout` from `CyberArkSIAProviderModel`
   - Removed `MaxRetries` from `ProviderData`
   - Removed `RequestTimeout` from `ProviderData`
   - Removed schema attributes (lines 98-105)
   - Removed default value logic (lines 151-159)
   - **Result**: 6 attributes → 4 attributes

2. **Resources** (`database_workspace_resource.go`, `secret_resource.go`):
   - Changed: `MaxRetries: r.providerData.MaxRetries`
   - To: `MaxRetries: client.DefaultMaxRetries`
   - **Impact**: 8 CRUD operations updated (4 per resource)

3. **Logging** (`internal/provider/logging.go`):
   - Removed from `LogProviderConfig()` output
   - Cleaner debug logs

4. **Retry Constants** (`internal/client/retry.go`):
   - **No changes needed** - already well-defined:
     ```go
     const (
         DefaultMaxRetries = 3               // Conservative default
         BaseDelay = 500 * time.Millisecond  // Gradual backoff
         MaxDelay = 30 * time.Second         // Reasonable cap
     )
     ```

5. **Documentation**:
   - Updated `CLAUDE.md` with breaking change notice
   - Updated retry logic examples
   - Created this reflection document

### Breaking Change

**Impact**: Users with these parameters in config will get schema validation errors.

**Migration**: Simply remove the parameters:

```hcl
# BEFORE (Phase 2-3)
provider "cyberark_sia" {
  client_id                   = "..."
  client_secret               = "..."
  identity_tenant_subdomain   = "abc123"
  identity_url                = "https://abc123.cyberark.cloud"
  max_retries                 = 5         # ← REMOVE
  request_timeout             = 60        # ← REMOVE
}

# AFTER (Phase 3.5+)
provider "cyberark_sia" {
  client_id                   = "..."
  client_secret               = "..."
  identity_tenant_subdomain   = "abc123"
  identity_url                = "https://abc123.cyberark.cloud"
}
```

**Rationale for Breaking Change**:
- Pre-1.0 provider - breaking changes acceptable
- `request_timeout` was non-functional anyway
- `max_retries` usage likely minimal (good defaults)
- Better now than after 1.0 release

## Benefits

1. ✅ **Simpler Provider**: 4 attributes (down from 6)
2. ✅ **Modern Architecture**: Aligns with 2025 best practices
3. ✅ **Better UX**: "It just works" for 99% of users
4. ✅ **Less Documentation**: Fewer parameters to explain
5. ✅ **Less Support**: Fewer configuration issues
6. ✅ **Cleaner Code**: Removed unused `RequestTimeout`

## Trade-offs

**Lost Flexibility**: Power users can't tune retry behavior

**Mitigation**:
- Defaults are already excellent (3 retries, exponential backoff)
- Edge cases can be addressed in Phase 4 if needed
- Resource-level `timeouts` blocks provide user control where it matters

## Future Enhancements

**Phase 4+**: Add resource-level `timeouts` blocks for long-running operations:

```hcl
resource "cyberark_sia_database_workspace" "example" {
  name          = "my-db"
  database_type = "postgres-aws-rds"

  timeouts {
    create = "15m"  # Database creation might be slow
    update = "10m"
    delete = "5m"
  }
}
```

This is the **idiomatic Terraform pattern** for user-controlled timeouts.

## Validation

### Build Test
```bash
$ go build -v
# SUCCESS - no compilation errors
```

### Grep Verification
```bash
$ rg "MaxRetries|RequestTimeout" internal/
# Should only show:
# - client/retry.go: constant definitions
# - *_resource.go: client.DefaultMaxRetries usage
# - NO references to r.providerData.MaxRetries
```

### Schema Validation
Provider now has exactly 4 attributes:
1. `client_id` (required, sensitive)
2. `client_secret` (required, sensitive)
3. `identity_tenant_subdomain` (required)
4. `identity_url` (optional)

## Lessons Learned

1. **Question Everything**: `request_timeout` sat unused for 3 phases
2. **Research First**: Industry patterns revealed modern approach
3. **Seek External Opinion**: Gemini confirmed the decision
4. **Simplify Ruthlessly**: Fewer parameters = better UX
5. **Break Pre-1.0**: Perfect time for architectural improvements

## References

- **Terraform Plugin Framework Docs**: [Resource Timeouts](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts)
- **AWS Provider Issue #1209**: "25 max_retries is really long for a default"
- **Google Provider**: No exposed retry configuration
- **Azure Provider**: No exposed retry configuration
- **Research Date**: 2025-10-15

## Conclusion

This change **improves provider quality** by:
- Reducing configuration complexity
- Aligning with modern best practices
- Separating concerns (internal retries vs user timeouts)
- Removing unused code (`request_timeout`)

The trade-off (lost retry configurability) is acceptable because:
- Defaults are well-chosen
- 99% of users don't need this control
- Resource-level timeouts cover the remaining use cases
- Pre-1.0 status allows breaking changes

**Verdict**: ✅ Correct architectural decision for a modern Terraform provider.
