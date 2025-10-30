// Package models defines Terraform state models
package models

import (
	"context"
	"fmt"
	"sort"
	"strings"

	uapcommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/common/models"
	uapsiacommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/common/models"
	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// changeInfoAttrTypes defines the attribute types for ChangeInfo objects (created_by, updated_on)
var changeInfoAttrTypes = map[string]attr.Type{
	"user":      types.StringType,
	"timestamp": types.StringType,
}

// ChangeInfoAttrTypes returns the attribute types for ChangeInfo objects
// Used for creating ObjectNull values in provider resource operations
func ChangeInfoAttrTypes() map[string]attr.Type {
	return changeInfoAttrTypes
}

// createChangeInfoObject creates a types.Object from user and timestamp strings
// Returns ObjectNull if user is empty, otherwise returns ObjectValue with the provided data
func createChangeInfoObject(user, timestamp string) types.Object {
	if user == "" {
		return types.ObjectNull(changeInfoAttrTypes)
	}

	attrs := map[string]attr.Value{
		"user":      types.StringValue(user),
		"timestamp": types.StringValue(timestamp),
	}

	objVal, diags := types.ObjectValue(changeInfoAttrTypes, attrs)
	if diags.HasError() {
		// Log error but return null object to avoid blocking operations
		// In practice, this should never happen with valid string inputs
		return types.ObjectNull(changeInfoAttrTypes)
	}

	return objVal
}

// DatabasePolicyModel represents the Terraform state for cyberarksia_database_policy resource
type DatabasePolicyModel struct {
	// Inline assignments (required by API - at least 1 of each)
	// Note: Singular tfsdk names match familiar Terraform patterns (aws_security_group ingress/egress)
	TargetDatabase []InlineDatabaseAssignmentModel `tfsdk:"target_database"` // 24 bytes (slice)
	Principal      []InlinePrincipalModel          `tfsdk:"principal"`       // 24 bytes (slice)

	// Optional time frame
	TimeFrame *TimeFrameModel `tfsdk:"time_frame"` // 8 bytes (pointer)

	// Conditions (nested block)
	Conditions *ConditionsModel `tfsdk:"conditions"` // 8 bytes (pointer)

	// Optional tags (max 20)
	PolicyTags types.List `tfsdk:"policy_tags"` // 8 bytes - []string

	// Computed metadata
	CreatedBy types.Object `tfsdk:"created_by"` // 8 bytes
	UpdatedOn types.Object `tfsdk:"updated_on"` // 8 bytes

	// Identifying attributes
	ID       types.String `tfsdk:"id"`        // types.String (likely 8 bytes) - Same as PolicyID (UUID)
	PolicyID types.String `tfsdk:"policy_id"` // types.String - UUID from API (computed)

	// Required metadata
	Name                     types.String `tfsdk:"name"`                      // types.String - 1-200 chars, unique
	Status                   types.String `tfsdk:"status"`                    // types.String - "active"|"suspended"
	DelegationClassification types.String `tfsdk:"delegation_classification"` // types.String - "restricted"|"unrestricted"

	// Optional metadata
	Description  types.String `tfsdk:"description"`   // types.String - max 200 chars
	TimeZone     types.String `tfsdk:"time_zone"`     // types.String - max 50 chars, default "GMT"
	LastModified types.String `tfsdk:"last_modified"` // types.String
}

// TimeFrameModel represents policy validity period
type TimeFrameModel struct {
	FromTime types.String `tfsdk:"from_time"` // ISO 8601
	ToTime   types.String `tfsdk:"to_time"`   // ISO 8601
}

// ConditionsModel represents policy access conditions
type ConditionsModel struct {
	AccessWindow       *AccessWindowModel `tfsdk:"access_window"`        // 8 bytes (pointer)
	MaxSessionDuration types.Int64        `tfsdk:"max_session_duration"` // 8 bytes - 1-24 hours
	IdleTime           types.Int64        `tfsdk:"idle_time"`            // 8 bytes - 1-120 minutes, default 10
}

// AccessWindowModel represents time-based access restrictions
type AccessWindowModel struct {
	DaysOfTheWeek types.Set    `tfsdk:"days_of_the_week"` // Set of int, 0=Sunday through 6=Saturday (order automatically normalized)
	FromHour      types.String `tfsdk:"from_hour"`        // "HH:MM"
	ToHour        types.String `tfsdk:"to_hour"`          // "HH:MM"
}

// ChangeInfoModel represents user and timestamp for policy changes
type ChangeInfoModel struct {
	User      types.String `tfsdk:"user"`
	Timestamp types.String `tfsdk:"timestamp"` // ISO 8601
}

// InlineDatabaseAssignmentModel represents an inline target database assignment
type InlineDatabaseAssignmentModel struct {
	// Profile blocks (mutually exclusive based on authentication_method)
	DBAuthProfile         *DBAuthProfileModel         `tfsdk:"db_auth_profile"`          // 8 bytes (pointer)
	LDAPAuthProfile       *LDAPAuthProfileModel       `tfsdk:"ldap_auth_profile"`        // 8 bytes (pointer)
	OracleAuthProfile     *OracleAuthProfileModel     `tfsdk:"oracle_auth_profile"`      // 8 bytes (pointer)
	MongoAuthProfile      *MongoAuthProfileModel      `tfsdk:"mongo_auth_profile"`       // 8 bytes (pointer)
	SQLServerAuthProfile  *SQLServerAuthProfileModel  `tfsdk:"sqlserver_auth_profile"`   // 8 bytes (pointer)
	RDSIAMUserAuthProfile *RDSIAMUserAuthProfileModel `tfsdk:"rds_iam_user_auth_profile"` // 8 bytes (pointer)

	DatabaseWorkspaceID  types.String `tfsdk:"database_workspace_id"`  // types.String
	AuthenticationMethod types.String `tfsdk:"authentication_method"` // types.String
}

// InlinePrincipalModel represents an inline principal assignment
type InlinePrincipalModel struct {
	PrincipalID         types.String `tfsdk:"principal_id"`
	PrincipalType       types.String `tfsdk:"principal_type"`
	PrincipalName       types.String `tfsdk:"principal_name"`
	SourceDirectoryName types.String `tfsdk:"source_directory_name"`
	SourceDirectoryID   types.String `tfsdk:"source_directory_id"`
}

// ToSDK converts Terraform state model to ARK SDK policy struct
func (m *DatabasePolicyModel) ToSDK() *uapsiadbmodels.ArkUAPSIADBAccessPolicy {
	policy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		ArkUAPSIACommonAccessPolicy: uapsiacommonmodels.ArkUAPSIACommonAccessPolicy{
			ArkUAPCommonAccessPolicy: uapcommonmodels.ArkUAPCommonAccessPolicy{
				Metadata: uapcommonmodels.ArkUAPMetadata{
					PolicyID:    m.PolicyID.ValueString(),
					Name:        m.Name.ValueString(),
					Description: m.Description.ValueString(),
					Status: uapcommonmodels.ArkUAPPolicyStatus{
						Status: m.Status.ValueString(),
					},
					PolicyEntitlement: uapcommonmodels.ArkUAPPolicyEntitlement{
						TargetCategory: "DB",
						LocationType:   "FQDN/IP",
						PolicyType:     "Recurring",
					},
					TimeZone: m.TimeZone.ValueString(),
				},
				DelegationClassification: m.DelegationClassification.ValueString(),
			},
		},
	}

	// Convert policy tags
	if !m.PolicyTags.IsNull() && !m.PolicyTags.IsUnknown() {
		var tags []string
		m.PolicyTags.ElementsAs(context.Background(), &tags, false)
		policy.Metadata.PolicyTags = tags
	}

	// Convert time frame
	if m.TimeFrame != nil {
		policy.Metadata.TimeFrame = uapcommonmodels.ArkUAPTimeFrame{
			FromTime: m.TimeFrame.FromTime.ValueString(),
			ToTime:   m.TimeFrame.ToTime.ValueString(),
		}
	}

	// Convert conditions
	if m.Conditions != nil {
		policy.Conditions = convertConditionsToSDK(m.Conditions)
	}

	return policy
}

