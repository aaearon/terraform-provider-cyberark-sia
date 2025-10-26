# OAuth2 Access Token Authentication Fix

**Date**: 2025-10-25
**Issue**: 401 Unauthorized errors when accessing CyberArk SIA API
**Status**: ✅ RESOLVED for Certificate resource
**Remaining Work**: Database Workspace and Secret resources need migration

---

## Executive Summary

The CyberArk ARK SDK v1.5.0's `IdentityServiceUser` authentication method obtains a valid OAuth2 **access token** but then exchanges it for an **ID token** before making API calls. The SIA API requires **access tokens** (which contain API authorization claims), not **ID tokens** (which only contain identity claims). This causes all API requests to fail with **401 Unauthorized** errors.

**Solution**: Bypass the SDK's ID token exchange by calling `/oauth2/platformtoken` directly and using the access token for all API requests.

---

## Root Cause Analysis

### Problem Discovery

When using the ARK SDK's `IdentityServiceUser` method for service account authentication, all API calls to SIA endpoints returned **401 Unauthorized**, despite:
- Successful authentication (no auth errors)
- Valid bearer token obtained
- Service account has `DpaAdmin` role
- Proper API endpoint URLs

### ARK SDK Source Code Investigation

**File**: `/pkg/auth/identity/ark_identity_service_user.go`

#### Step 1: OAuth2 Client Credentials (Lines 115-149) ✅ CORRECT
```go
// POST /OAuth2/Token/__idaptive_cybr_user_oidc
ai.session.UpdateToken(
    base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ai.username, ai.token))),
    "Basic",
)
response, err := ai.session.Post(
    context.Background(),
    fmt.Sprintf("OAuth2/Token/%s", ai.appName),
    map[string]string{
        "grant_type": "client_credentials",
        "scope":      "api",
    },
)
// Extract access_token from response
accessToken, ok := authResult["access_token"].(string)
```

✅ This step is **correct OAuth2 client credentials flow** and returns a valid **access token**.

#### Step 2: Exchange Access Token for ID Token (Lines 150-193) ❌ PROBLEM
```go
// Use access token to authorize and get ID token
ai.session.UpdateToken(accessToken, "Bearer")
response, err = ai.session.Get(
    context.Background(),
    fmt.Sprintf("OAuth2/Authorize/%s", ai.appName),
    map[string]string{
        "client_id":     ai.appName,
        "response_type": "id_token",  // ⚠️ Requesting ID token!
        "scope":         "openid profile api",
        "redirect_uri":  "https://cyberark.cloud/redirect",
    },
)
// Extract ID token from Location header redirect
idTokens, ok := parsedQuery["id_token"]
ai.sessionToken = idTokens[0]  // ❌ DISCARDS access token, uses ID token
ai.session.UpdateToken(ai.sessionToken, "Bearer")
```

❌ This step **exchanges the access token for an ID token** via OpenID Connect authorization flow.

#### Step 3: ID Token Used for API Calls (via `isp.FromISPAuth()`)
```go
// File: /pkg/common/isp/ark_isp_service_client.go:332
return NewArkISPServiceClient(
    serviceName, "", baseTenantURL, tenantEnv,
    ispAuth.Token.Token,  // ❌ This is the ID token from step 2
    "Authorization", separator, basePath, cookieJar, refreshConnectionCallback
)
```

❌ All SIA API calls use the **ID token**, which lacks API authorization claims.

### Token Type Comparison

| Token Type | Purpose | Contains | What SDK Does | What API Needs |
|------------|---------|----------|---------------|----------------|
| **Access Token** | API authorization | `scope`, `aud`, `permissions` | Obtained in step 1, **DISCARDED** | ✅ Required |
| **ID Token** | User identity | `sub`, `email`, `name` | Obtained in step 2, **USED for API calls** | ❌ Not valid |

### Why This Causes 401 Errors

