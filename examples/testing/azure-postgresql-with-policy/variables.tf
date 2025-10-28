# ============================================================================
# SIA Provider Configuration
# ============================================================================

variable "sia_username" {
  description = "CyberArk SIA service account username (from .env file)"
  type        = string
  sensitive   = true
}

variable "sia_client_secret" {
  description = "CyberArk SIA client secret (from .env file)"
  type        = string
  sensitive   = true
}

# ============================================================================
# Azure Configuration
# ============================================================================

variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "azure_region" {
  description = "Azure region for resources"
  type        = string
  default     = "westus2"
}

variable "resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = "rg-sia-policy-test"
}

# ============================================================================
# PostgreSQL Configuration
# ============================================================================

variable "postgres_admin_username" {
  description = "PostgreSQL administrator username"
  type        = string
  default     = "pgadmin"
}

variable "postgres_admin_password" {
  description = "PostgreSQL administrator password (must be strong)"
  type        = string
  sensitive   = true
}

# ============================================================================
# Test Principal Configuration (Tim Schindler)
# ============================================================================

variable "test_principal_id" {
  description = "UUID of Tim Schindler's user account in SIA"
  type        = string
  # Format: "c2c7bcc6-9560-44e0-8dff-5be221cd37ee"
}

variable "test_principal_email" {
  description = "Email address of test principal"
  type        = string
  default     = "tim.schindler@cyberark.cloud.40562"
}

variable "cyberark_cloud_directory_id" {
  description = "CyberArk Cloud Directory UUID"
  type        = string
  default     = "09B9A9B0-6CE8-465F-AB03-65766D33B05E"
}

# ============================================================================
# Service Account Principal (for inline policy assignment)
# ============================================================================

variable "service_account_principal_id" {
  description = "UUID of service account principal in SIA"
  type        = string
  # This should be the UUID of the sia_username account
}