// FromSDK populates Terraform state model from ARK SDK policy struct
func (m *DatabasePolicyModel) FromSDK(ctx context.Context, policy *uapsiadbmodels.ArkUAPSIADBAccessPolicy) error {
	m.ID = types.StringValue(policy.Metadata.PolicyID)
	m.PolicyID = types.StringValue(policy.Metadata.PolicyID)
	m.Name = types.StringValue(policy.Metadata.Name)
	m.Description = types.StringValue(policy.Metadata.Description)
	// Keep API values as-is (API returns "Active"/"Suspended" capitalized)
	// Normalize to lowercase to match user config (API returns titlecase)
	m.Status = types.StringValue(strings.ToLower(policy.Metadata.Status.Status))
	m.TimeZone = types.StringValue(policy.Metadata.TimeZone)
	// Normalize to lowercase to match user config (API returns titlecase)
	m.DelegationClassification = types.StringValue(strings.ToLower(policy.DelegationClassification))

	// Convert policy tags
	if len(policy.Metadata.PolicyTags) > 0 {
		tagValues := make([]attr.Value, len(policy.Metadata.PolicyTags))
		for i, tag := range policy.Metadata.PolicyTags {
			tagValues[i] = types.StringValue(tag)
		}
		tagList, diags := types.ListValue(types.StringType, tagValues)
		if diags.HasError() {
			return fmt.Errorf("failed to convert policy tags: %v", diags.Errors())
		}
		m.PolicyTags = tagList
	} else {
		m.PolicyTags = types.ListNull(types.StringType)
	}

	// Convert time frame
	if policy.Metadata.TimeFrame.FromTime != "" || policy.Metadata.TimeFrame.ToTime != "" {
		m.TimeFrame = &TimeFrameModel{
			FromTime: types.StringValue(policy.Metadata.TimeFrame.FromTime),
			ToTime:   types.StringValue(policy.Metadata.TimeFrame.ToTime),
		}
	}

	// Convert conditions
	m.Conditions = convertConditionsFromSDK(ctx, &policy.Conditions)

	// Computed fields - convert to types.Object to handle unknown values properly
	m.CreatedBy = createChangeInfoObject(policy.Metadata.CreatedBy.User, policy.Metadata.CreatedBy.Time)
	m.UpdatedOn = createChangeInfoObject(policy.Metadata.UpdatedOn.User, policy.Metadata.UpdatedOn.Time)

	return nil
}

