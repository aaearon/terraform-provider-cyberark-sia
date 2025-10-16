# Example: Onboard On-Premise Oracle Database to CyberArk SIA
#
# This example demonstrates:
# - Registering an existing on-premise Oracle database with SIA
# - Using explicit configuration values (no cloud provider needed)
# - Local authentication method
# - Manual configuration for non-cloud environments

terraform {
  required_providers {
    cyberark_sia = {
      source  = "local/cyberark-sia"
      version = "~> 1.0"
    }
  }
}

# Configure CyberArk SIA Provider
provider "cyberark_sia" {
  client_id                 = var.cyberark_client_id
  client_secret             = var.cyberark_client_secret
  identity_tenant_subdomain = var.cyberark_tenant_subdomain

  # Optional: Override identity URL for GovCloud or custom deployments
  # identity_url = var.cyberark_identity_url
}

# Register on-premise Oracle database with CyberArk SIA
resource "cyberark_sia_database_workspace" "oracle_prod" {
  name             = "oracle-production-db"
  database_type    = "oracle"
  database_version = "19.3.0" # Oracle 19c

  # On-premise database connection details
  address       = "oracle-prod.internal.example.com"
  port          = 1521
  database_name = "PRODDB" # Oracle SID or Service Name

  # Authentication method
  authentication_method = "local"

  # Explicitly set as on-premise (default, but shown for clarity)
  cloud_provider = "on_premise"

  description = "Production Oracle database in primary datacenter"

  tags = {
    Environment = "production"
    Location    = "DataCenter-East"
    Criticality = "High"
    Database    = "Oracle"
    Compliance  = "SOX"
  }
}

# Example: Oracle RAC (Real Application Clusters) configuration
resource "cyberark_sia_database_workspace" "oracle_rac" {
  name             = "oracle-rac-cluster"
  database_type    = "oracle"
  database_version = "19.12.0"

  # RAC SCAN address (Single Client Access Name)
  address       = "oracle-rac-scan.internal.example.com"
  port          = 1521
  database_name = "RACDB"

  authentication_method = "local"
  cloud_provider        = "on_premise"

  description = "Oracle RAC cluster for high availability workloads"

  tags = {
    Environment = "production"
    Location    = "DataCenter-West"
    Criticality = "Critical"
    HAEnabled   = "true"
    Database    = "Oracle-RAC"
  }
}

# Example: Oracle with specific service name
resource "cyberark_sia_database_workspace" "oracle_service" {
  name             = "oracle-hr-service"
  database_type    = "oracle"
  database_version = "21.3.0"

  address       = "oracle-hr.corp.example.com"
  port          = 1521
  database_name = "HRPDB.EXAMPLE.COM" # Pluggable database service name

  authentication_method = "local"
  cloud_provider        = "on_premise"

  description = "Oracle HR application pluggable database"

  tags = {
    Environment = "production"
    Application = "HR"
    Department  = "Human Resources"
    Database    = "Oracle-PDB"
  }
}

# Example: Development Oracle database
resource "cyberark_sia_database_workspace" "oracle_dev" {
  name             = "oracle-development"
  database_type    = "oracle"
  database_version = "19.3.0"

  address       = "oracle-dev.internal.example.com"
  port          = 1521
  database_name = "DEVDB"

  authentication_method = "local"
  cloud_provider        = "on_premise"

  description = "Development Oracle database for testing and QA"

  tags = {
    Environment = "development"
    Location    = "DataCenter-East"
    Criticality = "Low"
    Purpose     = "Testing"
  }
}

# Variables
variable "cyberark_client_id" {
  description = "CyberArk ISPSS client ID"
  type        = string
  sensitive   = true
}

variable "cyberark_client_secret" {
  description = "CyberArk ISPSS client secret"
  type        = string
  sensitive   = true
}

variable "cyberark_identity_url" {
  description = "CyberArk Identity URL (optional - only needed for GovCloud or custom deployments)"
  type        = string
  default     = ""
}

variable "cyberark_tenant_subdomain" {
  description = "CyberArk tenant subdomain (e.g., 'abc123' from abc123.cyberark.cloud)"
  type        = string
}

# Outputs
output "oracle_prod_id" {
  description = "SIA database target ID for production Oracle"
  value       = cyberark_sia_database_workspace.oracle_prod.id
}

output "oracle_rac_id" {
  description = "SIA database target ID for Oracle RAC cluster"
  value       = cyberark_sia_database_workspace.oracle_rac.id
}

output "oracle_hr_id" {
  description = "SIA database target ID for HR Oracle service"
  value       = cyberark_sia_database_workspace.oracle_service.id
}

output "oracle_dev_id" {
  description = "SIA database target ID for development Oracle"
  value       = cyberark_sia_database_workspace.oracle_dev.id
}

# Notes for on-premise deployments:
#
# 1. Network connectivity: Ensure SIA connectors can reach the database
#    - Firewall rules must allow TCP connections on port 1521 (or custom port)
#    - DNS resolution for database hostnames must work from connector network
#
# 2. Oracle listener configuration:
#    - Verify listener.ora is configured correctly
#    - Ensure tnsnames.ora has correct service names (if using service names)
#
# 3. Authentication prerequisites:
#    - Database user account must exist with appropriate privileges
#    - Account credentials will be managed by SIA strong accounts (separate resource)
#
# 4. High availability considerations:
#    - For Oracle RAC, use SCAN address for load balancing
#    - For Data Guard, use primary database address
#    - Configure appropriate connection retry and timeout settings
