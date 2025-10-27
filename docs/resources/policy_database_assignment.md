# cyberarksia_policy_database_assignment

Manages the assignment of a database workspace to an existing SIA access policy. This resource follows the AWS Security Group Rule pattern - managing individual database assignments to a policy rather than managing the entire policy.

## ⚠️ Important: Resource Conflicts (Similar to aws_security_group_rule)

This resource manages individual database assignments within a policy. Like AWS Security Group Rules, concurrent modifications to the same policy from multiple sources can cause conflicts.

### Important Limitation

**Do not manage the same policy from multiple Terraform workspaces.** The last write will overwrite previous changes, potentially causing data loss.

#### Safe vs Unsafe Patterns

- ✅ **SAFE**: Multiple assignment resources in the SAME workspace
- ✅ **SAFE**: UI-managed databases (automatically preserved by read-modify-write pattern)
- ✅ **SAFE**: Different policies managed by different workspaces
- ⚠️ **UNSAFE**: Multiple Terraform workspaces modifying the same policy
- ⚠️ **UNSAFE**: Concurrent manual UI changes to managed policies

### Best Practices

1. **Manage all database assignments for a policy in a single Terraform workspace**
2. **Use modules for code organization** (not separate workspaces)
3. **Coordinate with team members** before manual UI changes to managed policies
4. **Review plan output carefully** before applying changes
5. **Use the data source** to reference policies managed elsewhere

This is an accepted limitation of the separate-resource pattern, following the same model as AWS, GCP, and Azure providers.

## ⚠️ Critical: Policy Location Type Constraint

**Policies are locked to a single location type** based on the `locationType` field set when the policy is created. You **cannot mix different location types** in the same policy.

### Location Types

1. **`FQDN_IP`** - On-premise databases (cloud_provider: `on_premise`)
2. **`AWS`** - AWS RDS/Aurora databases (cloud_provider: `aws`)
3. **`AZURE`** - Azure SQL/PostgreSQL databases (cloud_provider: `azure`)
4. **`GCP`** - Google Cloud SQL databases (cloud_provider: `gcp`)
5. **`ATLAS`** - MongoDB Atlas databases (cloud_provider: `atlas`)

### API Error Examples

If you attempt to assign a database with a different location type than the policy allows:

```
Error: The only allowed key in the targets dictionary is "FQDN/IP".
```

This means the policy has `locationType: "FQDN_IP"` and you're trying to assign an AWS/Azure/GCP database.

### Solution: Use Separate Policies Per Location Type

```hcl
# On-premise database policy
data "cyberarksia_access_policy" "onprem_databases" {
  name = "On-Premise DB Access"  # locationType: FQDN_IP
}

resource "cyberarksia_policy_database_assignment" "postgres_onprem" {
  policy_id             = data.cyberarksia_access_policy.onprem_databases.id
  database_workspace_id = cyberarksia_database_workspace.postgres_local.id
  # ... postgres_local has cloud_provider = "on_premise"
}

# AWS database policy (separate!)
data "cyberarksia_access_policy" "aws_databases" {
  name = "AWS RDS Access"  # locationType: AWS
}

resource "cyberarksia_policy_database_assignment" "rds_postgres" {
  policy_id             = data.cyberarksia_access_policy.aws_databases.id
  database_workspace_id = cyberarksia_database_workspace.rds_instance.id
  # ... rds_instance has cloud_provider = "aws"
}
```

**Important**: This constraint is enforced by the CyberArk SIA API and is set at **policy creation time**. It cannot be changed. If you need to support multiple location types, create separate policies for each.

## Prerequisites

Before using this resource, you must have:

1. **Existing SIA Access Policy**: Create policies via the CyberArk Identity UI or API
2. **Database Workspace**: Use the `cyberarksia_database_workspace` resource
3. **Secret Resource**: Database workspaces require secrets for authentication
4. **Sufficient IAM Permissions**: Service account must have policy modification permissions

## Example Usage

### Basic Database Authentication

