# GROUP Principal with Azure AD
#
# This example demonstrates assigning a GROUP principal from Azure AD.
# Groups provide centralized principal management - add/remove users in Azure AD
# and changes automatically propagate to policy access.

resource "cyberarksia_database_policy" "team_policy" {
  name                       = "Engineering-Teams-Policy"
  status                     = "active"
  delegation_classification  = "Unrestricted"
  description                = "Database access for engineering teams"

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}

# Database Administrators group
resource "cyberarksia_database_policy_principal_assignment" "dba_group" {
  policy_id               = cyberarksia_database_policy.team_policy.policy_id
  principal_id            = "dba-team@example.com"
  principal_type          = "GROUP"
  principal_name          = "Database Administrators"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# Backend Engineers group
resource "cyberarksia_database_policy_principal_assignment" "backend_group" {
  policy_id               = cyberarksia_database_policy.team_policy.policy_id
  principal_id            = "backend-engineers@example.com"
  principal_type          = "GROUP"
  principal_name          = "Backend Engineering Team"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# Data Analytics group with limited hours
resource "cyberarksia_database_policy" "analytics_limited" {
  name                       = "Analytics-Business-Hours"
  status                     = "active"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 4

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}

resource "cyberarksia_database_policy_principal_assignment" "analytics_group" {
  policy_id               = cyberarksia_database_policy.analytics_limited.policy_id
  principal_id            = "data-analytics@example.com"
  principal_type          = "GROUP"
  principal_name          = "Data Analytics Team"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}
