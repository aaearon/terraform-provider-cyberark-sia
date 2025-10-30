# ROLE Principal (No Directory Required)
#
# This example demonstrates assigning a ROLE principal to a database policy.
# ROLEs do not require source_directory_name or source_directory_id.

resource "cyberarksia_database_policy" "admin_policy" {
  name                      = "Database-Admin-Role-Policy"
  status                    = "active"
  delegation_classification = "Restricted"
  description               = "Policy for database administrator role"

  conditions {
    max_session_duration = 12 # 12 hours for admin work
    idle_time            = 60 # 60 minutes
  }
}

# Database Administrator role
resource "cyberarksia_database_policy_principal_assignment" "admin_role" {
  policy_id      = cyberarksia_database_policy.admin_policy.policy_id
  principal_id   = "database-administrator"
  principal_type = "ROLE"
  principal_name = "Database Administrator Role"

  # Note: source_directory_name and source_directory_id are NOT required for ROLE type
  # They can be omitted entirely
}

# Read-Only Database role
resource "cyberarksia_database_policy" "readonly_policy" {
  name                      = "Database-ReadOnly-Role-Policy"
  status                    = "active"
  delegation_classification = "Unrestricted"

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}

resource "cyberarksia_database_policy_principal_assignment" "readonly_role" {
  policy_id      = cyberarksia_database_policy.readonly_policy.policy_id
  principal_id   = "database-reader"
  principal_type = "ROLE"
  principal_name = "Database Read-Only Role"
}

# DevOps role with time restrictions
resource "cyberarksia_database_policy" "devops_policy" {
  name                      = "DevOps-On-Call-Policy"
  status                    = "active"
  delegation_classification = "Restricted"

  conditions {
    max_session_duration = 6

    access_window {
      days_of_the_week = [0, 1, 2, 3, 4, 5, 6] # All days
      from_hour        = "00:00"
      to_hour          = "23:59"
    }
  }
}

resource "cyberarksia_database_policy_principal_assignment" "devops_oncall" {
  policy_id      = cyberarksia_database_policy.devops_policy.policy_id
  principal_id   = "devops-oncall"
  principal_type = "ROLE"
  principal_name = "DevOps On-Call Role"
}
