# Suspended database access policy
# Use to temporarily disable access without deleting the policy

resource "cyberarksia_database_policy" "suspended" {
  name                       = "Temporarily-Disabled-Policy"
  description                = "Policy temporarily suspended for maintenance"
  status                     = "suspended" # Access denied while suspended
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 4
    idle_time            = 10
  }
}

# Example: Policy that can be toggled between Active and Suspended
variable "policy_enabled" {
  description = "Enable or disable the policy"
  type        = bool
  default     = true
}

resource "cyberarksia_database_policy" "toggleable" {
  name                       = "Toggleable-Access-Policy"
  status                     = "suspended"
  delegation_classification  = "Unrestricted"

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}