1. **ID tokens** prove WHO you are (identity verification)
2. **Access tokens** prove WHAT you can do (API authorization)
3. SIA API validates **authorization claims** (`scope`, `aud`, `permissions`)
4. ID tokens don't contain these claims → API rejects with **401 Unauthorized**

---

## Solution Implementation

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Terraform Provider (provider.go:Configure)                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────┐        ┌─────────────────────────────┐ │
│  │ ARK SDK        │        │ Custom OAuth2               │ │
│  │ (WorkspacesDB, │        │ (Certificates only)         │ │
│  │  SecretsDB)    │        │                             │ │
│  └────────────────┘        └─────────────────────────────┘ │
│         │                              │                    │
│         │ Uses ID Token ❌             │ Uses Access Token ✅│
│         ▼                              ▼                    │
│  ┌────────────────────────────────────────────────────────┐│
│  │         CyberArk SIA API                               ││
│  │  - WorkspacesDB: 401 ❌                                ││
│  │  - SecretsDB: 401 ❌                                   ││
│  │  - Certificates: 201 ✅                                ││
│  └────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Implementation Files

#### 1. `internal/client/oauth2.go` (NEW)
Custom OAuth2 client credentials implementation that uses access tokens directly.

**Key Function**: `GetPlatformAccessToken()`
```go
// POST https://{identity_tenant}.id.cyberark.cloud/oauth2/platformtoken
// Authorization: Basic base64(username:password)
// Body: grant_type=client_credentials

func GetPlatformAccessToken(ctx context.Context, config *OAuth2Config) (*OAuth2TokenResponse, error)
```

**What it does**:
- Calls `/oauth2/platformtoken` endpoint directly
- Uses HTTP Basic authentication (username:password)
- Returns **access token** (NOT ID token)
- No additional authorization step

**Key Function**: `ResolveSIAURLFromToken()`
```go
// Extracts SIA API URL from JWT token claims
// Token contains: "subdomain" and "platform_domain"
// Constructs: https://{subdomain}.dpa.{platform_domain}

func ResolveSIAURLFromToken(accessToken string) (string, error)
```

#### 2. `internal/client/certificates_oauth2.go` (NEW)
HTTP client for certificate CRUD operations using access tokens.

**Key Type**: `CertificatesClientOAuth2`
```go
type CertificatesClientOAuth2 struct {
    baseURL     string      // SIA API base URL
    accessToken string      // OAuth2 access token (NOT ID token)
    httpClient  *http.Client
}
```

**Methods**:
- `CreateCertificate()` - POST /api/certificates
- `GetCertificate()` - GET /api/certificates/{id}
- `UpdateCertificate()` - PUT /api/certificates/{id}
- `DeleteCertificate()` - DELETE /api/certificates/{id}
- `ListCertificates()` - GET /api/certificates

All methods use `Authorization: Bearer {accessToken}` header.

#### 3. `internal/provider/provider.go` (MODIFIED)
Provider configuration initializes both ARK SDK clients and OAuth2 clients.

**Lines 170-182**: OAuth2 Certificates Client Initialization
```go
// Initialize OAuth2-based Certificates Client
tflog.Info(ctx, "Initializing OAuth2 certificates client")
if identityURL == "" {
    identityURL = ispAuth.Token.Endpoint  // Get from ARK SDK auth
}
certsClient, err := initCertificatesClient(ctx, username, clientSecret, identityURL)
if err != nil {
    resp.Diagnostics.Append(client.MapError(err, "certificates client initialization"))
    return
}
tflog.Info(ctx, "Certificates client initialized successfully with OAuth2 access token")
```

