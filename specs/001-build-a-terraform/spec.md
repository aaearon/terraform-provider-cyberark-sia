# Feature Specification: Terraform Provider for CyberArk Secure Infrastructure Access

**Feature Branch**: `001-build-a-terraform`
**Created**: 2025-10-15
**Status**: Draft
**Input**: User description: "Build a Terraform provider for CyberArk's Secure Infrastructure Access product that can be used to manage the resource lifecycle of databases within the Secure Infrastructure Access product. Users will leverage the provider to add databases to Secure Infrastructure Access at the same time they create them using infrastructure as code principals. The Terraform provider needs to support all the same databases that Secure Infrastructure Access supports. The databases could be in AWS or Azure or on-premise. If it is AWS or Azure, then the module must support databases that are offered in those cloud providers! It should also support managing the lifecycle of strong accounts for databases. Strong accounts for databases can either be stored directly within Secure Infrastructure Access OR can be stored in another CyberArk product, Privilege Cloud. For this provider, we only need to support the lifecycle of strong accounts for databases that are stored directly in Secure Infrastructure Access!"

## Clarifications

### Session 2025-10-15

- Q: Should the provider support CRUD operations on Target Sets themselves, or only reference existing Target Sets, or defer Target Set assignment entirely? → A: No Target Set assignment in initial release - defer to manual SIA configuration
- Q: How should the provider manage bearer token lifecycle given 15-minute expiration? → A: Token caching with auto-refresh before expiration (proactive renewal at ~80% lifetime)
- Q: Should the provider validate database connectivity during onboarding or only register connection details? → A: Provider registers database details without connectivity validation, SIA validates on first access
- Q: What should happen when strong account credentials are updated while SIA is actively using them? → A: Provider updates credentials immediately; SIA handles session lifecycle (new sessions use new credentials)
- Q: How should the provider handle partial state failures (database created by AWS/Azure but SIA onboarding fails)? → A: Terraform apply fails with clear error; database exists in cloud, user must resolve manually

## User Scenarios & Testing

### User Story 1 - Onboard Existing Database to SIA via IaC (Priority: P1)

As an infrastructure engineer, I want to onboard existing databases to CyberArk SIA using the same Infrastructure as Code workflow where I define my database infrastructure, so that databases are automatically secured by SIA without manual console operations, eliminating security gaps between database creation and security onboarding.

**Why this priority**: This is the core value proposition - enabling secure-by-default database onboarding. The SIA provider registers existing databases (created by AWS/Azure/on-premise providers or already deployed) with SIA. Without this capability, there's a security window between database creation and SIA registration where credentials may be exposed.

**Independent Test**: Can be fully tested by creating a database using the AWS provider (e.g., PostgreSQL RDS), then using the SIA provider to onboard it to SIA, and verifying it appears in the SIA console as a managed target with proper configuration. Delivers immediate security value by automating the onboarding workflow.

**Acceptance Scenarios**:

1. **Given** I have a Terraform configuration where the AWS provider creates an RDS PostgreSQL database, **When** I add a CyberArk SIA provider resource that references the RDS database connection details, **Then** after applying the configuration, the database exists in AWS (created by AWS provider) and is onboarded to SIA (registered by SIA provider) with correct connection parameters
2. **Given** I have a Terraform configuration where the Azure provider creates an Azure SQL database, **When** I add a SIA provider resource that references the Azure SQL database connection details, **Then** the database is created in Azure (by Azure provider) and onboarded to SIA (by SIA provider) with proper Azure-specific configuration
3. **Given** I have an existing on-premise database with known connection details, **When** I define a SIA provider resource specifying the database address and parameters, **Then** the database target is onboarded to SIA without requiring any database creation
4. **Given** I have existing databases of multiple types (PostgreSQL, MySQL, MariaDB, MongoDB, Oracle, SQL Server, Db2) with known connection details, **When** I define SIA provider resources for each, **Then** all supported database types are correctly onboarded to SIA with type-specific connection parameters

---

### User Story 2 - Manage Strong Account Lifecycle for Databases (Priority: P2)

As an infrastructure engineer, I want to declaratively define and manage the strong accounts that SIA uses to provision ephemeral credentials for database access, so that credential management is automated and versioned alongside my infrastructure code.

**Why this priority**: Strong accounts are essential for SIA's ephemeral credential provisioning, but they are a supporting capability. This comes after the core database registration capability because databases must exist before strong accounts can manage access to them.

