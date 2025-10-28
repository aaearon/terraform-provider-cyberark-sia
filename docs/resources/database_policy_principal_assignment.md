---
page_title: "cyberarksia_database_policy_principal_assignment Resource - terraform-provider-cyberark-sia"
subcategory: ""
description: |-
  Manages assignment of principals (users, groups, roles) to CyberArk SIA database access policies.
---

# cyberarksia_database_policy_principal_assignment (Resource)

Manages the assignment of a principal (user, group, or role) to a database access policy. Each resource represents a single principal assignment.

**Modular Assignment Pattern**: This resource follows the AWS Security Group Rule pattern, where each principal assignment is managed independently. This enables distributed team workflows and prevents conflicts when multiple teams manage different aspects of the same policy.

**Team Workflow**: Platform teams can manage principal assignments while application teams independently manage database assignments via `cyberarksia_database_policy_assignment`.

## Composite ID Format

**Format**: `policy-id:principal-id:principal-type` (3-part)

**Why 3 parts?**: Principal IDs can be duplicated across different types (e.g., a user named "admin" and a role named "admin" can both exist). The principal type ensures unique identification.

**Example**: `a1b2c3d4-e5f6-7890-abcd-ef1234567890:alice@example.com:USER`

## Example Usage

### USER Principal with Azure AD

```terraform
resource "cyberarksia_database_policy" "db_admins" {
  name                       = "Database-Administrators"
  status                     = "active"
  delegation_classification  = "Restricted"

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}

resource "cyberarksia_database_policy_principal_assignment" "alice" {
  policy_id               = cyberarksia_database_policy.db_admins.policy_id
  principal_id            = "alice@example.com"
  principal_type          = "USER"
  principal_name          = "Alice Smith"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}
```

### USER Principal with LDAP

```terraform
resource "cyberarksia_database_policy_principal_assignment" "bob_ldap" {
  policy_id               = cyberarksia_database_policy.db_admins.policy_id
  principal_id            = "bob"
  principal_type          = "USER"
  principal_name          = "Bob Johnson"
  source_directory_name   = "CorporateLDAP"
  source_directory_id     = "ldap-directory-001"
}
```

### GROUP Principal

```terraform
resource "cyberarksia_database_policy_principal_assignment" "dba_group" {
  policy_id               = cyberarksia_database_policy.db_admins.policy_id
  principal_id            = "dba-team@example.com"
  principal_type          = "GROUP"
  principal_name          = "DBA Team"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}
```

### ROLE Principal (No Directory Required)

```terraform
resource "cyberarksia_database_policy_principal_assignment" "admin_role" {
  policy_id      = cyberarksia_database_policy.db_admins.policy_id
  principal_id   = "database-admin"
  principal_type = "ROLE"
  principal_name = "Database Administrator Role"

  # Note: source_directory_name and source_directory_id are optional for ROLE type
}
```

### Multiple Principals on Same Policy

```terraform
# Security team manages policy
resource "cyberarksia_database_policy" "production_access" {
  name                       = "Production-Database-Access"
  status                     = "active"
  delegation_classification  = "Restricted"

  conditions {
    max_session_duration = 4
    idle_time            = 15

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}

# Multiple teams can independently add their principals
resource "cyberarksia_database_policy_principal_assignment" "team_a_users" {
  for_each = toset([
    "user1@example.com",
    "user2@example.com",
  ])

  policy_id               = cyberarksia_database_policy.production_access.policy_id
  principal_id            = each.value
  principal_type          = "USER"
  principal_name          = each.value
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

resource "cyberarksia_database_policy_principal_assignment" "team_b_group" {
  policy_id               = cyberarksia_database_policy.production_access.policy_id
  principal_id            = "team-b@example.com"
  principal_type          = "GROUP"
  principal_name          = "Team B Members"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}
```

## Schema

### Required

- `policy_id` (String) Policy ID to assign principal to. **ForceNew**: Changing this creates a new assignment.
- `principal_id` (String) Principal identifier (max 40 characters). Format depends on principal type:
  - USER: Email address or username
  - GROUP: Group identifier
  - ROLE: Role name
  **ForceNew**: Changing this creates a new assignment.
- `principal_type` (String) Principal type. Valid values: `USER`, `GROUP`, `ROLE`. **ForceNew**: Changing this creates a new assignment.
- `principal_name` (String) Display name for the principal (max 512 characters).

### Optional (Conditional)

- `source_directory_name` (String) Source directory name (max 50 characters). **Required for USER and GROUP**, optional for ROLE.
- `source_directory_id` (String) Source directory identifier. **Required for USER and GROUP**, optional for ROLE.

### Read-Only (Computed)

- `id` (String) Composite identifier in format `policy-id:principal-id:principal-type`.
- `last_modified` (String) Timestamp of last modification (ISO 8601 format).

## Import

Import using the 3-part composite ID format:

```bash
# Format: policy-id:principal-id:principal-type
terraform import cyberarksia_database_policy_principal_assignment.alice \
  "a1b2c3d4-e5f6-7890-abcd-ef1234567890:alice@example.com:USER"
```

**Important**: The principal type must exactly match the type in CyberArk SIA (case-sensitive: `USER`, `GROUP`, or `ROLE`).

