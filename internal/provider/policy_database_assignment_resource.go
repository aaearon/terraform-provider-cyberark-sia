package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/client"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/validators"
	dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
	uapcommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/common/models"
	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PolicyDatabaseAssignmentResource{}
var _ resource.ResourceWithImportState = &PolicyDatabaseAssignmentResource{}

func NewPolicyDatabaseAssignmentResource() resource.Resource {
	return &PolicyDatabaseAssignmentResource{}
}

// PolicyDatabaseAssignmentResource defines the resource implementation.
type PolicyDatabaseAssignmentResource struct {
	providerData *ProviderData
}

func (r *PolicyDatabaseAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_database_assignment"
}

func (r *PolicyDatabaseAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the assignment of a database workspace to an existing SIA access policy. " +
			"This resource follows the AWS Security Group Rule pattern - manage individual database assignments " +
			"to a policy rather than managing the entire policy.\n\n" +
			"Policies can be created using the `cyberarksia_database_policy` resource or managed through the SIA UI. " +
			"Use the `cyberarksia_access_policy` data source to reference existing policies.\n\n" +
			"**IMPORTANT**: Multiple assignments to the same policy within a single Terraform workspace are supported. " +
			"However, managing the same policy from multiple Terraform workspaces can cause conflicts. " +
			"See the resource documentation for best practices.",

		Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SIA access policy to add the database to. Use the `cyberarksia_access_policy` data source for lookup by name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_workspace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the database workspace to add to the policy.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method for this database. Valid values: `db_auth`, `ldap_auth`, `oracle_auth`, `mongo_auth`, `sqlserver_auth`, `rds_iam_user_auth`.",
				Required:            true,
				Validators: []validator.String{
					validators.AuthenticationMethod(),
				},
				// Note: Updates to authentication_method are supported
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the format `policy-id:database-id`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last modification to the policy assignment.",
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{
			"db_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "Database authentication profile. Use when `authentication_method` is `db_auth`. **Required** if authentication_method is `db_auth`.",
				Attributes: map[string]schema.Attribute{
					"roles": schema.ListAttribute{
						MarkdownDescription: "List of database roles to assign to the user. **Required** when this profile is used.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"ldap_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "LDAP authentication profile. Use when `authentication_method` is `ldap_auth`. **Required** if authentication_method is `ldap_auth`.",
				Attributes: map[string]schema.Attribute{
					"assign_groups": schema.ListAttribute{
						MarkdownDescription: "List of LDAP groups to assign to the user. **Required** when this profile is used.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"oracle_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "Oracle authentication profile. Use when `authentication_method` is `oracle_auth`. **Required** if authentication_method is `oracle_auth`.",
				Attributes: map[string]schema.Attribute{
					"roles": schema.ListAttribute{
						MarkdownDescription: "List of Oracle roles to assign to the user. **Required** when this profile is used.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"dba_role": schema.BoolAttribute{
						MarkdownDescription: "Grant DBA role to the user.",
						Optional:            true,
					},
					"sysdba_role": schema.BoolAttribute{
						MarkdownDescription: "Grant SYSDBA role to the user.",
						Optional:            true,
					},
					"sysoper_role": schema.BoolAttribute{
						MarkdownDescription: "Grant SYSOPER role to the user.",
						Optional:            true,
					},
				},
			},
			"mongo_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "MongoDB authentication profile. Use when `authentication_method` is `mongo_auth`.",
				Attributes: map[string]schema.Attribute{
					"global_builtin_roles": schema.ListAttribute{
						MarkdownDescription: "List of global built-in roles to assign.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"database_builtin_roles": schema.MapAttribute{
						MarkdownDescription: "Map of database names to their built-in roles.",
						Optional:            true,
						ElementType:         types.ListType{ElemType: types.StringType},
					},
					"database_custom_roles": schema.MapAttribute{
						MarkdownDescription: "Map of database names to their custom roles.",
						Optional:            true,
						ElementType:         types.ListType{ElemType: types.StringType},
					},
				},
			},
			"sqlserver_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "SQL Server authentication profile. Use when `authentication_method` is `sqlserver_auth`.",
				Attributes: map[string]schema.Attribute{
					"global_builtin_roles": schema.ListAttribute{
						MarkdownDescription: "List of global built-in roles to assign.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"global_custom_roles": schema.ListAttribute{
						MarkdownDescription: "List of global custom roles to assign.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"database_builtin_roles": schema.MapAttribute{
						MarkdownDescription: "Map of database names to their built-in roles.",
						Optional:            true,
						ElementType:         types.ListType{ElemType: types.StringType},
					},
					"database_custom_roles": schema.MapAttribute{
						MarkdownDescription: "Map of database names to their custom roles.",
						Optional:            true,
						ElementType:         types.ListType{ElemType: types.StringType},
					},
				},
			},
			"rds_iam_user_auth_profile": schema.SingleNestedBlock{
				MarkdownDescription: "RDS IAM User authentication profile. Use when `authentication_method` is `rds_iam_user_auth`. **Required** if authentication_method is `rds_iam_user_auth`.",
				Attributes: map[string]schema.Attribute{
					"db_user": schema.StringAttribute{
						MarkdownDescription: "The database user for RDS IAM authentication. **Required** when this profile is used.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *PolicyDatabaseAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerData = providerData
}

func (r *PolicyDatabaseAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.PolicyDatabaseAssignmentModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	LogOperationStart(ctx, "create", "policy_database_assignment")

	policyID := data.PolicyID.ValueString()
	databaseID := data.DatabaseWorkspaceID.ValueString()
	authMethod := data.AuthenticationMethod.ValueString()

	// Step 1: Fetch existing policy (READ-MODIFY-WRITE pattern)
	tflog.Debug(ctx, "Fetching policy", map[string]interface{}{
		"policy_id": policyID,
	})

	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "fetch policy"))
		return
	}

	// DEBUG: Log fetched policy structure
	tflog.Debug(ctx, "Fetched policy structure", map[string]interface{}{
		"policy_id":        policy.Metadata.PolicyID,
		"policy_name":      policy.Metadata.Name,
		"targets_count":    len(policy.Targets),
		"principals_count": len(policy.Principals),
		"delegation_class": policy.DelegationClassification,
	})

	// Log each workspace type and instance count
	for wsType, targets := range policy.Targets {
		tflog.Debug(ctx, "Workspace type in policy", map[string]interface{}{
			"workspace_type":  wsType,
			"instances_count": len(targets.Instances),
		})
	}

	// Step 2: Fetch database workspace (get InstanceName, InstanceType, InstanceID, Platform)
	tflog.Debug(ctx, "Fetching database workspace", map[string]interface{}{
		"database_id": databaseID,
	})

	// Convert string to int for database fetch
	databaseIDInt, err := strconv.Atoi(databaseID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Database ID",
			fmt.Sprintf("Database workspace ID must be a valid integer: %s", err.Error()))
		return
	}

	database, err := r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
		ID: databaseIDInt,
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "fetch database workspace"))
		return
	}

	// Determine workspace type from database platform
	workspaceType := determineWorkspaceType(database.Platform)
	tflog.Debug(ctx, "Determined workspace type", map[string]interface{}{
		"platform":       database.Platform,
		"workspace_type": workspaceType,
	})

	// Step 3: Check if database already exists in policy (IDEMPOTENCY)
	if policy.Targets == nil {
		policy.Targets = make(map[string]uapsiadbmodels.ArkUAPSIADBTargets)
	}

	// DEBUG: Log fetched policy targets structure
	tflog.Debug(ctx, "Fetched policy targets", map[string]interface{}{
		"targets_count": len(policy.Targets),
		"target_types": func() []string {
			keys := make([]string, 0, len(policy.Targets))
			for k := range policy.Targets {
				keys = append(keys, k)
			}
			return keys
		}(),
	})

	// Search for existing database in the policy
	existingTarget := findDatabaseInPolicy(policy, strconv.Itoa(database.ID))
	if existingTarget != nil {
		tflog.Info(ctx, "Database already exists in policy - adopting existing configuration", map[string]interface{}{
			"policy_id":   policyID,
			"database_id": databaseID,
		})

		// IDEMPOTENT: Adopt existing configuration
		data.ID = types.StringValue(buildCompositeID(policyID, databaseID))
		data.LastModified = types.StringValue(time.Now().UTC().Format(time.RFC3339))

		// Update state with existing configuration
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		LogOperationSuccess(ctx, "create", "policy_database_assignment", data.ID.ValueString())
		return
	}

	// Step 4: Build ArkUAPSIADBInstanceTarget with profile
	instanceTarget := &uapsiadbmodels.ArkUAPSIADBInstanceTarget{
		InstanceName:         database.Name,
		InstanceType:         database.ProviderDetails.Family,
		InstanceID:           strconv.Itoa(database.ID),
		AuthenticationMethod: authMethod,
	}

	// Set the appropriate profile based on authentication method
	switch authMethod {
	case "db_auth":
		if data.DBAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "db_auth_profile block is required when authentication_method is 'db_auth'")
			return
		}
		var roles []string
		resp.Diagnostics.Append(data.DBAuthProfile.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		instanceTarget.DBAuthProfile = &uapsiadbmodels.ArkUAPSIADBDBAuthProfile{Roles: roles}

	case "ldap_auth":
		if data.LDAPAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "ldap_auth_profile block is required when authentication_method is 'ldap_auth'")
			return
		}
		var assignGroups []string
		resp.Diagnostics.Append(data.LDAPAuthProfile.AssignGroups.ElementsAs(ctx, &assignGroups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		instanceTarget.LDAPAuthProfile = &uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile{AssignGroups: assignGroups}

	case "oracle_auth":
		if data.OracleAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "oracle_auth_profile block is required when authentication_method is 'oracle_auth'")
			return
		}
		var roles []string
		resp.Diagnostics.Append(data.OracleAuthProfile.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		instanceTarget.OracleAuthProfile = &uapsiadbmodels.ArkUAPSIADBOracleAuthProfile{
			Roles:       roles,
			DbaRole:     data.OracleAuthProfile.DbaRole.ValueBool(),
			SysdbaRole:  data.OracleAuthProfile.SysdbaRole.ValueBool(),
			SysoperRole: data.OracleAuthProfile.SysoperRole.ValueBool(),
		}

	case "mongo_auth":
		if data.MongoAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "mongo_auth_profile block is required when authentication_method is 'mongo_auth'")
			return
		}
		mongoProfile := &uapsiadbmodels.ArkUAPSIADBMongoAuthProfile{}

		// Global builtin roles
		if !data.MongoAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalRoles []string
			resp.Diagnostics.Append(data.MongoAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.GlobalBuiltinRoles = globalRoles
		}

		// Database builtin roles
		if !data.MongoAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltinRoles := make(map[string][]string)
			resp.Diagnostics.Append(data.MongoAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltinRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.DatabaseBuiltinRoles = dbBuiltinRoles
		}

		// Database custom roles
		if !data.MongoAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustomRoles := make(map[string][]string)
			resp.Diagnostics.Append(data.MongoAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustomRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.DatabaseCustomRoles = dbCustomRoles
		}

		instanceTarget.MongoAuthProfile = mongoProfile

	case "sqlserver_auth":
		if data.SQLServerAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "sqlserver_auth_profile block is required when authentication_method is 'sqlserver_auth'")
			return
		}
		sqlProfile := &uapsiadbmodels.ArkUAPSIADBSqlServerAuthProfile{}

		// Global builtin roles
		if !data.SQLServerAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalBuiltin []string
			resp.Diagnostics.Append(data.SQLServerAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalBuiltin, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.GlobalBuiltinRoles = globalBuiltin
		}

		// Global custom roles
		if !data.SQLServerAuthProfile.GlobalCustomRoles.IsNull() {
			var globalCustom []string
			resp.Diagnostics.Append(data.SQLServerAuthProfile.GlobalCustomRoles.ElementsAs(ctx, &globalCustom, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.GlobalCustomRoles = globalCustom
		}

		// Database builtin roles
		if !data.SQLServerAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltin := make(map[string][]string)
			resp.Diagnostics.Append(data.SQLServerAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltin, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.DatabaseBuiltinRoles = dbBuiltin
		}

		// Database custom roles
		if !data.SQLServerAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustom := make(map[string][]string)
			resp.Diagnostics.Append(data.SQLServerAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustom, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.DatabaseCustomRoles = dbCustom
		}

		instanceTarget.SQLServerAuthProfile = sqlProfile

	case "rds_iam_user_auth":
		if data.RDSIAMUserAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "rds_iam_user_auth_profile block is required when authentication_method is 'rds_iam_user_auth'")
			return
		}
		instanceTarget.RDSIAMUserAuthProfile = &uapsiadbmodels.ArkUAPSIADBRDSIAMUserAuthProfile{
			DBUser: data.RDSIAMUserAuthProfile.DBUser.ValueString(),
		}

	default:
		resp.Diagnostics.AddError("Unsupported Authentication Method",
			fmt.Sprintf("Authentication method %q is not yet implemented", authMethod))
		return
	}

	tflog.Debug(ctx, "Built instance target", map[string]interface{}{
		"instance_name": instanceTarget.InstanceName,
		"instance_type": instanceTarget.InstanceType,
		"instance_id":   instanceTarget.InstanceID,
		"auth_method":   instanceTarget.AuthenticationMethod,
	})

	// Step 5: Append to appropriate workspace type targets (PRESERVE EXISTING)
	targets := policy.Targets[workspaceType]
	if targets.Instances == nil {
		targets.Instances = []uapsiadbmodels.ArkUAPSIADBInstanceTarget{}
	}
	targets.Instances = append(targets.Instances, *instanceTarget)
	policy.Targets[workspaceType] = targets

	// Step 6: Update policy with modified workspace type
	tflog.Debug(ctx, "Updating policy with new database assignment")

	// CRITICAL: API only accepts ONE workspace type in Targets per update
	// Send full policy structure (metadata, principals, conditions) but ONLY the workspace type we modified
	updatePolicy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		ArkUAPSIACommonAccessPolicy: policy.ArkUAPSIACommonAccessPolicy,
		Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
			workspaceType: policy.Targets[workspaceType], // ONLY the workspace type we're modifying
		},
	}

	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		_, updateErr := r.providerData.UAPClient.Db().UpdatePolicy(updatePolicy)
		return updateErr
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "update policy"))
		return
	}

	// Step 7: Store composite ID and timestamp
	data.ID = types.StringValue(buildCompositeID(policyID, databaseID))
	data.LastModified = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	LogOperationSuccess(ctx, "create", "policy_database_assignment", data.ID.ValueString())
}

