# Basic certificate upload example

terraform {
  required_providers {
    cyberark_sia = {
      source = "cyberark/cyberark-sia"
    }
  }
}

provider "cyberark_sia" {
  client_id                  = var.cyberark_client_id
  client_secret              = var.cyberark_client_secret
  identity_tenant_subdomain  = var.cyberark_subdomain
}

# Upload a TLS certificate for PostgreSQL databases
resource "cyberark_sia_certificate" "postgres_tls" {
  cert_name        = "postgres-production-tls"
  cert_body        = file("${path.module}/certs/postgres.pem")
  cert_description = "TLS certificate for production PostgreSQL databases"
  domain_name      = "db.example.com"

  labels = {
    environment = "production"
    database    = "postgres"
    owner       = "platform-team"
  }
}

# Output the certificate ID for reference
output "postgres_cert_id" {
  value       = cyberark_sia_certificate.postgres_tls.id
  description = "Certificate ID for PostgreSQL TLS cert"
}

# Output certificate metadata
output "postgres_cert_expiry" {
  value       = cyberark_sia_certificate.postgres_tls.expiration_date
  description = "Certificate expiration date"
}

# Output certificate metadata details
output "postgres_cert_metadata" {
  value = {
    issuer        = cyberark_sia_certificate.postgres_tls.metadata.issuer
    subject       = cyberark_sia_certificate.postgres_tls.metadata.subject
    serial_number = cyberark_sia_certificate.postgres_tls.metadata.serial_number
  }
  description = "Certificate X.509 metadata"
}