```hcl
# Lookup existing policy
data "cyberarksia_access_policy" "db_admins" {
  name = "Database Administrators"
}

# Create database workspace with secret
resource "cyberarksia_secret" "postgres_admin" {
  name                = "postgres-admin"
  authentication_type = "local"
  username            = "admin"
  password            = var.db_password
}

resource "cyberarksia_database_workspace" "production" {
  name          = "prod-postgres"
  database_type = "postgres"
  address       = "postgres.example.com"
  port          = 5432
  secret_id     = cyberarksia_secret.postgres_admin.id
}

# Assign database to policy with db_auth
resource "cyberarksia_policy_database_assignment" "postgres_admin" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.production.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["db_reader", "db_writer", "db_admin"]
  }
}
```

### LDAP Authentication

```hcl
resource "cyberarksia_policy_database_assignment" "ldap_auth" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.production.id
  authentication_method  = "ldap_auth"

  ldap_auth_profile {
    assign_groups = ["DBAdmins", "Developers", "DataEngineers"]
  }
}
```

### Oracle Authentication with Privileged Roles

```hcl
resource "cyberarksia_policy_database_assignment" "oracle_dba" {
  policy_id              = data.cyberarksia_access_policy.oracle_admins.id
  database_workspace_id  = cyberarksia_database_workspace.oracle_prod.id
  authentication_method  = "oracle_auth"

  oracle_auth_profile {
    roles        = ["CONNECT", "RESOURCE", "DBA"]
    dba_role     = true
    sysdba_role  = false
    sysoper_role = false
  }
}
```

### MongoDB Authentication

```hcl
resource "cyberarksia_policy_database_assignment" "mongo_auth" {
  policy_id              = data.cyberarksia_access_policy.mongo_users.id
  database_workspace_id  = cyberarksia_database_workspace.mongo_cluster.id
  authentication_method  = "mongo_auth"

  mongo_auth_profile {
    # Global roles apply across all databases
    global_builtin_roles = ["readWriteAnyDatabase", "dbAdminAnyDatabase"]

    # Database-specific builtin roles
    database_builtin_roles = {
      "myapp_db"   = ["readWrite", "dbAdmin"]
      "analytics"  = ["read"]
    }

    # Database-specific custom roles
    database_custom_roles = {
      "myapp_db" = ["customAppRole"]
    }
  }
}
```

### SQL Server Authentication

```hcl
resource "cyberarksia_policy_database_assignment" "sqlserver_auth" {
  policy_id              = data.cyberarksia_access_policy.sql_admins.id
  database_workspace_id  = cyberarksia_database_workspace.sqlserver_prod.id
  authentication_method  = "sqlserver_auth"

  sqlserver_auth_profile {
    # Global roles
    global_builtin_roles = ["db_datareader", "db_datawriter"]
    global_custom_roles  = ["CustomAdminRole"]

    # Database-specific roles
    database_builtin_roles = {
      "master"     = ["db_owner"]
      "production" = ["db_datareader", "db_datawriter", "db_ddladmin"]
    }

    database_custom_roles = {
      "production" = ["AppSpecificRole"]
    }
  }
}
```

### AWS RDS IAM Authentication

```hcl
resource "cyberarksia_policy_database_assignment" "rds_iam" {
  policy_id              = data.cyberarksia_access_policy.rds_users.id
  database_workspace_id  = cyberarksia_database_workspace.rds_postgres.id
  authentication_method  = "rds_iam_user_auth"

  rds_iam_user_auth_profile {
    db_user = "rds_iam_user"
  }
}
```

### Multiple Databases in Same Policy

```hcl
# Multiple assignments to the same policy (SAFE - same workspace)
resource "cyberarksia_policy_database_assignment" "postgres" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.postgres.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["db_admin"]
  }
}

resource "cyberarksia_policy_database_assignment" "mysql" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.mysql.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["admin"]
  }
}

resource "cyberarksia_policy_database_assignment" "oracle" {
  policy_id              = data.cyberarksia_access_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.oracle.id
  authentication_method  = "oracle_auth"

  oracle_auth_profile {
    roles      = ["DBA"]
    dba_role   = true
  }
}
```

