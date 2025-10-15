# terraform-provider-cyberark-sia Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-15 (Phase 2.5)

## Active Technologies
- **Go**: 1.25.0 (confirmed in Phase 2)
- **ARK SDK**: github.com/cyberark/ark-sdk-golang v1.5.0
- **Terraform Plugin Framework**: v1.16.1 (Plugin Framework v6)
- **Terraform Plugin Log**: v0.9.0

## Project Structure
```
terraform-provider-cyberark-sia/
├── internal/
│   ├── provider/         # Terraform provider implementation
│   ├── client/          # ARK SDK wrappers, retry, error handling
│   └── models/          # Data models (Phase 3+)
├── examples/            # Terraform HCL examples (Phase 3+)
├── docs/                # Documentation
│   ├── sdk-integration.md     # ARK SDK reference
│   └── phase2-reflection.md   # Phase 2 lessons learned
├── specs/               # Feature specifications
└── tests/               # Acceptance tests (Phase 3+)
```

## Commands

### Build & Test
```bash
# Build provider
go build -v

# Run tests (unit)
go test ./...

# Run tests (acceptance - requires TF_ACC=1)
TF_ACC=1 go test ./... -v

# Run specific package tests
go test ./internal/client/... -v
```

### Code Quality
```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...
gofmt -w .
```

### Development Workflow
```bash
# Install locally for testing
go install

# Clean build artifacts
go clean

# Update dependencies
go mod tidy
go mod download
```

## Code Style

### Go Standards
- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before commits
- Write godoc comments for exported functions

### Terraform Provider Patterns
- Use Terraform Plugin Framework v6
- Mark sensitive attributes with `Sensitive: true`
- Use `terraform-plugin-log/tflog` for structured logging
- **NEVER log sensitive data** (passwords, tokens, secrets)

### Error Handling (Phase 2.5 Improvements)
- Use `internal/client.MapError()` for Terraform diagnostics
- Wrap operations with `internal/client.RetryWithBackoff()`
- Classify errors by type (auth, permission, network, etc.)
- Provide actionable error messages with guidance

### Testing Strategy
- **Primary**: Acceptance tests (test against real SIA API)
- **Selective**: Unit tests for complex validators and helpers
- Use `TF_ACC=1` environment variable for acceptance tests
- Mock only when necessary (prefer real integration tests)

## ARK SDK Integration Patterns

### Authentication
```go
// Enable token caching for auto-refresh
ispAuth := auth.NewArkISPAuth(true)

// Authenticate (note: first param is *ArkProfile, NOT context.Context)
_, err := ispAuth.Authenticate(nil, profile, secret, false, false)
```

### Error Handling
```go
// ARK SDK v1.5.0 returns standard error interface (no structured types)
// Use multi-layer detection:
// 1. Standard Go errors (net.Error, context errors)
// 2. Pattern matching (case-insensitive, ordered by specificity)
// 3. Fallback for unknown errors
```

### Retry Logic
```go
// Wrap SDK calls with exponential backoff
err := client.RetryWithBackoff(ctx, config, func() error {
    return siaAPI.WorkspacesDB().AddDatabase(...)
})
```

### Logging
```go
// Use structured logging with context
tflog.Info(ctx, "Operation succeeded", map[string]interface{}{
    "operation": "create",
    "resource_id": id,
})

// NEVER log: password, client_secret, aws_secret_access_key, tokens
```

## Recent Changes

### Phase 2.5 (2025-10-15) - Technical Debt Resolution
- **Enhanced Error Handling**: Robust error classification with fallback
- **Improved Retry Logic**: Better error detection with `net.Error` support
- **Type Safety**: Removed `interface{}` from ProviderData
- **Comprehensive Tests**: 95% coverage for error/retry logic
- **SDK Research**: Documented ARK SDK v1.5.0 packages and limitations
- **Logging**: Added retry attempt logging at WARN level
- **Documentation**: Created `sdk-integration.md` and `phase2-reflection.md`

### Phase 2 (2025-10-15) - Foundation Complete
- **Provider Setup**: Schema with 7 auth attributes + environment fallback
- **Authentication**: ISPSS OAuth2 via ARK SDK with caching
- **SIA Client**: Wrapper for WorkspacesDB() and SecretsDB() access
- **Error Mapping**: ARK SDK errors → actionable Terraform diagnostics
- **Retry Logic**: Exponential backoff with 30s max delay
- **Logging**: Structured logging with sensitive data protection
- **Testing**: Acceptance test scaffolding

### Phase 1 (2025-10-15) - Project Initialization
- Go module initialized
- Dependency management configured
- Project structure created

## Known ARK SDK Limitations (v1.5.0)

1. **No Context Support**: `Authenticate()` doesn't accept `context.Context`
2. **No Structured Errors**: Returns generic `error` interface
3. **No HTTP Status Codes**: Status codes embedded in error strings
4. **Token Expiration**: 15-minute bearer tokens (SDK handles refresh)

See `docs/sdk-integration.md` for detailed SDK integration patterns.

## Next Steps

- **Phase 3**: Implement database_target resource (User Story 1)
- **Phase 4**: Implement strong_account resource (User Story 2)
- **Phase 5**: Lifecycle enhancements (User Story 3)
- **Phase 6**: Documentation and polish

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->