**Independent Test**: Can be tested by creating a database target in SIA (manually or via Story 1), then using Terraform to create and manage a strong account for that target. Verify the strong account can be created, updated (e.g., credential rotation), and deleted through Terraform operations.

**Acceptance Scenarios**:

1. **Given** I have a database registered in SIA, **When** I define a strong account resource in Terraform with credentials and assign it to the database, **Then** the strong account is created in SIA (stored directly in SIA, not Privilege Cloud) and associated with the correct database target
2. **Given** I need to rotate strong account credentials, **When** I update the credential value in my Terraform configuration and apply changes, **Then** the strong account in SIA is updated with the new credentials
3. **Given** I have a strong account defined for a local authentication method (for MariaDB, MongoDB, MySQL, Oracle, or PostgreSQL), **When** I apply the configuration, **Then** SIA can use this account to create ephemeral local users for database access
4. **Given** I have a strong account defined for a domain authentication method (for SQL Server or Db2), **When** I apply the configuration, **Then** SIA can use this account to create ephemeral domain users for database access
5. **Given** I have a strong account defined with AWS IAM credentials for RDS databases (MariaDB, MySQL, or PostgreSQL on RDS), **When** I apply the configuration, **Then** SIA can use these credentials to generate ephemeral RDS IAM credentials

---

### User Story 3 - Update and Delete Database Targets (Priority: P3)

As an infrastructure engineer, I want to manage the complete lifecycle of database targets in SIA through Terraform, including updates and deletions, so that my security configuration stays synchronized with infrastructure changes and decommissioned databases are automatically removed from SIA.

**Why this priority**: Create operations (P1, P2) are more critical than update/delete for initial deployment. This capability ensures ongoing infrastructure hygiene but is less urgent than initial provisioning.

**Independent Test**: Can be tested by first creating a database target (using Story 1), then modifying its configuration (e.g., changing connection port or authentication method) and verifying the update is reflected in SIA. Finally, deleting the Terraform resource and confirming the target is removed from SIA.

**Acceptance Scenarios**:

1. **Given** I have a database target registered in SIA via Terraform, **When** I modify the target configuration (e.g., connection port or authentication method) and apply changes, **Then** the database target in SIA reflects the updated configuration
2. **Given** I have a database target registered in SIA via Terraform, **When** I remove the resource from my Terraform configuration and apply the change, **Then** the database target is removed from SIA
3. **Given** I decommission a database and remove both the database resource and SIA target resource from Terraform, **When** I apply the configuration, **Then** both the database and its SIA registration are cleanly removed
4. **Given** I need to update strong account credentials for a database, **When** I modify the strong account resource in Terraform and apply, **Then** the credentials are updated in SIA immediately (SIA manages session lifecycle; new connections use new credentials)

---

### Edge Cases

- What happens when a database is created by AWS/Azure provider but the SIA onboarding fails? (Terraform apply fails with clear error message; database exists in cloud but not in SIA; user must either fix SIA issue and re-run apply, or manually clean up the database)
- How does the system handle SIA being unavailable during Terraform apply? (Retry logic, error handling)
- What happens when trying to onboard a database type that SIA does not support? (Validation error before apply)
- What happens when trying to onboard a database version below SIA's minimum requirements? (Version validation)
- How does the provider handle credential conflicts when multiple strong accounts target the same database?
- What happens when a database target is manually deleted from SIA outside of Terraform? (State drift detection)
- What happens when trying to delete a strong account that's referenced by active policies or target assignments?
- What happens when the database connection details change in the source provider (AWS/Azure) after SIA onboarding?
- What happens when cloud provider credentials (AWS access keys, Azure service principals) expire or are rotated?
- What happens when a user onboards a database with invalid connection details (database doesn't exist or is unreachable)? (Note: Provider won't detect this during onboarding; SIA will report connectivity errors when access is first attempted)

## Requirements

### Functional Requirements

#### Database Target Management

- **FR-001**: Users MUST be able to onboard existing databases as SIA targets for all SIA-supported database types: SQL Server, Db2, MariaDB, MongoDB, MySQL, Oracle, and PostgreSQL
- **FR-001a**: System MUST support SIA's minimum database version requirements:
  - Db2: 11.5.0.0 and later
  - MariaDB: 10.0.0 and later
  - MongoDB: 4.0 and later
  - MySQL: 5.7.0 and later
  - Oracle: 18c and later
  - PostgreSQL: 12.0.0 and later
  - SQL Server: 13.0.0 and later
