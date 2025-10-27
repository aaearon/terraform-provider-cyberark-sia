# CRUD Testing Framework

> **⚠️ CANONICAL REFERENCE**: For hands-on testing, use [`examples/testing/TESTING-GUIDE.md`](../examples/testing/TESTING-GUIDE.md)
> This document provides conceptual framework and historical reference. For actual testing templates and workflows, refer to [`examples/testing/TESTING-GUIDE.md`](../examples/testing/TESTING-GUIDE.md).

**Purpose**: Comprehensive framework for testing all CRUD operations (Create, Read, Update, Delete) across all Terraform provider resources.

**Last Updated**: 2025-10-27 (Added Policy Database Assignment)

---

## Overview

This framework provides a systematic approach to validate that all provider resources work correctly through their full lifecycle. It tests real API integration (not mocks) against a live CyberArk SIA tenant.

## Test Directory Structure

```
/tmp/sia-crud-validation/
├── provider.tf          # Provider configuration with credentials
├── main.tf              # All resource definitions (certificate, secret, database_workspace, policy_database_assignment)
├── outputs.tf           # Validation outputs
├── test-cert.pem        # Generated test certificate
└── README.md            # Testing instructions
```

## Quick Start

### 1. Generate Test Certificate

```bash
cd /tmp/sia-crud-validation
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out test-cert.pem \
  -days 365 -nodes -subj "/CN=crud-test-cert/O=CRUDValidation/C=US"
```

### 2. Create Provider Configuration

**File**: `provider.tf`

```hcl
terraform {
  required_providers {
    cyberarksia = {
      source  = "terraform.local/local/cyberark-sia"
      version = "0.1.0"
    }
  }
}

provider "cyberarksia" {
  username      = "your-username@cyberark.cloud.XXXX"
  client_secret = "your-client-secret"
}
```

**Note**: Get credentials from `.env` file in project root.

### 3. Create Resource Definitions

**File**: `main.tf`

```hcl
# CRUD Validation Test Configuration
# Tests all four resources: certificate, secret, database_workspace, policy_database_assignment

# DATA SOURCE: Lookup existing access policy for assignment testing
data "cyberarksia_access_policy" "test_policy" {
  name = "Terraform-Test-Policy"
}

# 1. Certificate Resource - Base dependency
resource "cyberarksia_certificate" "test_cert" {
  cert_name        = "crud-test-cert-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  cert_description = "CRUD validation test certificate"
  cert_body        = file("${path.module}/test-cert.pem")
  cert_type        = "PEM"

  labels = {
    environment = "test"
    purpose     = "crud-validation"
    created_at  = formatdate("YYYY-MM-DD", timestamp())
  }
}

# 2. Secret Resource - Required for database workspace
resource "cyberarksia_secret" "test_secret" {
  name                = "crud-test-secret-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  authentication_type = "local"
  username            = "testuser"
  password            = "TestPassword123!"

  tags = {
    environment = "test"
    purpose     = "crud-validation"
  }
}

# 3. Database Workspace Resource - References both certificate and secret
resource "cyberarksia_database_workspace" "test_db" {
  name                          = "crud-test-db-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  database_type                 = "postgres"
  cloud_provider                = "on_premise"  # REQUIRED
  address                       = "test.postgres.local"
  port                          = 5432
  secret_id                     = cyberarksia_secret.test_secret.id
  certificate_id                = cyberarksia_certificate.test_cert.id
  enable_certificate_validation = false  # Disable for test environment

  tags = {
    environment = "test"
    purpose     = "crud-validation"
  }
}

# 4. Policy Database Assignment Resource - Assigns database to access policy
resource "cyberarksia_policy_database_assignment" "test_assignment" {
  policy_id              = data.cyberarksia_access_policy.test_policy.id
  database_workspace_id  = cyberarksia_database_workspace.test_db.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["test_reader", "test_writer"]
  }

  depends_on = [cyberarksia_database_workspace.test_db]
}
```

### 4. Create Outputs

**File**: `outputs.tf`

