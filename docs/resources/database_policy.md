---
page_title: "cyberarksia_database_policy Resource - terraform-provider-cyberark-sia"
subcategory: ""
description: |-
  Manages a CyberArk SIA database access policy including metadata and access conditions.
---

# cyberarksia_database_policy (Resource)

Manages a CyberArk SIA database access policy including metadata and access conditions. This resource manages policy-level configuration only.

**Modular Assignment Pattern**: Use `cyberarksia_database_policy_principal_assignment` to assign principals (users/groups/roles) and `cyberarksia_database_policy_assignment` to assign database workspaces to the policy.

**Team Workflow**: This pattern enables distributed workflows where security teams manage policies and principals, while application teams independently manage their database assignments.

## Example Usage

### Basic Policy

```terraform
resource "cyberarksia_database_policy" "basic" {
  name                       = "Basic-Database-Policy"
  status                     = "Active"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 8  # 8 hours
    idle_time            = 30 # 30 minutes
  }
}
```

### Policy with Access Window

```terraform
resource "cyberarksia_database_policy" "business_hours" {
  name                       = "Weekday-Business-Hours"
  description                = "Database access limited to business hours"
  status                     = "Active"
  delegation_classification  = "Unrestricted"
  time_zone                  = "America/New_York"

  conditions {
    max_session_duration = 4  # 4 hours
    idle_time            = 15 # 15 minutes

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}
```

### Policy with Tags and Time Frame

```terraform
resource "cyberarksia_database_policy" "project" {
  name                       = "Q1-Migration-Project"
  description                = "Temporary access for Q1 database migration"
  status                     = "Active"
  delegation_classification  = "Restricted"
  time_zone                  = "GMT"

  time_frame {
    from_time = "2024-01-01T00:00:00Z"
    to_time   = "2024-03-31T23:59:59Z"
  }

  policy_tags = [
    "project:migration",
    "temporary:true",
    "team:platform",
  ]

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}
```

## Schema

### Required

