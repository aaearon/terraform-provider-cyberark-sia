# Basic database access policy with minimal configuration
# This policy is valid indefinitely and has default conditions

resource "cyberarksia_database_policy" "basic" {
  name                       = "Basic-Database-Policy"
  status                     = "active"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 8  # 8 hours
    idle_time            = 30 # 30 minutes
  }
}

# Output the policy ID for use in assignments
output "policy_id" {
  value       = cyberarksia_database_policy.basic.policy_id
  description = "The ID of the created policy"
}
