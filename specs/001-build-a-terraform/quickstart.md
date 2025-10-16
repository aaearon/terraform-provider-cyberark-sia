# Quickstart Guide: Terraform Provider for CyberArk SIA

**Target Audience**: Infrastructure engineers using Terraform to manage database infrastructure
**Time to Complete**: 15 minutes
**Prerequisites**: Terraform >= 1.0, CyberArk SIA tenant, ISPSS service account credentials

---

## Prerequisites

### 1. CyberArk Identity Security Platform (ISPSS) Setup

You need an ISPSS service account with appropriate role memberships for SIA API access:

- **Role Required**: SIA Database Administrator (or equivalent with database target/strong account management permissions)
- **Client Credentials**: `client_id` and `client_secret` from Identity Administration portal
- **Tenant Information**: Your CyberArk Identity URL and tenant subdomain

### 2. Terraform Installation

```bash
# Verify Terraform is installed (version 1.0 or later)
terraform version

# Expected output:
# Terraform v1.x.x
```

### 3. Database Infrastructure

You should have an existing database to onboard to SIA:
- AWS RDS instance
- Azure SQL database
- On-premise database with network connectivity to SIA connectors

**Important**: This provider does NOT create databases. It only registers existing databases with SIA.

---

## Quick Start Example

### Step 1: Configure the Provider

Create a new directory and add `provider.tf`:

```hcl
terraform {
  required_providers {
    cyberark_sia = {
      source  = "local/cyberark-sia"  # Will be registry.terraform.io/cyberark/sia after publication
      version = "~> 1.0"
    }
  }
}

provider "cyberark_sia" {
  client_id                 = var.cyberark_client_id     # From environment or tfvars
  client_secret             = var.cyberark_client_secret # Sensitive
  identity_url              = "https://example.cyberark.cloud"
  identity_tenant_subdomain = "example"
}

variable "cyberark_client_id" {
  description = "ISPSS service account client ID"
  type        = string
}

variable "cyberark_client_secret" {
  description = "ISPSS service account client secret"
  type        = string
  sensitive   = true
}
```

### Step 2: Define Credentials in `terraform.tfvars`

Create `terraform.tfvars` (add to `.gitignore`):

```hcl
cyberark_client_id     = "your-client-id@cyberark.cloud.12345"
cyberark_client_secret = "your-client-secret-value"
```

**Security Note**: NEVER commit `terraform.tfvars` to version control. Use environment variables or secret managers in CI/CD:

```bash
export TF_VAR_cyberark_client_id="your-client-id"
export TF_VAR_cyberark_client_secret="your-client-secret"
```

---

### Step 3: Onboard an Existing AWS RDS PostgreSQL Database

Create `main.tf`:

```hcl
# Reference existing RDS instance (created by AWS provider or manually)
data "aws_db_instance" "production" {
  db_instance_identifier = "production-postgres"
}

# Register the database with SIA
resource "cyberark_sia_database_target" "production_postgres" {
  name             = "production-postgres-db"
  database_type    = "postgresql"
  database_version = "14.2.0"  # Must match actual RDS version
  address          = data.aws_db_instance.production.endpoint
  port             = data.aws_db_instance.production.port
  database_name    = "app_database"

  authentication_method = "local"  # Or "aws_iam" for RDS IAM auth

  cloud_provider = "aws"
  aws_region     = "us-east-1"
  aws_account_id = data.aws_caller_identity.current.account_id

  description = "Production PostgreSQL database for application"
  tags = {
    Environment = "production"
    Team        = "platform"
  }
}

# Create a strong account for SIA to provision ephemeral access
resource "cyberark_sia_strong_account" "postgres_admin" {
  name               = "postgres-admin-account"
  database_target_id = cyberark_sia_database_target.production_postgres.id

  authentication_type = "local"
  username            = "sia_service_account"
  password            = var.postgres_admin_password  # Sensitive

  description = "Strong account for SIA ephemeral credential provisioning"
  tags = {
    Environment = "production"
  }
}

variable "postgres_admin_password" {
  description = "Password for SIA strong account"
  type        = string
  sensitive   = true
}

data "aws_caller_identity" "current" {}
```