## Argument Reference

### Required Arguments

- `policy_id` (String) - The ID of the SIA access policy. Use the `cyberarksia_access_policy` data source to lookup policies by name. **Forces replacement** when changed.
- `database_workspace_id` (String) - The ID of the database workspace to assign. Reference a `cyberarksia_database_workspace` resource. **Forces replacement** when changed.
- `authentication_method` (String) - The authentication method to use for this database. **Valid values**: `db_auth`, `ldap_auth`, `oracle_auth`, `mongo_auth`, `sqlserver_auth`, `rds_iam_user_auth`.

### Profile Blocks

Exactly **ONE** of the following profile blocks must be configured, matching the `authentication_method`:

#### `db_auth_profile` Block

Used when `authentication_method = "db_auth"` (standard database authentication).

- `roles` (List of String, **Required**) - Database roles to grant. Examples: `["db_reader", "db_writer", "db_admin"]`

#### `ldap_auth_profile` Block

Used when `authentication_method = "ldap_auth"` (LDAP/Active Directory group-based authentication).

- `assign_groups` (List of String, **Required**) - LDAP/AD groups to assign. Examples: `["DBAdmins", "Developers", "DataEngineers"]`

#### `oracle_auth_profile` Block

Used when `authentication_method = "oracle_auth"` (Oracle-specific authentication with privileged roles).

- `roles` (List of String, **Required**) - Oracle database roles. Examples: `["CONNECT", "RESOURCE", "DBA"]`
- `dba_role` (Boolean, Optional) - Grant DBA role (standard administrative privileges). Default: `false`
- `sysdba_role` (Boolean, Optional) - Grant SYSDBA role (highest privilege, instance management). Default: `false`
- `sysoper_role` (Boolean, Optional) - Grant SYSOPER role (operational privileges, limited admin). Default: `false`

#### `mongo_auth_profile` Block

Used when `authentication_method = "mongo_auth"` (MongoDB authentication with global and database-specific roles).

- `global_builtin_roles` (List of String, Optional) - Global MongoDB builtin roles that apply across all databases. Examples: `["readWriteAnyDatabase", "dbAdminAnyDatabase", "userAdminAnyDatabase"]`
- `database_builtin_roles` (Map of List of String, Optional) - Database-specific MongoDB builtin roles. Keys are database names, values are lists of roles. Example: `{"mydb" = ["readWrite", "dbAdmin"]}`
- `database_custom_roles` (Map of List of String, Optional) - Database-specific custom roles. Keys are database names, values are lists of custom role names. Example: `{"mydb" = ["customAppRole"]}`

**Note**: At least one of the three attributes must be specified.

#### `sqlserver_auth_profile` Block

Used when `authentication_method = "sqlserver_auth"` (SQL Server authentication with global and database-specific roles).

- `global_builtin_roles` (List of String, Optional) - Global SQL Server builtin roles. Examples: `["db_datareader", "db_datawriter", "db_owner"]`
- `global_custom_roles` (List of String, Optional) - Global custom database roles.
- `database_builtin_roles` (Map of List of String, Optional) - Database-specific builtin roles. Keys are database names. Example: `{"master" = ["db_owner"], "production" = ["db_datareader"]}`
- `database_custom_roles` (Map of List of String, Optional) - Database-specific custom roles. Keys are database names.

**Note**: At least one of the four attributes must be specified.

#### `rds_iam_user_auth_profile` Block

Used when `authentication_method = "rds_iam_user_auth"` (AWS RDS IAM authentication).

- `db_user` (String, **Required**) - The AWS RDS IAM database user name. This user must be created in RDS with IAM authentication enabled.

**AWS RDS IAM Prerequisites**:
1. Database workspace must have `cloud_provider = "aws"` and `region` set
2. RDS instance must have IAM authentication enabled
3. IAM database user must be created: `CREATE USER iam_user IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';`
4. IAM role/user must have `rds-db:connect` permission

### Computed Attributes

