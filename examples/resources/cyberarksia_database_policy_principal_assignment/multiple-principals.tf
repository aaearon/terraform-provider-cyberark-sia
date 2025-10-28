# Multiple Principals on Same Policy
#
# This example demonstrates managing multiple principals (mixed types) on a single policy.
# Shows distributed team workflow where different teams manage their own principal assignments.

resource "cyberarksia_database_policy" "production_database" {
  name                       = "Production-Database-Access"
  status                     = "active"
  delegation_classification  = "Restricted"
  description                = "Production database access with strict conditions"
  time_zone                  = "America/New_York"

  conditions {
    max_session_duration = 4 # 4 hours max

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday
      from_hour        = "09:00"
      to_hour          = "17:00"
    }

    idle_time = 15 # 15 minutes
  }

  policy_tags = [
    "environment:production",
    "compliance:required",
  ]
}

# Individual USER principals for senior engineers
locals {
  senior_engineers = {
    alice = { email = "alice@example.com", name = "Alice Smith" },
    bob   = { email = "bob@example.com",   name = "Bob Johnson" },
  }
}

resource "cyberarksia_database_policy_principal_assignment" "senior_engineers" {
  for_each = local.senior_engineers

  policy_id               = cyberarksia_database_policy.production_database.policy_id
  principal_id            = each.value.email
  principal_type          = "USER"
  principal_name          = each.value.name
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# GROUP for DBA team
resource "cyberarksia_database_policy_principal_assignment" "dba_team" {
  policy_id               = cyberarksia_database_policy.production_database.policy_id
  principal_id            = "dba-team@example.com"
  principal_type          = "GROUP"
  principal_name          = "Database Administration Team"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# GROUP for backend engineers
resource "cyberarksia_database_policy_principal_assignment" "backend_team" {
  policy_id               = cyberarksia_database_policy.production_database.policy_id
  principal_id            = "backend-team@example.com"
  principal_type          = "GROUP"
  principal_name          = "Backend Engineering Team"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# ROLE for automated systems
resource "cyberarksia_database_policy_principal_assignment" "automation_role" {
  policy_id      = cyberarksia_database_policy.production_database.policy_id
  principal_id   = "database-automation"
  principal_type = "ROLE"
  principal_name = "Database Automation Role"
  # No directory fields needed for ROLE
}

# ROLE for emergency access
resource "cyberarksia_database_policy_principal_assignment" "emergency_role" {
  policy_id      = cyberarksia_database_policy.production_database.policy_id
  principal_id   = "emergency-access"
  principal_type = "ROLE"
  principal_name = "Emergency Database Access Role"
}

# Output all assignment IDs for reference
output "principal_assignments" {
  description = "All principal assignment composite IDs"
  value = {
    senior_engineers = { for k, v in cyberarksia_database_policy_principal_assignment.senior_engineers : k => v.id }
    dba_team         = cyberarksia_database_policy_principal_assignment.dba_team.id
    backend_team     = cyberarksia_database_policy_principal_assignment.backend_team.id
    automation_role  = cyberarksia_database_policy_principal_assignment.automation_role.id
    emergency_role   = cyberarksia_database_policy_principal_assignment.emergency_role.id
  }
}

# Show principal counts
output "principal_summary" {
  description = "Summary of principals by type"
  value = {
    users  = length(cyberarksia_database_policy_principal_assignment.senior_engineers)
    groups = 2
    roles  = 2
    total  = length(cyberarksia_database_policy_principal_assignment.senior_engineers) + 4
  }
}
