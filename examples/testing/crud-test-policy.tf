# CRUD Test Template for cyberarksia_database_policy
# This template validates CREATE → READ → UPDATE → DELETE workflow

terraform {
  required_providers {
    cyberarksia = {
      source  = "terraform.local/local/cyberark-sia"
      version = "0.1.0"
    }
  }
}

# ============================================================================
# STEP 1: CREATE - Create a minimal policy
# ============================================================================

resource "cyberarksia_database_policy" "test" {
  name                       = "CRUD-Test-Policy-${formatdate("YYYYMMDD-hhmmss", timestamp())}"
  description                = "CRUD validation test policy"
  status                     = "Active"
  delegation_classification  = "Unrestricted"
  time_zone                  = "GMT"

  conditions {
    max_session_duration = 4
    idle_time            = 10
  }
}

# ============================================================================
# STEP 2: READ - Verify state matches configuration
# ============================================================================

output "create_validation" {
  value = {
    policy_id                 = cyberarksia_database_policy.test.policy_id
    name                      = cyberarksia_database_policy.test.name
    description               = cyberarksia_database_policy.test.description
    status                    = cyberarksia_database_policy.test.status
    delegation_classification = cyberarksia_database_policy.test.delegation_classification
    time_zone                 = cyberarksia_database_policy.test.time_zone
    max_session_duration      = cyberarksia_database_policy.test.conditions.max_session_duration
    idle_time                 = cyberarksia_database_policy.test.conditions.idle_time
    has_created_by            = cyberarksia_database_policy.test.created_by != null
    has_updated_on            = cyberarksia_database_policy.test.updated_on != null
  }
  description = "CREATE validation - Verify all attributes populated correctly"
}

# ============================================================================
# VALIDATION CHECKLIST - CREATE
# ============================================================================
# [ ] policy_id is UUID format
# [ ] name matches input
# [ ] description matches input
# [ ] status is "Active"
# [ ] delegation_classification is "Unrestricted"
# [ ] time_zone is "GMT"
# [ ] max_session_duration is 4
# [ ] idle_time is 10
# [ ] created_by block is populated
# [ ] updated_on block is populated

# ============================================================================
# STEP 3: UPDATE - Modify description, status, and conditions
# ============================================================================
# To test UPDATE, modify the resource block above:
# 1. Change description to "Updated CRUD test policy"
# 2. Change status to "Suspended"
# 3. Change max_session_duration to 8
# 4. Change idle_time to 30
# 5. Add access_window block (uncomment below)
# 6. Run: terraform apply

# Uncomment to test UPDATE with access window:
/*
resource "cyberarksia_database_policy" "test" {
  name                       = "CRUD-Test-Policy-${formatdate("YYYYMMDD-hhmmss", timestamp())}"
  description                = "Updated CRUD test policy"  # CHANGED
  status                     = "Suspended"                 # CHANGED
  delegation_classification  = "Unrestricted"
  time_zone                  = "GMT"

  conditions {
    max_session_duration = 8   # CHANGED
    idle_time            = 30  # CHANGED

    # ADDED access_window
    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]  # Monday-Friday
      from_hour        = "09:00"
      to_hour          = "17:00"
    }
  }
}
*/

output "update_validation" {
  value = {
    description_changed = cyberarksia_database_policy.test.description
    status_changed      = cyberarksia_database_policy.test.status
    max_session_updated = cyberarksia_database_policy.test.conditions.max_session_duration
    idle_time_updated   = cyberarksia_database_policy.test.conditions.idle_time
    has_access_window   = cyberarksia_database_policy.test.conditions.access_window != null
  }
  description = "UPDATE validation - Verify changes applied correctly"
}

# ============================================================================
# VALIDATION CHECKLIST - UPDATE
# ============================================================================
# [ ] description changed to "Updated CRUD test policy"
# [ ] status changed to "Suspended"
# [ ] max_session_duration changed to 8
# [ ] idle_time changed to 30
# [ ] access_window block populated (if added)
# [ ] policy_id remains unchanged
# [ ] name remains unchanged (ForceNew if changed)
# [ ] updated_on timestamp changed

# ============================================================================
# STEP 4: IMPORT - Test import functionality
# ============================================================================
# 1. Remove policy from state: terraform state rm cyberarksia_database_policy.test
# 2. Import using policy_id: terraform import cyberarksia_database_policy.test <policy_id_from_output>
# 3. Run terraform plan - should show no changes

# ============================================================================
# VALIDATION CHECKLIST - IMPORT
# ============================================================================
# [ ] Import command succeeds
# [ ] terraform plan shows no changes
# [ ] All attributes match state before removal
# [ ] created_by and updated_on are populated

# ============================================================================
# STEP 5: DELETE - Clean up test policy
# ============================================================================
# Run: terraform destroy -auto-approve
# Verify: Policy removed from SIA UI

# ============================================================================
# VALIDATION CHECKLIST - DELETE
# ============================================================================
# [ ] terraform destroy succeeds
# [ ] Policy no longer appears in SIA UI
# [ ] No orphaned resources remain

# ============================================================================
# ADVANCED TESTING - Policy with all features
# ============================================================================

# Uncomment to test policy with all attributes:
/*
resource "cyberarksia_database_policy" "complete" {
  name                       = "Complete-Test-Policy-${formatdate("YYYYMMDD-hhmmss", timestamp())}"
  description                = "Complete test with all attributes"
  status                     = "Active"
  delegation_classification  = "Restricted"
  time_zone                  = "America/New_York"

  time_frame {
    from_time = "2024-01-01T00:00:00Z"
    to_time   = "2024-12-31T23:59:59Z"
  }

  policy_tags = [
    "test:crud",
    "environment:test",
    "managed-by:terraform",
  ]

  conditions {
    max_session_duration = 8
    idle_time            = 20

    access_window {
      days_of_the_week = [1, 2, 3, 4, 5]
      from_hour        = "08:00"
      to_hour          = "18:00"
    }
  }
}

output "complete_validation" {
  value = {
    has_time_frame    = cyberarksia_database_policy.complete.time_frame != null
    has_tags          = length(cyberarksia_database_policy.complete.policy_tags) > 0
    has_access_window = cyberarksia_database_policy.complete.conditions.access_window != null
  }
}
*/

# ============================================================================
# TESTING NOTES
# ============================================================================
# 1. Run each step sequentially (CREATE → READ → UPDATE → IMPORT → DELETE)
# 2. Validate outputs after each terraform apply
# 3. Check SIA UI to verify policy creation/updates
# 4. Use unique names with timestamps to avoid conflicts
# 5. Clean up test policies after validation
