# Example: Onboard Azure SQL Server to CyberArk SIA
#
# This example demonstrates:
# - Creating an Azure SQL Server using azurerm provider
# - Registering the database with SIA using database_target resource
# - Using Terraform references for Azure-specific configuration
# - Active Directory authentication method

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    cyberark_sia = {
      source  = "local/cyberark-sia"
      version = "~> 1.0"
    }
  }
}

# Configure Azure Provider
provider "azurerm" {
  features {}
}

# Configure CyberArk SIA Provider
provider "cyberark_sia" {
  client_id                 = var.cyberark_client_id
  client_secret             = var.cyberark_client_secret
  identity_url              = var.cyberark_identity_url
  identity_tenant_subdomain = var.cyberark_tenant_subdomain
}

# Get current Azure client configuration
data "azurerm_client_config" "current" {}

# Resource group for SQL Server
resource "azurerm_resource_group" "sql" {
  name     = "rg-sql-production"
  location = "East US"

  tags = {
    Environment = "production"
    ManagedBy   = "Terraform"
  }
}

# Azure SQL Server
resource "azurerm_mssql_server" "main" {
  name                         = "sql-server-prod-${random_string.suffix.result}"
  resource_group_name          = azurerm_resource_group.sql.name
  location                     = azurerm_resource_group.sql.location
  version                      = "12.0" # SQL Server 2019
  administrator_login          = "sqladmin"
  administrator_login_password = var.sql_admin_password

  # Azure AD authentication
  azuread_administrator {
    login_username = "sql-admin-group"
    object_id      = var.azure_ad_admin_object_id
  }

  # Security settings
  minimum_tls_version               = "1.2"
  public_network_access_enabled     = false
  outbound_network_restriction_enabled = false

  tags = {
    Environment = "production"
    ManagedBy   = "Terraform"
    SIAManaged  = "true"
  }
}

# SQL Database
resource "azurerm_mssql_database" "app" {
  name      = "app-database"
  server_id = azurerm_mssql_server.main.id

  sku_name = "S0" # Standard tier

  max_size_gb = 50

  tags = {
    Environment = "production"
    Application = "api-backend"
  }
}

# Firewall rule for SIA connector access
resource "azurerm_mssql_firewall_rule" "sia_connector" {
  name             = "sia-connector-access"
  server_id        = azurerm_mssql_server.main.id
  start_ip_address = var.sia_connector_ip
  end_ip_address   = var.sia_connector_ip
}

# Register the Azure SQL Server with CyberArk SIA
resource "cyberark_sia_database_target" "azure_sql" {
  name             = "production-sqlserver"
  database_type    = "sqlserver"
  database_version = "13.0.0" # SQL Server 2016+

  # Use Terraform reference to get FQDN dynamically
  address = azurerm_mssql_server.main.fully_qualified_domain_name
  port    = 1433

  database_name = azurerm_mssql_database.app.name

  # Authentication method - Active Directory for enterprise scenarios
  authentication_method = "domain"

  # Azure-specific metadata
  cloud_provider        = "azure"
  azure_tenant_id       = data.azurerm_client_config.current.tenant_id
  azure_subscription_id = data.azurerm_client_config.current.subscription_id

  description = "Production SQL Server for application services"

  tags = {
    Environment  = "production"
    Team         = "platform"
    Application  = "api-backend"
    CloudProvider = "Azure"
  }

  # Ensure SQL Server is created before SIA registration
  depends_on = [azurerm_mssql_server.main]
}

# Random suffix for globally unique SQL Server name
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
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
  description = "CyberArk Identity URL"
  type        = string
}

variable "cyberark_tenant_subdomain" {
  description = "CyberArk tenant subdomain"
  type        = string
}

variable "sql_admin_password" {
  description = "SQL Server administrator password"
  type        = string
  sensitive   = true
}

variable "azure_ad_admin_object_id" {
  description = "Azure AD admin group object ID"
  type        = string
}

variable "sia_connector_ip" {
  description = "IP address of SIA connector for firewall rule"
  type        = string
}

# Outputs
output "database_target_id" {
  description = "SIA database target ID"
  value       = cyberark_sia_database_target.azure_sql.id
}

output "sql_server_fqdn" {
  description = "SQL Server fully qualified domain name"
  value       = azurerm_mssql_server.main.fully_qualified_domain_name
}

output "database_name" {
  description = "Database name"
  value       = azurerm_mssql_database.app.name
}