func (r *PolicyDatabaseAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.PolicyDatabaseAssignmentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	LogOperationStart(ctx, "read", "policy_database_assignment")

	// Step 1: Parse composite ID
	policyID, databaseID, err := parseCompositeID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Failed to parse resource ID: %s", err.Error()),
		)
		return
	}

	// Step 2: Fetch policy
	tflog.Debug(ctx, "Fetching policy for read", map[string]interface{}{
		"policy_id": policyID,
	})

	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "fetch policy"))
		return
	}

	// Step 3: Search for database in policy (drift detection)
	target, _, found := findDatabaseInPolicyWithType(policy, databaseID)
	if !found {
		tflog.Warn(ctx, "Database not found in policy - resource deleted outside Terraform", map[string]interface{}{
			"policy_id":   policyID,
			"database_id": databaseID,
		})
		LogDriftDetected(ctx, "policy_database_assignment", data.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	// Step 4: Verify database workspace still exists
	// Convert string to int for database fetch
	databaseIDInt, err := strconv.Atoi(databaseID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Database ID",
			fmt.Sprintf("Database workspace ID must be a valid integer: %s", err.Error()))
		return
	}

	_, err = r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
		ID: databaseIDInt,
	})
	if err != nil {
		// Database workspace deleted - remove assignment
		tflog.Warn(ctx, "Database workspace not found - removing assignment", map[string]interface{}{
			"database_id": databaseID,
		})
		LogDriftDetected(ctx, "policy_database_assignment", data.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	// Note: Platform drift detection removed - all databases use "FQDN/IP" target set
	// regardless of cloud provider, so platform changes don't affect assignment validity

	// Step 5: Update state with current configuration
	data.PolicyID = types.StringValue(policyID)
	data.DatabaseWorkspaceID = types.StringValue(databaseID)
	data.AuthenticationMethod = types.StringValue(target.AuthenticationMethod)
	data.LastModified = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	// Update profile based on authentication method
	switch target.AuthenticationMethod {
	case "db_auth":
		if target.DBAuthProfile != nil {
			rolesList, diags := types.ListValueFrom(ctx, types.StringType, target.DBAuthProfile.Roles)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.DBAuthProfile = &models.DBAuthProfileModel{Roles: rolesList}
		}

	case "ldap_auth":
		if target.LDAPAuthProfile != nil {
			assignGroupsList, diags := types.ListValueFrom(ctx, types.StringType, target.LDAPAuthProfile.AssignGroups)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.LDAPAuthProfile = &models.LDAPAuthProfileModel{AssignGroups: assignGroupsList}
		}

	case "oracle_auth":
		if target.OracleAuthProfile != nil {
			rolesList, diags := types.ListValueFrom(ctx, types.StringType, target.OracleAuthProfile.Roles)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.OracleAuthProfile = &models.OracleAuthProfileModel{
				Roles:       rolesList,
				DbaRole:     types.BoolValue(target.OracleAuthProfile.DbaRole),
				SysdbaRole:  types.BoolValue(target.OracleAuthProfile.SysdbaRole),
				SysoperRole: types.BoolValue(target.OracleAuthProfile.SysoperRole),
			}
		}

	case "mongo_auth":
		if target.MongoAuthProfile != nil {
			mongoModel := &models.MongoAuthProfileModel{}

			// Global builtin roles
			if len(target.MongoAuthProfile.GlobalBuiltinRoles) > 0 {
				globalList, diags := types.ListValueFrom(ctx, types.StringType, target.MongoAuthProfile.GlobalBuiltinRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				mongoModel.GlobalBuiltinRoles = globalList
			}

			// Database builtin roles
			if len(target.MongoAuthProfile.DatabaseBuiltinRoles) > 0 {
				dbBuiltinMap, diags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, target.MongoAuthProfile.DatabaseBuiltinRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				mongoModel.DatabaseBuiltinRoles = dbBuiltinMap
			}

			// Database custom roles
			if len(target.MongoAuthProfile.DatabaseCustomRoles) > 0 {
				dbCustomMap, diags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, target.MongoAuthProfile.DatabaseCustomRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				mongoModel.DatabaseCustomRoles = dbCustomMap
			}

			data.MongoAuthProfile = mongoModel
		}

	case "sqlserver_auth":
		if target.SQLServerAuthProfile != nil {
			sqlModel := &models.SQLServerAuthProfileModel{}

			// Global builtin roles
			if len(target.SQLServerAuthProfile.GlobalBuiltinRoles) > 0 {
				globalBuiltinList, diags := types.ListValueFrom(ctx, types.StringType, target.SQLServerAuthProfile.GlobalBuiltinRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				sqlModel.GlobalBuiltinRoles = globalBuiltinList
			}

			// Global custom roles
			if len(target.SQLServerAuthProfile.GlobalCustomRoles) > 0 {
				globalCustomList, diags := types.ListValueFrom(ctx, types.StringType, target.SQLServerAuthProfile.GlobalCustomRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				sqlModel.GlobalCustomRoles = globalCustomList
			}

			// Database builtin roles
			if len(target.SQLServerAuthProfile.DatabaseBuiltinRoles) > 0 {
				dbBuiltinMap, diags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, target.SQLServerAuthProfile.DatabaseBuiltinRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				sqlModel.DatabaseBuiltinRoles = dbBuiltinMap
			}

			// Database custom roles
			if len(target.SQLServerAuthProfile.DatabaseCustomRoles) > 0 {
				dbCustomMap, diags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, target.SQLServerAuthProfile.DatabaseCustomRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				sqlModel.DatabaseCustomRoles = dbCustomMap
			}

			data.SQLServerAuthProfile = sqlModel
		}

	case "rds_iam_user_auth":
		if target.RDSIAMUserAuthProfile != nil {
			data.RDSIAMUserAuthProfile = &models.RDSIAMUserAuthProfileModel{
				DBUser: types.StringValue(target.RDSIAMUserAuthProfile.DBUser),
			}
		}
	}

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	LogOperationSuccess(ctx, "read", "policy_database_assignment", data.ID.ValueString())
}

func (r *PolicyDatabaseAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.PolicyDatabaseAssignmentModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	LogOperationStart(ctx, "update", "policy_database_assignment")

	// Step 1: Parse composite ID
	policyID, databaseID, err := parseCompositeID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Failed to parse resource ID: %s", err.Error()),
		)
		return
	}

	// Step 2: Fetch policy (READ-MODIFY-WRITE pattern)
	tflog.Debug(ctx, "Fetching policy for update", map[string]interface{}{
		"policy_id": policyID,
	})

	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})
	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "fetch policy"))
		return
	}

	// Step 3: Find database by InstanceID
	target, workspaceType, found := findDatabaseInPolicyWithType(policy, databaseID)
	if !found {
		resp.Diagnostics.AddError(
			"Database Not Found in Policy",
			fmt.Sprintf("Database %s not found in policy %s. The resource may have been deleted outside Terraform.", databaseID, policyID),
		)
		return
	}

	// Step 4: Update authentication method and profile in place (PRESERVE OTHER DATABASES)
	target.AuthenticationMethod = data.AuthenticationMethod.ValueString()

	// Update profile based on authentication method
	// Clear all profiles first
	target.DBAuthProfile = nil
	target.LDAPAuthProfile = nil
	target.OracleAuthProfile = nil
	target.MongoAuthProfile = nil
	target.SQLServerAuthProfile = nil
	target.RDSIAMUserAuthProfile = nil

	authMethod := data.AuthenticationMethod.ValueString()
	switch authMethod {
	case "db_auth":
		if data.DBAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "db_auth_profile block is required when authentication_method is 'db_auth'")
			return
		}
		var roles []string
		resp.Diagnostics.Append(data.DBAuthProfile.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.DBAuthProfile = &uapsiadbmodels.ArkUAPSIADBDBAuthProfile{Roles: roles}

	case "ldap_auth":
		if data.LDAPAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "ldap_auth_profile block is required when authentication_method is 'ldap_auth'")
			return
		}
		var assignGroups []string
		resp.Diagnostics.Append(data.LDAPAuthProfile.AssignGroups.ElementsAs(ctx, &assignGroups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.LDAPAuthProfile = &uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile{AssignGroups: assignGroups}

	case "oracle_auth":
		if data.OracleAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "oracle_auth_profile block is required when authentication_method is 'oracle_auth'")
			return
		}
		var roles []string
		resp.Diagnostics.Append(data.OracleAuthProfile.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.OracleAuthProfile = &uapsiadbmodels.ArkUAPSIADBOracleAuthProfile{
			Roles:       roles,
			DbaRole:     data.OracleAuthProfile.DbaRole.ValueBool(),
			SysdbaRole:  data.OracleAuthProfile.SysdbaRole.ValueBool(),
			SysoperRole: data.OracleAuthProfile.SysoperRole.ValueBool(),
		}

	case "mongo_auth":
		if data.MongoAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "mongo_auth_profile block is required when authentication_method is 'mongo_auth'")
			return
		}
		mongoProfile := &uapsiadbmodels.ArkUAPSIADBMongoAuthProfile{}

		if !data.MongoAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalRoles []string
			resp.Diagnostics.Append(data.MongoAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.GlobalBuiltinRoles = globalRoles
		}

		if !data.MongoAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltinRoles := make(map[string][]string)
			resp.Diagnostics.Append(data.MongoAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltinRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.DatabaseBuiltinRoles = dbBuiltinRoles
		}

		if !data.MongoAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustomRoles := make(map[string][]string)
			resp.Diagnostics.Append(data.MongoAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustomRoles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			mongoProfile.DatabaseCustomRoles = dbCustomRoles
		}

		target.MongoAuthProfile = mongoProfile

	case "sqlserver_auth":
		if data.SQLServerAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "sqlserver_auth_profile block is required when authentication_method is 'sqlserver_auth'")
			return
		}
		sqlProfile := &uapsiadbmodels.ArkUAPSIADBSqlServerAuthProfile{}

		if !data.SQLServerAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalBuiltin []string
			resp.Diagnostics.Append(data.SQLServerAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalBuiltin, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.GlobalBuiltinRoles = globalBuiltin
		}

		if !data.SQLServerAuthProfile.GlobalCustomRoles.IsNull() {
			var globalCustom []string
			resp.Diagnostics.Append(data.SQLServerAuthProfile.GlobalCustomRoles.ElementsAs(ctx, &globalCustom, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.GlobalCustomRoles = globalCustom
		}

		if !data.SQLServerAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltin := make(map[string][]string)
			resp.Diagnostics.Append(data.SQLServerAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltin, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.DatabaseBuiltinRoles = dbBuiltin
		}

		if !data.SQLServerAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustom := make(map[string][]string)
			resp.Diagnostics.Append(data.SQLServerAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustom, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sqlProfile.DatabaseCustomRoles = dbCustom
		}

		target.SQLServerAuthProfile = sqlProfile

	case "rds_iam_user_auth":
		if data.RDSIAMUserAuthProfile == nil {
			resp.Diagnostics.AddError("Missing Profile", "rds_iam_user_auth_profile block is required when authentication_method is 'rds_iam_user_auth'")
			return
		}
		target.RDSIAMUserAuthProfile = &uapsiadbmodels.ArkUAPSIADBRDSIAMUserAuthProfile{
			DBUser: data.RDSIAMUserAuthProfile.DBUser.ValueString(),
		}

	default:
		resp.Diagnostics.AddError("Unsupported Authentication Method",
			fmt.Sprintf("Authentication method %q is not yet implemented", authMethod))
		return
	}

	// Update the target in the policy's targets map
	targets := policy.Targets[workspaceType]
	for i := range targets.Instances {
		if targets.Instances[i].InstanceID == databaseID {
			targets.Instances[i] = *target
			break
		}
	}
	policy.Targets[workspaceType] = targets

	tflog.Debug(ctx, "Updated database assignment in policy", map[string]interface{}{
		"policy_id":   policyID,
		"database_id": databaseID,
		"auth_method": authMethod,
	})

	// Step 5: Write policy back with modified workspace type
	// CRITICAL: API only accepts ONE workspace type in Targets per update
	// Send full policy structure (metadata, principals, conditions) but ONLY the workspace type we modified
	updatePolicy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		ArkUAPSIACommonAccessPolicy: policy.ArkUAPSIACommonAccessPolicy,
		Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
			workspaceType: policy.Targets[workspaceType], // ONLY the workspace type we're modifying
		},
	}

	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		_, updateErr := r.providerData.UAPClient.Db().UpdatePolicy(updatePolicy)
		return updateErr
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "update policy"))
		return
	}

	// Update timestamp
	data.LastModified = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	LogOperationSuccess(ctx, "update", "policy_database_assignment", data.ID.ValueString())
}

