package provider

import (
	"context"
	"testing"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestBuildAuthenticationProfile_DBAuth tests building a db_auth profile
func TestBuildAuthenticationProfile_DBAuth(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create test data with db_auth profile
	data := &models.PolicyDatabaseAssignmentModel{
		DBAuthProfile: &models.DBAuthProfileModel{
			Roles: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("db_reader"),
				types.StringValue("db_writer"),
			}),
		},
	}

	// Build profile
	profile := BuildAuthenticationProfile(ctx, "db_auth", data, &diags)

	// Verify no errors
	if diags.HasError() {
		t.Fatalf("Expected no errors, got: %v", diags.Errors())
	}

	// Verify profile type
	dbProfile, ok := profile.(*uapsiadbmodels.ArkUAPSIADBDBAuthProfile)
	if !ok {
		t.Fatalf("Expected *ArkUAPSIADBDBAuthProfile, got %T", profile)
	}

	// Verify roles
	if len(dbProfile.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(dbProfile.Roles))
	}
	if dbProfile.Roles[0] != "db_reader" {
		t.Errorf("Expected first role 'db_reader', got %q", dbProfile.Roles[0])
	}
	if dbProfile.Roles[1] != "db_writer" {
		t.Errorf("Expected second role 'db_writer', got %q", dbProfile.Roles[1])
	}
}

// TestBuildAuthenticationProfile_DBAuth_Missing tests error when db_auth profile is missing
func TestBuildAuthenticationProfile_DBAuth_Missing(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create test data WITHOUT db_auth profile
	data := &models.PolicyDatabaseAssignmentModel{
		DBAuthProfile: nil,
	}

	// Build profile
	_ = BuildAuthenticationProfile(ctx, "db_auth", data, &diags)

	// Verify error occurred
	if !diags.HasError() {
		t.Fatal("Expected error for missing db_auth profile, got none")
	}

	// Verify error message contains expected text
	errors := diags.Errors()
	if len(errors) == 0 {
		t.Fatal("Expected error diagnostic, got none")
	}

	errorSummary := errors[0].Summary()
	if errorSummary != "Missing db_auth Profile" {
		t.Errorf("Expected error summary 'Missing db_auth Profile', got %q", errorSummary)
	}

	// Profile should be nil on error (already verified by error check above)
}

// TestBuildAuthenticationProfile_LDAPAuth tests building an ldap_auth profile
func TestBuildAuthenticationProfile_LDAPAuth(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create test data with ldap_auth profile
	data := &models.PolicyDatabaseAssignmentModel{
		LDAPAuthProfile: &models.LDAPAuthProfileModel{
			AssignGroups: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("admins"),
				types.StringValue("developers"),
			}),
		},
	}

	// Build profile
	profile := BuildAuthenticationProfile(ctx, "ldap_auth", data, &diags)

	// Verify no errors
	if diags.HasError() {
		t.Fatalf("Expected no errors, got: %v", diags.Errors())
	}

	// Verify profile type
	ldapProfile, ok := profile.(*uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile)
	if !ok {
		t.Fatalf("Expected *ArkUAPSIADBLDAPAuthProfile, got %T", profile)
	}

	// Verify groups
	if len(ldapProfile.AssignGroups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(ldapProfile.AssignGroups))
	}
	if ldapProfile.AssignGroups[0] != "admins" {
		t.Errorf("Expected first group 'admins', got %q", ldapProfile.AssignGroups[0])
	}
}

// TestBuildAuthenticationProfile_UnsupportedMethod tests error for unsupported auth method
func TestBuildAuthenticationProfile_UnsupportedMethod(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := &models.PolicyDatabaseAssignmentModel{}

	// Build profile with invalid method
	_ = BuildAuthenticationProfile(ctx, "invalid_method", data, &diags)

	// Verify error occurred
	if !diags.HasError() {
		t.Fatal("Expected error for unsupported auth method, got none")
	}

	// Verify error message
	errors := diags.Errors()
	if len(errors) == 0 {
		t.Fatal("Expected error diagnostic, got none")
	}

	errorSummary := errors[0].Summary()
	if errorSummary != "Unsupported Authentication Method" {
		t.Errorf("Expected error summary 'Unsupported Authentication Method', got %q", errorSummary)
	}

	// Profile should be nil on error (already verified by error check above)
}

