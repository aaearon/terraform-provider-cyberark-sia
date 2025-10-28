# Database access policy with full conditions
# Restricts access to weekdays 9 AM - 5 PM in GMT timezone

resource "cyberarksia_database_policy" "with_conditions" {
  name                       = "Weekday-Business-Hours-Policy"
  description                = "Database access limited to business hours"
  status                     = "active"
  delegation_classification  = "Unrestricted"
  time_zone                  = "GMT"

  conditions {
    max_session_duration = 4  # 4 hours
    idle_time            = 15 # 15 minutes

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday (0=Sunday, 6=Saturday)
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}

# Example with US Eastern timezone
resource "cyberarksia_database_policy" "us_eastern" {
  name                       = "US-Eastern-Business-Hours"
  description                = "Access during US Eastern business hours"
  status                     = "active"
  delegation_classification  = "Unrestricted"
  time_zone                  = "America/New_York"

  conditions {
    max_session_duration = 8
    idle_time            = 20

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]
      from_hour        = "08:00"
      to_hour          = "18:00"
    }
  }
}