- `name` (String) Policy name (1-200 characters, unique per tenant). **ForceNew**: Changing this creates a new policy.
- `status` (String) Policy status. Valid values: `Active` (enabled), `Suspended` (disabled). **Note**: `Expired`, `Validating`, and `Error` are server-managed statuses and cannot be set by users.
- `conditions` (Block, Required) Policy access conditions. See [Conditions](#conditions) below.

### Optional

- `description` (String) Policy description (max 200 characters).
- `delegation_classification` (String) Delegation classification. Valid values: `Restricted`, `Unrestricted`. Default: `Unrestricted`.
- `time_zone` (String) Timezone for access window conditions (max 50 characters). Supports IANA timezone names (e.g., `America/New_York`) or GMT offsets (e.g., `GMT+05:00`). Default: `GMT`.
- `policy_tags` (List of String) List of tags for policy organization (max 20 tags).
- `time_frame` (Block, Optional) Policy validity period. If not specified, policy is valid indefinitely. See [Time Frame](#time_frame) below.

### Read-Only (Computed)

- `id` (String) Policy identifier (same as `policy_id`).
- `policy_id` (String) Unique policy identifier (UUID, API-generated).
- `last_modified` (String) Timestamp of the last modification to the policy.
- `created_by` (Block, Computed) User who created the policy. See [Change Info](#change_info) below.
- `updated_on` (Block, Computed) Last user who updated the policy. See [Change Info](#change_info) below.

<a id="conditions"></a>
### Conditions

The `conditions` block supports:

- `max_session_duration` (Number, Required) Maximum session duration in hours (1-24).
- `idle_time` (Number, Optional) Session idle timeout in minutes (1-120). Default: 10.
- `access_window` (Block, Optional) Time-based access restrictions. See [Access Window](#access_window) below.

<a id="access_window"></a>
### Access Window

The `access_window` block within `conditions` supports:

- `days_of_the_week` (List of Number, Required) Days access is allowed (0=Sunday through 6=Saturday). Example: `[1, 2, 3, 4, 5]` for weekdays.
- `from_hour` (String, Required) Start time in HH:MM format (e.g., `09:00`).
- `to_hour` (String, Required) End time in HH:MM format (e.g., `17:00`).

<a id="time_frame"></a>
### Time Frame

The `time_frame` block supports:

- `from_time` (String, Required) Start time (ISO 8601 format, e.g., `2024-01-01T00:00:00Z`).
- `to_time` (String, Required) End time (ISO 8601 format, e.g., `2024-12-31T23:59:59Z`).

<a id="change_info"></a>
### Change Info

The `created_by` and `updated_on` blocks contain:

- `user` (String) Username.
- `timestamp` (String) Timestamp (ISO 8601 format).

## Import

Policies can be imported using the policy ID:

```shell
terraform import cyberarksia_database_policy.example 12345678-1234-1234-1234-123456789012
```

**Finding Policy IDs**: Use the `cyberarksia_access_policy` data source to lookup policies by name:

```terraform
data "cyberarksia_access_policy" "existing" {
  name = "Existing-Policy-Name"
}

# Import using data source
resource "cyberarksia_database_policy" "imported" {
  # ... configuration
}

# Run: terraform import cyberarksia_database_policy.imported ${data.cyberarksia_access_policy.existing.id}
```

## Constraints and Validation

### Policy Name
- **Length**: 1-200 characters
- **Uniqueness**: Must be unique per tenant
- **ForceNew**: Changing the name creates a new policy (existing policy is deleted)

### Status Values
- **Valid**: `Active`, `Suspended` (user-controllable)
- **Invalid**: `Expired`, `Validating`, `Error` (server-managed, cannot be set)

### Session Limits
- **max_session_duration**: 1-24 hours (validated by provider)
- **idle_time**: 1-120 minutes (validated by provider)
- **Default idle_time**: 10 minutes if not specified

### Access Window
- **days_of_the_week**: Values 0-6 (0=Sunday, 6=Saturday)
- **from_hour/to_hour**: HH:MM format (00:00-23:59)
- **Validation**: Provider validates format; API enforces from_hour < to_hour

### Tags
- **Maximum**: 20 tags per policy (validated by API)
- **Format**: Free-form strings (common pattern: `key:value`)

### Time Zone
- **Formats**: IANA timezone names (`America/New_York`), GMT offsets (`GMT+05:00`)
- **Length**: Max 50 characters
- **Default**: `GMT`

## Relationships

### With Principal Assignments

Principals (users/groups/roles) are assigned to policies using the `cyberarksia_database_policy_principal_assignment` resource:

```terraform
resource "cyberarksia_database_policy" "admins" {
  name   = "Database-Admins"
  status = "Active"

  conditions {
    max_session_duration = 8
  }
}

resource "cyberarksia_database_policy_principal_assignment" "alice" {
  policy_id               = cyberarksia_database_policy.admins.policy_id
  principal_id            = "alice@example.com"
  principal_type          = "USER"
  principal_name          = "Alice Smith"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345"
}
```

### With Database Assignments

Database workspaces are assigned to policies using the `cyberarksia_database_policy_assignment` resource:

```terraform
resource "cyberarksia_database_policy_assignment" "prod_db" {
  policy_id              = cyberarksia_database_policy.admins.policy_id
  database_workspace_id  = cyberarksia_database_workspace.prod.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["db_reader", "db_writer"]
  }
}
```

## Behavior Notes

### Policy Updates (In-Place vs ForceNew)

**In-Place Updates** (no replacement):
- `description`
- `status` (`Active` ↔ `Suspended`)
- `delegation_classification`
- `time_zone`
- `policy_tags`
- `time_frame`
- All `conditions` attributes

**ForceNew** (requires replacement):
- `name` - Changing the policy name creates a new policy

### Read-Modify-Write Pattern

When updating policy metadata, the provider uses a read-modify-write pattern to preserve principals and database assignments:

1. **Fetch**: GET policy by ID (includes all principals and targets)
2. **Modify**: Update only metadata/conditions fields
3. **Preserve**: Keep all existing principals and targets unchanged
4. **Write**: PUT updated policy back to API

This ensures updates to policy metadata don't affect existing assignments managed by other resources or the SIA UI.

### Cascade Delete

When a policy is deleted:
- **API Behavior**: The API automatically removes all principals and database assignments
- **Terraform Behavior**: Assignment resources in state will show as "deleted" on next refresh
- **Best Practice**: Delete assignment resources first, then delete policy

### Concurrent Modifications

**Known Limitation**: The SIA API uses last-write-wins behavior with no optimistic locking. Managing the same policy from multiple Terraform workspaces can cause conflicts.

**Mitigation**:
- Manage all resources for a policy within a single Terraform workspace
- Use modules to organize resources
- Coordinate changes across teams

This is the same limitation as AWS resources like `aws_security_group_rule` and `google_project_iam_member`.

## Common Patterns

### Suspended Policy Toggle

Use a variable to enable/disable policies without deletion:

```terraform
variable "maintenance_mode" {
  type    = bool
  default = false
}

resource "cyberarksia_database_policy" "app" {
  name   = "Application-Database-Access"
  status = var.maintenance_mode ? "Suspended" : "Active"

  conditions {
    max_session_duration = 4
  }
}
```

### Environment-Specific Policies

Create policies per environment with dynamic naming:

```terraform
variable "environment" {
  type = string
}

resource "cyberarksia_database_policy" "env" {
  name                       = "${var.environment}-database-policy"
  status                     = "Active"
  delegation_classification  = var.environment == "production" ? "Restricted" : "Unrestricted"

  policy_tags = [
    "environment:${var.environment}",
    "managed-by:terraform",
  ]

  conditions {
    max_session_duration = var.environment == "production" ? 4 : 8
    idle_time            = var.environment == "production" ? 15 : 30
  }
}
```

### 24/7 Access Policy

Configure policies for always-available access:

```terraform
resource "cyberarksia_database_policy" "oncall" {
  name        = "On-Call-24x7-Access"
  description = "24/7 access for on-call engineers"
  status      = "Active"

  policy_tags = ["access-type:24x7"]

  conditions {
    max_session_duration = 12

    access_window {
      days_of_the_week = [0, 1, 2, 3, 4, 5, 6] # All days
      from_hour        = "00:00"
      to_hour          = "23:59"
    }
  }
}
```

## Troubleshooting

### Policy Not Found After Creation

**Symptom**: `terraform plan` shows policy needs to be recreated after successful `terraform apply`.

**Cause**: UAP service not provisioned on tenant.

**Resolution**:
1. Verify UAP service availability:
   ```bash
   curl -s "https://platform-discovery.cyberark.cloud/api/v2/services/subdomain/{tenant}" | jq '.uap'
   ```
2. If UAP service is missing, contact CyberArk support to provision it

### "Policy Already Exists" Error

**Symptom**: Error during `terraform apply`: `Policy with name "..." already exists`

**Cause**: Policy name must be unique per tenant.

**Resolution**:
- Choose a different policy name, OR
- Import the existing policy using `terraform import`

### Status Update Failing

**Symptom**: Cannot change policy status to `Expired`, `Validating`, or `Error`.

**Cause**: These are server-managed statuses.

**Resolution**: Only use `Active` or `Suspended` for the `status` attribute.