- `id` (String) - Composite identifier in format `policy-id:database-id`. Used for resource tracking and import.
- `last_modified` (String) - Timestamp of last modification in RFC3339 format (e.g., `2025-10-27T10:30:00Z`).

## Authentication Method Reference

| Authentication Method | Database Types | Profile Block | Use Case |
|-----------------------|----------------|---------------|----------|
| `db_auth` | PostgreSQL, MySQL, MariaDB, SQL Server (non-Windows), MongoDB | `db_auth_profile` | Standard database user authentication with role assignment |
| `ldap_auth` | PostgreSQL, MySQL (with LDAP plugin) | `ldap_auth_profile` | LDAP/Active Directory group-based authentication |
| `oracle_auth` | Oracle | `oracle_auth_profile` | Oracle-specific authentication with DBA/SYSDBA/SYSOPER roles |
| `mongo_auth` | MongoDB | `mongo_auth_profile` | MongoDB role-based access control (RBAC) with global and database-specific roles |
| `sqlserver_auth` | SQL Server (Windows Auth) | `sqlserver_auth_profile` | SQL Server Windows authentication with global and database-specific roles |
| `rds_iam_user_auth` | AWS RDS (PostgreSQL, MySQL, MariaDB) | `rds_iam_user_auth_profile` | AWS RDS IAM authentication (passwordless, temporary credentials) |

## Import

Policy database assignments can be imported using the composite ID format: `policy-id:database-id`

### Finding the Import ID

You need both the policy ID and database workspace ID:

```bash
# Get policy ID from CyberArk UI or API
# Policy ID format: UUID (e.g., 12345678-1234-1234-1234-123456789012)

# Get database ID from Terraform state
terraform state show cyberarksia_database_workspace.production
# Look for the "id" attribute

# Construct import ID
IMPORT_ID="${POLICY_ID}:${DATABASE_ID}"
```

### Import Command

```bash
# Import existing assignment
terraform import cyberarksia_policy_database_assignment.example \
  "12345678-1234-1234-1234-123456789012:42"
```

### Import Example

```hcl
# 1. Create placeholder resource in configuration
resource "cyberarksia_policy_database_assignment" "imported_db" {
  policy_id              = "12345678-1234-1234-1234-123456789012"
  database_workspace_id  = "42"
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["placeholder"]  # Will be overwritten by import
  }
}

# 2. Run import command
# terraform import cyberarksia_policy_database_assignment.imported_db "12345678-1234-1234-1234-123456789012:42"

# 3. Run terraform plan to see current configuration
# terraform plan

# 4. Update configuration to match imported state
```

### Import Notes

- **Full State Reconstruction**: Import fetches the policy from the API and reconstructs the complete state including authentication method and profile configuration
- **Profile Detection**: The correct profile block is automatically detected based on the authentication method stored in the policy
- **Drift Detection**: After import, run `terraform plan` to verify the configuration matches the imported state
- **Takeover from UI**: You can import databases that were manually added via the CyberArk UI

## Behavior and Lifecycle

### Create

1. **Idempotency Check**: Checks if the database already exists in the policy
   - If exists: Adopts the existing configuration into Terraform state (takeover from UI)
   - If not exists: Adds the database to the policy
2. **Read-Modify-Write**: Fetches the full policy, adds the database, writes the entire policy back
3. **Preservation**: All other databases in the policy (UI-managed or other Terraform resources) are preserved

### Read (Drift Detection)

1. **Database Removal Detection**: If the database is not found in the policy, the resource is marked as deleted (drift)
2. **Platform Drift Detection**: If the database workspace's platform changes (e.g., ON-PREMISE → AWS), the resource is removed from state to force replacement
3. **State Refresh**: Updates authentication method and profile from current API state

### Update

1. **In-Place Updates**: Authentication method and profile configuration can be updated without replacement
2. **Profile Switching**: When changing authentication methods, the old profile is cleared and the new profile is set
3. **Preservation**: All other databases in the policy remain unchanged

### Delete

