# Certificate Resource Implementation Summary

**Date**: 2025-10-24
**Status**: ✅ COMPLETE - All CRUD operations implemented

## Overview

Complete Terraform resource implementation for CyberArk SIA certificate management at `/home/tim/terraform-provider-cyberark-sia/internal/provider/resource_certificate.go`.

## Implementation Details

### A. Resource Schema ✅ COMPLETE (22 attributes)

All 22 attributes from data-model.md implemented:

**Input Attributes (9)**:
1. `id` - Computed, alias for certificate_id
2. `certificate_id` - Computed, unique identifier
3. `cert_name` - Optional string
4. `cert_body` - **Required**, **Sensitive**, with `UseStateForUnknown()` modifier
5. `cert_description` - Optional string
6. `cert_type` - Optional string (default "PEM", validates "PEM"|"DER")
7. `cert_password` - Optional, **Sensitive**, with `UseStateForUnknown()` modifier
8. `domain_name` - Optional string
9. `labels` - Optional map[string]string (max 10)

**Computed Attributes (7)**:
10. `expiration_date` - Computed string (ISO 8601)
11. `checksum` - Computed string (SHA256 hash)
12. `version` - Computed int64 (drift detection)
13. `tenant_id` - Computed string
14. `created_by` - Computed string (user email)
15. `last_updated_by` - Computed string (nullable)
16. `updated_time` - Computed string (drift detection)

**Nested Metadata Block (6 fields)**:
17-22. `metadata` - Computed nested block:
   - `issuer` - Computed string (DN)
   - `subject` - Computed string (DN)
   - `valid_from` - Computed string (Unix timestamp)
   - `valid_to` - Computed string (Unix timestamp)
   - `serial_number` - Computed string (decimal format)
   - `subject_alternative_name` - Computed list of strings

### B. CertificateModel Struct ✅ COMPLETE

```go
type CertificateModel struct {
    ID            types.String `tfsdk:"id"`
    CertificateID types.String `tfsdk:"certificate_id"`
    CertName      types.String `tfsdk:"cert_name"`
    CertBody      types.String `tfsdk:"cert_body"`        // SENSITIVE
    CertDescription types.String `tfsdk:"cert_description"`
    CertType      types.String `tfsdk:"cert_type"`
    CertPassword  types.String `tfsdk:"cert_password"`    // SENSITIVE
    DomainName    types.String `tfsdk:"domain_name"`
    Labels        types.Map    `tfsdk:"labels"`
    ExpirationDate types.String `tfsdk:"expiration_date"`
    Checksum      types.String `tfsdk:"checksum"`
    Version       types.Int64  `tfsdk:"version"`
    TenantID      types.String `tfsdk:"tenant_id"`
    CreatedBy     types.String `tfsdk:"created_by"`
    LastUpdatedBy types.String `tfsdk:"last_updated_by"`
    UpdatedTime   types.String `tfsdk:"updated_time"`
    Metadata      types.Object `tfsdk:"metadata"`         // Nested block
}
```

### C. CRUD Methods ✅ ALL IMPLEMENTED

#### 1. Create() ✅ COMPLETE
**Implementation**: Lines 265-365

**Features**:
- Validates PEM certificate before API call (client.ValidatePEMCertificate)
- Builds CertificateCreateRequest with required cert_body
- Calls client.CreateCertificate() with retry logic
- Maps 8 CREATE response fields to state
- Error handling via client.MapCertificateError()
- Structured logging (NEVER logs cert_body/cert_password)
- Supports all optional fields (cert_name, description, type, password, domain, labels)

**Response Mapping** (8 fields from CREATE):
- certificate_id, tenant_id, cert_body, cert_name, cert_description
- domain_name, expiration_date, labels

**Missing from CREATE** (populated by Read):
- metadata, checksum, version, created_by, last_updated_by, updated_time

#### 2. Read() ✅ COMPLETE
**Implementation**: Lines 368-517

**Features**:
- Calls client.GetCertificate() to fetch full certificate
- Handles 404 Not Found → removes from state (drift detection)
- Maps all 14 GET response fields including nested metadata
- Drift detection via version, checksum, updated_time
- Properly constructs metadata types.Object from CertificateMetadata
- Handles nullable fields (last_updated_by, SANs, labels)

**Response Mapping** (14 fields from GET):
- All CREATE fields (8) PLUS:
- checksum, version, created_by, last_updated_by, updated_time, metadata

**Metadata Nested Block Mapping**:
```go
metadataAttrTypes := map[string]attr.Type{
    "issuer":                   types.StringType,
    "subject":                  types.StringType,
    "valid_from":               types.StringType,
    "valid_to":                 types.StringType,
    "serial_number":            types.StringType,
    "subject_alternative_name": types.ListType{ElemType: types.StringType},
}
metadataObj, _ := types.ObjectValue(metadataAttrTypes, metadataValues)
```

#### 3. Update() ✅ COMPLETE
**Implementation**: Lines 520-657

