terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = "~> 0.1"
    }
  }
}

provider "cyberarksia" {
  # Credentials can be set via environment variables:
  # export CYBERARK_USERNAME="service-account@cyberark.cloud.12345"
  # export CYBERARK_CLIENT_SECRET="your-secret"
  # export CYBERARK_IDENTITY_URL="https://abc123.cyberark.cloud" # Optional - auto-resolved from username
}