```hcl
# Certificate Outputs
output "certificate_id" {
  description = "Created certificate ID"
  value       = cyberarksia_certificate.test_cert.id
}

output "certificate_expiration" {
  description = "Certificate expiration date"
  value       = cyberarksia_certificate.test_cert.expiration_date
}

output "certificate_metadata_issuer" {
  description = "Certificate issuer from metadata"
  value       = cyberarksia_certificate.test_cert.metadata.issuer
}

# Secret Outputs
output "secret_id" {
  description = "Created secret ID"
  value       = cyberarksia_secret.test_secret.id
}

output "secret_created_at" {
  description = "Secret creation timestamp"
  value       = cyberarksia_secret.test_secret.created_at
}

# Database Workspace Outputs
output "database_workspace_id" {
  description = "Created database workspace ID"
  value       = cyberarksia_database_workspace.test_db.id
}

# Policy Database Assignment Outputs
output "policy_assignment_id" {
  description = "Policy database assignment composite ID"
  value       = cyberarksia_policy_database_assignment.test_assignment.id
}

output "policy_assignment_platform" {
  description = "Database platform (computed)"
  value       = cyberarksia_policy_database_assignment.test_assignment.platform
}

# Validation Summary
output "validation_summary" {
  description = "Summary of created resources for validation"
  value = {
    policy_found         = data.cyberarksia_access_policy.test_policy.id != ""
    certificate_created  = cyberarksia_certificate.test_cert.id != ""
    secret_created       = cyberarksia_secret.test_secret.id != ""
    database_created     = cyberarksia_database_workspace.test_db.id != ""
    assignment_created   = cyberarksia_policy_database_assignment.test_assignment.id != ""
    dependencies_working = cyberarksia_database_workspace.test_db.secret_id == cyberarksia_secret.test_secret.id
    assignment_working   = cyberarksia_policy_database_assignment.test_assignment.database_workspace_id == cyberarksia_database_workspace.test_db.id
    total_resources      = 4
  }
}
```

---

## Provider Development Workflow

### Build and Install Provider

```bash
# From project root
cd /home/tim/terraform-provider-cyberark-sia

# Build provider
go build -v

# Install to Go bin
go install

# Copy to Terraform plugin directory (for local testing)
cp ~/go/bin/terraform-provider-cyberark-sia \
   ~/.terraform.d/plugins/terraform.local/local/cyberark-sia/0.1.0/linux_amd64/terraform-provider-cyberark-sia
```

**Note**: You need to copy to BOTH locations:
- `~/go/bin/` - For `go install`
- `~/.terraform.d/plugins/.../` - For Terraform to find it

---

## Testing Workflow

### Phase 1: CREATE Test

**Objective**: Validate resource creation and initial state population

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Preview changes
terraform plan

# Create resources
terraform apply -auto-approve

# Verify outputs
terraform output
```

**Expected Results**:
- ✅ 4 resources created successfully (+ 1 data source lookup)
- ✅ All computed fields populated (IDs, timestamps, metadata)
- ✅ No warnings about unknown attributes
- ✅ Dependencies working (database_workspace references secret_id and certificate_id)
- ✅ Policy assignment references database_workspace and policy
- ✅ Outputs show all resource IDs

**Validation Checklist**:
- [ ] Policy data source found "Terraform-Test-Policy"
- [ ] Certificate ID is numeric string
- [ ] Certificate expiration_date is ISO 8601 timestamp
- [ ] Certificate metadata object is populated (issuer, subject, etc.)
- [ ] Secret ID is UUID format
- [ ] Secret created_at timestamp is populated
- [ ] Database workspace ID is numeric
- [ ] Database workspace references correct secret_id and certificate_id
- [ ] Policy assignment composite ID format: `{policy_id}:{database_id}`
- [ ] Policy assignment platform computed correctly (e.g., "ON-PREMISE")
- [ ] Policy assignment references correct policy_id and database_workspace_id
- [ ] validation_summary shows all true (8 checks, total_resources = 4)

### Phase 2: READ Test (State Refresh)

**Objective**: Validate drift detection and state refresh

```bash
# Refresh state from API
terraform refresh

# Verify no changes detected
terraform plan
```

**Expected Results**:
- ✅ No changes detected
- ✅ State matches remote API
- ✅ All computed fields still populated

**Validation Checklist**:
- [ ] `terraform plan` shows "No changes"
- [ ] All outputs remain unchanged
- [ ] State file has all attributes

### Phase 3: UPDATE Test

**Objective**: Validate in-place updates without resource replacement

**Step 1: Modify Configuration**

Edit `main.tf` to change:
- Certificate: description and labels
- Secret: tags
- Database workspace: tags

**Example Changes**:
```hcl
# Certificate
cert_description = "CRUD validation test certificate - UPDATED"
labels = {
  environment = "test"
  purpose     = "crud-validation"
  created_at  = "2025-10-27"
  updated     = "true"  # NEW
}

# Secret
tags = {
  environment = "test"
  purpose     = "crud-validation"
  updated     = "true"  # NEW
}

# Database Workspace
tags = {
  environment = "test"
  purpose     = "crud-validation"
  updated     = "true"   # NEW
  version     = "v2"     # NEW
}
```

**Step 2: Apply Updates**

```bash
# Preview update plan
terraform plan

