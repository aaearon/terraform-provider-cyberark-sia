# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Changed
- Increased total acceptance test count from 33 to 69 tests
- Policy resource testing coverage increased from 0% to comprehensive (36 new tests)
- Validator test coverage increased from 22.9% to 100.0% (197+ new test cases)
- Total unit test functions increased from 34 to 47

## [0.1.0] - 2025-01-27

### BREAKING CHANGES
- **Certificate Resource**: Removed 6 fabricated attributes that don't exist in the CyberArk SIA Certificates API
  - Removed `created_by` (user who created certificate - not returned by API)
  - Removed `last_updated_by` (user who last updated - not returned by API)
  - Removed `version` (version number - not returned by API)
  - Removed `checksum` (SHA256 hash - not returned by API)
  - Removed `updated_time` (last modification timestamp - not returned by API)
  - Removed `cert_password` (API only supports public keys, not encrypted certificates)
  - **Migration**: Remove references to these attributes from your Terraform configurations and outputs
  - **Impact**: Existing state files with these attributes will cause plan errors until refreshed

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

## [0.1.0] - TBD

Initial development release.

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

---

## Version History Notes

This provider was developed using a test-driven approach with comprehensive planning and specification documents available in the `specs/` directory.

### Development Phases
- **Phase 1**: Project foundation and authentication
- **Phase 2**: Certificate resource implementation
- **Phase 3**: Database workspace resource (renamed from database_target)

For detailed architectural decisions and implementation insights, see [docs/development-history.md](docs/development-history.md).