// convertConditionsToSDK converts Terraform conditions to SDK conditions
func convertConditionsToSDK(c *ConditionsModel) uapsiacommonmodels.ArkUAPSIACommonConditions {
	conditions := uapsiacommonmodels.ArkUAPSIACommonConditions{
		ArkUAPConditions: uapcommonmodels.ArkUAPConditions{
			MaxSessionDuration: int(c.MaxSessionDuration.ValueInt64()),
		},
		IdleTime: int(c.IdleTime.ValueInt64()),
	}

	// Convert access window if present
	if c.AccessWindow != nil {
		var days []int
		if !c.AccessWindow.DaysOfTheWeek.IsNull() && !c.AccessWindow.DaysOfTheWeek.IsUnknown() {
			// Convert set to slice
			var daysInt64 []int64
			c.AccessWindow.DaysOfTheWeek.ElementsAs(context.Background(), &daysInt64, false)

			// Sort to ensure canonical order (eliminates plan/state mismatch)
			sort.Slice(daysInt64, func(i, j int) bool { return daysInt64[i] < daysInt64[j] })

			// Convert []int64 to []int for SDK
			days = make([]int, len(daysInt64))
			for i, day := range daysInt64 {
				days[i] = int(day)
			}
		}

		conditions.AccessWindow = uapcommonmodels.ArkUAPTimeCondition{
			DaysOfTheWeek: days,
			FromHour:      c.AccessWindow.FromHour.ValueString(),
			ToHour:        c.AccessWindow.ToHour.ValueString(),
		}
	}

	return conditions
}

// convertConditionsFromSDK converts SDK conditions to Terraform conditions
func convertConditionsFromSDK(ctx context.Context, c *uapsiacommonmodels.ArkUAPSIACommonConditions) *ConditionsModel {
	if c == nil {
		return nil
	}

	conditions := &ConditionsModel{
		MaxSessionDuration: types.Int64Value(int64(c.MaxSessionDuration)),
		IdleTime:           types.Int64Value(int64(c.IdleTime)),
	}

	// Convert access window if present
	if len(c.AccessWindow.DaysOfTheWeek) > 0 || c.AccessWindow.FromHour != "" || c.AccessWindow.ToHour != "" {
		// Convert API response to slice
		daysInt64 := make([]int64, len(c.AccessWindow.DaysOfTheWeek))
		for i, day := range c.AccessWindow.DaysOfTheWeek {
			daysInt64[i] = int64(day)
		}

		// Sort to ensure canonical order (eliminates plan/state mismatch)
		sort.Slice(daysInt64, func(i, j int) bool { return daysInt64[i] < daysInt64[j] })

		// Create set from sorted days
		daysSet, _ := types.SetValueFrom(ctx, types.Int64Type, daysInt64)

		conditions.AccessWindow = &AccessWindowModel{
			DaysOfTheWeek: daysSet,
			FromHour:      types.StringValue(c.AccessWindow.FromHour),
			ToHour:        types.StringValue(c.AccessWindow.ToHour),
		}
	}

	return conditions
}
