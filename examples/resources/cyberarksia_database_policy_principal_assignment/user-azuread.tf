# USER Principal with Azure AD Directory
#
# This example demonstrates assigning a USER principal from Azure AD to a database policy.
# Required fields for USER type: source_directory_name and source_directory_id

resource "cyberarksia_database_policy" "example" {
  name                       = "Example-Database-Policy"
  status                     = "active"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 8  # 8 hours
    idle_time            = 30 # 30 minutes
  }
}

resource "cyberarksia_database_policy_principal_assignment" "alice_azuread" {
  policy_id               = cyberarksia_database_policy.example.policy_id
  principal_id            = "alice@example.com"
  principal_type          = "USER"
  principal_name          = "Alice Smith"
  source_directory_name   = "AzureAD"
  source_directory_id     = "12345678-1234-1234-1234-123456789012"
}

# Output the composite ID for reference
output "assignment_id" {
  description = "Composite ID in format policy-id:principal-id:principal-type"
  value       = cyberarksia_database_policy_principal_assignment.alice_azuread.id
}