### Step 4: Initialize and Apply

```bash
# Initialize Terraform (download provider)
terraform init

# Review the execution plan
terraform plan

# Apply the configuration
terraform apply

# Expected output:
# cyberark_sia_database_target.production_postgres: Creating...
# cyberark_sia_database_target.production_postgres: Creation complete after 3s [id=target-uuid-123]
# cyberark_sia_strong_account.postgres_admin: Creating...
# cyberark_sia_strong_account.postgres_admin: Creation complete after 2s [id=account-uuid-456]
#
# Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
```

---

## Common Scenarios

### Scenario 1: Azure SQL Server

```hcl
resource "azurerm_mssql_server" "main" {
  name                         = "production-sqlserver"
  resource_group_name          = azurerm_resource_group.main.name
  location                     = "East US"
  version                      = "12.0"
  administrator_login          = "sqladmin"
  administrator_login_password = var.sql_admin_password
}

resource "cyberark_sia_database_target" "azure_sql" {
  name             = "azure-production-sql"
  database_type    = "sqlserver"
  database_version = "13.0.0"  # SQL Server 2016+
  address          = azurerm_mssql_server.main.fully_qualified_domain_name
  port             = 1433
  database_name    = "master"

  authentication_method = "domain"  # Active Directory auth

  cloud_provider        = "azure"
  azure_tenant_id       = data.azurerm_client_config.current.tenant_id
  azure_subscription_id = data.azurerm_client_config.current.subscription_id

  tags = {
    ManagedBy = "Terraform"
  }
}

resource "cyberark_sia_strong_account" "sql_domain_account" {
  name               = "sql-domain-admin"
  database_target_id = cyberark_sia_database_target.azure_sql.id

  authentication_type = "domain"
  username            = "CORP\\sqladmin"
  password            = var.domain_admin_password
  domain              = "corp.example.com"
}

data "azurerm_client_config" "current" {}
```

---

### Scenario 2: On-Premise Oracle Database

```hcl
resource "cyberark_sia_database_target" "onprem_oracle" {
  name             = "onprem-oracle-prod"
  database_type    = "oracle"
  database_version = "19.3.0"  # Oracle 19c
  address          = "oracle-prod.internal.example.com"
  port             = 1521
  database_name    = "PRODDB"

  authentication_method = "local"

  cloud_provider = "on_premise"  # Explicitly on-premise

  description = "On-premise Oracle database in datacenter"
  tags = {
    Location = "DataCenter-East"
    Criticality = "High"
  }
}

resource "cyberark_sia_strong_account" "oracle_admin" {
  name               = "oracle-sys-account"
  database_target_id = cyberark_sia_database_target.onprem_oracle.id

  authentication_type = "local"
  username            = "SYS AS SYSDBA"
  password            = var.oracle_sys_password

  rotation_enabled       = true
  rotation_interval_days = 90  # Rotate every 90 days
}
```

---

### Scenario 3: AWS RDS with IAM Authentication

```hcl
resource "aws_db_instance" "iam_enabled" {
  identifier                 = "rds-iam-postgres"
  engine                     = "postgres"
  iam_database_authentication_enabled = true
  # ... other config
}

resource "cyberark_sia_database_target" "rds_iam" {
  name             = "rds-postgres-iam"
  database_type    = "postgresql"
  database_version = "14.2.0"
  address          = aws_db_instance.iam_enabled.endpoint
  port             = aws_db_instance.iam_enabled.port

  authentication_method = "aws_iam"  # RDS IAM authentication

  cloud_provider = "aws"
  aws_region     = aws_db_instance.iam_enabled.region
  aws_account_id = data.aws_caller_identity.current.account_id
}

resource "cyberark_sia_strong_account" "rds_iam_account" {
  name               = "rds-iam-admin"
  database_target_id = cyberark_sia_database_target.rds_iam.id

  authentication_type = "aws_iam"
  aws_access_key_id   = var.aws_access_key
  aws_secret_access_key = var.aws_secret_key

  description = "AWS IAM credentials for RDS IAM auth"
}
```