1. **Selective Removal**: Removes only the specified database from the policy
2. **Preservation**: All other databases (UI-managed or other Terraform resources) remain in the policy
3. **Idempotency**: Delete succeeds even if the database or policy is already deleted (graceful handling)

### Retry Logic

All policy update operations use exponential backoff retry (3 attempts, 500ms-30s delays) to handle transient API errors:
- Network timeouts
- Rate limiting (429)
- Temporary service unavailability (503)

**Note**: Persistent 409 Conflict errors indicate multi-workspace contention and require organizational fixes.

## Limitations and Known Issues

### Multi-Workspace Conflicts

**Status**: Documented Limitation (Not a Bug)

Managing the same policy from multiple Terraform workspaces will cause race conditions. This is an accepted limitation of the separate-resource pattern (same as AWS Security Group Rules).

**Mitigation**:
- Manage all assignments for a policy in a single workspace
- Use Terraform modules for code organization (not separate workspaces)
- Document workspace ownership in team runbooks

### ForceNew Attributes

The following attributes **force resource replacement** when changed:
- `policy_id` - Cannot move assignment to different policy (delete + recreate)
- `database_workspace_id` - Cannot change which database is assigned (delete + recreate)

To change these attributes, Terraform will:
1. Remove the database from the old policy/assignment
2. Create a new assignment with the new values

### Profile Configuration Requirements

- **Exactly one profile block** must be configured per resource
- The profile type must match the `authentication_method`
- Changing `authentication_method` requires updating the profile block in the same apply

**Example of valid update**:
```hcl
# Before
authentication_method = "db_auth"
db_auth_profile {
  roles = ["reader"]
}

# After (valid - both changed together)
authentication_method = "ldap_auth"
ldap_auth_profile {
  assign_groups = ["DBAdmins"]
}
```

## Troubleshooting

### Common Errors

#### "Policy not found"

**Error**: `Policy {id} not found. Ensure the policy exists.`

**Solution**:
- Verify the policy ID is correct (check CyberArk UI)
- Ensure the service account has permission to read policies
- Use the `cyberarksia_access_policy` data source to lookup policies by name

#### "Database workspace not found"

**Error**: `Database workspace {id} not found. Ensure the database exists.`

**Solution**:
- Ensure the `cyberarksia_database_workspace` resource exists and is created before the assignment
- Check that the database workspace ID is correct
- Verify the database hasn't been deleted outside Terraform

#### "Insufficient permissions to modify policy"

**Error**: `Insufficient permissions to modify policy {id}.`

**Solution**:
- Grant policy modification permissions to the service account
- Contact CyberArk administrator to update IAM permissions
- Verify the service account has the correct role assignments

#### "Policy update conflict (409)"

**Error**: `Policy update conflict (409) - The policy was modified concurrently`

**Solution**:
- **Check for multi-workspace conflicts**: Ensure no other Terraform workspace is managing this policy
- **Wait and retry**: Transient conflicts auto-retry (3 attempts)
- **Coordinate changes**: If persistent, coordinate with team members to avoid concurrent modifications

#### "Profile type does not match authentication method"

**Error**: `Authentication method 'db_auth' requires db_auth_profile block`

**Solution**:
- Ensure the profile block matches the authentication method
- When updating authentication methods, update the profile block in the same apply
- Use `terraform plan` to preview changes before applying

### Debug Logging

Enable debug logging to troubleshoot issues:

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform-debug.log
terraform apply
```

Look for log entries with `policy_database_assignment` operations:
- `"Operation starting"` - Operation initiation
- `"Operation succeeded"` - Successful completion
- `"Drift detected"` - Resource modified outside Terraform
- `"Platform drift detected"` - Database workspace platform changed

## See Also

- [cyberarksia_access_policy Data Source](../data-sources/access_policy.md) - Lookup policies by ID or name
- [cyberarksia_database_workspace Resource](./database_workspace.md) - Create database workspaces
- [cyberarksia_secret Resource](./secret.md) - Manage database secrets
- [CyberArk SIA Documentation](https://docs.cyberark.com/) - Official CyberArk documentation
