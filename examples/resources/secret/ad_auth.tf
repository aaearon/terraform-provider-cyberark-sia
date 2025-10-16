# Example: Strong Account with Active Directory Authentication
# This example demonstrates creating a strong account using Active Directory (domain) authentication
# for a SQL Server database target.

# Provider configuration
terraform {
  required_providers {
    cyberark_sia = {
      source = "local/cyberark-sia"
    }
  }
}

provider "cyberark_sia" {
  # Authentication credentials should be set via environment variables:
  # export CYBERARK_CLIENT_ID="your-client-id"
  # export CYBERARK_CLIENT_SECRET="your-client-secret"
  # export CYBERARK_TENANT_SUBDOMAIN="your-subdomain"
  # Optional (for GovCloud or custom deployments):
  # export CYBERARK_IDENTITY_URL="https://your-tenant.cyberarkgov.cloud"
}

# Database target for SQL Server with AD authentication
resource "cyberark_sia_database_workspace" "sqlserver" {
  name                  = "production-sqlserver"
  database_type         = "sqlserver"
  address               = "sql-prod.corp.example.com"
  port                  = 1433
  authentication_method = "ad_ephemeral_user"
  cloud_provider        = "azure"
  region                = "eastus"
  description           = "Production SQL Server database with AD authentication"

  tags = {
    environment = "production"
    team        = "database"
    auth_type   = "active_directory"
  }
}

# Strong account with Active Directory authentication
resource "cyberark_sia_secret" "sqlserver_service" {
  name               = "sqlserver-service-account"
  # NOTE: Secrets are standalone - no workspace_id needed = cyberark_sia_database_workspace.sqlserver.id
  authentication_type = "domain"

  # Active Directory authentication credentials
  username = "svc_sia_sqlserver"
  password = var.ad_service_password # Sensitive - should be stored in variable or secret manager
  domain   = "corp.example.com"      # Active Directory domain

  description = "AD service account for SQL Server ephemeral access provisioning"

  # Optional: Enable automatic credential rotation
  # Note: Ensure AD password policy allows programmatic rotation
  rotation_enabled      = false
  rotation_interval_days = 60 # Aligned with AD password policy

  tags = {
    purpose        = "ephemeral-access"
    managed_by     = "terraform"
    ad_domain      = "corp.example.com"
    security_tier  = "high"
  }
}

# Variables for sensitive data
variable "ad_service_password" {
  description = "Password for the Active Directory service account"
  type        = string
  sensitive   = true
}

# Outputs (safe - does not expose sensitive data)
output "strong_account_id" {
  description = "ID of the created strong account"
  value       = cyberark_sia_secret.sqlserver_service.id
}

output "strong_account_username" {
  description = "Username of the AD service account (non-sensitive)"
  value       = cyberark_sia_secret.sqlserver_service.username
}

output "ad_domain" {
  description = "Active Directory domain"
  value       = cyberark_sia_secret.sqlserver_service.domain
}

output "last_modified" {
  description = "Timestamp when the strong account was last modified"
  value       = cyberark_sia_secret.sqlserver_service.last_modified
}