---

## Importing Existing Resources

If you have databases already registered in SIA:

```bash
# Import database target
terraform import cyberark_sia_database_target.existing <sia-target-id>

# Import strong account
terraform import cyberark_sia_strong_account.existing <sia-account-id>
```

After import, run `terraform plan` to ensure state matches configuration.

---

## Best Practices

### 1. Credential Management

**DO**:
- Store sensitive values in Terraform variables marked `sensitive = true`
- Use environment variables or secret managers (AWS Secrets Manager, Azure Key Vault)
- Enable Terraform state encryption (S3 backend with encryption, Terraform Cloud)

**DON'T**:
- Hardcode passwords in `.tf` files
- Commit `terraform.tfvars` to version control
- Share state files unencrypted

```hcl
# Good: Use data sources for secrets
data "aws_secretsmanager_secret_version" "db_password" {
  secret_id = "prod/database/admin-password"
}

resource "cyberark_sia_strong_account" "secure" {
  # ...
  password = data.aws_secretsmanager_secret_version.db_password.secret_string
}
```

### 2. Dependency Management

Ensure proper resource ordering:

```hcl
resource "cyberark_sia_database_target" "db" {
  # Database target must be created first
}

resource "cyberark_sia_strong_account" "account" {
  database_target_id = cyberark_sia_database_target.db.id
  # Implicit dependency via reference ^
}
```

### 3. Validation Before Apply

```bash
# Validate configuration syntax
terraform validate

# Format code
terraform fmt

# Review plan output
terraform plan -out=tfplan

# Apply with saved plan
terraform apply tfplan
```

### 4. Handling Partial State Failures

If database created by AWS/Azure but SIA onboarding fails:

```bash
# Option 1: Fix SIA issue and re-run
terraform apply

# Option 2: Remove database and start over
terraform destroy
```

Provider will show clear error messages with resolution steps.

---

## Troubleshooting

### Authentication Errors

**Error**: `Authentication failed: Invalid client_id or client_secret`

**Solution**:
1. Verify credentials in `terraform.tfvars`
2. Check client ID format: `client-id@cyberark.cloud.tenant-id`
3. Ensure service account has SIA role memberships

```bash
# Test authentication independently
export ARK_SECRET="your-client-secret"
# Use ARK SDK CLI to verify auth works
```

---

### Database Version Validation

**Error**: `Database version 10.5.0 is below minimum for mariadb (10.0.0 required)`

**Solution**:
- Check actual database version
- Update `database_version` in Terraform config
- Ensure version meets SIA requirements (see spec.md FR-001a)

---

### Resource Not Found After Apply

**Error**: `Resource not found during refresh`

**Possible Causes**:
1. Database target deleted in SIA console outside Terraform
2. Network connectivity issues between Terraform and SIA API

**Solution**:
```bash
# Refresh state
terraform refresh

# If deleted, remove from state and re-create
terraform state rm cyberark_sia_database_target.missing
terraform apply
```

---

## Next Steps

1. **Explore Examples**: Review `/examples/` directory for complete configurations
2. **Read Documentation**: See `/docs/` for detailed resource reference
3. **Configure CI/CD**: Integrate Terraform into deployment pipelines
4. **Set Up Monitoring**: Use Terraform outputs to feed into monitoring systems

---

## Additional Resources

- **Provider Documentation**: https://registry.terraform.io/providers/cyberark/sia (post-publication)
- **CyberArk SIA Docs**: https://docs.cyberark.com/sia
- **Terraform Best Practices**: https://www.terraform.io/docs/cloud/guides/recommended-practices
- **ARK SDK Documentation**: https://github.com/cyberark/ark-sdk-golang

---

## Support

- **Provider Issues**: https://github.com/aaearon/terraform-provider-cyberark-sia/issues
- **CyberArk Support**: https://www.cyberark.com/customer-support/
- **Community**: CyberArk Commons forums

---

## Complete Example

For a full working example combining AWS RDS provisioning + SIA onboarding + strong account creation, see `examples/complete/full_workflow.tf` in the provider repository.