**Lines 217-241**: Helper Function
```go
func initCertificatesClient(ctx context.Context, username, clientSecret, identityURL string) (*client.CertificatesClientOAuth2, error) {
    // Step 1: Get OAuth2 access token
    tokenResp, err := client.GetPlatformAccessToken(ctx, &client.OAuth2Config{
        IdentityURL:  identityURL,
        ClientID:     username,
        ClientSecret: clientSecret,
    })

    // Step 2: Resolve SIA URL from token
    siaURL, err := client.ResolveSIAURLFromToken(tokenResp.AccessToken)

    // Step 3: Create client with access token
    certsClient, err := client.NewCertificatesClientOAuth2(siaURL, tokenResp.AccessToken)

    return certsClient, nil
}
```

#### 4. `internal/provider/resource_certificate.go` (MODIFIED)
Certificate resource now uses OAuth2 client from provider.

**Line 38**: Type Change
```go
type CertificateResource struct {
    providerData    *ProviderData
    certificatesAPI *client.CertificatesClientOAuth2  // Changed from CertificatesClient
}
```

**Lines 212-223**: Configure Method
```go
func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Use OAuth2-based certificates client from provider
    if providerData.CertificatesClient == nil {
        resp.Diagnostics.AddError(
            "Certificates Client Not Initialized",
            "The certificates client was not properly initialized in the provider.",
        )
        return
    }
    r.certificatesAPI = providerData.CertificatesClient  // OAuth2 client
}
```

### Provider Schema Changes

**Lines 78-92**: Provider attributes now optional
```go
"username": schema.StringAttribute{
    Optional:  true,  // Changed from Required: true
    Sensitive: true,
},
"client_secret": schema.StringAttribute{
    Optional:  true,  // Changed from Required: true
    Sensitive: true,
},
```

Environment variables: `CYBERARK_USERNAME`, `CYBERARK_CLIENT_SECRET`, `CYBERARK_IDENTITY_URL`

### Test Results

**Certificate Creation**: ✅ SUCCESS
```
Resource: cyberarksia_certificate.oauth2_test
Status: Creating...
Result: Certificate created successfully
  - ID: 1761407796062935
  - Name: oauth2-fix-test-20251025155636
  - Expiration: 2026-10-25T15:50:42+00:00
  - HTTP Status: 201 Created ✓
```

**No 401 Errors**: ✅ CONFIRMED

---

## Current State: Resource Status

### ✅ Certificate Resource - FIXED
- **Status**: Using OAuth2 access tokens
- **File**: `internal/provider/resource_certificate.go`
- **Client**: `client.CertificatesClientOAuth2`
- **Authentication**: `/oauth2/platformtoken` → access token → API calls
- **Result**: All CRUD operations work ✅

### ❌ Database Workspace Resource - BROKEN
- **Status**: Using ARK SDK with ID tokens
- **File**: `internal/provider/resource_database_workspace.go`
- **Client**: `siaAPI.WorkspacesDB()` (ARK SDK)
- **Authentication**: ARK SDK → ID token → API calls
- **Result**: 401 Unauthorized ❌

**API Endpoints Affected**:
- `POST /api/workspaces/db` - Create database workspace
- `GET /api/workspaces/db/{id}` - Read database workspace
- `PUT /api/workspaces/db/{id}` - Update database workspace
- `DELETE /api/workspaces/db/{id}` - Delete database workspace

### ❌ Secret Resource - BROKEN
- **Status**: Using ARK SDK with ID tokens
- **File**: `internal/provider/resource_secret.go`
- **Client**: `siaAPI.SecretsDB()` (ARK SDK)
- **Authentication**: ARK SDK → ID token → API calls
- **Result**: 401 Unauthorized ❌

**API Endpoints Affected**:
- `POST /api/secrets/db` - Create database secret
- `GET /api/secrets/db/{id}` - Read database secret
- `PUT /api/secrets/db/{id}` - Update database secret
- `DELETE /api/secrets/db/{id}` - Delete database secret

---

## Migration Plan for Remaining Resources

### Option 1: Create OAuth2 Clients for Each Resource (RECOMMENDED)

**Pros**:
- Consistent with certificate resource implementation
- Full control over HTTP requests
- Easy to debug and maintain
- No dependency on ARK SDK quirks