func (r *PolicyDatabaseAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.PolicyDatabaseAssignmentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	LogOperationStart(ctx, "delete", "policy_database_assignment")

	// Step 1: Parse composite ID
	policyID, databaseID, err := parseCompositeID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Failed to parse resource ID: %s", err.Error()),
		)
		return
	}

	// Step 2: Fetch policy (READ-MODIFY-WRITE pattern)
	tflog.Debug(ctx, "Fetching policy for delete", map[string]interface{}{
		"policy_id": policyID,
	})

	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})
	if err != nil {
		// If policy not found, resource is already gone - success
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Info(ctx, "Policy not found - considering delete successful")
			return
		}
		resp.Diagnostics.Append(client.MapError(err, "fetch policy"))
		return
	}

	// Step 3: Find and remove ONLY our database from targets (PRESERVE OTHER DATABASES)
	_, workspaceType, found := findDatabaseInPolicyWithType(policy, databaseID)
	if !found {
		// Database not in policy - already deleted, consider success
		tflog.Info(ctx, "Database not found in policy - considering delete successful", map[string]interface{}{
			"policy_id":   policyID,
			"database_id": databaseID,
		})
		return
	}

	// Remove the database from the instances array
	targets := policy.Targets[workspaceType]
	newInstances := make([]uapsiadbmodels.ArkUAPSIADBInstanceTarget, 0, len(targets.Instances))
	for _, instance := range targets.Instances {
		if instance.InstanceID != databaseID {
			newInstances = append(newInstances, instance)
		}
	}
	targets.Instances = newInstances
	policy.Targets[workspaceType] = targets

	tflog.Debug(ctx, "Removed database from policy targets", map[string]interface{}{
		"policy_id":       policyID,
		"database_id":     databaseID,
		"workspace_type":  workspaceType,
		"remaining_count": len(newInstances),
	})

	// Step 4: Write policy back (API only accepts ONE workspace type at a time)
	// CRITICAL: API requires Targets to contain exactly ONE workspace type
	updatePolicy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		ArkUAPSIACommonAccessPolicy: policy.ArkUAPSIACommonAccessPolicy,
		Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
			workspaceType: policy.Targets[workspaceType],
		},
	}

	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		_, updateErr := r.providerData.UAPClient.Db().UpdatePolicy(updatePolicy)
		return updateErr
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "update policy after delete"))
		return
	}

	LogOperationSuccess(ctx, "delete", "policy_database_assignment", data.ID.ValueString())
}

