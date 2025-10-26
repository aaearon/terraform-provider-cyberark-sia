# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **CRITICAL**: Fixed 401 Unauthorized errors for certificate resource by implementing OAuth2 access token authentication
  - ARK SDK's `IdentityServiceUser` method was exchanging access tokens for ID tokens
  - SIA API requires access tokens (API authorization claims), not ID tokens (identity claims only)
  - Created custom OAuth2 client (`internal/client/oauth2.go`) that uses `/oauth2/platformtoken` endpoint
  - Implemented `CertificatesClientOAuth2` that uses access tokens directly
  - See `docs/oauth2-authentication-fix.md` for detailed analysis and proof
- **IN PROGRESS**: Secret resource migrated to OAuth2 access token authentication (T029-T034 complete)
  - Removed ARK SDK dependencies from `internal/provider/secret_resource.go`
  - Implemented `SecretsClient` wrapper using generic `RestClient`
  - All CRUD operations (Create, Read, Update, Delete) now use OAuth2 access tokens
  - Following certificate resource pattern (manual TF↔API model conversion)

### Changed
- Provider schema: `username` and `client_secret` are now optional (can use environment variables)
- Certificate resource now uses OAuth2 access token authentication instead of ARK SDK
- Secret resource now uses custom OAuth2 client instead of ARK SDK (partial - needs testing)
- Removed `Sensitive: true` from `cert_body` attribute (public certificates are not sensitive)

### Known Issues
- ⚠️ **Database Workspace resource still uses ARK SDK (ID tokens) - will have 401 errors**
- ⚠️ Secret resource migration incomplete - needs acceptance testing (T035)
- ⚠️ ARK SDK cleanup pending (T057-T069) - old wrapper files still present
- Migration to OAuth2 access tokens required for database workspace resource (see `docs/oauth2-authentication-fix.md`)

## [Unreleased - Prior Changes]

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
