# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Future changes will be documented here

## [0.2.0] - 2025-11-01

### BREAKING CHANGES
- **Provider Schema**: Renamed provider attribute `client_secret` to `password` for improved clarity
  - Update your provider configuration blocks: change `client_secret` to `password`
  - Example: `provider "cyberarksia" { password = var.password }`
- **Environment Variable**: Renamed `CYBERARK_CLIENT_SECRET` to `CYBERARK_PASSWORD`
  - Update environment variables in CI/CD pipelines and local development
  - Update `.env` files and Terraform variable references
- **Documentation**: Updated security recommendations to reference CyberArk Conjur instead of Terraform Cloud/HashiCorp Vault

### Migration Guide
1. **Update Provider Configuration**: Change `client_secret` attribute to `password` in all `provider "cyberarksia"` blocks
2. **Update Environment Variables**: Rename `CYBERARK_CLIENT_SECRET` to `CYBERARK_PASSWORD` in:
   - Shell environment (`export CYBERARK_PASSWORD="..."`)
   - CI/CD pipeline secrets
   - Terraform Cloud/Enterprise workspace variables
   - `.env` files
3. **Update Terraform Variables**: Rename any variables like `cyberark_client_secret` to `cyberark_password`
4. **Update Scripts**: Search for `CYBERARK_CLIENT_SECRET` in automation scripts and update to `CYBERARK_PASSWORD`

### Changed
- All examples updated to use `password` attribute
- All documentation updated to reference `password` terminology
- Error messages now use "username or password" terminology instead of "client_id or client_secret"

## [0.1.2] - 2025-11-01

### Fixed
- Provider description now includes comprehensive feature list (ZSP/JIT access, 60+ database engines, OAuth2)
- Documentation regenerated to ensure all 6 resources appear in Terraform Registry
  - `cyberarksia_database_workspace` (was missing from Registry)
  - `cyberarksia_secret` (was missing from Registry)
  - `cyberarksia_database_policy_principal_assignment` (was missing from Registry)
  - All resource and data source documentation updated with current schema
- Makefile now detects OS/architecture automatically (works on macOS, Linux, Windows)

### Added
- Complete end-to-end workflow example in `examples/complete/end-to-end-workflow/`
  - Demonstrates secret management, database workspaces, policies, and assignments
  - Shows both inline and modular assignment patterns
  - Includes comprehensive README with troubleshooting guide
- Missing `examples/resources/database_workspace/` with multiple cloud provider examples
- `.pre-commit-config.yaml` already existed (verified complete)

## [0.1.1] - 2025-10-30

### Fixed
- Binary naming to match repository rename (terraform-provider-cyberarksia)
- Terraform Registry installation now works correctly

## [0.1.0] - 2025-10-30

### Added
- Comprehensive acceptance test coverage for policy resources
  - `database_policy_resource_test.go`: 12 tests covering CRUD, conditions, time frames, inline assignments, validation, and ForceNew behavior
  - `database_policy_principal_assignment_resource_test.go`: 10 tests covering principal types (USER, GROUP, ROLE), composite IDs, and assignments
  - `policy_database_assignment_resource_test.go`: 14 tests covering all 6 authentication methods (db_auth, ldap_auth, oracle_auth, mongo_auth, sqlserver_auth, rds_iam_user_auth) and composite IDs
- Complete profile factory test coverage
  - Added tests for all 4 remaining authentication methods: OracleAuth, MongoAuth, SQLServerAuth, RDSIAMUserAuth
  - Total coverage: 14 tests for all 6 authentication profile types
- Complete validator test coverage (100% coverage)
  - `policy_status_validator_test.go`: 16 test cases validating "active"/"suspended" status values
  - `principal_type_validator_test.go`: 22 test cases validating USER/GROUP/ROLE types
  - `location_type_validator_test.go`: 20 test cases validating "FQDN/IP" location type
  - `database_engine_validator_test.go`: 67 test cases covering 60+ database engines (AWS, Azure, GCP, on-premise, Atlas)
  - `uuid_validator_test.go`: 27 test cases validating UUID v4 format
  - `email_like_validator_test.go`: 45 test cases validating email-like principal names