func (r *PolicyDatabaseAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: policy-id:database-id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions (Tasks 13-14)

// buildCompositeID creates a composite ID from policy ID and database ID
func buildCompositeID(policyID, dbID string) string {
	return fmt.Sprintf("%s:%s", policyID, dbID)
}

// parseCompositeID splits a composite ID into policy ID and database ID
func parseCompositeID(id string) (policyID, dbID string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid composite ID format: expected 'policy-id:database-id', got '%s'", id)
	}
	return parts[0], parts[1], nil
}

// determineWorkspaceType returns the policy workspace type for database targets
// ALL database workspaces use "FQDN/IP" target set regardless of cloud provider
// The cloud_provider field is metadata only and doesn't affect policy target sets
// This is validated by:
// 1. SDK annotation in ark_uap_sia_db_access_policy.go: choices:"FQDN/IP"
// 2. UI behavior: Azure databases successfully use "FQDN/IP" target set
// 3. API validation: Only "FQDN/IP" key is allowed in targets dictionary
func determineWorkspaceType(platform string) string {
	return "FQDN/IP"
}

// mapDatabaseTypeToInstanceType converts database_type to policy InstanceType
// Handles compound types like "postgres-aws-rds" by extracting base engine
// Returns capitalized format expected by ArkUAPSIADBInstanceTarget
//
// Transformation logic (validated by Gemini 2025-10-27):
// 1. Extract base engine (substring before first hyphen)
// 2. Map to capitalized InstanceType value
// 3. Fallback to "Unknown" for unrecognized types
func mapDatabaseTypeToInstanceType(databaseType string) string {
	// Extract base engine (e.g., "postgres" from "postgres-aws-rds")
	baseEngine := databaseType
	if idx := strings.Index(databaseType, "-"); idx > 0 {
		baseEngine = databaseType[:idx]
	}

	// Map to capitalized InstanceType expected by policy API
	switch strings.ToLower(baseEngine) {
	case "postgres":
		return "Postgres"
	case "mysql":
		return "MySQL"
	case "mssql":
		return "MSSQL"
	case "oracle":
		return "Oracle"
	case "mariadb":
		return "MariaDB"
	case "db2":
		return "DB2"
	case "mongo":
		return "Mongo"
	default:
		return "Unknown" // Graceful fallback for new/unrecognized types
	}
}

