# USER Principal with LDAP Directory
#
# This example demonstrates assigning a USER principal from an LDAP directory.
# Shows integration with corporate LDAP infrastructure.

resource "cyberarksia_database_policy" "corporate_access" {
  name                      = "Corporate-LDAP-Access"
  status                    = "active"
  delegation_classification = "Restricted"
  description               = "Policy for corporate users via LDAP"

  conditions {
    max_session_duration = 4  # 4 hours
    idle_time            = 15 # 15 minutes

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5] # Monday-Friday
      from_hour        = "08:00"
      to_hour          = "18:00"
    }
  }
}

resource "cyberarksia_database_policy_principal_assignment" "bob_ldap" {
  policy_id             = cyberarksia_database_policy.corporate_access.policy_id
  principal_id          = "bob.johnson"
  principal_type        = "USER"
  principal_name        = "Bob Johnson"
  source_directory_name = "CorporateLDAP"
  source_directory_id   = "ldap-directory-001"
}

# Multiple LDAP users
resource "cyberarksia_database_policy_principal_assignment" "ldap_users" {
  for_each = tomap({
    carol = { id = "carol.davis", name = "Carol Davis" },
    dave  = { id = "dave.martinez", name = "Dave Martinez" },
  })

  policy_id             = cyberarksia_database_policy.corporate_access.policy_id
  principal_id          = each.value.id
  principal_type        = "USER"
  principal_name        = each.value.name
  source_directory_name = "CorporateLDAP"
  source_directory_id   = "ldap-directory-001"
}
