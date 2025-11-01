# ==============================================================================
# CyberArk SIA CRUD Testing Template - Provider Configuration
# ==============================================================================
# This is a CANONICAL TEMPLATE - do not modify directly.
#
# For complete testing workflow, see:
#   examples/testing/TESTING-GUIDE.md
#
# Usage:
#   1. Copy this file to /tmp/sia-crud-validation/
#   2. Update with your credentials (from project root .env file)
#   3. Follow TESTING-GUIDE.md for complete instructions
# ==============================================================================

terraform {
  required_providers {
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = ">= 0.1.0"
    }
  }
}

provider "cyberarksia" {
  # Get credentials from .env file in project root:
  # - CYBERARK_USERNAME=your-username@cyberark.cloud.XXXX
  # - CYBERARK_PASSWORD=your-password

  username = "your-username@cyberark.cloud.XXXX"
  password = "<your-password-here>"
}
