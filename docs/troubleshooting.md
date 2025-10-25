# Troubleshooting Guide: CyberArk SIA Terraform Provider

## Partial State Failures

### Scenario: Database Created but SIA Onboarding Failed

**Problem**: Your AWS RDS or Azure SQL database was successfully created, but the SIA database workspace registration failed. This leaves you with a database in AWS/Azure but no corresponding entry in SIA.

**Symptoms**:
```
Error: Failed to create database workspace
│ SIA API error: [specific error message]
│
│ Recommended action: [recovery guidance]
```

**Root Causes**:
1. **Authentication Issues**: ISPSS credentials expired or invalid
2. **Permission Issues**: Service account lacks required SIA roles
3. **Network Issues**: SIA API unreachable from Terraform execution environment
4. **API Validation**: Database type/version not supported by SIA
5. **Resource Limits**: SIA tenant quota exceeded

### Recovery Options

#### Option 1: Fix the Issue and Re-apply (Recommended)

This is the safest approach as it maintains Terraform state consistency.

**Steps**:
1. Diagnose and fix the underlying issue (see specific error guidance below)
2. Run `terraform apply` again
3. Terraform will detect that the AWS/Azure database exists and skip its creation
4. Terraform will retry the SIA database workspace creation
5. Verify both resources are now in state: `terraform state list`

**Example**:
```bash
# Fix authentication issue (e.g., renew ISPSS credentials)
export TF_VAR_cyberark_client_secret="new-secret-value"

# Re-apply configuration
terraform apply

# Expected output:
# aws_db_instance.main: Refreshing state... [id=mydb]
# cyberark_sia_database_workspace.main: Creating...
# cyberark_sia_database_workspace.main: Creation complete [id=target-123]
```

#### Option 2: Import Existing Database (If SIA Entry Was Created Manually)

If you manually registered the database in SIA console during troubleshooting:

**Steps**:
1. Find the SIA database workspace ID from the SIA console
2. Import it into Terraform state:
   ```bash
   terraform import cyberark_sia_database_workspace.main <sia-target-id>
   ```
3. Run `terraform plan` to verify state matches configuration
4. Run `terraform apply` if needed to reconcile differences

#### Option 3: Start Over (Nuclear Option)

Only use this if Options 1 and 2 don't work.

**Steps**:
1. Remove the cloud database (if not needed):
   ```bash
   # For AWS
   aws rds delete-db-instance --db-instance-identifier mydb --skip-final-snapshot

   # For Azure
   az sql db delete --name mydb --resource-group myrg --server myserver
   ```
2. Remove the resource from Terraform state:
   ```bash
   terraform state rm aws_db_instance.main
   terraform state rm cyberark_sia_database_workspace.main
   ```
3. Run `terraform apply` to recreate everything

## Specific Error Scenarios

### Error: Authentication Failed

**Error Message**:
```
Authentication failed: Invalid client_id or client_secret
```

**Resolution**:
1. Verify ISPSS credentials are correct
2. Check if service account is enabled in Identity
3. Ensure credentials haven't expired
4. Re-run with updated credentials

**Prevention**:
- Use secret managers (AWS Secrets Manager, Azure Key Vault) for credential storage
- Set up credential rotation workflows
- Monitor service account expiration dates

### Error: Insufficient Permissions

**Error Message**:
```
Insufficient permissions: Service account lacks required role memberships
```

**Resolution**:
1. Log in to CyberArk Identity Admin Portal
2. Navigate to Users & Roles
3. Find your ISPSS service account
4. Add required roles:
   - `SIA Database Administrator` (or equivalent)
   - `SIA Secrets Manager` (for secrets)
5. Wait 2-5 minutes for permissions to propagate
6. Re-run `terraform apply`

**Prevention**:
- Document required roles in your infrastructure documentation
- Use Infrastructure as Code to manage Identity roles (if supported)
- Implement least-privilege access control

### Error: Database Type Not Supported

**Error Message**:
```
Database version 9.5.0 is below minimum for postgresql (10.0.0 required)
```

**Resolution**:
1. Check SIA documentation for supported database versions
2. Upgrade your database to a supported version:
   ```hcl
   resource "aws_db_instance" "main" {
     engine_version = "14.7" # Updated to supported version
   }
   ```
3. Run `terraform apply`

