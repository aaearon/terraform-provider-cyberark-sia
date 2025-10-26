# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
