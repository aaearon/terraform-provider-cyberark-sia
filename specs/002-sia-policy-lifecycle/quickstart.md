# Quickstart: Database Policy Management

**Feature**: Database Policy Management - Modular Assignment Pattern
**Resources**: 3 resources for complete policy lifecycle management

---

## Prerequisites

1. ✅ **UAP Service Provisioned**: Your CyberArk tenant must have UAP (Unified Access Policy) service enabled
   - **How to verify**: Check with your CyberArk tenant administrator
   - **If not enabled**: Contact CyberArk support to provision UAP service for your tenant
   - **Note**: The provider will return DNS lookup errors if UAP is not available

2. ✅ **Service Account Credentials**: Obtain ISP credentials for authentication
   - Username: `your-service-account@cyberark.cloud.XXXX`
   - Client Secret: From CyberArk tenant admin

3. ✅ **Identity Directory**: Configure identity provider (AzureAD, LDAP, Okta, etc.) in SIA UI

4. ✅ **Database Workspaces**: Existing database workspaces to assign to policies

---

## Step 1: Create Database Policy

Create a policy with metadata and access conditions:

```hcl
resource "cyberarksia_database_policy" "db_admins" {
  name                       = "Database-Admins"
  description                = "Admin access to production databases"
  status                     = "active"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 8  # 8-hour sessions
    idle_time            = 30  # 30-minute idle timeout

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]  # Monday-Friday
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}
```

**Apply**:
```bash
terraform apply
```

**Result**: Policy created with ID, no principals or databases assigned yet.

---

## Step 2: Assign Principals

Assign users, groups, and roles to the policy:

### USER Principal (with directory)
```hcl
resource "cyberarksia_database_policy_principal_assignment" "alice" {
  policy_id             = cyberarksia_database_policy.db_admins.id
  principal_id          = "alice@example.com"
  principal_name        = "Alice Smith"
  principal_type        = "USER"
  source_directory_name = "AzureAD"
  source_directory_id   = "dir-12345"
}
```

### GROUP Principal
```hcl
resource "cyberarksia_database_policy_principal_assignment" "db_admins_group" {
  policy_id             = cyberarksia_database_policy.db_admins.id
  principal_id          = "db-admins"
  principal_name        = "DB Administrators"
  principal_type        = "GROUP"
  source_directory_name = "AzureAD"
  source_directory_id   = "dir-12345"
}
```

### ROLE Principal (no directory required)
```hcl
resource "cyberarksia_database_policy_principal_assignment" "admin_role" {
  policy_id      = cyberarksia_database_policy.db_admins.id
  principal_id   = "admin"
  principal_name = "Admin Role"
  principal_type = "ROLE"
}
```

---

## Step 3: Assign Databases

Assign database workspaces to the policy:

```hcl
resource "cyberarksia_database_policy_assignment" "prod_postgres" {
  policy_id              = cyberarksia_database_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.prod_pg.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["db_admin", "db_writer"]
  }
}

resource "cyberarksia_database_policy_assignment" "prod_mysql" {
  policy_id              = cyberarksia_database_policy.db_admins.id
  database_workspace_id  = cyberarksia_database_workspace.prod_mysql.id
  authentication_method  = "ldap_auth"

  ldap_auth_profile {
    assign_groups = ["dbadmins", "developers"]
  }
}
```

---

## Step 4: Import Existing Policies

Import policies created in SIA UI. **Important**: Import order matters to ensure proper state dependencies.

### Import Workflow (Recommended Order)

**Step 1: Import the Policy First**
```bash
# Find policy ID using data source or SIA UI
terraform import cyberarksia_database_policy.existing_policy <policy-id>

# Verify import
terraform plan
# Should show no changes if configuration matches
```

**Step 2: Import Principal Assignments**
```bash
# Import each principal assignment (3-part composite ID)
terraform import cyberarksia_database_policy_principal_assignment.alice \
  "<policy-id>:<principal-id>:USER"

terraform import cyberarksia_database_policy_principal_assignment.admins \
  "<policy-id>:<group-id>:GROUP"

terraform import cyberarksia_database_policy_principal_assignment.db_role \
  "<policy-id>:<role-name>:ROLE"
```

**Step 3: Import Database Assignments**
```bash
# Import each database assignment (2-part composite ID)
terraform import cyberarksia_database_policy_assignment.prod_db \
  "<policy-id>:<database-workspace-id>"
```

### Finding Import IDs

**Policy ID**: Use the `cyberarksia_access_policy` data source:
```hcl
data "cyberarksia_access_policy" "existing" {
  name = "Production Database Access"
}

output "policy_id" {
  value = data.cyberarksia_access_policy.existing.id
}
```

