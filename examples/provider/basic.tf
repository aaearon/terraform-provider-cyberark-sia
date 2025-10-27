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
  username      = "service-account@cyberark.cloud.12345"
  client_secret = var.cyberark_client_secret
}

# Define the client_secret as a sensitive variable
variable "cyberark_client_secret" {
  description = "CyberArk Identity service account client secret"
  type        = string
  sensitive   = true
}
