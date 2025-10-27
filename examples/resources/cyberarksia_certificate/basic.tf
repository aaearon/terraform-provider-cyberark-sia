# Basic certificate upload example

terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = ">= 0.1.0"
    }
  }
}

provider "cyberarksia" {
  username      = var.cyberark_username
  client_secret = var.cyberark_client_secret
}

# Upload a TLS certificate for PostgreSQL databases
resource "cyberarksia_certificate" "postgres_tls" {
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
  value       = cyberarksia_certificate.postgres_tls.id
  description = "Certificate ID for PostgreSQL TLS cert"
}

# Output certificate metadata
output "postgres_cert_expiry" {
  value       = cyberarksia_certificate.postgres_tls.expiration_date
  description = "Certificate expiration date"
}

# Output certificate metadata details
output "postgres_cert_metadata" {
  value = {
    issuer        = cyberarksia_certificate.postgres_tls.metadata.issuer
    subject       = cyberarksia_certificate.postgres_tls.metadata.subject
    serial_number = cyberarksia_certificate.postgres_tls.metadata.serial_number
  }
  description = "Certificate X.509 metadata"
}
