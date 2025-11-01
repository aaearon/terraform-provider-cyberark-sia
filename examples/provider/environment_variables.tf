# Provider Configuration Using Environment Variables
#
# This example demonstrates configuring the provider using environment variables,
# which is the recommended approach for CI/CD and production environments.
#
# Set these environment variables before running Terraform:
#   export CYBERARK_USERNAME="service-account@cyberark.cloud.XXXXX"
#   export CYBERARK_PASSWORD="<your-password-here>"

terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = ">= 0.1.0"
    }
  }
}

provider "cyberarksia" {
  # Credentials are automatically read from environment variables:
  # - CYBERARK_USERNAME
  # - CYBERARK_PASSWORD
  #
  # No explicit configuration needed when using environment variables
}
