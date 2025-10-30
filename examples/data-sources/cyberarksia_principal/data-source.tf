# CyberArk SIA Principal Data Source Examples

terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = "~> 0.1"
    }
  }
}

provider "cyberarksia" {
  username      = "service-account@cyberark.cloud.12345"
  client_secret = var.cyberark_secret
}

# Example 1: Basic Cloud Directory user lookup
data "cyberarksia_principal" "cloud_user" {
  name = "admin@cyberark.cloud.12345"
}

output "cloud_user_details" {
  value = {
    id           = data.cyberarksia_principal.cloud_user.id
    type         = data.cyberarksia_principal.cloud_user.principal_type
    display_name = data.cyberarksia_principal.cloud_user.display_name
    directory    = data.cyberarksia_principal.cloud_user.directory_name
    email        = data.cyberarksia_principal.cloud_user.email
  }
}

# Example 2: Federated Directory (Entra ID) user lookup
data "cyberarksia_principal" "entra_user" {
  name = "john.doe@company.com"
}

output "entra_user_id" {
  value = data.cyberarksia_principal.entra_user.id
}

# Example 3: Group lookup with type filter
data "cyberarksia_principal" "db_admins" {
  name = "Database Administrators"
  type = "GROUP"
}

output "group_details" {
  value = {
    id           = data.cyberarksia_principal.db_admins.id
    type         = data.cyberarksia_principal.db_admins.principal_type
    display_name = data.cyberarksia_principal.db_admins.display_name
    directory    = data.cyberarksia_principal.db_admins.directory_name
  }
}

# Example 4: Active Directory user lookup
data "cyberarksia_principal" "ad_user" {
  name = "SchindlerT@cyberiam.tech"
}

# Example 5: User lookup with explicit type filter
data "cyberarksia_principal" "specific_user" {
  name = "service-account@cyberark.cloud.12345"
  type = "USER"
}

# Example 6: Role lookup
data "cyberarksia_principal" "system_role" {
  name = "System Administrator"
  type = "ROLE"
}

# Example 7: Integration with policy assignment
data "cyberarksia_principal" "policy_user" {
  name = "tim.schindler@cyberark.cloud.40562"
}

data "cyberarksia_database_policy" "production_policy" {
  name = "Production Database Access"
}

resource "cyberarksia_database_policy_principal_assignment" "user_access" {
  policy_id             = data.cyberarksia_database_policy.production_policy.id
  principal_id          = data.cyberarksia_principal.policy_user.id
  principal_type        = data.cyberarksia_principal.policy_user.principal_type
  principal_name        = data.cyberarksia_principal.policy_user.name
  source_directory_name = data.cyberarksia_principal.policy_user.directory_name
  source_directory_id   = data.cyberarksia_principal.policy_user.directory_id
}

# Example 8: Multiple principals for group-based access
data "cyberarksia_principal" "dev_group" {
  name = "Developers"
  type = "GROUP"
}

data "cyberarksia_principal" "qa_group" {
  name = "QA Engineers"
  type = "GROUP"
}

resource "cyberarksia_database_policy_principal_assignment" "dev_access" {
  policy_id             = data.cyberarksia_database_policy.production_policy.id
  principal_id          = data.cyberarksia_principal.dev_group.id
  principal_type        = data.cyberarksia_principal.dev_group.principal_type
  principal_name        = data.cyberarksia_principal.dev_group.name
  source_directory_name = data.cyberarksia_principal.dev_group.directory_name
  source_directory_id   = data.cyberarksia_principal.dev_group.directory_id
}

resource "cyberarksia_database_policy_principal_assignment" "qa_access" {
  policy_id             = data.cyberarksia_database_policy.production_policy.id
  principal_id          = data.cyberarksia_principal.qa_group.id
  principal_type        = data.cyberarksia_principal.qa_group.principal_type
  principal_name        = data.cyberarksia_principal.qa_group.name
  source_directory_name = data.cyberarksia_principal.qa_group.directory_name
  source_directory_id   = data.cyberarksia_principal.qa_group.directory_id
}

# Variables
variable "cyberark_secret" {
  description = "CyberArk service account secret"
  type        = string
  sensitive   = true
}