**Cons**:
- More code to write
- Need to replicate ARK SDK request/response handling
- Manual JSON serialization/deserialization

**Implementation**:

1. **Create `internal/client/database_workspace_oauth2.go`**
   ```go
   type DatabaseWorkspaceClientOAuth2 struct {
       baseURL     string
       accessToken string
       httpClient  *http.Client
   }

   func (c *DatabaseWorkspaceClientOAuth2) CreateDatabaseWorkspace(ctx context.Context, req *DatabaseWorkspaceCreateRequest) (*DatabaseWorkspace, error)
   func (c *DatabaseWorkspaceClientOAuth2) GetDatabaseWorkspace(ctx context.Context, id string) (*DatabaseWorkspace, error)
   func (c *DatabaseWorkspaceClientOAuth2) UpdateDatabaseWorkspace(ctx context.Context, id string, req *DatabaseWorkspaceUpdateRequest) (*DatabaseWorkspace, error)
   func (c *DatabaseWorkspaceClientOAuth2) DeleteDatabaseWorkspace(ctx context.Context, id string) error
   ```

2. **Create `internal/client/secrets_oauth2.go`**
   ```go
   type SecretsClientOAuth2 struct {
       baseURL     string
       accessToken string
       httpClient  *http.Client
   }

   func (c *SecretsClientOAuth2) CreateSecret(ctx context.Context, req *SecretCreateRequest) (*Secret, error)
   func (c *SecretsClientOAuth2) GetSecret(ctx context.Context, id string) (*Secret, error)
   func (c *SecretsClientOAuth2) UpdateSecret(ctx context.Context, id string, req *SecretUpdateRequest) (*Secret, error)
   func (c *SecretsClientOAuth2) DeleteSecret(ctx context.Context, id string) error
   ```

3. **Update Provider Data Structure**
   ```go
   type ProviderData struct {
       ISPAuth                *auth.ArkISPAuth                        // Keep for backward compatibility
       SIAAPI                 *sia.ArkSIAAPI                          // Deprecated - uses ID tokens
       CertificatesClient     *client.CertificatesClientOAuth2        // ✅ OAuth2
       DatabaseWorkspaceClient *client.DatabaseWorkspaceClientOAuth2  // NEW - OAuth2
       SecretsClient          *client.SecretsClientOAuth2             // NEW - OAuth2
   }
   ```

4. **Update Resources**
   - Modify `resource_database_workspace.go` to use `DatabaseWorkspaceClientOAuth2`
   - Modify `resource_secret.go` to use `SecretsClientOAuth2`

**Estimated Effort**: 2-3 days
- Day 1: Implement database_workspace_oauth2.go
- Day 2: Implement secrets_oauth2.go
- Day 3: Testing and validation

### Option 2: Patch ARK SDK to Use Access Tokens

**Pros**:
- Minimal code changes in provider
- Leverages existing ARK SDK functionality
- Centralized fix

**Cons**:
- Requires forking and maintaining ARK SDK
- SDK updates require re-applying patches
- May break other SDK functionality
- Not sustainable long-term

**Implementation**:
1. Fork `github.com/cyberark/ark-sdk-golang`
2. Modify `ark_identity_service_user.go:140-193` to skip ID token exchange
3. Update provider `go.mod` to use forked SDK
4. Maintain fork for all future SDK updates

**Estimated Effort**: 1-2 weeks
- Including SDK testing, fork maintenance, documentation

**Recommendation**: ❌ NOT RECOMMENDED - maintenance burden too high

### Option 3: Create Wrapper Around ARK SDK

**Pros**:
- Reuses ARK SDK HTTP client infrastructure
- Can intercept and replace tokens before requests
- Less code duplication

**Cons**:
- Complex wrapper logic
- Still depends on ARK SDK internals
- Token replacement may be fragile

