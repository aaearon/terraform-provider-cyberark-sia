# Look up principals first (needed for inline principal blocks in policies)
data "cyberarksia_principal" "developer_group" {
  name = "developers@example.com"
  type = "GROUP"
}

data "cyberarksia_principal" "oncall_group" {
  name = "oncall-engineers@example.com"
  type = "GROUP"
}

# Create a database access policy with time-based conditions
resource "cyberarksia_database_policy" "developer_access" {
  name        = "Developer Database Access"
  status      = "active"
  description = "Allow developers to access databases during business hours"

  time_zone   = "America/New_York"
  policy_tags = ["developers", "business-hours", "production"]

  # Time frame block (validity period for the policy)
  time_frame {
    from_time = "2025-01-01T00:00:00Z"
    to_time   = "2025-12-31T23:59:59Z"
  }

  # Conditions block (session limits and access windows)
  conditions {
    max_session_duration = 8  # Maximum 8 hour sessions
    idle_time            = 10 # 10 minute idle timeout

    # Business hours access window
    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday (0=Sunday, 6=Saturday)
      from_hour        = "08:00"
      to_hour          = "18:00"
    }
  }

  # Inline database assignment (WHO gets access to WHAT)
  target_database {
    database_workspace_id = cyberarksia_database_workspace.production_postgres.id
    authentication_method = "db_auth"

    db_auth_profile {
      roles = ["app_developer"]
    }
  }

  target_database {
    database_workspace_id = cyberarksia_database_workspace.production_mysql.id
    authentication_method = "db_auth"

    db_auth_profile {
      roles = ["app_developer"]
    }
  }

  # Inline principal assignment (WHO can use this policy)
  principal {
    principal_id          = data.cyberarksia_principal.developer_group.principal_id
    principal_type        = data.cyberarksia_principal.developer_group.principal_type
    principal_name        = data.cyberarksia_principal.developer_group.principal_name
    source_directory_id   = data.cyberarksia_principal.developer_group.source_directory_id
    source_directory_name = data.cyberarksia_principal.developer_group.source_directory_name
  }
}

# Policy for 24/7 access (e.g., for on-call engineers)
resource "cyberarksia_database_policy" "oncall_access" {
  name        = "On-Call Engineer Access"
  status      = "active"
  description = "24/7 access for on-call engineers to analytics database"

  policy_tags = ["oncall", "24x7", "production"]

  # Time frame (full year, no restrictions)
  time_frame {
    from_time = "2025-01-01T00:00:00Z"
    to_time   = "2025-12-31T23:59:59Z"
  }

  # Conditions (no access_window = 24/7 access)
  conditions {
    max_session_duration = 12 # Longer sessions for on-call work
    idle_time            = 30 # Longer idle timeout
  }

  # Analytics database with RDS IAM authentication
  target_database {
    database_workspace_id = cyberarksia_database_workspace.rds_iam_database.id
    authentication_method = "rds_iam_user_auth"

    rds_iam_user_auth_profile {
      db_user = "oncall_analyst"
    }
  }

  # On-call engineer group
  principal {
    principal_id          = data.cyberarksia_principal.oncall_group.principal_id
    principal_type        = data.cyberarksia_principal.oncall_group.principal_type
    principal_name        = data.cyberarksia_principal.oncall_group.principal_name
    source_directory_id   = data.cyberarksia_principal.oncall_group.source_directory_id
    source_directory_name = data.cyberarksia_principal.oncall_group.source_directory_name
  }
}