- **FR-002**: System MUST support onboarding database targets hosted in AWS, Azure, and on-premise environments
- **FR-003**: System MUST support AWS RDS database services (RDS for MySQL, PostgreSQL, MariaDB, Oracle, SQL Server)
- **FR-004**: System MUST support Azure SQL Database services (Azure SQL Database, Azure Database for MySQL, PostgreSQL, MariaDB)
- **FR-005**: Users MUST be able to specify database connection parameters including address/hostname, port, database name, and authentication method
- **FR-006**: System MUST validate database type against SIA's supported database list before attempting registration
- **FR-008**: System MUST support updating database target configuration after initial creation
- **FR-009**: System MUST support deletion of database targets from SIA
- **FR-010**: System MUST handle Terraform state synchronization when resources are modified outside of Terraform

#### Strong Account Management

- **FR-011**: Users MUST be able to define strong accounts stored directly in SIA (not Privilege Cloud)
- **FR-012**: System MUST support three strong account authentication methods:
  - Local strong account (for MariaDB, MongoDB, MySQL, Oracle, PostgreSQL)
  - Active Directory strong account (for SQL Server, Db2)
  - AWS IAM credentials (for RDS MariaDB, MySQL, PostgreSQL)
- **FR-013**: Users MUST be able to associate strong accounts with specific database targets
- **FR-014**: System MUST securely handle credential input for strong accounts (sensitive data handling)
- **FR-015**: Users MUST be able to update strong account credentials through Terraform
- **FR-015a**: System MUST update strong account credentials in SIA immediately without validating active session state (SIA determines session handling: new sessions use updated credentials, existing sessions behavior is SIA's responsibility)
- **FR-016**: System MUST support deletion of strong accounts when no longer needed

#### Integration and IaC Workflow

- **FR-017**: System MUST integrate with CyberArk SIA REST API for all resource operations
- **FR-018**: System MUST support Terraform's standard resource lifecycle (Create, Read, Update, Delete)
- **FR-019**: System MUST support Terraform import functionality for existing SIA resources
- **FR-020**: System MUST provide clear error messages when API operations fail
- **FR-021**: System MUST authenticate to SIA API using CyberArk Identity Security Platform Shared Services (ISPSS) OAuth2 client credentials flow via the platformtoken endpoint
- **FR-022**: Users MUST be able to configure provider authentication with ISPSS service account credentials (client_id and client_secret)
- **FR-023**: System MUST obtain and use bearer tokens from the platformtoken endpoint for authenticating SIA API requests
- **FR-023a**: System MUST cache bearer tokens and proactively refresh them before expiration (at approximately 80% of token lifetime, ~12 minutes for 15-minute tokens) to prevent mid-operation authentication failures
- **FR-024**: System MUST allow parallel resource creation when resources have no dependencies
- **FR-025**: Users MUST be able to use Terraform variables and data sources to reference database connection details from other providers (AWS, Azure)

#### Data Validation and Error Handling

- **FR-026**: System MUST validate required fields before making API calls to SIA
- **FR-027**: System MUST provide descriptive error messages that map SIA API errors to user-actionable guidance
- **FR-027a**: When SIA onboarding fails after database creation by another provider, system MUST fail the Terraform apply with a clear error message explaining the partial state (database exists in cloud provider but not onboarded to SIA) and provide guidance for manual resolution
- **FR-028**: System MUST detect and report configuration drift when SIA resources are modified outside Terraform
- **FR-029**: System MUST validate port numbers are within valid range (1-65535)
- **FR-030**: System MUST validate database type is supported before attempting operations
- **FR-031**: System MUST NOT attempt direct database connectivity validation during onboarding (connection details are registered with SIA; SIA validates connectivity when access is first attempted)

### Key Entities

- **Database Target**: Represents an existing database that has been onboarded/registered to SIA for secure access management (Note: The database itself is created by AWS/Azure/on-premise infrastructure, this entity only represents its registration in SIA)
  - Attributes: unique identifier, database type (SQL Server, PostgreSQL, etc.), database version, address/hostname, port, database name, authentication method, connection metadata (cloud provider specifics)
  - Relationships: Associated with zero or more strong accounts

- **Strong Account**: Represents credentials that SIA uses to provision ephemeral access to database targets
  - Attributes: unique identifier, account name, authentication method type (local/AD/IAM), credentials (username/password or IAM keys), storage location (SIA service), associated database targets
  - Relationships: Associated with one or more database targets, may reference cloud provider credentials

- **Authentication Method**: Configuration defining how SIA authenticates to database targets
  - Types: Ephemeral local user, Ephemeral domain user, RDS IAM ephemeral credentials
  - Attributes: method type, strong account reference, permissions configuration

## Success Criteria

### Measurable Outcomes

- **SC-001**: Infrastructure engineers can define database creation (via AWS/Azure providers) and SIA onboarding (via SIA provider) in a single Terraform configuration file and apply it with one command, with the complete workflow (database creation + SIA onboarding) completing in under 5 minutes for standard database configurations
- **SC-002**: Changes to database target configuration (e.g., port updates, authentication method changes) propagate to SIA within 30 seconds of Terraform apply completion
- **SC-003**: Strong account credential rotation operations (Terraform apply with updated credentials) complete within 1 minute
- **SC-004**: The provider successfully handles parallel creation of at least 10 database targets without errors or race conditions
- **SC-005**: Configuration drift detection accurately identifies 100% of manual changes made to SIA resources outside of Terraform
- **SC-006**: Error messages provide sufficient detail that users can resolve 90% of configuration issues without consulting SIA API documentation
- **SC-012**: Partial state failures (database created but SIA onboarding failed) provide clear error messages with specific resolution steps (fix SIA issue and re-run, or manually clean up database)
- **SC-007**: Terraform state refresh operations complete within 10 seconds for configurations managing up to 50 database targets
- **SC-011**: Long-running Terraform operations (>15 minutes) complete successfully without authentication interruptions due to token expiration
- **SC-008**: Strong account creation reduces manual credential management time by at least 75% compared to manual SIA console operations
- **SC-009**: The provider supports all 7 SIA-supported database types (SQL Server, Db2, MariaDB, MongoDB, MySQL, Oracle, PostgreSQL) with type-specific configuration options
- **SC-010**: Infrastructure teams report that using the provider reduces security gaps (time between database creation and SIA onboarding) from hours/days to near-zero (automated in same Terraform apply)

### Assumptions

- SIA REST API is available and stable for programmatic access
- Users have appropriate permissions in SIA to create and manage database targets and strong accounts
- Cloud provider APIs (AWS, Azure) are accessible for database resource provisioning
- Terraform execution environment has network connectivity to SIA API (and ISPSS platformtoken endpoint) but does NOT require direct connectivity to database targets
- Users understand basic Terraform workflow and Infrastructure as Code principles
- Strong account credentials provided by users have sufficient privileges for SIA's ephemeral credential generation (permission validation is out of scope - the provider assumes credentials are valid)
- Users have access to an ISPSS service account with appropriate role memberships for SIA API access
- Database platforms/types must already exist in SIA configuration before targets can be registered (platform creation may be out of scope)
- Default authentication method for each database type follows SIA documentation standards (local for most, domain for SQL Server/Db2)
- SIA manages active session behavior when strong account credentials are updated (provider does not track or manage session lifecycle)

### Out of Scope

- **Creating or provisioning database infrastructure** - Databases are created using AWS provider (for RDS), Azure provider (for Azure SQL), or deployed on-premise by other means. This provider ONLY handles onboarding existing databases to SIA as targets.
- **Database connectivity validation** - Provider does not test direct connectivity to database targets during onboarding. SIA performs connectivity validation when database access is first attempted through SIA connectors.
- Managing SIA connector infrastructure or deployment
- **Creating, managing, or assigning Target Sets** - Target Set configuration and database target assignment to Target Sets are performed through SIA console in initial release. Provider focuses on database target and strong account lifecycle only.
- Creating or managing SIA access policies or policy rules
- Integration with CyberArk Privilege Cloud for strong accounts (explicitly noted as out of scope per requirements)
- Monitoring or logging of database access through SIA (that's SIA's operational concern)
- Managing database-level permissions or user accounts beyond what SIA requires for strong accounts
- Supporting non-database targets (Windows, Linux, Kubernetes) - this spec is database-only
- Automatic discovery or import of existing databases not yet onboarded to SIA
- Validation of strong account database permissions (provider assumes credentials have necessary privileges)
- Management of ISPSS service accounts or roles (users must create these through Identity Administration portal)
- Database version upgrades or migrations (provider expects database version meets SIA minimums)