### Import Validation

The import command will validate:
- ✅ Composite ID format (3 parts separated by colons)
- ✅ Policy exists
- ✅ Principal exists on the policy with matching type
- ✅ Principal type is valid (USER, GROUP, or ROLE)

**Error Examples**:

```bash
# ❌ Wrong format (2 parts)
terraform import cyberarksia_database_policy_principal_assignment.alice \
  "policy-id:alice@example.com"
# Error: Invalid composite ID format: expected 'policy-id:principal-id:principal-type', got 'policy-id:alice@example.com'

# ❌ Invalid principal type
terraform import cyberarksia_database_policy_principal_assignment.alice \
  "policy-id:alice@example.com:user"
# Error: Invalid principal type 'user', must be USER, GROUP, or ROLE

# ❌ Principal not found on policy
terraform import cyberarksia_database_policy_principal_assignment.bob \
  "policy-id:bob@example.com:USER"
# Error: Principal bob@example.com (type: USER) not found on policy policy-id
```

## Read-Modify-Write Pattern

**How it works**: This resource uses a read-modify-write pattern to preserve principals managed by:
- Other Terraform workspaces
- The CyberArk SIA UI
- Other automation tools

**Algorithm**:
1. **Fetch** the full policy including all principals
2. **Modify** only the single principal managed by this resource
3. **Write** back the complete policy with all principals intact

**Concurrent Modification**: If multiple Terraform workspaces modify the same policy simultaneously, the last write wins. Coordinate workspace execution to avoid conflicts.

## Validation Rules

### Principal Type Validation

- **USER and GROUP**: Must provide both `source_directory_name` and `source_directory_id`
- **ROLE**: `source_directory_name` and `source_directory_id` are optional

**Example Error**:

```terraform
resource "cyberarksia_database_policy_principal_assignment" "invalid" {
  policy_id      = cyberarksia_database_policy.test.policy_id
  principal_id   = "alice@example.com"
  principal_type = "USER"
  principal_name = "Alice"
  # ❌ Missing required fields for USER type
}
```

**Error Message**: `source_directory_name and source_directory_id are required for USER and GROUP principal types`

### Duplicate Prevention

Attempting to assign the same principal (ID + type combination) to a policy multiple times will fail:

```terraform
resource "cyberarksia_database_policy_principal_assignment" "alice1" {
  policy_id      = cyberarksia_database_policy.test.policy_id
  principal_id   = "alice@example.com"
  principal_type = "USER"
  # ... other fields
}

resource "cyberarksia_database_policy_principal_assignment" "alice2" {
  policy_id      = cyberarksia_database_policy.test.policy_id
  principal_id   = "alice@example.com"  # ❌ Same ID
  principal_type = "USER"                # ❌ Same type
  # ... other fields
}
```

**Error Message**: `Principal alice@example.com (type: USER) is already assigned to policy <policy-id>`

## Known Limitations

### 1. Multi-Workspace Conflicts

Managing the same policy from multiple Terraform workspaces can cause race conditions (same limitation as `aws_security_group_rule`, `google_project_iam_member`).

**Mitigation**: Manage all principal assignments for a policy in a single workspace. Use modules for organization:

```terraform
module "db_admin_principals" {
  source    = "./modules/principal-assignments"
  policy_id = cyberarksia_database_policy.db_admins.policy_id

  users = [
    { id = "alice@example.com", name = "Alice", directory = "AzureAD", directory_id = "12345" },
    { id = "bob@example.com",   name = "Bob",   directory = "AzureAD", directory_id = "12345" },
  ]
}
```

### 2. External Directory Validation

The provider does NOT validate that source directories exist or that principals are valid in those directories. Invalid references will be caught by the CyberArk SIA API during assignment.

**Why**: Source directories are managed outside Terraform (in identity providers like Azure AD, LDAP).

### 3. Principal Name Pattern

Principal names must match the pattern `^[\w.+\-@#]+$` (alphanumeric, dots, plus, hyphen, at-sign, hash). This excludes spaces and Unicode characters.

**Valid Examples**:
- ✅ `alice@example.com`
- ✅ `bob.johnson`
- ✅ `admin-role#1`

**Invalid Examples**:
- ❌ `Alice Smith` (space)
- ❌ `user@domain.com (Admin)` (parentheses)
- ❌ `José García` (Unicode)

## Edge Cases

### Principal Already Assigned

If a principal is already assigned to a policy (via UI or other means), creating a Terraform resource for it will fail with a clear error. Use `terraform import` to bring it under Terraform management.

### Policy Deleted Outside Terraform

If the policy is deleted outside Terraform, the next `terraform plan` will show the principal assignment as needing recreation. The creation will fail because the policy doesn't exist. Delete the principal assignment resource or recreate the policy first.

### Source Directory Deleted

If the source directory is deleted in the identity provider (Azure AD, LDAP), the principal assignment will remain in the policy but may not function correctly. CyberArk SIA will show warnings in the UI. Remove and recreate the assignment with a valid directory.

## See Also

- `cyberarksia_database_policy` - Manage policy metadata and conditions
- `cyberarksia_database_policy_assignment` - Assign database workspaces to policies
- `cyberarksia_access_policy` (Data Source) - Lookup existing policies by name or ID
