# Complete Principal Assignment Example
#
# This comprehensive example demonstrates:
# - All three principal types (USER, GROUP, ROLE)
# - Multiple directory sources (Azure AD, LDAP)
# - Modular assignment pattern
# - Import scenarios
# - Team workflow separation

# ============================================================================
# POLICY DEFINITION (managed by security team)
# ============================================================================

resource "cyberarksia_database_policy" "enterprise_database" {
  name                       = "Enterprise-Database-Access-Policy"
  status                     = "active"
  delegation_classification  = "Restricted"
  description                = "Comprehensive policy for enterprise database access"
  time_zone                  = "America/New_York"

  # Policy valid for Q1 2024
  time_frame {
    from_time = "2024-01-01T00:00:00Z"
    to_time   = "2024-03-31T23:59:59Z"
  }

  # Strict access conditions
  conditions {
    max_session_duration = 4 # 4 hours

    # Business hours only
    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday
      from_hour        = "08:00"
      to_hour          = "18:00"
    }

    idle_time = 15 # 15 minutes
  }

  policy_tags = [
    "environment:production",
    "compliance:sox",
    "team:platform",
    "quarter:q1-2024",
  ]
}

# ============================================================================
# USER PRINCIPALS - Azure AD (managed by platform team)
# ============================================================================

locals {
  azuread_users = {
    alice = {
      email = "alice.smith@example.com"
      name  = "Alice Smith - Senior DBA"
    }
    bob = {
      email = "bob.johnson@example.com"
      name  = "Bob Johnson - Platform Engineer"
    }
    carol = {
      email = "carol.davis@example.com"
      name  = "Carol Davis - DevOps Lead"
    }
  }

  azuread_directory_id = "12345678-1234-1234-1234-123456789012"
}

resource "cyberarksia_database_policy_principal_assignment" "azuread_users" {
  for_each = local.azuread_users

  policy_id               = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id            = each.value.email
  principal_type          = "USER"
  principal_name          = each.value.name
  source_directory_name   = "AzureAD-Production"
  source_directory_id     = local.azuread_directory_id
}

# ============================================================================
# USER PRINCIPALS - LDAP (managed by infrastructure team)
# ============================================================================

locals {
  ldap_users = {
    dave = {
      username = "dave.martinez"
      name     = "Dave Martinez - System Administrator"
    }
    eve = {
      username = "eve.wilson"
      name     = "Eve Wilson - Database Architect"
    }
  }
}

resource "cyberarksia_database_policy_principal_assignment" "ldap_users" {
  for_each = local.ldap_users

  policy_id               = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id            = each.value.username
  principal_type          = "USER"
  principal_name          = each.value.name
  source_directory_name   = "CorporateLDAP"
  source_directory_id     = "ldap-prod-001"
}

# ============================================================================
# GROUP PRINCIPALS (managed by team leads)
# ============================================================================

# DBA Team
resource "cyberarksia_database_policy_principal_assignment" "dba_group" {
  policy_id               = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id            = "dba-team@example.com"
  principal_type          = "GROUP"
  principal_name          = "Database Administration Team"
  source_directory_name   = "AzureAD-Production"
  source_directory_id     = local.azuread_directory_id
}

# Backend Engineers
resource "cyberarksia_database_policy_principal_assignment" "backend_group" {
  policy_id               = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id            = "backend-engineers@example.com"
  principal_type          = "GROUP"
  principal_name          = "Backend Engineering Team"
  source_directory_name   = "AzureAD-Production"
  source_directory_id     = local.azuread_directory_id
}

# Data Analytics Team (LDAP group)
resource "cyberarksia_database_policy_principal_assignment" "analytics_group" {
  policy_id               = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id            = "data-analytics"
  principal_type          = "GROUP"
  principal_name          = "Data Analytics Team"
  source_directory_name   = "CorporateLDAP"
  source_directory_id     = "ldap-prod-001"
}

