# Database access policy with tags and metadata

resource "cyberarksia_database_policy" "tagged" {
  name                       = "Production-Database-Policy"
  description                = "Policy for production database access"
  status                     = "Active"
  delegation_classification  = "Restricted"
  time_zone                  = "GMT"

  policy_tags = [
    "environment:production",
    "team:database-admins",
    "compliance:soc2",
    "region:us-west",
  ]

  conditions {
    max_session_duration = 4
    idle_time            = 15

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}

# Example: Dynamic tags based on environment
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "team" {
  description = "Team name"
  type        = string
  default     = "platform"
}

resource "cyberarksia_database_policy" "dynamic_tags" {
  name                       = "${var.environment}-${var.team}-policy"
  description                = "Access policy for ${var.team} team in ${var.environment}"
  status                     = "Active"
  delegation_classification  = "Unrestricted"

  policy_tags = [
    "environment:${var.environment}",
    "team:${var.team}",
    "managed-by:terraform",
  ]

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }
}
