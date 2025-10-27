# Certificate Resource Examples

This directory contains examples for the `cyberark_sia_certificate` resource.

## Prerequisites

1. CyberArk SIA tenant with access
2. Service account with certificate management permissions
3. Valid PEM-encoded certificate file

## Basic Usage

Upload a TLS certificate to CyberArk SIA:

```hcl
resource "cyberark_sia_certificate" "example" {
  cert_name        = "my-database-cert"
  cert_body        = file("path/to/certificate.pem")
  cert_description = "TLS certificate for database connections"
}
```

## Features Demonstrated

### basic.tf
- Upload a certificate from a local PEM file
- Add metadata via labels
- Output certificate ID and expiration date
- Access certificate X.509 metadata

### with_database.tf
- Upload certificate and associate with database workspace
- Enable TLS certificate validation for secure connections
- Share one certificate across multiple database workspaces
- Conditional certificate usage based on environment
- Certificate rotation workflow for renewal scenarios

## Certificate Requirements

- **Format**: PEM or DER encoded
- **Content**: Public certificate only (no private keys)
- **Validation**: Must be valid X.509 certificate
- **Expiration**: No client-side expiration check (deferred to API)

## Common Use Cases

1. **Database TLS**: Upload certificates for database workspace encryption
2. **Certificate Rotation**: Replace expiring certificates (Phase 5)
3. **Multi-Database**: Share one certificate across multiple workspaces

## Using Certificates with Database Workspaces

### Basic Association

Upload a certificate and reference it in a database workspace:

```hcl
# 1. Upload certificate
resource "cyberark_sia_certificate" "db_cert" {
  cert_name = "postgres-tls"
  cert_body = file("certs/postgres.pem")
}

# 2. Reference in database workspace
resource "cyberark_sia_database_workspace" "postgres" {
  name           = "prod-postgres"
  database_type  = "postgres"
  address        = "postgres.example.com"
  port           = 5432

  # Associate certificate
  certificate_id                = cyberark_sia_certificate.db_cert.id
  enable_certificate_validation = true
}
```

### Certificate Validation

- **`certificate_id`**: References uploaded certificate for TLS connections
- **`enable_certificate_validation`**: Enforces TLS certificate validation (default: true)
- **Without `certificate_id`**: Database uses system CA bundle for TLS validation

### Sharing Certificates

One certificate can be used by multiple database workspaces:

```hcl
resource "cyberark_sia_certificate" "shared_cert" {
  cert_name = "wildcard-db-cert"
  cert_body = file("certs/wildcard.pem")
}

resource "cyberark_sia_database_workspace" "db1" {
  name           = "postgres-primary"
  certificate_id = cyberark_sia_certificate.shared_cert.id
  # ...
}

resource "cyberark_sia_database_workspace" "db2" {
  name           = "postgres-replica"
  certificate_id = cyberark_sia_certificate.shared_cert.id  # Same cert
  # ...
}
```

### Updating Certificate Associations

You can add, update, or remove certificate references:

```hcl
# Add certificate to existing workspace
resource "cyberark_sia_database_workspace" "db" {
  name           = "my-database"
  certificate_id = cyberark_sia_certificate.new_cert.id  # Add
  # ...
}

# Remove certificate reference
resource "cyberark_sia_database_workspace" "db" {
  name           = "my-database"
  certificate_id = null  # Remove
  # ...
}
```

### Error Handling

If you reference a non-existent certificate:

```
Error: Certificate Not Found

The specified certificate (ID: cert-123) does not exist or is invalid.

Ensure the certificate exists before associating it with this database workspace.
You can verify the certificate exists with:
  terraform state show cyberark_sia_certificate.<name>
```

**Solution**: Ensure certificate resource is created before database workspace references it (Terraform handles dependency automatically).

## Notes

- `cert_body` is marked as **sensitive** - won't appear in console output
- `cert_body` **must persist in state** for updates (Phase 5)
- `cert_password` is write-only - only needed for encrypted certificates

## Phase Status

✅ **Phase 3 (User Story 1)**: CREATE and READ operations implemented
✅ **Phase 4 (User Story 2)**: Database workspace certificate association
⏳ **Phase 5 (User Story 3)**: UPDATE operation (certificate rotation)
⏳ **Phase 6 (User Story 4)**: DELETE operation and IMPORT support

## Related Resources

- `cyberark_sia_database_workspace`: References certificates via `certificate_id`

## Testing

Run `terraform validate` to check syntax:

```bash
cd examples/resources/cyberark_sia_certificate
terraform validate
```

For actual deployment, ensure you have:
1. Provider credentials configured
2. Certificate file at the specified path
3. Appropriate permissions in CyberArk SIA