**Principal ID & Type**: Check SIA UI → Policy → "Assigned To" tab
- USER: email address (e.g., `alice@example.com`)
- GROUP: group name (e.g., `db-admins`)
- ROLE: role name (e.g., `DatabaseAdministrator`)

**Database Workspace ID**: From Terraform state:
```bash
terraform state show cyberarksia_database_workspace.production
# Look for "id" attribute
```

### Import Validation Checklist

After importing all resources:
- [ ] Run `terraform plan` - should show no changes
- [ ] All policy attributes match configuration
- [ ] All principals appear in state
- [ ] All database assignments appear in state
- [ ] Computed fields populated (created_by, updated_on, last_modified)

---

## Multi-Team Workflow

**Scenario**: Security team manages policies/principals, app teams manage databases

### Security Team Workspace

```hcl
# security-team/main.tf
resource "cyberarksia_database_policy" "prod_access" {
  name   = "Production-Access"
  status = "active"

  conditions {
    max_session_duration = 4
    idle_time            = 15
  }
}

# Assign principals
resource "cyberarksia_database_policy_principal_assignment" "security_team" {
  policy_id             = cyberarksia_database_policy.prod_access.id
  principal_id          = "security-team"
  principal_name        = "Security Team"
  principal_type        = "GROUP"
  source_directory_name = "AzureAD"
  source_directory_id   = "dir-12345"
}
```

### App Team A Workspace

```hcl
# app-team-a/main.tf
# Reference policy created by security team
data "cyberarksia_access_policy" "prod_access" {
  name = "Production-Access"
}

# Assign Team A databases
resource "cyberarksia_database_policy_assignment" "team_a_db" {
  policy_id              = data.cyberarksia_access_policy.prod_access.id
  database_workspace_id  = cyberarksia_database_workspace.team_a_db.id
  authentication_method  = "db_auth"

  db_auth_profile {
    roles = ["app_reader"]
  }
}
```

### App Team B Workspace

```hcl
# app-team-b/main.tf
# Reference same policy
data "cyberarksia_access_policy" "prod_access" {
  name = "Production-Access"
}

# Assign Team B databases (independent of Team A)
resource "cyberarksia_database_policy_assignment" "team_b_db" {
  policy_id              = data.cyberarksia_access_policy.prod_access.id
  database_workspace_id  = cyberarksia_database_workspace.team_b_db.id
  authentication_method  = "ldap_auth"

  ldap_auth_profile {
    assign_groups = ["team_b_devs"]
  }
}
```

**Benefits**:
- ✅ Security team controls policy conditions and principals
- ✅ App teams independently manage their database assignments
- ✅ No coordination required between app teams
- ✅ Policy changes affect all teams automatically

---

## Common Patterns

### 1. Suspend Policy Temporarily
```hcl
resource "cyberarksia_database_policy" "maintenance" {
  name   = "Maintenance-Access"
  status = "suspended"  # Disable without deleting
  # ...
}
```

### 2. Time-Limited Policy
```hcl
resource "cyberarksia_database_policy" "contractor" {
  name = "Contractor-Access"

  time_frame {
    from_time = "2025-01-01T00:00:00Z"
    to_time   = "2025-12-31T23:59:59Z"
  }
  # ...
}
```

### 3. After-Hours Access
```hcl
resource "cyberarksia_database_policy" "on_call" {
  name = "On-Call-Access"

  conditions {
    max_session_duration = 12

    access_window {
      days_of_the_week = [0, 6]  # Weekends only
      from_hour        = "00:00"
      to_hour          = "23:59"
    }
  }
}
```

---

## Troubleshooting

### UAP Service Not Provisioned
**Error**: DNS lookup failed for UAP endpoint

**Solution**: Verify UAP service is provisioned on your tenant. Contact CyberArk support if not available.

### Duplicate Policy Name
**Error**: Policy with name already exists

**Solution**: Policy names must be unique per tenant. Choose a different name or import the existing policy.

### Principal Directory Not Found
**Error**: Source directory not found

**Solution**: Configure identity directory in SIA UI before assigning principals. Verify `source_directory_id` is correct.

### Concurrent Modification
**Warning**: Last-write-wins behavior

**Solution**: Coordinate Terraform workspaces. Manage all assignments for a policy in a single workspace when possible.

---

## Next Steps

1. **Review CRUD Testing**: See `examples/testing/TESTING-GUIDE.md` for validation workflow
2. **Explore Examples**: Check `examples/resources/` for all 6 authentication methods
3. **Read Documentation**: See `docs/resources/` for complete attribute references
4. **Generate Tasks**: Run `/speckit.tasks` to create implementation task list

**Estimated Implementation**: 2-3 weeks for full feature completion