// findDatabaseInPolicy searches all workspace types for a database by InstanceID
// Returns the target if found, nil otherwise
// Used for idempotency checking and READ operations
func findDatabaseInPolicy(policy *uapsiadbmodels.ArkUAPSIADBAccessPolicy, databaseID string) *uapsiadbmodels.ArkUAPSIADBInstanceTarget {
	if policy.Targets == nil {
		return nil
	}

	// Search all workspace types
	for _, targets := range policy.Targets {
		for i := range targets.Instances {
			if targets.Instances[i].InstanceID == databaseID {
				return &targets.Instances[i]
			}
		}
	}

	return nil
}

// findDatabaseInPolicyWithType searches for a database and returns both the target and workspace type
// Returns (target, workspaceType, found)
func findDatabaseInPolicyWithType(policy *uapsiadbmodels.ArkUAPSIADBAccessPolicy, databaseID string) (*uapsiadbmodels.ArkUAPSIADBInstanceTarget, string, bool) {
	if policy.Targets == nil {
		return nil, "", false
	}

	// Search all workspace types
	for workspaceType, targets := range policy.Targets {
		for i := range targets.Instances {
			if targets.Instances[i].InstanceID == databaseID {
				return &targets.Instances[i], workspaceType, true
			}
		}
	}

	return nil, "", false
}
