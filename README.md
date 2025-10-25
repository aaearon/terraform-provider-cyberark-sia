# Terraform Provider for CyberArk Secure Infrastructure Access (SIA)

A Terraform provider for managing CyberArk Secure Infrastructure Access (SIA) resources, enabling infrastructure-as-code workflows for database access control and certificate management.

## Features

- **Certificate Management**: Create, update, and delete TLS/SSL certificates for database connections
- **Database Workspace Management**: Configure and manage database targets with certificate-based authentication
- **Multiple Database Engines**: Support for PostgreSQL, MySQL, Oracle, SQL Server, MongoDB, and more
- **Cloud Platform Support**: AWS RDS, Azure SQL, GCP Cloud SQL, MongoDB Atlas, and on-premise databases
- **Secure by Default**: Built-in certificate validation and encryption support

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for development)
- CyberArk Identity tenant with SIA enabled
- Valid CyberArk service account credentials

## Installation

### From Source

```bash
git clone https://github.com/aaearon/terraform-provider-cyberark-sia
cd terraform-provider-cyberark-sia
go build -v
```

### Local Development Installation

```bash
make install
```

This installs the provider to `~/.terraform.d/plugins/` for local testing.

## Quick Start

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
  username      = "service-account@cyberark.cloud.1234"
  client_secret = var.client_secret
}

# Create a TLS certificate
resource "cyberarksia_certificate" "postgres_cert" {
  cert_name        = "prod-postgres-tls"
  cert_description = "TLS certificate for production PostgreSQL"
  cert_body        = file("certs/postgres.pem")
  cert_type        = "PEM"

  labels = {
    environment = "production"
    database    = "postgres"
  }
}

# Configure a database workspace
resource "cyberarksia_database_workspace" "prod_postgres" {
  name                          = "prod-postgres-db"
  database_type                 = "postgres"
  address                       = "prod-postgres.example.com"
  port                          = 5432
  cloud_provider                = "aws"
  region                        = "us-west-2"
  network_name                  = "PRODUCTION"
  certificate_id                = cyberarksia_certificate.postgres_cert.id
  enable_certificate_validation = true
}
```

## Documentation

- **[Examples](examples/)**: Complete configuration examples for various scenarios
- **[SDK Integration Guide](docs/sdk-integration.md)**: ARK SDK patterns and best practices
- **[Development History](docs/development-history.md)**: Architectural decisions and implementation insights
- **[Troubleshooting](docs/troubleshooting.md)**: Common issues and solutions

## Supported Resources

### `cyberarksia_certificate`

Manages TLS/SSL certificates for database connections.

**Features:**
- PEM and DER format support
- Automatic X.509 metadata extraction
- Label-based organization
- Version tracking and drift detection

See [examples/resources/cyberark_sia_certificate/](examples/resources/cyberark_sia_certificate/) for usage examples.

### `cyberarksia_database_workspace`

Manages database workspace configurations for secure access.

**Supported Database Engines:**
- PostgreSQL (including AWS RDS, Azure Database, GCP Cloud SQL)
- MySQL/MariaDB
- Oracle
- Microsoft SQL Server
- MongoDB (including Atlas)
- Snowflake
- And 60+ more engine types

**Features:**
- Certificate-based TLS/mTLS authentication
- Multi-cloud support (AWS, Azure, GCP, Atlas, on-premise)
- Network segmentation
- Authentication method configuration

See [examples/resources/database_workspace/](examples/resources/database_workspace/) for usage examples.

## Known Limitations

### `last_updated_by` Field Warning

The `last_updated_by` computed field on certificate resources may show a Terraform warning during apply operations:

```
Warning: Provider returned invalid result object after apply
After the apply operation, the provider still indicated an unknown value for cyberarksia_certificate.*.last_updated_by
```

**Impact**: This is a cosmetic warning that does not prevent resource creation or other CRUD operations. The field is correctly set to `null` for newly created certificates and populated after update operations.

**Root Cause**: Terraform Plugin Framework limitation in handling nullable computed fields during plan/apply cycle.

**Workaround**: None required - all operations function correctly despite the warning.

## Development

### Building

```bash
go build -v
```

### Testing

```bash
# Run acceptance tests (requires SIA credentials)
TF_ACC=1 go test ./... -v

# Run unit tests only
go test ./internal/client/... -v
```

### Development Guidelines

See [CLAUDE.md](CLAUDE.md) for development conventions, code style, and contribution guidelines.

## Project Structure

```
terraform-provider-cyberark-sia/
├── internal/
│   ├── client/          # ARK SDK wrappers, retry logic, error handling
│   ├── provider/        # Terraform provider implementation
│   ├── models/          # Data models
│   └── validators/      # Custom validators
├── examples/            # Terraform HCL examples
├── docs/                # Documentation
├── specs/               # Feature specifications and planning docs
└── tests/               # Additional test resources
```

## Contributing

Contributions are welcome! Please:

1. Review [CLAUDE.md](CLAUDE.md) for development guidelines
2. Ensure all tests pass (`go test ./...`)
3. Run `go fmt ./...` before committing
4. Follow existing code patterns and conventions

## Acknowledgments

Built using:
- [HashiCorp Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- [CyberArk ARK SDK for Go](https://github.com/cyberark/ark-sdk-golang)

## Support

For issues, questions, or contributions:
- [GitHub Issues](https://github.com/aaearon/terraform-provider-cyberark-sia/issues)
- [Specifications](specs/) - Feature planning and design documentation
