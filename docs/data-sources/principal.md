# cyberarksia_principal Data Source

Looks up a principal (user, group, or role) by name from CyberArk Identity directories.

Use this data source to retrieve principal information for policy assignments without manual UUID lookups. Supports Cloud Directory (CDS), Federated Directory (FDS/Entra ID), and Active Directory (AdProxy).

## Features

- **Universal Support**: Works with all directory types (CDS, FDS, AdProxy)
- **All Principal Types**: Supports USER, GROUP, and ROLE principals
- **Hybrid Lookup**: Automatically optimizes between fast and comprehensive lookup strategies
- **Case-Insensitive**: Principal name matching is case-insensitive
- **Type Filtering**: Optional filter by principal type

## Example Usage

### Basic User Lookup

```hcl
data "cyberarksia_principal" "admin_user" {
  name = "admin@cyberark.cloud.12345"
}

output "user_id" {
  value = data.cyberarksia_principal.admin_user.id
}
```

### Group Lookup with Type Filter

```hcl
data "cyberarksia_principal" "db_admins" {
  name = "Database Administrators"
  type = "GROUP"
}
```

### Integration with Policy Assignment

```hcl
# Look up the principal
data "cyberarksia_principal" "db_user" {
  name = "tim.schindler@cyberark.cloud.40562"
}

# Assign principal to policy
resource "cyberarksia_database_policy_principal_assignment" "user_access" {
  policy_id               = cyberarksia_database_policy.prod.policy_id
  principal_id            = data.cyberarksia_principal.db_user.id
  principal_type          = data.cyberarksia_principal.db_user.principal_type
  principal_name          = data.cyberarksia_principal.db_user.name
  source_directory_name   = data.cyberarksia_principal.db_user.directory_name
  source_directory_id     = data.cyberarksia_principal.db_user.directory_id
}
```

### Federated Directory User

```hcl
data "cyberarksia_principal" "entra_user" {
  name = "john.doe@company.com"
}
```

### Active Directory User

```hcl
data "cyberarksia_principal" "ad_user" {
  name = "SchindlerT@cyberiam.tech"
}
```

## Argument Reference

* `name` - (Required) The principal's SystemName (unique identifier). Examples:
  - Cloud Directory: `user@cyberark.cloud.12345`
  - Federated Directory: `john.doe@company.com`
  - Active Directory: `SchindlerT@cyberiam.tech`
  - Group: `Database Administrators`

* `type` - (Optional) Filter by principal type. Valid values:
  - `USER` - User principals only
  - `GROUP` - Group principals only
  - `ROLE` - Role principals only

  If omitted, searches all principal types.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The principal's unique identifier (UUID).
* `principal_type` - The type of principal: `USER`, `GROUP`, or `ROLE`.
* `directory_name` - The localized, human-readable directory name (e.g., `CyberArk Cloud Directory`, `Federation with company.com`, `Active Directory (domain.com)`).
* `directory_id` - The directory's unique identifier (UUID).
* `display_name` - The principal's human-readable display name.
* `email` - The principal's email address. Only present for USER principals (null for GROUP/ROLE).
* `description` - The principal's description. May be null/empty depending on the principal type and configuration.

## Directory Type Support

### Cloud Directory (CDS)

Native CyberArk cloud users. These principals are created directly in the CyberArk Identity tenant.

```hcl
data "cyberarksia_principal" "cloud_user" {
  name = "service-account@cyberark.cloud.12345"
}
```

**Lookup Performance**: < 1 second (fast path via UserByName API)

### Federated Directory (FDS)

Users from external identity providers (Entra ID, Okta, etc.) federated with CyberArk Identity.

```hcl
data "cyberarksia_principal" "federated_user" {
  name = "user@company.com"
}
```

**Lookup Performance**: < 1 second (fast path via UserByName API)

**Note**: The `directory_name` will show the localized federation name (e.g., "Federation with company.com"), not the generic "FDS".

### Active Directory (AdProxy)

On-premises Active Directory users synchronized via AdProxy connector.

```hcl
data "cyberarksia_principal" "ad_user" {
  name = "UserName@domain.local"
}
```

**Lookup Performance**: < 2 seconds (fallback path via ListDirectoriesEntities API)

**Note**: The `directory_name` will show a recognizable name like "Active Directory (domain.com)".

## Performance Characteristics

The data source uses a hybrid lookup strategy that automatically optimizes performance:

### Fast Path (Phase 1)
- **When**: Looking up USER principals
- **Method**: Direct UserByName API call
- **Performance**: ~100ms
- **Directories**: CDS, FDS

### Fallback Path (Phase 2)
- **When**: Looking up GROUP/ROLE principals, or USER not found in Phase 1
- **Method**: Comprehensive entity scan with client-side filtering
- **Performance**: ~1-2 seconds
- **Directories**: All (CDS, FDS, AdProxy)

### Scalability
- Tested with 200 entities
- Supports up to 10,000 principals per tenant
- No caching (each Terraform run re-queries the API)

## Error Messages

### Principal Not Found

```
Error: Principal Not Found
Principal 'nonexistent@example.com' not found in any directory
```

**Resolution**: Verify the principal name spelling and ensure the principal exists in one of the configured directories.

### Authentication Failed

```
Error: Authentication Failed
Failed to authenticate with CyberArk Identity: invalid credentials
```

**Resolution**: Check your provider configuration (username/client_secret).

### API Connectivity Error

```
Error: API Error
Failed to query CyberArk Identity API: connection timeout
```

**Resolution**: Verify network connectivity to the CyberArk Identity tenant and check firewall/proxy settings.

## Implementation Details

### Case-Insensitive Matching

Principal names are matched case-insensitively. The following are equivalent:

```hcl
name = "user@example.com"
name = "USER@EXAMPLE.COM"
name = "User@Example.Com"
```

### SystemName vs DisplayName

The `name` parameter matches against the **SystemName** (the unique identifier), not the DisplayName (the human-readable name).

- **SystemName**: `tim.schindler@cyberark.cloud.40562` (use this)
- **DisplayName**: `Tim Schindler` (don't use this)

### Null Field Handling

Some fields may be null depending on the principal type:

- `email`: Only present for USER principals (null for GROUP/ROLE)
- `description`: May be null/empty for any principal type

### State Management

Data sources in Terraform are stateless and read-only. Each `terraform plan` or `terraform apply` re-queries the API. There is no caching between invocations.

## See Also

* [cyberarksia_database_policy_principal_assignment](../resources/database_policy_principal_assignment.md) - Resource for assigning principals to policies
* [cyberarksia_access_policy](./access_policy.md) - Data source for looking up policies
