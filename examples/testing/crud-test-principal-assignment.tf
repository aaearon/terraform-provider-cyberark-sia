# CRUD Test Template: cyberarksia_database_policy_principal_assignment
#
# This template tests all CRUD operations for principal assignment resource.
# Follow the testing workflow in TESTING-GUIDE.md for step-by-step validation.
#
# Prerequisites:
# 1. Provider configured with valid credentials
# 2. Existing policy OR create one (see database_policy test below)
# 3. Valid identity directory (Azure AD, LDAP) with test principals
# 4. Principal IDs and directory information
#
# Test Duration: 10-15 minutes
# Cost: FREE (no cloud resources required)

terraform {
  required_providers {
    cyberarksia = {
      source  = "terraform.local/local/cyberark-sia"
      version = "0.1.0"
    }
  }
}

provider "cyberarksia" {
  username      = var.sia_username
  client_secret = var.sia_client_secret
}

variable "sia_username" {
  description = "SIA service account username (format: user@cyberark.cloud.XXXX)"
  type        = string
}

variable "sia_client_secret" {
  description = "SIA service account client secret"
  type        = string
  sensitive   = true
}

# Test Variables - UPDATE THESE with your values
variable "test_policy_id" {
  description = "Existing policy ID to assign principals to (or use the policy created below)"
  type        = string
  default     = "" # Leave empty to create new policy
}

variable "azuread_directory_id" {
  description = "Azure AD directory ID for USER/GROUP principals"
  type        = string
  default     = "12345678-1234-1234-1234-123456789012" # UPDATE THIS
}

variable "test_user_email" {
  description = "Test USER principal email"
  type        = string
  default     = "test-user@example.com" # UPDATE THIS
}

variable "test_group_email" {
  description = "Test GROUP principal email"
  type        = string
  default     = "test-group@example.com" # UPDATE THIS
}

# ============================================================================
# STEP 1: Create Test Policy (optional - use if test_policy_id not provided)
# ============================================================================

resource "cyberarksia_database_policy" "test" {
  count = var.test_policy_id == "" ? 1 : 0

  name                      = "CRUD-Test-Policy-${formatdate("YYYYMMDD-HHmmss", timestamp())}"
  status                    = "active"
  delegation_classification = "unrestricted"
  description               = "Temporary policy for principal assignment CRUD testing"

  conditions {
    max_session_duration = 8
    idle_time            = 30
  }

  # ⚠️ API REQUIREMENT: At least 1 target_database and 1 principal required
  # TODO: Update these with actual values for your environment

  target_database {
    database_workspace_id = "YOUR_DATABASE_WORKSPACE_ID_HERE"
    authentication_method = "db_auth"
    db_auth_profile {
      roles = ["pg_read_all_settings"]
    }
  }

  principal {
    principal_id          = "YOUR_PRINCIPAL_UUID_HERE"
    principal_type        = "USER"
    principal_name        = "initial.user@example.com"
    source_directory_name = "CyberArk Cloud Directory"
    source_directory_id   = "YOUR_DIRECTORY_ID_HERE"
  }

  # Allow assignment resources to manage additional principals
  lifecycle {
    ignore_changes = [principal]
  }
}

locals {
  # Use provided policy_id or created policy's ID
  policy_id = var.test_policy_id != "" ? var.test_policy_id : cyberarksia_database_policy.test[0].policy_id
}

# ============================================================================
# STEP 2: CREATE - Assign principals to policy
# ============================================================================

# Test USER principal (requires directory)
resource "cyberarksia_database_policy_principal_assignment" "test_user" {
  policy_id             = local.policy_id
  principal_id          = var.test_user_email
  principal_type        = "USER"
  principal_name        = "Test User - CRUD Testing"
  source_directory_name = "AzureAD-Test"
  source_directory_id   = var.azuread_directory_id
}

# Test GROUP principal (requires directory)
resource "cyberarksia_database_policy_principal_assignment" "test_group" {
  policy_id             = local.policy_id
  principal_id          = var.test_group_email
  principal_type        = "GROUP"
  principal_name        = "Test Group - CRUD Testing"
  source_directory_name = "AzureAD-Test"
  source_directory_id   = var.azuread_directory_id
}

# Test ROLE principal (no directory required)
resource "cyberarksia_database_policy_principal_assignment" "test_role" {
  policy_id      = local.policy_id
  principal_id   = "crud-test-role"
  principal_type = "ROLE"
  principal_name = "CRUD Test Role"
  # Note: No source_directory fields for ROLE
}

# ============================================================================
# STEP 3: READ - Validation Outputs
# ============================================================================

