# Comprehensive CRUD Testing Guide

> **STATUS**: ⚠️ **CANONICAL REFERENCE** ⚠️ (Mandatory)

This guide is the **single source of truth** for CRUD testing of the CyberArk SIA Terraform provider.

---

## About This Guide

### Document Authority

- **Location**: `examples/testing/TESTING-GUIDE.md`
- **Referenced by**: `CLAUDE.md`, `docs/testing-framework.md`, all template files
- **Maintainers**: Must update when resource schemas or testing requirements change
- **Version**: See git history for changes

### When to Update This Guide

Update this guide when:
1. ✅ Adding a new resource type
2. ✅ Changing resource schemas or dependencies
3. ✅ Discovering new validation requirements
4. ✅ Adding new troubleshooting scenarios
5. ✅ Updating provider configuration requirements

### When NOT to Follow This Guide

This guide is for **integration testing** (real API). For:
- **Unit testing**: See `internal/*/` test files
- **Acceptance testing**: See `tests/` (future)
- **CI/CD testing**: See `.github/workflows/` (future)

### Reporting Issues

If templates or procedures are outdated:
1. Create issue in GitHub repo
2. Update this guide first (it's the source of truth)
3. Then update templates to match

---

## Resources Tested

This test validates all four CyberArk SIA Terraform provider resources:
1. **Certificate** - TLS/mTLS certificates
2. **Secret** - Database credentials
3. **Database Workspace** - Database connection configurations
4. **Policy Database Assignment** - Assign databases to access policies

---

## Prerequisites

- ✅ CyberArk SIA tenant with UAP service provisioned
- ✅ Valid credentials (username + client_secret)
- ✅ Existing access policy named "Terraform-Test-Policy"
- ✅ Provider built and installed

---

## Quick Start

### 1. Setup Test Environment

```bash
# Create test directory
mkdir -p /tmp/sia-crud-validation
cd /tmp/sia-crud-validation

# Copy templates from project
cp ~/terraform-provider-cyberark-sia/examples/testing/crud-test-*.tf .

# Generate test certificate
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out test-cert.pem \
  -days 365 -nodes -subj "/CN=crud-test-full/O=CRUDValidation/C=US"
```

### 2. Configure Provider

Edit `provider.tf` with your credentials (from project root `.env` file):

```hcl
provider "cyberarksia" {
  username      = "your-username@cyberark.cloud.XXXX"
  client_secret = "your-client-secret"
}
```

### 3. Build and Install Provider

```bash
cd ~/terraform-provider-cyberark-sia
go build -v
go install
```

### 4. Initialize Terraform

```bash
cd /tmp/sia-crud-validation
terraform init
```

### 5. Run Complete CRUD Test

```bash
# CREATE - Create all 4 resources
terraform apply -auto-approve

# READ - Verify state matches API
terraform plan  # Should show "No changes"

# UPDATE - Modify tags/labels in main.tf, then:
terraform apply -auto-approve

# DELETE - Clean up all resources
terraform destroy -auto-approve
```

---

## Resource Dependencies

```
┌─────────────────────────┐
│ Data Source:            │
│ access_policy (lookup)  │◄────┐
└─────────────────────────┘     │
                                │
┌─────────────────────────┐     │
│ Resource:               │     │
│ certificate (TLS cert)  │     │
└──────────┬──────────────┘     │
           │                     │
           │  ┌─────────────────────────┐
           │  │ Resource:               │
           │  │ secret (DB credentials) │
           │  └──────────┬──────────────┘
           │             │
           │             │
           ▼             ▼
┌─────────────────────────────────────┐
│ Resource:                           │
│ database_workspace                  │
│ (references cert + secret)          │
└──────────┬──────────────────────────┘
           │
           │
           ▼
┌─────────────────────────────────────┐
│ Resource:                           │
│ policy_database_assignment          │
│ (assigns database to policy)        │
└─────────────────────────────────────┘
```

---

## Expected Output

After `terraform apply`, you should see:

```
Outputs:

validation_summary = {
  "assignment_created" = true
  "assignment_has_database" = true
  "assignment_has_policy" = true
  "certificate_created" = true
  "database_created" = true
  "database_has_certificate" = true
  "database_has_secret" = true
  "policy_found" = true
  "secret_created" = true
  "total_resources_created" = 4
}

test_completion_message = "✅ All 4 resources created successfully! Review validation_summary for dependency verification."
```

---

## Validation Checklists

### Certificate Resource
- [ ] Certificate ID is numeric string
- [ ] `expiration_date` is ISO 8601 timestamp
- [ ] `metadata` object populated (issuer, subject, valid_from, valid_to)
- [ ] Labels saved correctly

### Secret Resource
- [ ] Secret ID is UUID format
- [ ] `created_at` timestamp populated
- [ ] `authentication_type` matches input
- [ ] Tags saved correctly

### Database Workspace Resource
- [ ] Database ID is numeric
- [ ] `secret_id` matches created secret
- [ ] `certificate_id` matches created certificate
- [ ] `database_type` set correctly
- [ ] `cloud_provider` defaults to "on_premise"

### Policy Database Assignment Resource
- [ ] Composite ID format: `{policy_id}:{database_id}`
- [ ] `policy_id` matches policy data source
- [ ] `database_workspace_id` matches created database
- [ ] `authentication_method` set to "db_auth"
- [ ] `platform` computed from database workspace
- [ ] `db_auth_profile.roles` saved correctly

### Data Source
- [ ] Policy found by name
- [ ] Policy ID retrieved
- [ ] Policy status shown

---

## Troubleshooting

### Error: Policy Not Found
**Symptom**: `Error: Policy not found: Terraform-Test-Policy`

**Solution**: Ensure "Terraform-Test-Policy" exists in your tenant, or modify the policy name in `main.tf`:
```hcl
data "cyberarksia_access_policy" "test_policy" {
  name = "Your-Policy-Name-Here"
}
```

### Error: No UAP Service
**Symptom**: DNS lookup failures or "service not available" errors

**Solution**: Verify tenant has UAP service provisioned:
```bash
curl -s "https://platform-discovery.cyberark.cloud/api/v2/services/subdomain/{your-tenant}" | jq '.jit // .dpa'
```

If UAP/JIT/DPA is not in the response, contact CyberArk support to provision the service.

### Error: Provider Binary Not Found
**Symptom**: `Error: Failed to query available provider packages`

**Solution**: Rebuild and reinstall provider:
```bash
cd ~/terraform-provider-cyberark-sia
go clean && go build -v && go install
```

### Error: Schema Validation Failed
**Symptom**: `Error: Missing required argument` or `Unsupported argument`

**Solution**: Reinitialize Terraform:
```bash
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Error: Lock File Checksum Mismatch
**Symptom**: `cached package does not match any of the checksums recorded`

**Solution**:
```bash
rm .terraform.lock.hcl
terraform init
```

---

## UPDATE Testing

To test UPDATE operations, modify `main.tf`:

### Certificate Updates
```hcl
cert_description = "CRUD validation test certificate - UPDATED"
labels = {
  environment = "test"
  purpose     = "crud-validation"
  suite       = "full"
  created_at  = formatdate("YYYY-MM-DD", timestamp())
  updated     = "true"  # NEW
}
```

### Secret Updates
```hcl
tags = {
  environment = "test"
  purpose     = "crud-validation"
  suite       = "full"
  updated     = "true"  # NEW
}
```

### Database Workspace Updates
```hcl
tags = {
  environment = "test"
  purpose     = "crud-validation"
  suite       = "full"
  updated     = "true"  # NEW
}
```

### Policy Assignment Updates
```hcl
db_auth_profile {
  roles = ["crud_test_admin", "crud_test_auditor"]  # CHANGED
}
```

Then apply:
```bash
terraform apply -auto-approve
```

**Expected result**: `0 to add, 4 to change, 0 to destroy`

---

## Clean Up

```bash
# Remove all test resources
terraform destroy -auto-approve

# Remove test directory (optional)
cd ~ && rm -rf /tmp/sia-crud-validation
```

---

## Files in Test Directory

After setup, your test directory should contain:

```
/tmp/sia-crud-validation/
├── provider.tf       # Provider configuration (from template)
├── main.tf           # All 4 resources + policy data source (from template)
├── outputs.tf        # Comprehensive validation outputs (from template)
├── test-cert.pem     # Generated test certificate
└── key.pem           # Certificate private key (not used by provider)
```

---

## Best Practices

### 1. Always Use Templates
**DO NOT** create ad-hoc test configurations. Always start from:
- `examples/testing/crud-test-provider.tf`
- `examples/testing/crud-test-main.tf`
- `examples/testing/crud-test-outputs.tf`

### 2. Use Timestamped Working Directories
```bash
# Good - includes timestamp
mkdir -p /tmp/sia-crud-validation-$(date +%Y%m%d-%H%M%S)

# Also good - standard name
mkdir -p /tmp/sia-crud-validation
```

### 3. Never Modify Templates Directly
Templates in `examples/testing/` are canonical references. Copy to working directory first.

### 4. Always Rebuild After Code Changes
```bash
cd ~/terraform-provider-cyberark-sia
go build -v && go install
```

### 5. Reinitialize After Provider Updates
```bash
cd /tmp/sia-crud-validation
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### 6. Verify Outputs After Each Operation
```bash
terraform output validation_summary
```

---

## See Also

- [`CLAUDE.md`](../../CLAUDE.md) - Development guidelines (references this guide)
- [`docs/testing-framework.md`](../../docs/testing-framework.md) - Conceptual testing framework
- [`docs/resources/policy_database_assignment.md`](../../docs/resources/policy_database_assignment.md) - Policy assignment documentation
- [`examples/resources/`](../resources/) - Per-resource usage examples