// TestSetProfileOnInstanceTarget_DBAuth tests setting db_auth profile on instance target
func TestSetProfileOnInstanceTarget_DBAuth(t *testing.T) {
	// Create instance target
	instanceTarget := &uapsiadbmodels.ArkUAPSIADBInstanceTarget{
		InstanceName:         "test-db",
		InstanceType:         "Postgres",
		AuthenticationMethod: "db_auth",
	}

	// Create profile
	profile := &uapsiadbmodels.ArkUAPSIADBDBAuthProfile{
		Roles: []string{"db_reader", "db_writer"},
	}

	// Set profile
	SetProfileOnInstanceTarget(instanceTarget, "db_auth", profile)

	// Verify profile was set
	if instanceTarget.DBAuthProfile == nil {
		t.Fatal("Expected DBAuthProfile to be set, got nil")
	}

	if len(instanceTarget.DBAuthProfile.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(instanceTarget.DBAuthProfile.Roles))
	}

	// Verify other profiles are nil
	if instanceTarget.LDAPAuthProfile != nil {
		t.Error("Expected LDAPAuthProfile to be nil")
	}
	if instanceTarget.OracleAuthProfile != nil {
		t.Error("Expected OracleAuthProfile to be nil")
	}
	if instanceTarget.MongoAuthProfile != nil {
		t.Error("Expected MongoAuthProfile to be nil")
	}
	if instanceTarget.SQLServerAuthProfile != nil {
		t.Error("Expected SQLServerAuthProfile to be nil")
	}
	if instanceTarget.RDSIAMUserAuthProfile != nil {
		t.Error("Expected RDSIAMUserAuthProfile to be nil")
	}
}

// TestSetProfileOnInstanceTarget_ClearsOtherProfiles tests that setting a profile clears others
func TestSetProfileOnInstanceTarget_ClearsOtherProfiles(t *testing.T) {
	// Create instance target with existing profiles
	instanceTarget := &uapsiadbmodels.ArkUAPSIADBInstanceTarget{
		InstanceName:         "test-db",
		InstanceType:         "Postgres",
		AuthenticationMethod: "ldap_auth",
		DBAuthProfile:        &uapsiadbmodels.ArkUAPSIADBDBAuthProfile{Roles: []string{"old"}},
		LDAPAuthProfile:      &uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile{AssignGroups: []string{"old"}},
	}

	// Create new LDAP profile
	profile := &uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile{
		AssignGroups: []string{"new_group"},
	}

	// Set profile (should clear db_auth and update ldap_auth)
	SetProfileOnInstanceTarget(instanceTarget, "ldap_auth", profile)

	// Verify old DB profile was cleared
	if instanceTarget.DBAuthProfile != nil {
		t.Error("Expected DBAuthProfile to be cleared, but it wasn't")
	}

	// Verify LDAP profile was set
	if instanceTarget.LDAPAuthProfile == nil {
		t.Fatal("Expected LDAPAuthProfile to be set, got nil")
	}

	if len(instanceTarget.LDAPAuthProfile.AssignGroups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(instanceTarget.LDAPAuthProfile.AssignGroups))
	}

	if instanceTarget.LDAPAuthProfile.AssignGroups[0] != "new_group" {
		t.Errorf("Expected group 'new_group', got %q", instanceTarget.LDAPAuthProfile.AssignGroups[0])
	}
}

// TODO: Add comprehensive tests for remaining auth methods:
// - TestBuildAuthenticationProfile_OracleAuth
// - TestBuildAuthenticationProfile_MongoAuth
// - TestBuildAuthenticationProfile_SQLServerAuth
// - TestBuildAuthenticationProfile_RDSIAMUserAuth

// TODO: Add tests for ParseAuthenticationProfile function:
// - TestParseAuthenticationProfile_DBAuth
// - TestParseAuthenticationProfile_LDAPAuth
// - TestParseAuthenticationProfile_OracleAuth
// - TestParseAuthenticationProfile_MongoAuth
// - TestParseAuthenticationProfile_SQLServerAuth
// - TestParseAuthenticationProfile_RDSIAMUserAuth

// TODO: Add round-trip tests:
// - TestProfileRoundTrip_DBAuth (Build -> Parse -> Build should be idempotent)
// - Similar tests for other auth methods
