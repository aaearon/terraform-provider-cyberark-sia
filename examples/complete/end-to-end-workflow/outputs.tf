output "postgres_workspace_id" {
  description = "ID of the PostgreSQL database workspace"
  value       = cyberarksia_database_workspace.production_postgres.id
}

output "developer_policy_id" {
  description = "ID of the developer access policy"
  value       = cyberarksia_database_policy.developer_access.id
}

output "oncall_policy_id" {
  description = "ID of the on-call access policy"
  value       = cyberarksia_database_policy.oncall_access.id
}

output "access_summary" {
  description = "Summary of configured access"
  value = {
    developer_group = {
      policy      = cyberarksia_database_policy.developer_access.name
      databases   = ["customers (PostgreSQL)", "orders (MySQL)", "staging-customers (PostgreSQL)"]
      access_time = "Monday-Friday, 8 AM - 6 PM EST"
    }
    senior_developers = {
      policy      = cyberarksia_database_policy.developer_access.name
      note        = "Added via modular assignment pattern"
      access_time = "Monday-Friday, 8 AM - 6 PM EST"
    }
    oncall_group = {
      policy      = cyberarksia_database_policy.oncall_access.name
      databases   = ["analytics (PostgreSQL with RDS IAM)"]
      access_time = "24/7"
    }
  }
}
