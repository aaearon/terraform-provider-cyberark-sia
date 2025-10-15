# Example: Strong Account with Local Authentication
# This example demonstrates creating a strong account using local (username/password) authentication
# for a PostgreSQL database target.

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

# Database target (from Phase 3 example)
resource "cyberark_sia_database_workspace" "postgres" {
  name                  = "production-postgres"
  database_type         = "postgresql"
  address               = "prod-db.example.com"
  port                  = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider        = "aws"
  region                = "us-east-1"
  description           = "Production PostgreSQL database"

  tags = {
    environment = "production"
    team        = "platform"
  }
}

# Strong account with local authentication
resource "cyberark_sia_secret" "postgres_admin" {
  name               = "postgres-admin-account"
  # NOTE: Secrets are standalone - no workspace_id needed = cyberark_sia_database_workspace.postgres.id
  authentication_type = "local"

  # Local authentication credentials
  username = "sia_admin"
  password = var.postgres_admin_password # Sensitive - should be stored in variable or secret manager

  description = "Administrator account for PostgreSQL ephemeral access provisioning"

  # Optional: Enable automatic credential rotation
  rotation_enabled      = false # Set to true for automatic rotation
  rotation_interval_days = 90   # Only used if rotation_enabled = true

  tags = {
    purpose     = "ephemeral-access"
    managed_by  = "terraform"
  }
}

# Variables for sensitive data
variable "postgres_admin_password" {
  description = "Password for the PostgreSQL admin account"
  type        = string
  sensitive   = true
}

# Outputs (safe - does not expose sensitive data)
output "strong_account_id" {
  description = "ID of the created strong account"
  value       = cyberark_sia_secret.postgres_admin.id
}

output "strong_account_name" {
  description = "Name of the strong account"
  value       = cyberark_sia_secret.postgres_admin.name
}

output "created_at" {
  description = "Timestamp when the strong account was created"
  value       = cyberark_sia_secret.postgres_admin.created_at
}