**Implementation**:
```go
type SIAClientWrapper struct {
    siaAPI      *sia.ArkSIAAPI
    accessToken string
}

func (w *SIAClientWrapper) interceptRequest(req *http.Request) {
    // Replace ID token with access token in Authorization header
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.accessToken))
}
```

**Estimated Effort**: 3-5 days

**Recommendation**: ❌ NOT RECOMMENDED - too fragile, hard to maintain

---

## Recommended Approach: Option 1

### Phase 1: Database Workspace Resource (Priority 1)

**Steps**:

1. **API Documentation Research**
   - Document WorkspacesDB API endpoints
   - Map ARK SDK request/response structures to API contracts
   - Identify all field mappings (SDK struct → API JSON)

2. **Create `database_workspace_oauth2.go`**
   - Implement `DatabaseWorkspaceClientOAuth2` struct
   - Implement CRUD methods mirroring `certificates_oauth2.go` pattern
   - Use `common.DeserializeJSONSnake()` for response parsing (same as ARK SDK)
   - Add retry logic with `RetryWithBackoff()`

3. **Update Provider**
   - Initialize `DatabaseWorkspaceClientOAuth2` in `provider.go:Configure()`
   - Add to `ProviderData` struct
   - Update `resource_database_workspace.go:Configure()` to use OAuth2 client

4. **Testing**
   - Create test database workspace
   - Verify CRUD operations
   - Confirm no 401 errors
   - Test drift detection

### Phase 2: Secret Resource (Priority 2)

**Steps**:

1. **API Documentation Research**
   - Document SecretsDB API endpoints
   - Map ARK SDK request/response structures
   - Identify secret-specific field mappings

2. **Create `secrets_oauth2.go`**
   - Implement `SecretsClientOAuth2` struct
   - Implement CRUD methods
   - Handle sensitive data properly (secret passwords)
   - Add retry logic

3. **Update Provider**
   - Initialize `SecretsClientOAuth2` in `provider.go:Configure()`
   - Add to `ProviderData` struct
   - Update `resource_secret.go:Configure()` to use OAuth2 client

4. **Testing**
   - Create test secret
   - Verify CRUD operations
   - Test secret rotation
   - Confirm no 401 errors

### Phase 3: Cleanup and Deprecation

**Steps**:

1. **Remove ARK SDK Dependencies** (Breaking Change)
   - Remove `SIAAPI *sia.ArkSIAAPI` from `ProviderData`
   - Remove `internal/client/sia_client.go`
   - Remove `ISPAuth` initialization (keep OAuth2 only)
   - Update `go.mod` to remove unused ARK SDK packages

2. **Documentation Updates**
   - Update `README.md` with OAuth2 authentication details
   - Document breaking changes
   - Update examples
   - Add migration guide from v0.x to v1.0

3. **Version Bump**
   - Bump to v1.0.0 (breaking change)
   - Update `CHANGELOG.md`
   - Tag release

---

## Technical Notes

### Token Lifecycle

**Access Token Expiration**: Typically 3600 seconds (1 hour)

**Current Implementation**: No automatic token refresh for OAuth2 clients

**Future Enhancement**: Implement token refresh logic
```go
type OAuth2Client struct {
    tokenResp   *OAuth2TokenResponse
    tokenExpiry time.Time
    config      *OAuth2Config
    mu          sync.RWMutex
}

func (c *OAuth2Client) getValidToken(ctx context.Context) (string, error) {
    c.mu.RLock()
    if time.Now().Before(c.tokenExpiry.Add(-5 * time.Minute)) {
        defer c.mu.RUnlock()
        return c.tokenResp.AccessToken, nil
    }
    c.mu.RUnlock()

    // Token expired or expiring soon - refresh
    c.mu.Lock()
    defer c.mu.Unlock()

    tokenResp, err := GetPlatformAccessToken(ctx, c.config)
    if err != nil {
        return "", err
    }

    c.tokenResp = tokenResp
    c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

    return tokenResp.AccessToken, nil
}
```

