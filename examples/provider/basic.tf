# Basic Provider Configuration
#
# This example demonstrates the simplest way to configure the CyberArk SIA provider
# using explicit credentials.

terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = ">= 0.1.0"
    }
  }
}

provider "cyberarksia" {
  username = "service-account@cyberark.cloud.12345"
  password = var.cyberark_password
}

# Define the password as a sensitive variable
variable "cyberark_password" {
  description = "CyberArk Identity service account password"
  type        = string
  sensitive   = true
}
