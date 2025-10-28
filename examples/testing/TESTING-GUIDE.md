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

This test validates all CyberArk SIA Terraform provider resources:
1. **Certificate** - TLS/mTLS certificates
2. **Secret** - Database credentials
3. **Database Workspace** - Database connection configurations
4. **Database Policy** - Access policy metadata and conditions (NEW)
5. **Policy Database Assignment** - Assign databases to access policies

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

## Cloud Provider Integration Testing

### Overview

Testing with cloud-managed databases (AWS RDS, Azure PostgreSQL, GCP Cloud SQL) requires additional setup for cloud provider authentication and resource provisioning.

### Key Finding: Database Target Sets

**CRITICAL**: ALL database workspaces use `"FQDN/IP"` target set in policies, **regardless of cloud_provider attribute**.

| Cloud Provider | Target Set (Actual) |
|----------------|---------------------|
| on_premise | FQDN/IP |
| aws | FQDN/IP |
| azure | FQDN/IP |
| gcp | FQDN/IP |
| atlas | FQDN/IP |

**Impact**: The `cloud_provider` field is **metadata only** for database workspaces. Policy assignment always uses `"FQDN/IP"` target set.

### Azure PostgreSQL Testing

**Canonical Test Template**: `examples/testing/azure-postgresql/`

This directory contains the validated Azure PostgreSQL CRUD test configuration from 2025-10-27.
Use this as the reference implementation for all future cloud provider testing.

#### Quick Start

```bash
cd examples/testing/azure-postgresql
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init && terraform apply
# Verify success, then clean up
terraform destroy
```

#### Prerequisites
1. Azure CLI authentication: `az login`
2. Valid subscription with unrestricted regions
3. SIA credentials in project root `.env` file
4. Existing SIA policy with `locationType: "FQDN/IP"`

#### Recommended Configuration
```hcl
# Azure PostgreSQL Flexible Server (B1ms - cheapest option)
resource "azurerm_postgresql_flexible_server" "test" {
  name                = "psql-sia-test-${random_string.suffix.result}"
  resource_group_name = azurerm_resource_group.test.name
  location            = "westus2"  # Check subscription restrictions

  sku_name   = "B_Standard_B1ms"  # ~$0.017/hour
  storage_mb = 32768              # 32 GB minimum
  version    = "16"

  administrator_login    = var.admin_username
  administrator_password = var.admin_password

  public_network_access_enabled = true  # Required for SIA connectivity
}

# SIA Database Workspace
resource "cyberarksia_database_workspace" "azure_postgres" {
  name                          = "azure-postgres-test-${random_string.suffix.result}"
  database_type                 = "postgres-azure-managed"
  cloud_provider                = "azure"  # Metadata only
  region                        = var.azure_region
  address                       = azurerm_postgresql_flexible_server.test.fqdn
  port                          = 5432
  secret_id                     = cyberarksia_secret.admin.id
  enable_certificate_validation = false  # Simplify testing
  certificate_id                = cyberarksia_certificate.azure_cert.id
}

# Policy Assignment (uses "FQDN/IP" target set for ALL databases)
resource "cyberarksia_policy_database_assignment" "azure_postgres" {
  policy_id              = data.cyberarksia_access_policy.test_policy.id
  database_workspace_id  = cyberarksia_database_workspace.azure_postgres.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["pg_read_all_settings"]
  }
}
```

#### Cost Estimate
- **Hourly**: $0.017 USD (B1ms compute)
- **Test Duration**: ~15-30 minutes
- **Test Cost**: < $0.01 USD

#### Common Issues

**Issue**: LocationIsOfferRestricted
```
Error: Subscriptions are restricted from provisioning in location 'eastus'
```
**Solution**: Change region to `westus2` or check `az account list-locations` for allowed regions.

**Issue**: Firewall Rule Already Exists
**Solution**: Import existing rule:
```bash
terraform import azurerm_postgresql_flexible_server_firewall_rule.allow_azure \
  "/subscriptions/.../firewallRules/AllowAzureServices"
```

#### Azure Certificate

Azure PostgreSQL uses **Microsoft RSA Root Certificate Authority 2017**:

```hcl
resource "cyberarksia_certificate" "azure_cert" {
  cert_name = "azure-postgres-ssl-cert"
  cert_type = "PEM"
  cert_body = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIFqDCCA5CgAwIBAgIQHtOXCV/YtLNHcB6qvn9FszANBgkqhkiG9w0BAQwFADBl
    [... full certificate content ...]
    -----END CERTIFICATE-----
  EOT
}
```

**Download**: https://www.microsoft.com/pkiops/certs/Microsoft%20RSA%20Root%20Certificate%20Authority%202017.crt

### AWS RDS Testing (Template)

**Prerequisites**:
1. AWS CLI authentication: `aws configure`
2. Valid AWS credentials with RDS permissions

**Recommended Configuration**:
- RDS Instance Class: `db.t3.micro` (cheapest)
- Storage: 20 GB minimum
- Public accessibility: Yes (for testing)
- Certificate: AWS RDS CA bundle

### GCP Cloud SQL Testing (Template)

**Prerequisites**:
1. gcloud CLI authentication: `gcloud auth login`
2. Valid GCP project with Cloud SQL API enabled

**Recommended Configuration**:
- Tier: `db-f1-micro` (cheapest)
- Storage: 10 GB minimum
- Public IP: Yes (for testing)
- Certificate: GCP server CA

### Testing Workflow

1. **Setup Cloud Resources** (5-10 minutes)
   ```bash
   terraform apply -target=azurerm_postgresql_flexible_server.test
   terraform apply -target=azurerm_postgresql_flexible_server_database.testdb
   ```

2. **Create SIA Resources** (< 1 minute)
   ```bash
   terraform apply -target=cyberarksia_certificate.azure_cert
   terraform apply -target=cyberarksia_secret.admin
   terraform apply -target=cyberarksia_database_workspace.azure_postgres
   ```

3. **Test Policy Assignment** (< 1 minute)
   ```bash
   terraform apply -target=cyberarksia_policy_database_assignment.azure_postgres
   ```

4. **Verify in SIA UI**
   - Navigate to policy: Terraform-Test-Policy
   - Confirm database appears in targets
   - Verify authentication method and profile

5. **Cleanup** (2-5 minutes)
   ```bash
   terraform destroy -auto-approve
   ```

### Cost Management

**Best Practices**:
- Use smallest instance sizes (B1ms, db.t3.micro, db-f1-micro)
- Add `auto_delete = "true"` tag to all resources
- Set up automatic cleanup GitHub Actions
- Stop/delete resources immediately after testing

**Estimated Costs per Test**:
- Azure PostgreSQL B1ms: < $0.01 USD (15 min test)
- AWS RDS db.t3.micro: < $0.01 USD (15 min test)
- GCP Cloud SQL db-f1-micro: < $0.01 USD (15 min test)

---

## See Also

- [`CLAUDE.md`](../../CLAUDE.md) - Development guidelines (references this guide)
- [`docs/testing-framework.md`](../../docs/testing-framework.md) - Conceptual testing framework
- [`docs/resources/policy_database_assignment.md`](../../docs/resources/policy_database_assignment.md) - Policy assignment documentation
- [`examples/resources/`](../resources/) - Per-resource usage examples
- `/tmp/sia-azure-test-20251027-185657/TEST-RESULTS.md` - Azure PostgreSQL test results (2025-10-27)