output "validation_summary" {
  description = "CRUD test validation summary"
  value = {
    policy_id = local.policy_id

    principals_assigned = {
      user = {
        id            = cyberarksia_database_policy_principal_assignment.test_user.id
        composite_id  = cyberarksia_database_policy_principal_assignment.test_user.id
        principal_id  = cyberarksia_database_policy_principal_assignment.test_user.principal_id
        type          = cyberarksia_database_policy_principal_assignment.test_user.principal_type
        has_directory = cyberarksia_database_policy_principal_assignment.test_user.source_directory_id != ""
      }
      group = {
        id            = cyberarksia_database_policy_principal_assignment.test_group.id
        composite_id  = cyberarksia_database_policy_principal_assignment.test_group.id
        principal_id  = cyberarksia_database_policy_principal_assignment.test_group.principal_id
        type          = cyberarksia_database_policy_principal_assignment.test_group.principal_type
        has_directory = cyberarksia_database_policy_principal_assignment.test_group.source_directory_id != ""
      }
      role = {
        id            = cyberarksia_database_policy_principal_assignment.test_role.id
        composite_id  = cyberarksia_database_policy_principal_assignment.test_role.id
        principal_id  = cyberarksia_database_policy_principal_assignment.test_role.principal_id
        type          = cyberarksia_database_policy_principal_assignment.test_role.principal_type
        has_directory = false
      }
    }

    principal_count = {
      users  = 1
      groups = 1
      roles  = 1
      total  = 3
    }
  }
}

output "composite_id_format" {
  description = "Composite ID format examples for import testing"
  value = {
    user  = cyberarksia_database_policy_principal_assignment.test_user.id
    group = cyberarksia_database_policy_principal_assignment.test_group.id
    role  = cyberarksia_database_policy_principal_assignment.test_role.id

    format_explanation = "Format: policy-id:principal-id:principal-type (3 parts)"
  }
}

output "import_commands" {
  description = "Commands to test import functionality"
  value = {
    user  = "terraform import cyberarksia_database_policy_principal_assignment.test_user \"${cyberarksia_database_policy_principal_assignment.test_user.id}\""
    group = "terraform import cyberarksia_database_policy_principal_assignment.test_group \"${cyberarksia_database_policy_principal_assignment.test_group.id}\""
    role  = "terraform import cyberarksia_database_policy_principal_assignment.test_role \"${cyberarksia_database_policy_principal_assignment.test_role.id}\""
  }
}

# ============================================================================
# STEP 4: UPDATE Test Instructions (Manual)
# ============================================================================

# To test UPDATE (in-place modification):
# 1. Edit principal_name for any assignment above
# 2. Run: terraform apply
# 3. Verify: Changed name appears in SIA UI
# 4. Verify: terraform plan shows no changes after apply
#
# Example:
# principal_name = "Test User - Updated Name"

# ============================================================================
# STEP 5: DELETE Test Instructions (Manual)
# ============================================================================

# To test DELETE:
# 1. Run: terraform destroy -target=cyberarksia_database_policy_principal_assignment.test_role
# 2. Verify: Role removed from policy in SIA UI
# 3. Verify: USER and GROUP principals still present (read-modify-write pattern)
# 4. Run: terraform destroy (full cleanup)

# ============================================================================
# Testing Checklist
# ============================================================================

# ✅ CREATE Tests:
# - [ ] USER principal created with directory fields
# - [ ] GROUP principal created with directory fields
# - [ ] ROLE principal created without directory fields
# - [ ] All principals appear in SIA UI under policy's "Assigned To" section
# - [ ] Composite IDs follow format: policy-id:principal-id:principal-type

# ✅ READ Tests:
# - [ ] terraform refresh shows no changes
# - [ ] terraform plan shows no changes
# - [ ] All outputs display correct information
# - [ ] Composite ID format validated

# ✅ UPDATE Tests:
# - [ ] Modify principal_name → terraform apply succeeds
# - [ ] Changed name appears in SIA UI
# - [ ] terraform plan shows no changes after update
# - [ ] ForceNew attributes (policy_id, principal_id, principal_type) trigger recreation

# ✅ DELETE Tests:
# - [ ] Delete single principal → other principals preserved
# - [ ] terraform destroy removes all principals
# - [ ] SIA UI shows principals removed from policy
# - [ ] No errors during cleanup

# ✅ Import Tests:
# - [ ] terraform import with 3-part ID succeeds for USER
# - [ ] terraform import with 3-part ID succeeds for GROUP
# - [ ] terraform import with 3-part ID succeeds for ROLE
# - [ ] Invalid ID format shows clear error message

# ✅ Validation Tests:
# - [ ] USER without directory fields fails with clear error
# - [ ] GROUP without directory fields fails with clear error
# - [ ] ROLE without directory fields succeeds
# - [ ] Duplicate principal (same ID + type) fails with clear error
# - [ ] Invalid principal_type fails validation

# ============================================================================
# Next Steps
# ============================================================================

# After CRUD testing completes:
# 1. Review validation_summary output
# 2. Test import commands
# 3. Verify all checklist items
# 4. Run full cleanup: terraform destroy -auto-approve
# 5. Document any issues or unexpected behavior