**Critical Features**:
- **ISSUE #4 FIX**: cert_body retrieval with fallback to state
  ```go
  certBody := plan.CertBody.ValueString()
  if certBody == "" || plan.CertBody.IsUnknown() {
      certBody = state.CertBody.ValueString()  // CRITICAL fallback
  }
  ```
- **ISSUE #4 FIX**: cert_password fallback pattern (same as cert_body)
- Validates certificate if cert_body changed
- Builds CertificateUpdateRequest with cert_body (REQUIRED for all updates)
- Calls client.UpdateCertificate() with retry logic
- Maps 8 UPDATE response fields to state
- Persists cert_body and cert_password in state (UseStateForUnknown)

**API Requirement** (Issue #4):
- cert_body is REQUIRED for ALL updates (even metadata-only)
- API returns 400 Bad Request if cert_body missing
- Provider MUST persist cert_body in state using `UseStateForUnknown()` modifier

#### 4. Delete() ✅ COMPLETE
**Implementation**: Lines 660-690

**Features**:
- Calls client.DeleteCertificate()
- Handles errors via client.MapCertificateError()
- Special handling for CERTIFICATE_IN_USE (409 Conflict)
- Structured logging
- State automatically removed by Terraform framework on success

**Error Handling**:
- 409 CERTIFICATE_IN_USE → Actionable error listing dependent workspaces
- 404 Not Found → Treated as success (already deleted)
- Other errors → Mapped to appropriate Terraform diagnostics

#### 5. ImportState() ✅ COMPLETE
**Implementation**: Lines 693-701

**Features**:
- Uses passthrough pattern: `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`
- Import by certificate_id, then calls Read() to populate full state
- Structured logging

**Import Syntax**:
```bash
terraform import cyberark_sia_certificate.example <certificate_id>
```

**Import Behavior**:
1. Fetch full certificate from GET /api/certificates/{id}
2. Populate all 22 fields including cert_body (API returns it)
3. cert_body is stored in state (required for future updates)

### D. Security & Validation ✅ COMPLETE

**1. Sensitive Data Protection**:
- `cert_body` marked as `Sensitive: true`
- `cert_password` marked as `Sensitive: true`
- NEVER logged in tflog or error messages
- Stored encrypted in Terraform remote state

**2. Certificate Validation**:
- Client-side PEM validation via client.ValidatePEMCertificate()
- Validates:
  - PEM format (must contain "-----BEGIN CERTIFICATE-----")
  - PEM decode to ASN.1 DER
  - X.509 parse (valid certificate structure)
  - Private key check (MUST NOT contain private key material)
  - ❌ NO EXPIRATION CHECK (deferred to SIA API per Issue #13)

**3. Schema Validators**:
- `cert_name`: stringvalidator.LengthBetween(1, 255)
- `cert_type`: stringvalidator.OneOf("PEM", "DER")
- `labels`: mapvalidator.SizeAtMost(10)

**4. Error Handling**:
- Uses client.MapCertificateError() for all client errors
- Certificate-specific errors:
  - CERTIFICATE_IN_USE (409)
  - DUPLICATE_NAME (409)
  - INVALID_CERTIFICATE (400)
- Fallback to generic MapError() for standard errors

### E. Provider Registration ✅ VERIFIED

**File**: `/home/tim/terraform-provider-cyberark-sia/internal/provider/provider.go`

```go
// Line 195
func (p *CyberArkSIAProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewDatabaseWorkspaceResource,
        NewSecretResource,
        NewCertificateResource,  // ✅ REGISTERED
    }
}
```

**Provider Configure**: Lines 236-262
- Initializes CertificatesClient on-demand in resource Configure()
- Uses shared ISPAuth from provider for authentication

### F. Compilation Status ✅ SUCCESS

```bash
$ go build -v
github.com/aaearon/terraform-provider-cyberark-sia/internal/provider
github.com/aaearon/terraform-provider-cyberark-sia
```

**No Errors**: All code compiles successfully

## Key Implementation Patterns

### 1. cert_body State Persistence (Issue #4 Fix)
```go
// Schema definition (line 117-125)
"cert_body": schema.StringAttribute{
    Required:  true,
    Sensitive: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(), // CRITICAL!
    },
},

// Update() fallback logic (line 536-539)
certBody := plan.CertBody.ValueString()
if certBody == "" || plan.CertBody.IsUnknown() {
    certBody = state.CertBody.ValueString()  // Fallback
}
```

### 2. Metadata Nested Block Mapping
```go
// Read() metadata conversion (lines 462-505)
metadataModel := CertificateMetadataModel{
    Issuer:       types.StringValue(certificate.Metadata.Issuer),
    Subject:      types.StringValue(certificate.Metadata.Subject),
    ValidFrom:    types.StringValue(certificate.Metadata.ValidFrom),
    ValidTo:      types.StringValue(certificate.Metadata.ValidTo),
    SerialNumber: types.StringValue(certificate.Metadata.SerialNumber),
}

// Convert to types.Object
metadataObj, _ := types.ObjectValue(metadataAttrTypes, metadataValues)
state.Metadata = metadataObj
```

### 3. Error Handling Pattern
```go
// All CRUD methods use consistent error handling
certificate, err := r.certificatesAPI.CreateCertificate(ctx, createReq)
if err != nil {
    resp.Diagnostics.Append(client.MapCertificateError(err, "create certificate"))
    return
}
```

### 4. Structured Logging (Security)
```go
tflog.Info(ctx, "Creating certificate", map[string]interface{}{
    "cert_name": createReq.CertName,
    // NEVER log cert_body or cert_password!
})
```

## Alignment with Specifications

### Data Model (specs/002-end-users-need/data-model.md)
✅ All 22 attributes implemented exactly as specified
✅ Nested metadata block with 6 fields
✅ Sensitive fields marked correctly (cert_body, cert_password)
✅ Plan modifiers applied (UseStateForUnknown for cert_body/cert_password)

### API Contracts (specs/002-end-users-need/contracts/certificates-api.md)
✅ CREATE: Maps 8 response fields
✅ READ: Maps 14 response fields including metadata
✅ UPDATE: Maps 8 response fields, requires cert_body
✅ DELETE: Handles 409 CERTIFICATE_IN_USE
✅ IMPORT: Uses passthrough pattern

### Client Layer (internal/client/certificates.go)
✅ All CRUD methods use correct ARK SDK patterns
✅ RetryWithBackoff for transient errors
✅ MapCertificateError for error mapping
✅ ValidatePEMCertificate for client-side validation

## Testing Checklist

### Unit Tests (Future)
- [ ] PEM validation logic
- [ ] Metadata object conversion
- [ ] Error mapping (CERTIFICATE_IN_USE, DUPLICATE_NAME)

### Acceptance Tests (TF_ACC=1)
- [ ] Create certificate with all fields
- [ ] Create certificate with minimal fields (cert_body only)
- [ ] Read certificate and verify drift detection
- [ ] Update certificate metadata (cert_name, description, labels)
- [ ] Update certificate content (cert_body replacement)
- [ ] Delete certificate (success case)
- [ ] Delete certificate in-use (409 error)
- [ ] Import existing certificate
- [ ] Certificate with SANs
- [ ] Certificate with labels (map handling)
- [ ] Certificate expiration edge case (Issue #13)

## Known Issues Resolved

### Issue #4: cert_body Required for Updates ✅ FIXED
**Problem**: API requires cert_body for ALL updates (even metadata-only)
**Solution**: Fallback to state.CertBody when plan.CertBody is unknown
**Implementation**: Lines 536-539

### Issue #13: Expiration Validation ✅ FIXED
**Problem**: Client-side expiration check blocks legitimate workflows
**Solution**: Removed expiration check from ValidatePEMCertificate()
**Rationale**: Defer expiration enforcement to SIA API

### Issue #3: Nested Metadata Block ✅ IMPLEMENTED
**Problem**: metadata is nested object with 6 fields
**Solution**: Proper types.Object mapping with all 6 metadata fields
**Implementation**: Lines 462-505

### Issue #14: Drift Detection ✅ IMPLEMENTED
**Problem**: Need version, checksum, updated_time for drift detection
**Solution**: All 3 fields mapped from GET response
**Implementation**: Lines 448-459, 507-513

## Files Modified

1. `/home/tim/terraform-provider-cyberark-sia/internal/provider/resource_certificate.go` (NEW - 701 lines)
   - Complete resource implementation with all CRUD methods
   - Schema with 22 attributes (16 top-level + 6 nested)
   - Security-hardened with sensitive field handling

2. `/home/tim/terraform-provider-cyberark-sia/internal/provider/provider.go` (VERIFIED)
   - Resource registered in Resources() method (line 195)

3. `/home/tim/terraform-provider-cyberark-sia/internal/client/certificates.go` (EXISTING - verified)
   - All CRUD methods use correct ARK SDK patterns
   - ValidatePEMCertificate() without expiration check

## Next Steps

1. **Acceptance Testing**: Write comprehensive tests for all CRUD operations
2. **Documentation**: Update provider documentation with certificate examples
3. **Integration Testing**: Test with database_workspace.certificate_id reference
4. **Edge Cases**: Test certificate chains, expiration scenarios, delete protection

## Summary

✅ **COMPLETE**: Full Terraform resource implementation for certificate management
✅ **22/22 Attributes**: All schema attributes from data-model.md implemented
✅ **5/5 CRUD Methods**: Create, Read, Update, Delete, ImportState all complete
✅ **Compilation**: Code compiles with no errors
✅ **Provider Registration**: Resource registered in provider
✅ **Security**: Sensitive fields protected, validation implemented
✅ **Error Handling**: Certificate-specific errors mapped to actionable diagnostics
✅ **Issue #4 Fixed**: cert_body fallback pattern implemented
✅ **Issue #13 Fixed**: Expiration validation deferred to API

**Status**: Ready for acceptance testing and integration with database_workspace resource.