# Apply updates
terraform apply -auto-approve

# Verify state
terraform show
```

**Expected Results**:
- ✅ Plan shows "0 to add, 4 to change, 0 to destroy"
- ✅ All updates are in-place (no replacements)
- ✅ Resource IDs unchanged
- ✅ Updated attributes reflected in state

**Validation Checklist**:
- [ ] Plan shows `~ update in-place` (NOT `-/+ destroy and then create`)
- [ ] Certificate description updated
- [ ] Certificate labels updated
- [ ] Certificate metadata still populated (not "unknown")
- [ ] Secret tags updated
- [ ] Database workspace tags updated
- [ ] Policy assignment roles updated (db_auth_profile.roles)
- [ ] All resource IDs remain the same (including composite ID)

### Phase 4: DELETE Test

**Objective**: Validate clean resource deletion

```bash
# Preview deletion
terraform plan -destroy

# Delete all resources
terraform destroy -auto-approve

# Verify state is empty
terraform show
```

**Expected Results**:
- ✅ All 4 resources deleted successfully
- ✅ Deletion in correct dependency order (assignment → workspace → secret/cert)
- ✅ State file is empty
- ✅ No orphaned resources in SIA

**Validation Checklist**:
- [ ] Destroy plan shows 4 resources to delete
- [ ] Deletion order: policy_assignment first, then database_workspace, then secret/certificate
- [ ] No dependency errors
- [ ] `terraform show` returns empty state
- [ ] Manual verification in SIA UI: all resources gone
- [ ] Policy still exists (not deleted, only assignment removed)

---

## Common Issues and Solutions

### Issue 1: Provider Binary Not Updated

**Symptom**: Changes to code not reflected in Terraform behavior

**Solution**:
```bash
# Clean rebuild
go clean
go build -v
go install

# Copy to Terraform directory
cp ~/go/bin/terraform-provider-cyberark-sia \
   ~/.terraform.d/plugins/terraform.local/local/cyberark-sia/0.1.0/linux_amd64/

# Reinitialize Terraform
cd /tmp/sia-crud-validation
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Issue 2: Lock File Checksum Mismatch

**Symptom**: "cached package does not match any of the checksums recorded"

**Solution**:
```bash
# Remove lock file and reinitialize
rm .terraform.lock.hcl
terraform init
```

### Issue 3: Schema Validation Errors

**Symptom**: "Missing required argument" or "Unsupported argument"

**Root Cause**: Provider binary is old or wasn't rebuilt after code changes

**Solution**: Follow Issue 1 solution above

### Issue 4: Unknown Attribute After Update

**Symptom**: "Provider returned invalid result object after apply"

**Root Cause**: Update() method not fetching full resource details after API call

**Solution Pattern** (for any resource):
```go
// After UPDATE API call, fetch full resource details
fullResource, err := r.apiClient.GetResource(ctx, resourceID)
if err != nil {
    resp.Diagnostics.Append(client.MapError(err, "read resource after update"))
    return
}

// Map full response to state (ensures all computed fields populated)
mapResourceToState(ctx, fullResource, &plan, resp)
```

---

## Testing Individual Resources

To test a single resource in isolation:

**Certificate Only**:
```hcl
resource "cyberarksia_certificate" "test" {
  cert_name = "test-cert-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  cert_body = file("${path.module}/test-cert.pem")
  cert_type = "PEM"
}

output "cert_id" {
  value = cyberarksia_certificate.test.id
}
```

**Secret Only**:
```hcl
resource "cyberarksia_secret" "test" {
  name                = "test-secret-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  authentication_type = "local"
  username            = "testuser"
  password            = "TestPassword123!"
}

output "secret_id" {
  value = cyberarksia_secret.test.id
}
```

**Database Workspace** (requires secret + certificate):
```hcl
# Must include secret and certificate resources as dependencies
# See full example in main.tf above
```

**Policy Database Assignment** (requires database + policy):
```hcl
data "cyberarksia_access_policy" "test_policy" {
  name = "Terraform-Test-Policy"
}

resource "cyberarksia_database_workspace" "test_db" {
  # ... database configuration ...
}

resource "cyberarksia_policy_database_assignment" "test" {
  policy_id              = data.cyberarksia_access_policy.test_policy.id
  database_workspace_id  = cyberarksia_database_workspace.test_db.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["test_reader"]
  }
}

output "assignment_id" {
  value = cyberarksia_policy_database_assignment.test.id
}
```

---

## Automation Script

For quick automated testing:

```bash
#!/bin/bash
# crud-test.sh - Automated CRUD testing

set -e

echo "=== Building Provider ==="
cd ~/terraform-provider-cyberark-sia
go build -v
go install
cp ~/go/bin/terraform-provider-cyberark-sia \
   ~/.terraform.d/plugins/terraform.local/local/cyberark-sia/0.1.0/linux_amd64/

echo "=== Initializing Test Environment ==="
cd /tmp/sia-crud-validation
rm -rf .terraform .terraform.lock.hcl
terraform init

echo "=== CREATE Test ==="
terraform apply -auto-approve

echo "=== READ Test ==="
terraform plan

echo "=== UPDATE Test ==="
# Modify configuration here or manually
terraform apply -auto-approve

echo "=== DELETE Test ==="
terraform destroy -auto-approve

echo "=== CRUD Tests Complete ==="
```

---

## Best Practices

### 1. Use Timestamps in Resource Names
```hcl
name = "test-resource-${formatdate("YYYYMMDDhhmmss", timestamp())}"
```
Prevents naming conflicts when re-running tests.

### 2. Always Rebuild Provider After Code Changes
```bash
go build -v && go install && cp ~/go/bin/terraform-provider-cyberark-sia ~/.terraform.d/plugins/terraform.local/local/cyberark-sia/0.1.0/linux_amd64/
```

### 3. Reinitialize After Provider Updates
```bash
rm -rf .terraform .terraform.lock.hcl && terraform init
```

### 4. Verify Outputs After Each Operation
```bash
terraform output
```

### 5. Check State for Computed Fields
```bash
terraform state show cyberarksia_certificate.test_cert
```

### 6. Use Descriptive Labels/Tags for Test Resources
```hcl
labels = {
  environment = "test"
  purpose     = "crud-validation"
  test_run    = "2025-10-27"
}
```

### 7. Always Clean Up After Testing
```bash
terraform destroy -auto-approve
cd ~ && rm -rf /tmp/sia-crud-validation
```

---

## What This Framework Tests

### Data Source Lookup
- ✅ Policy lookup by name works
- ✅ Policy ID retrieved correctly
- ✅ Can reference policy in resources

### Resource Creation (CREATE)
- ✅ API accepts all required fields
- ✅ API accepts all optional fields
- ✅ Computed fields populated after creation
- ✅ Dependencies resolved correctly
- ✅ Composite IDs generated correctly (policy assignments)
- ✅ State saved correctly

### Resource Reading (READ)
- ✅ GET API returns all fields
- ✅ State refresh works
- ✅ Drift detection works
- ✅ No perpetual diffs

### Resource Updates (UPDATE)
- ✅ PUT API accepts changes
- ✅ Updates applied in-place (no replacement)
- ✅ Computed fields remain populated
- ✅ Changed fields reflected in state
- ✅ Unchanged fields preserved

### Resource Deletion (DELETE)
- ✅ DELETE API succeeds
- ✅ Dependencies prevent premature deletion
- ✅ Correct deletion order (assignments before workspaces)
- ✅ State cleaned up
- ✅ No orphaned resources
- ✅ Policy not deleted (only assignment removed)

### Schema Validation
- ✅ All attributes match SDK fields
- ✅ No warnings about unknown attributes
- ✅ Required fields enforced
- ✅ Optional fields work
- ✅ Sensitive fields marked correctly

### Error Handling
- ✅ Validation errors are clear
- ✅ API errors mapped to diagnostics
- ✅ Retry logic works for transient errors
- ✅ Dependency errors handled

---

## Future Enhancements

### 1. Import Testing
Add import validation:
```bash
# Remove from state
terraform state rm cyberarksia_certificate.test_cert

# Import by ID
terraform import cyberarksia_certificate.test_cert <certificate_id>

# Verify
terraform plan
```

### 2. Concurrent Resource Testing
Test creating multiple resources in parallel.

### 3. Error Scenario Testing
- Invalid credentials
- Missing required fields
- API rate limiting
- Network failures

### 4. Performance Testing
- Time each operation
- Test with large resource counts

### 5. Drift Detection Testing
- Manual changes in SIA UI
- Verify Terraform detects drift

---

## Reference: Files in This Framework

1. **provider.tf** - Provider configuration
2. **main.tf** - Resource definitions
3. **outputs.tf** - Validation outputs
4. **test-cert.pem** - Test certificate
5. **README.md** - Testing instructions

All files should be version controlled EXCEPT credentials (use .env instead).

---

## Questions?

See:
- `docs/sdk-integration.md` - ARK SDK integration patterns
- `docs/troubleshooting.md` - Common issues
- `CLAUDE.md` - Development guidelines
- `examples/` - Example configurations