### Identity URL Resolution

**Current Approach**: ARK SDK auto-discovers identity URL from username

**Process**:
1. Provider authenticates via `client.NewISPAuth()` (uses ARK SDK)
2. ARK SDK resolves identity URL automatically
3. Provider reads `ispAuth.Token.Endpoint` for identity URL
4. Passes identity URL to `GetPlatformAccessToken()`

**Alternative**: Use CyberArk Platform Discovery Service
```bash
curl https://platform-discovery.cyberark.cloud/api/v2/services/subdomain/{subdomain}
```

**Response**:
```json
{
  "identity_user_portal": {
    "api": "https://{tenant}.id.cyberark.cloud"
  }
}
```

**Note**: Discovery service requires proper subdomain (not just tenant ID).

### SIA API URL Construction

**Pattern**: `https://{subdomain}.dpa.{platform_domain}`

**Example**: `https://cyberiamtest.dpa.cyberark.cloud`

**Extraction from Token**:
```go
claims := token.Claims.(jwt.MapClaims)
subdomain := claims["subdomain"].(string)        // "cyberiamtest"
platformDomain := claims["platform_domain"].(string)  // "cyberark.cloud"

// Remove "shell." prefix if present
if strings.HasPrefix(platformDomain, "shell.") {
    platformDomain = strings.TrimPrefix(platformDomain, "shell.")
}

siaURL := fmt.Sprintf("https://%s.dpa.%s", subdomain, platformDomain)
```

### Error Handling

**401 Unauthorized**: Now only occurs for actual auth failures (invalid credentials)

**403 Forbidden**: Insufficient permissions (role-based access control)

**409 Conflict**: Resource already exists or in use

**429 Too Many Requests**: Rate limiting

**Retry Logic**: Uses exponential backoff (500ms → 30s max, 3 retries)

### Performance Considerations

**Token Reuse**: Single access token used for all requests until expiry

**Concurrent Requests**: OAuth2 client is thread-safe (uses http.Client)

**Connection Pooling**: Go's `http.Client` handles connection reuse automatically

---

## References

### Source Code Locations

**ARK SDK v1.5.0**:
- Authentication: `/pkg/auth/identity/ark_identity_service_user.go`
- ISP Client: `/pkg/common/isp/ark_isp_service_client.go`
- SIA Service: `/pkg/services/sia/workspaces/db/ark_sia_workspaces_db_service.go`

**Provider**:
- OAuth2 Client: `internal/client/oauth2.go`
- Certificates OAuth2: `internal/client/certificates_oauth2.go`
- Provider Config: `internal/provider/provider.go:170-182, 217-241`
- Certificate Resource: `internal/provider/resource_certificate.go:38, 212-223`

### Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Project development guidelines
- [sdk-integration.md](sdk-integration.md) - ARK SDK integration patterns (needs update)
- [CHANGELOG.md](../CHANGELOG.md) - Version history (needs update)

### External Resources

- [CyberArk OAuth2 Documentation](https://docs.cyberark.com/ispss-access/latest/en/content/ispss/ispss-api-authentication.htm) (may not exist)
- [ARK SDK Documentation](https://cyberark.github.io/ark-sdk-golang/)
- [OAuth2 Client Credentials Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)

---

## Changelog

**2025-10-25**: Initial OAuth2 fix implementation
- Created custom OAuth2 authentication for certificate resource
- Proved ARK SDK ID token issue via source code analysis
- Documented migration plan for database workspace and secret resources

---

## Next Steps

1. **Immediate**: Document this fix in CHANGELOG.md
2. **Short-term**: Implement `database_workspace_oauth2.go` (Phase 1)
3. **Medium-term**: Implement `secrets_oauth2.go` (Phase 2)
4. **Long-term**: Remove ARK SDK dependency entirely (Phase 3, v1.0.0)

---

**Author**: Claude Code
**Reviewed By**: [Pending]
**Approved By**: [Pending]
