// Package models defines Terraform state models
package models

import (
	"context"
	"fmt"

	uapcommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/common/models"
	uapsiacommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/common/models"
	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatabasePolicyModel represents the Terraform state for cyberarksia_database_policy resource
type DatabasePolicyModel struct {
	// Identifying attributes
	ID       types.String `tfsdk:"id"`        // Same as PolicyID (UUID)
	PolicyID types.String `tfsdk:"policy_id"` // UUID from API (computed)

	// Required metadata
	Name                     types.String `tfsdk:"name"`                        // 1-200 chars, unique
	Status                   types.String `tfsdk:"status"`                      // "active"|"suspended"
	DelegationClassification types.String `tfsdk:"delegation_classification"`   // "restricted"|"unrestricted"

	// Optional metadata
	Description types.String `tfsdk:"description"` // max 200 chars
	TimeZone    types.String `tfsdk:"time_zone"`   // max 50 chars, default "GMT"

	// Optional time frame
	TimeFrame *TimeFrameModel `tfsdk:"time_frame"`

	// Optional tags (max 20)
	PolicyTags types.List `tfsdk:"policy_tags"` // []string

	// Inline assignments (required by API - at least 1 of each)
	// Note: Singular tfsdk names match familiar Terraform patterns (aws_security_group ingress/egress)
	TargetDatabase []InlineDatabaseAssignmentModel `tfsdk:"target_database"`
	Principal      []InlinePrincipalModel          `tfsdk:"principal"`

	// Conditions (nested block)
	Conditions *ConditionsModel `tfsdk:"conditions"`

	// Computed metadata
	CreatedBy    *ChangeInfoModel `tfsdk:"created_by"`
	UpdatedOn    *ChangeInfoModel `tfsdk:"updated_on"`
	LastModified types.String     `tfsdk:"last_modified"`
}

// TimeFrameModel represents policy validity period
type TimeFrameModel struct {
	FromTime types.String `tfsdk:"from_time"` // ISO 8601
	ToTime   types.String `tfsdk:"to_time"`   // ISO 8601
}

// ConditionsModel represents policy access conditions
type ConditionsModel struct {
	MaxSessionDuration types.Int64        `tfsdk:"max_session_duration"` // 1-24 hours
	IdleTime           types.Int64        `tfsdk:"idle_time"`            // 1-120 minutes, default 10
	AccessWindow       *AccessWindowModel `tfsdk:"access_window"`
}

// AccessWindowModel represents time-based access restrictions
type AccessWindowModel struct {
	DaysOfTheWeek types.Set    `tfsdk:"days_of_the_week"` // Set of int, 0=Sunday through 6=Saturday (order doesn't matter)
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
	DatabaseWorkspaceID  types.String `tfsdk:"database_workspace_id"`
	AuthenticationMethod types.String `tfsdk:"authentication_method"`

	// Profile blocks (mutually exclusive based on authentication_method)
	DBAuthProfile         *DBAuthProfileModel         `tfsdk:"db_auth_profile"`
	LDAPAuthProfile       *LDAPAuthProfileModel       `tfsdk:"ldap_auth_profile"`
	OracleAuthProfile     *OracleAuthProfileModel     `tfsdk:"oracle_auth_profile"`
	MongoAuthProfile      *MongoAuthProfileModel      `tfsdk:"mongo_auth_profile"`
	SQLServerAuthProfile  *SQLServerAuthProfileModel  `tfsdk:"sqlserver_auth_profile"`
	RDSIAMUserAuthProfile *RDSIAMUserAuthProfileModel `tfsdk:"rds_iam_user_auth_profile"`
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
	m.Status = types.StringValue(policy.Metadata.Status.Status)
	m.TimeZone = types.StringValue(policy.Metadata.TimeZone)
	m.DelegationClassification = types.StringValue(policy.DelegationClassification)

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

	// Computed fields
	if policy.Metadata.CreatedBy.User != "" {
		m.CreatedBy = &ChangeInfoModel{
			User:      types.StringValue(policy.Metadata.CreatedBy.User),
			Timestamp: types.StringValue(policy.Metadata.CreatedBy.Time),
		}
	}

	if policy.Metadata.UpdatedOn.User != "" {
		m.UpdatedOn = &ChangeInfoModel{
			User:      types.StringValue(policy.Metadata.UpdatedOn.User),
			Timestamp: types.StringValue(policy.Metadata.UpdatedOn.Time),
		}
	}

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
			// Convert Set to []int (Set automatically handles ordering)
			c.AccessWindow.DaysOfTheWeek.ElementsAs(context.Background(), &days, false)
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
		dayValues := make([]attr.Value, len(c.AccessWindow.DaysOfTheWeek))
		for i, day := range c.AccessWindow.DaysOfTheWeek {
			dayValues[i] = types.Int64Value(int64(day))
		}

		// Use SetValue instead of ListValue - order doesn't matter
		daysSet, _ := types.SetValue(types.Int64Type, dayValues)

		conditions.AccessWindow = &AccessWindowModel{
			DaysOfTheWeek: daysSet,
			FromHour:      types.StringValue(c.AccessWindow.FromHour),
			ToHour:        types.StringValue(c.AccessWindow.ToHour),
		}
	}

	return conditions
}
