# Example: Certificate with Database Workspace Association
#
# This example demonstrates how to:
# 1. Upload a TLS certificate to CyberArk SIA
# 2. Create a database workspace that references the certificate
# 3. Enable TLS certificate validation for secure database connections
#
# Prerequisites:
# - Valid PEM-encoded certificate file in certs/postgres.pem
# - CyberArk SIA provider configured with valid credentials

# Upload TLS certificate for PostgreSQL database
resource "cyberark_sia_certificate" "postgres_tls" {
  cert_name        = "postgres-production-tls"
  cert_body        = file("${path.module}/certs/postgres.pem")
  cert_description = "TLS certificate for production PostgreSQL databases"
  domain_name      = "db.example.com"

  labels = {
    environment = "production"
    database    = "postgres"
    owner       = "platform-team"
    compliance  = "pci-dss"
  }
}

# Create database workspace with certificate reference
resource "cyberark_sia_database_workspace" "prod_postgres" {
  name                          = "prod-postgres-db"
  database_type                 = "postgres"
  address                       = "postgres.example.com"
  port                          = 5432

  # Reference the uploaded certificate for TLS validation
  certificate_id                = cyberark_sia_certificate.postgres_tls.id
  enable_certificate_validation = true

  # Additional database workspace configuration
  network_name                  = "production-network"
  cloud_provider                = "aws"
  region                        = "us-east-1"

  tags = {
    environment = "production"
    managed_by  = "terraform"
    team        = "platform"
  }
}

# Output certificate details for verification
output "certificate_id" {
  value       = cyberark_sia_certificate.postgres_tls.id
  description = "ID of the uploaded TLS certificate"
}

output "certificate_expiration" {
  value       = cyberark_sia_certificate.postgres_tls.expiration_date
  description = "Certificate expiration date (ISO 8601)"
}

output "database_workspace_id" {
  value       = cyberark_sia_database_workspace.prod_postgres.id
  description = "ID of the database workspace"
}

# Example: Multiple databases sharing the same certificate
# Useful for wildcard certificates or databases with the same CA

resource "cyberark_sia_database_workspace" "prod_postgres_replica" {
  name                          = "prod-postgres-replica"
  database_type                 = "postgres"
  address                       = "postgres-replica.example.com"
  port                          = 5432

  # Reuse the same certificate
  certificate_id                = cyberark_sia_certificate.postgres_tls.id
  enable_certificate_validation = true

  tags = {
    environment = "production"
    role        = "replica"
    managed_by  = "terraform"
  }
}

# Example: Database workspace WITHOUT certificate (uses system CA bundle)
resource "cyberark_sia_database_workspace" "dev_postgres" {
  name                          = "dev-postgres-db"
  database_type                 = "postgres"
  address                       = "postgres-dev.example.com"
  port                          = 5432

  # No certificate_id specified - uses system CA bundle for TLS
  enable_certificate_validation = true

  tags = {
    environment = "development"
    managed_by  = "terraform"
  }
}

# Example: Conditional certificate usage
variable "enable_custom_tls" {
  description = "Enable custom TLS certificate for database connections"
  type        = bool
  default     = true
}

resource "cyberark_sia_database_workspace" "flexible_postgres" {
  name                          = "flexible-postgres-db"
  database_type                 = "postgres"
  address                       = "postgres-flexible.example.com"
  port                          = 5432

  # Conditionally reference certificate
  certificate_id                = var.enable_custom_tls ? cyberark_sia_certificate.postgres_tls.id : null
  enable_certificate_validation = true

  tags = {
    environment = "staging"
    managed_by  = "terraform"
  }
}

# Example: Certificate rotation workflow
# When renewing certificates, update the cert_body and apply
# All database workspaces referencing this certificate will automatically use the new certificate

# To rotate certificate:
# 1. Update certs/postgres.pem with renewed certificate
# 2. Run: terraform plan  (shows cert_body will update)
# 3. Run: terraform apply (updates certificate in-place, same ID)
# 4. All database workspaces continue working with new certificate