- LLM Testing Guide in CLAUDE.md for automated CRUD operation validation
  - Structured test plans for Certificate and Database Workspace resources
  - Validation checklists and expected outputs
  - Common testing patterns and automation sequences
- Initial provider implementation
- Certificate resource (`cyberarksia_certificate`)
  - Create, read, update, delete TLS/SSL certificates
  - Support for PEM and DER formats
  - Automatic X.509 metadata extraction
  - Label-based organization
- Database workspace resource (`cyberarksia_database_workspace`)
  - Configure database targets with 60+ supported engines
  - Multi-cloud support (AWS, Azure, GCP, Atlas, on-premise)
  - Certificate-based authentication
  - Network segmentation support
  - Authentication method configuration
- Database policy resource (`cyberarksia_database_policy`)
  - Access policies with session limits and time-based restrictions
  - Policy tags, time frames, and access windows
  - Support for inline principal and database assignments
- Database policy principal assignment resource (`cyberarksia_database_policy_principal_assignment`)
  - Assign users, groups, or roles to policies
  - Support for multiple directory types (Cloud Directory, Azure AD, LDAP)
- Policy database assignment resource (`cyberarksia_database_policy_database_assignment`)
  - Connect database workspaces to policies
  - Support for 6 authentication methods
- Secret resource (`cyberarksia_secret`)
  - Store database credentials (local auth, Active Directory, AWS IAM)
- Principal data source (`cyberarksia_principal`)
  - Look up users, groups, and roles by name
- Database policy data source (`cyberarksia_database_policy`)
  - Reference existing policies by name or ID
- Provider authentication using CyberArk Identity OAuth2
- ARK SDK integration with automatic token refresh
- Comprehensive error handling and retry logic with exponential backoff
- Acceptance test suite with 69 tests
- Example configurations for common use cases

### Changed
- Increased total acceptance test count from 33 to 69 tests
- Policy resource testing coverage increased from 0% to comprehensive (36 new tests)
- Validator test coverage increased from 22.9% to 100.0% (197+ new test cases)
- Total unit test functions increased from 34 to 47

### Security
- All sensitive fields (passwords, secrets, certificate bodies) properly marked as sensitive
- Certificate validation enabled by default
- Secure OAuth2 token handling with automatic refresh

### Documentation
- Complete resource documentation
- SDK integration guide
- Development guidelines
- Troubleshooting guide
- Multiple example configurations

### Added
- LLM Testing Guide in CLAUDE.md for automated CRUD operation validation
  - Structured test plans for Certificate and Database Workspace resources
  - Validation checklists and expected outputs
  - Common testing patterns and automation sequences
- Initial provider implementation
- Certificate resource (`cyberarksia_certificate`)
  - Create, read, update, delete TLS/SSL certificates
  - Support for PEM and DER formats
  - Automatic X.509 metadata extraction
  - Label-based organization
- Database workspace resource (`cyberarksia_database_workspace`)
  - Configure database targets with 60+ supported engines
  - Multi-cloud support (AWS, Azure, GCP, Atlas, on-premise)
  - Certificate-based authentication
  - Network segmentation support
  - Authentication method configuration
- Provider authentication using CyberArk Identity OAuth2
- ARK SDK integration with automatic token refresh
- Comprehensive error handling and retry logic with exponential backoff
- Acceptance test suite
- Example configurations for common use cases

---

## Version History Notes

This provider was developed using a test-driven approach with comprehensive planning and specification documents available in the `specs/` directory.

### Development Phases
- **Phase 1**: Project foundation and authentication
- **Phase 2**: Certificate resource implementation
- **Phase 3**: Database workspace resource (renamed from database_target)

For detailed architectural decisions and implementation insights, see [docs/development-history.md](docs/development-history.md).
