# Input variables for Azure + SIA integration test

variable "azure_subscription_id" {
  description = "Azure subscription ID (from az account show)"
  type        = string
  sensitive   = true
}

variable "azure_region" {
  description = "Azure region for PostgreSQL server (eastus is typically cheapest)"
  type        = string
  default     = "eastus"
}

variable "postgres_admin_username" {
  description = "PostgreSQL admin username"
  type        = string
  default     = "siaadmin"
}

variable "postgres_admin_password" {
  description = "PostgreSQL admin password (generated securely)"
  type        = string
  sensitive   = true
}

variable "sia_username" {
  description = "CyberArk SIA username (from .env)"
  type        = string
  sensitive   = true
}

variable "sia_client_secret" {
  description = "CyberArk SIA client secret (from .env)"
  type        = string
  sensitive   = true
}