# ============================================================================
# ROLE PRINCIPALS (managed by security team)
# ============================================================================

# Database Administrator role
resource "cyberarksia_database_policy_principal_assignment" "admin_role" {
  policy_id      = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id   = "database-administrator"
  principal_type = "ROLE"
  principal_name = "Database Administrator Role"
}

# Emergency Access role
resource "cyberarksia_database_policy_principal_assignment" "emergency_role" {
  policy_id      = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id   = "emergency-database-access"
  principal_type = "ROLE"
  principal_name = "Emergency Database Access Role"
}

# Automation role for CI/CD pipelines
resource "cyberarksia_database_policy_principal_assignment" "automation_role" {
  policy_id      = cyberarksia_database_policy.enterprise_database.policy_id
  principal_id   = "database-automation"
  principal_type = "ROLE"
  principal_name = "Database Automation Role"
}

# ============================================================================
# OUTPUTS
# ============================================================================

output "policy_info" {
  description = "Policy information"
  value = {
    policy_id   = cyberarksia_database_policy.enterprise_database.policy_id
    policy_name = cyberarksia_database_policy.enterprise_database.name
    status      = cyberarksia_database_policy.enterprise_database.status
  }
}

output "principal_assignments" {
  description = "All principal assignment composite IDs grouped by type"
  value = {
    azuread_users = { for k, v in cyberarksia_database_policy_principal_assignment.azuread_users : k => v.id }
    ldap_users    = { for k, v in cyberarksia_database_policy_principal_assignment.ldap_users : k => v.id }
    groups = {
      dba       = cyberarksia_database_policy_principal_assignment.dba_group.id
      backend   = cyberarksia_database_policy_principal_assignment.backend_group.id
      analytics = cyberarksia_database_policy_principal_assignment.analytics_group.id
    }
    roles = {
      admin      = cyberarksia_database_policy_principal_assignment.admin_role.id
      emergency  = cyberarksia_database_policy_principal_assignment.emergency_role.id
      automation = cyberarksia_database_policy_principal_assignment.automation_role.id
    }
  }
}

output "principal_summary" {
  description = "Summary of principals by type and source"
  value = {
    azuread_users = length(cyberarksia_database_policy_principal_assignment.azuread_users)
    ldap_users    = length(cyberarksia_database_policy_principal_assignment.ldap_users)
    groups        = 3
    roles         = 3
    total         = length(cyberarksia_database_policy_principal_assignment.azuread_users) + length(cyberarksia_database_policy_principal_assignment.ldap_users) + 6
  }
}

output "import_commands" {
  description = "Example import commands for existing principals"
  value = {
    user_example = "terraform import cyberarksia_database_policy_principal_assignment.azuread_users[\"alice\"] \"${cyberarksia_database_policy.enterprise_database.policy_id}:alice.smith@example.com:USER\""
    group_example = "terraform import cyberarksia_database_policy_principal_assignment.dba_group \"${cyberarksia_database_policy.enterprise_database.policy_id}:dba-team@example.com:GROUP\""
    role_example = "terraform import cyberarksia_database_policy_principal_assignment.admin_role \"${cyberarksia_database_policy.enterprise_database.policy_id}:database-administrator:ROLE\""
  }
}

# ============================================================================
# NOTES
# ============================================================================

# This example demonstrates:
# 1. ✅ Modular assignment pattern (each principal managed independently)
# 2. ✅ Team workflow separation (different teams manage their principals)
# 3. ✅ Multiple directory sources (Azure AD + LDAP)
# 4. ✅ All three principal types (USER, GROUP, ROLE)
# 5. ✅ for_each for managing multiple principals efficiently
# 6. ✅ Composite ID format for imports
# 7. ✅ Comprehensive outputs for visibility

# Team Responsibilities:
# - Security Team: Policy definition + ROLE principals
# - Platform Team: Azure AD USER principals
# - Infrastructure Team: LDAP USER principals
# - Team Leads: GROUP principals for their teams