**Prevention**:
- Always check [SIA supported databases documentation](https://docs.cyberark.com/sia) before provisioning
- Use Terraform variables for database versions with validation:
  ```hcl
  variable "postgres_version" {
    type = string
    validation {
      condition     = tonumber(split(".", var.postgres_version)[0]) >= 10
      error_message = "PostgreSQL version must be 10.0 or higher for SIA compatibility."
    }
  }
  ```

### Error: Network Connectivity Issues

**Error Message**:
```
SIA service unavailable: Connection timeout
```

**Resolution**:
1. Verify network connectivity from Terraform execution environment to SIA API:
   ```bash
   curl -I https://your-tenant.cyberark.cloud/api/health
   ```
2. Check firewall rules and security groups
3. Verify DNS resolution
4. Check proxy settings if applicable

**Prevention**:
- Run Terraform from a network location with reliable SIA API access
- Configure appropriate network timeouts in provider configuration:
  ```hcl
  provider "cyberark_sia" {
    request_timeout = 60 # seconds
  }
  ```

### Error: Conflict (Resource Already Exists)

**Error Message**:
```
Resource already exists: A database workspace with name 'prod-postgres' already exists
```

**Resolution**:
1. Check SIA console for existing database workspace
2. Either:
   - **Option A**: Import the existing resource:
     ```bash
     terraform import cyberark_sia_database_workspace.main <existing-target-id>
     ```
   - **Option B**: Rename your new resource:
     ```hcl
     resource "cyberark_sia_database_workspace" "main" {
       name = "prod-postgres-v2" # Changed name
     }
     ```

## State Drift Detection

### Scenario: Resource Modified Outside Terraform

**Problem**: Someone modified the SIA database workspace or secret directly in the SIA console, causing state drift.

**Detection**:
Run `terraform plan` regularly to detect drift:
```bash
terraform plan

# Output shows drift:
# Note: Objects have changed outside of Terraform
# cyberark_sia_database_workspace.main: Refreshing state... [id=target-123]
# Resource actions are indicated with the following symbols:
# ~ update in-place
```

**Resolution**:
1. Review the detected changes
2. Decide whether to:
   - **Accept the manual change**: Update your Terraform configuration to match
   - **Revert the manual change**: Run `terraform apply` to restore Terraform-managed values

### Scenario: Resource Deleted Outside Terraform

**Problem**: Someone deleted the SIA database workspace from the console.

**Detection**:
```bash
terraform plan

# Output:
# cyberark_sia_database_workspace.main: Refreshing state... [id=target-123]
# Warning: Database target not found, removing from state
```

**Resolution**:
1. Verify the resource was intentionally deleted
2. Run `terraform apply` to recreate it:
   ```bash
   terraform apply
   # Output: cyberark_sia_database_workspace.main will be created
   ```

## Best Practices for Preventing Issues

### 1. Use Terraform Workspaces for Environments

Separate state for dev, staging, and production:
```bash
terraform workspace new production
terraform workspace select production
terraform apply
```

### 2. Enable State Locking

Use remote state with locking to prevent concurrent modifications:
```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "sia/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-lock"
    encrypt        = true
  }
}
```

### 3. Implement CI/CD Validation

Run `terraform plan` in CI/CD before merging:
```yaml
# Example GitHub Actions workflow
- name: Terraform Plan
  run: terraform plan -out=tfplan
- name: Verify No Errors
  run: terraform show tfplan | grep -q "Error:" && exit 1 || exit 0
```

### 4. Regular State Backups

Backup Terraform state regularly:
```bash
# Automated backup
terraform state pull > backups/terraform-$(date +%Y%m%d-%H%M%S).tfstate
```

### 5. Use Terraform Modules

Encapsulate common patterns in reusable modules:
```hcl
module "sia_database" {
  source = "./modules/sia-rds-integration"

  environment    = "production"
  database_type  = "postgresql"
  database_version = "14.7"
}
```

## Getting Help

### Log Collection

When opening support tickets, include:

1. **Terraform Version**:
   ```bash
   terraform version
   ```

2. **Provider Version**:
   ```bash
   terraform providers
   ```

3. **Debug Logs** (redact sensitive data):
   ```bash
   TF_LOG=DEBUG terraform apply 2>&1 | tee terraform-debug.log
   ```

4. **State Information**:
   ```bash
   terraform state list
   terraform state show cyberark_sia_database_workspace.main
   ```

### Support Channels

- **Provider Issues**: https://github.com/aaearon/terraform-provider-cyberark-sia/issues
- **CyberArk SIA Support**: https://www.cyberark.com/customer-support/
- **Community Forums**: CyberArk Commons

## Additional Resources

- [CyberArk SIA Documentation](https://docs.cyberark.com/sia)
- [Terraform Provider Best Practices](https://www.terraform.io/docs/extend/best-practices/index.html)
- [ARK SDK Documentation](https://github.com/cyberark/ark-sdk-golang)

## Known Limitations

### Certificate `last_updated_by` Field Warning

**Symptom**:
```
Warning: Provider returned invalid result object after apply
After the apply operation, the provider still indicated an unknown value
for cyberarksia_certificate.*.last_updated_by
```

**Impact**: This is a cosmetic warning that does not prevent resource creation or any other CRUD operations. The field is correctly set to `null` for newly created certificates and will be populated after the first update operation.

**Root Cause**: Terraform Plugin Framework limitation in handling nullable computed fields during the plan/apply cycle. The framework automatically marks nullable computed attributes as "unknown" during planning, which conflicts with the API's behavior of returning `null` for this field on newly created certificates.

**Workaround**: None required - all operations function correctly despite the warning. You can safely ignore this message.

**Verification**: After apply completes, run `terraform show` to verify the certificate was created successfully:
```bash
terraform show | grep -A5 "cyberarksia_certificate"
```

All fields except `last_updated_by` will be populated. The field will remain `null` until the certificate is updated.

**Future Fix**: This will be resolved in a future version when either:
1. The Terraform Plugin Framework improves nullable computed field handling
2. The SIA API is updated to always return a value for this field
