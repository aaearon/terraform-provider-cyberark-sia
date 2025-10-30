# Complete database access policy with all options
# Demonstrates all available attributes including time_frame

resource "cyberarksia_database_policy" "complete" {
  name                      = "Complete-Example-Policy"
  description               = "Comprehensive example showing all policy attributes"
  status                    = "active"
  delegation_classification = "Restricted"
  time_zone                 = "America/New_York"

  # Policy validity period (optional)
  time_frame {
    from_time = "2024-01-01T00:00:00Z"
    to_time   = "2024-12-31T23:59:59Z"
  }

  # Policy tags (max 20)
  policy_tags = [
    "environment:production",
    "compliance:soc2",
    "compliance:hipaa",
    "team:security",
    "owner:platform-team",
    "cost-center:engineering",
  ]

  # Access conditions
  conditions {
    max_session_duration = 4  # Maximum 4-hour sessions
    idle_time            = 10 # 10-minute idle timeout

    # Time-based access restrictions
    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday through Friday
      from_hour        = "08:00"         # 8 AM
      to_hour          = "20:00"         # 8 PM
    }
  }
}

# Example: Policy for 24/7 access (weekend and weekday)
resource "cyberarksia_database_policy" "twentyfourseven" {
  name                      = "Always-Available-Policy"
  description               = "24/7 database access for on-call teams"
  status                    = "active"
  delegation_classification = "Unrestricted"

  policy_tags = [
    "access-type:24x7",
    "team:on-call",
  ]

  conditions {
    max_session_duration = 12 # Longer sessions for troubleshooting
    idle_time            = 60 # 1-hour idle timeout

    access_window {
      days_of_the_week = [0, 1, 2, 3, 4, 5, 6] # All days
      from_hour        = "00:00"               # Midnight
      to_hour          = "23:59"               # End of day
    }
  }
}

# Example: Short-term project policy with time_frame
resource "cyberarksia_database_policy" "project" {
  name                      = "Q1-Migration-Project"
  description               = "Temporary access for Q1 database migration"
  status                    = "active"
  delegation_classification = "Restricted"

  time_frame {
    from_time = "2024-01-01T00:00:00Z"
    to_time   = "2024-03-31T23:59:59Z"
  }

  policy_tags = [
    "project:migration",
    "temporary:true",
    "expires:2024-03-31",
  ]

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}

# Outputs
output "complete_policy_id" {
  value       = cyberarksia_database_policy.complete.policy_id
  description = "Policy ID for the complete example"
}

output "complete_policy_metadata" {
  value = {
    id         = cyberarksia_database_policy.complete.policy_id
    name       = cyberarksia_database_policy.complete.name
    status     = cyberarksia_database_policy.complete.status
    created_by = cyberarksia_database_policy.complete.created_by
    updated_on = cyberarksia_database_policy.complete.updated_on
  }
  description = "Complete policy metadata"
}
