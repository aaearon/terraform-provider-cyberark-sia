# Terraform Provider for CyberArk Secure Infrastructure Access (SIA)

A Terraform provider for managing CyberArk Secure Infrastructure Access (SIA) resources, enabling infrastructure-as-code workflows for database access control and certificate management.

## Features

- **Policy Management**: Create access policies with session limits and time-based restrictions
- **User & Group Assignment**: Grant access to specific users, groups, or roles - no manual UUID lookups needed
- **Principal Lookup**: Find users and groups by name across Cloud Directory, Azure AD, Active Directory
- **Database Configuration**: Configure database workspaces with certificate-based authentication
- **Database Assignment**: Connect databases to policies with 6 authentication methods
- **Certificate Management**: Manage TLS/SSL certificates for secure database connections
- **Secret Management**: Store database credentials (local auth, Active Directory, AWS IAM)
- **Multiple Database Engines**: PostgreSQL, MySQL, Oracle, SQL Server, MongoDB, Snowflake, and 60+ more
- **Multi-Cloud Support**: AWS RDS, Azure SQL, GCP Cloud SQL, MongoDB Atlas, and on-premise

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for development)
- CyberArk Identity tenant with SIA enabled
- Identity Security Platform Shared Services tenant with Unified Access Policies (UAP) enabled
- Valid CyberArk service account credentials with `DpaAdmin` role

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

## Authentication

The provider authenticates using a CyberArk service account. Simply provide the service account username and password - the provider handles OAuth2 authentication automatically.

**Service Account Setup:**
1. Create a service account in CyberArk Identity
2. Assign the **`DpaAdmin`** role (required for managing SIA resources)
3. Provide the username and password to the provider

**Important:** Never commit credentials to version control. Use environment variables or secure variable storage (e.g., Terraform Cloud, HashiCorp Vault).

## Quick Start

```hcl
terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = "0.1.0"
    }
  }
}

provider "cyberarksia" {
  username      = "service-account@cyberark.cloud.1234"  # OAuth2 client username
  client_secret = var.client_secret                     # OAuth2 client secret
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

See [examples/resources/cyberarksia_certificate/](examples/resources/cyberarksia_certificate/) for usage examples.

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

### `cyberarksia_secret`

Manages database authentication secrets for use with database workspaces.

**Supported Authentication Types:**
- Local database authentication (username/password)
- Domain authentication (Active Directory)
- AWS IAM authentication (for RDS)

**Features:**
- Secure credential storage
- Integration with database workspaces
- Support for domain-based authentication
- AWS IAM role ARN configuration

See [examples/resources/secret/](examples/resources/secret/) for usage examples.

### `cyberarksia_database_policy`

Create and manage access policies that control who can access which databases and when.

**What you can do:**
- Set session limits (max duration, idle timeout)
- Restrict access to specific time windows (e.g., business hours only, Monday-Friday 9-5)
- Set policy validity periods (e.g., temporary access for Q1 2024)
- Enable or suspend policies without deleting them
- Tag policies for organization

Think of policies as the rules. The other resources assign specific users and databases to those rules.

See [docs/resources/database_policy.md](docs/resources/database_policy.md) for usage examples.

### `cyberarksia_database_policy_principal_assignment`

Grant specific users, groups, or roles access to databases through policies.

**What you can do:**
- Assign individual users to policies
- Assign entire groups (like "Database Admins" or "Developers")
- Assign federated users from Azure AD, Okta, etc.
- No more hunting for user UUIDs - use the `cyberarksia_principal` data source to look them up by name

**Example use case:** Your security team creates a policy for production database access. You assign your DevOps team's group to that policy without needing to know any UUIDs.

See [docs/resources/database_policy_principal_assignment.md](docs/resources/database_policy_principal_assignment.md) for usage examples.

### `cyberarksia_database_policy_database_assignment`

Connect database workspaces to access policies with specific authentication settings.

**What you can do:**
- Assign databases to policies with different auth methods (standard DB auth, LDAP, AWS IAM, etc.)
- Specify which database roles users get when they connect
- Works with 6 authentication methods including passwordless AWS RDS IAM

**Example use case:** You have a production PostgreSQL database and a policy for developers. This resource connects them together and specifies that users get the `readonly` role.

See [docs/resources/policy_database_assignment.md](docs/resources/policy_database_assignment.md) and [examples/resources/cyberarksia_database_policy_database_assignment/](examples/resources/cyberarksia_database_policy_database_assignment/) for usage examples.

## Data Sources

Data sources let you look up existing resources without creating them.

### `cyberarksia_database_policy`

Look up existing access policies by name or ID.

**Why you'd use this:**
- Reference policies created outside Terraform (in the UI or by another team)
- Share policies across multiple Terraform workspaces

See [examples/data-sources/cyberarksia_database_policy/](examples/data-sources/cyberarksia_database_policy/) for usage examples.

### `cyberarksia_principal`

Look up users, groups, or roles by name - no more hunting for UUIDs.

**What you can do:**
- Find cloud users: `tim@cyberark.cloud.12345`
- Find federated users: `john.doe@company.com` (Azure AD, Okta, etc.)
- Find Active Directory users: `SchindlerT@domain.com`
- Find groups: `Database Administrators`

Returns the UUID and directory information you need for policy assignments.

**Example:**
```hcl
data "cyberarksia_principal" "db_team" {
  name = "Database Administrators"
  type = "GROUP"
}

resource "cyberarksia_database_policy_principal_assignment" "grant_access" {
  policy_id         = cyberarksia_database_policy.prod.policy_id
  principal_id      = data.cyberarksia_principal.db_team.id
  principal_type    = data.cyberarksia_principal.db_team.principal_type
  # ... other fields populated automatically from the data source
}
```

See [docs/data-sources/principal.md](docs/data-sources/principal.md) for usage examples.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding conventions, and pull request process.

### Quick Links

- **[Contributing Guide](CONTRIBUTING.md)** - Development setup and contribution guidelines
- **[Testing Guide](TESTING.md)** - Running tests and CRUD testing framework
- **[Design Decisions](docs/development/design-decisions.md)** - Active technologies, SDK limitations, breaking changes
- **[SDK Integration](docs/sdk-integration.md)** - ARK SDK patterns and field mappings
- **[Development Guidelines](CLAUDE.md)** - Code style, commands, and project structure
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

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

For comprehensive CRUD testing, see [TESTING.md](TESTING.md) and [examples/testing/TESTING-GUIDE.md](examples/testing/TESTING-GUIDE.md).

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

Contributions are welcome! To get started:

**Quick Setup:**
```bash
make tools-install         # Install dev tools
make pre-commit-install    # Enable automatic validation
make validate              # Verify setup works
```

**Before submitting:**
```bash
make validate              # Run all checks locally (mirrors CI)
```

**Full contributor guide:** [CONTRIBUTING.md](CONTRIBUTING.md)

**Development reference:** [CLAUDE.md](CLAUDE.md) (for LLM-assisted development)

## Acknowledgments

This provider is built on top of:
- **[CyberArk ARK SDK for Go](https://github.com/cyberark/ark-sdk-golang)** - Official Go SDK for CyberArk platform APIs. All provider API calls use this SDK for authentication, SIA workspace management, and UAP policy operations.
- [HashiCorp Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) - Framework for building Terraform providers with type-safe schemas and state management.

The provider implements custom OAuth2 authentication flows for CyberArk Identity platform integration.

## Support

For issues, questions, or contributions:
- [GitHub Issues](https://github.com/aaearon/terraform-provider-cyberark-sia/issues)
- [Specifications](specs/) - Feature planning and design documentation
