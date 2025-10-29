package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
var _ resource.Resource = &DatabasePolicyResource{}
var _ resource.ResourceWithImportState = &DatabasePolicyResource{}
var _ resource.ResourceWithValidateConfig = &DatabasePolicyResource{}

func NewDatabasePolicyResource() resource.Resource {
	return &DatabasePolicyResource{}
}

// DatabasePolicyResource defines the resource implementation.
type DatabasePolicyResource struct {
	providerData *ProviderData
}

func (r *DatabasePolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_policy"
}

func (r *DatabasePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CyberArk SIA database access policy including metadata and access conditions. " +
			"This resource manages policy-level configuration only. Use `cyberarksia_database_policy_principal_assignment` " +
			"to assign principals (users/groups/roles) and `cyberarksia_database_policy_assignment` to assign database workspaces.\n\n" +
			"**Pattern**: Follows the modular assignment pattern for distributed team workflows - security teams manage policies " +
			"and principals, application teams manage database assignments independently.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Policy identifier (same as policy_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "Unique policy identifier (UUID, API-generated).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Policy name (1-200 characters, unique per tenant). **ForceNew**: Changing this creates a new policy.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Policy description (max 200 characters).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(200),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Policy status. Valid values: `active` (enabled), `suspended` (disabled). **Note**: `expired`, `validating`, and `error` are server-managed statuses and cannot be set by users.",
				Required:            true,
				Validators: []validator.String{
					validators.PolicyStatus(),
				},
			},
			"delegation_classification": schema.StringAttribute{
				MarkdownDescription: "Delegation classification. Valid values: `restricted`/`Restricted`, `unrestricted`/`Unrestricted`. Default: `unrestricted`. Note: API returns capitalized values.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Unrestricted"),
				Validators: []validator.String{
					stringvalidator.OneOf("restricted", "unrestricted", "Restricted", "Unrestricted"),
				},
			},
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "Timezone for access window conditions (max 50 characters). Supports IANA timezone names (e.g., `America/New_York`) or GMT offsets (e.g., `GMT+05:00`). Default: `GMT`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GMT"),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(50),
				},
			},
			"policy_tags": schema.ListAttribute{
				MarkdownDescription: "List of tags for policy organization (max 20 tags).",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(20),
				},
			},
			"last_modified": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last modification to the policy.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.SingleNestedAttribute{
				MarkdownDescription: "Metadata about policy creation (set by API).",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						MarkdownDescription: "Username of the user who created the policy.",
						Computed:            true,
					},
					"timestamp": schema.StringAttribute{
						MarkdownDescription: "Creation timestamp in ISO 8601 format.",
						Computed:            true,
					},
				},
			},
			"updated_on": schema.SingleNestedAttribute{
				MarkdownDescription: "Metadata about the last policy update (set by API).",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						MarkdownDescription: "Username of the user who last updated the policy.",
						Computed:            true,
					},
					"timestamp": schema.StringAttribute{
						MarkdownDescription: "Last update timestamp in ISO 8601 format.",
						Computed:            true,
					},
				},
			},
		},

		Blocks: map[string]schema.Block{
			"target_database": schema.ListNestedBlock{
				MarkdownDescription: "Database workspace assignment (repeatable block). **Required**: At least 1 target_database block is required. " +
					"Follows familiar Terraform patterns (aws_security_group ingress/egress). " +
					"Use `lifecycle { ignore_changes = [target_database] }` if managing assignments via separate `cyberarksia_policy_database_assignment` resources.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_workspace_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the database workspace to assign.",
							Required:            true,
						},
						"authentication_method": schema.StringAttribute{
							MarkdownDescription: "Authentication method. Valid values: `db_auth`, `ldap_auth`, `oracle_auth`, `mongo_auth`, `sqlserver_auth`, `rds_iam_user_auth`.",
							Required:            true,
							Validators: []validator.String{
								validators.AuthenticationMethod(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"db_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "Database authentication profile. **Required** when `authentication_method` is `db_auth`.",
							Attributes: map[string]schema.Attribute{
								"roles": schema.ListAttribute{
									MarkdownDescription: "List of database roles to assign.",
									Optional:            true,
									ElementType:         types.StringType,
								},
							},
						},
						"ldap_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "LDAP authentication profile. **Required** when `authentication_method` is `ldap_auth`.",
							Attributes: map[string]schema.Attribute{
								"assign_groups": schema.ListAttribute{
									MarkdownDescription: "List of LDAP groups to assign.",
									Optional:            true,
									ElementType:         types.StringType,
								},
							},
						},
						"oracle_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "Oracle authentication profile. **Required** when `authentication_method` is `oracle_auth`.",
							Attributes: map[string]schema.Attribute{
								"roles": schema.ListAttribute{
									MarkdownDescription: "List of Oracle roles to assign.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"dba_role": schema.BoolAttribute{
									MarkdownDescription: "Grant DBA role.",
									Optional:            true,
								},
								"sysdba_role": schema.BoolAttribute{
									MarkdownDescription: "Grant SYSDBA role.",
									Optional:            true,
								},
								"sysoper_role": schema.BoolAttribute{
									MarkdownDescription: "Grant SYSOPER role.",
									Optional:            true,
								},
							},
						},
						"mongo_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "MongoDB authentication profile. **Required** when `authentication_method` is `mongo_auth`.",
							Attributes: map[string]schema.Attribute{
								"global_builtin_roles": schema.ListAttribute{
									MarkdownDescription: "List of global built-in roles.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"database_builtin_roles": schema.MapAttribute{
									MarkdownDescription: "Map of database names to built-in roles.",
									Optional:            true,
									ElementType:         types.ListType{ElemType: types.StringType},
								},
								"database_custom_roles": schema.MapAttribute{
									MarkdownDescription: "Map of database names to custom roles.",
									Optional:            true,
									ElementType:         types.ListType{ElemType: types.StringType},
								},
							},
						},
						"sqlserver_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "SQL Server authentication profile. **Required** when `authentication_method` is `sqlserver_auth`.",
							Attributes: map[string]schema.Attribute{
								"global_builtin_roles": schema.ListAttribute{
									MarkdownDescription: "List of global built-in roles.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"global_custom_roles": schema.ListAttribute{
									MarkdownDescription: "List of global custom roles.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"database_builtin_roles": schema.MapAttribute{
									MarkdownDescription: "Map of database names to built-in roles.",
									Optional:            true,
									ElementType:         types.ListType{ElemType: types.StringType},
								},
								"database_custom_roles": schema.MapAttribute{
									MarkdownDescription: "Map of database names to custom roles.",
									Optional:            true,
									ElementType:         types.ListType{ElemType: types.StringType},
								},
							},
						},
						"rds_iam_user_auth_profile": schema.SingleNestedBlock{
							MarkdownDescription: "RDS IAM User authentication profile. **Required** when `authentication_method` is `rds_iam_user_auth`.",
							Attributes: map[string]schema.Attribute{
								"db_user": schema.StringAttribute{
									MarkdownDescription: "Database user for RDS IAM authentication.",
									Optional:            true,
								},
							},
						},
					},
				},
			},
			"principal": schema.ListNestedBlock{
				MarkdownDescription: "Principal assignment (repeatable block). **Required**: At least 1 principal block is required. " +
					"Follows familiar Terraform patterns (aws_security_group ingress/egress). " +
					"Use `lifecycle { ignore_changes = [principal] }` if managing assignments via separate `cyberarksia_database_policy_principal_assignment` resources.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"principal_id": schema.StringAttribute{
							MarkdownDescription: "Principal identifier in UUID format (e.g., `c2c7bcc6-9560-44e0-8dff-5be221cd37ee`). This is the unique identifier returned by the SIA API.",
							Required:            true,
							Validators: []validator.String{
								validators.UUID(),
							},
						},
						"principal_type": schema.StringAttribute{
							MarkdownDescription: "Principal type. Valid values: `USER`, `GROUP`, `ROLE`.",
							Required:            true,
							Validators: []validator.String{
								validators.PrincipalType(),
							},
						},
						"principal_name": schema.StringAttribute{
							MarkdownDescription: "Principal name in email format (e.g., `user@example.com` or `tim.schindler@cyberark.cloud.40562`).",
							Required:            true,
							Validators: []validator.String{
								validators.EmailLike(),
							},
						},
						"source_directory_name": schema.StringAttribute{
							MarkdownDescription: "Source identity directory name (max 50 characters). **Required** for USER and GROUP types.",
							Optional:            true,
						},
						"source_directory_id": schema.StringAttribute{
							MarkdownDescription: "Source identity directory ID. **Required** for USER and GROUP types.",
							Optional:            true,
						},
					},
				},
			},
			"time_frame": schema.SingleNestedBlock{
				MarkdownDescription: "Policy validity period. **Optional**: If not specified, policy never expires (valid indefinitely). When specified, both `from_time` and `to_time` must be provided.",
				Attributes: map[string]schema.Attribute{
					"from_time": schema.StringAttribute{
						MarkdownDescription: "Start time (ISO 8601 format, e.g., `2024-01-01T00:00:00Z`). Required when `time_frame` block is present.",
						Optional:            true,
					},
					"to_time": schema.StringAttribute{
						MarkdownDescription: "End time (ISO 8601 format, e.g., `2024-12-31T23:59:59Z`). Required when `time_frame` block is present.",
						Optional:            true,
					},
				},
			},
			"conditions": schema.SingleNestedBlock{
				MarkdownDescription: "Policy access conditions (session limits, idle timeouts, time windows).",
				Attributes: map[string]schema.Attribute{
					"max_session_duration": schema.Int64Attribute{
						MarkdownDescription: "Maximum session duration in hours (1-24). **Required**.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 24),
						},
					},
					"idle_time": schema.Int64Attribute{
						MarkdownDescription: "Session idle timeout in minutes (1-120). Default: 10.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(10),
						Validators: []validator.Int64{
							int64validator.Between(1, 120),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"access_window": schema.SingleNestedBlock{
						MarkdownDescription: "Time-based access restrictions (days and hours).",
						Attributes: map[string]schema.Attribute{
							"days_of_the_week": schema.SetAttribute{
								MarkdownDescription: "Days access is allowed (0=Sunday through 6=Saturday). Specify days in any order - order is automatically normalized. Example: `[1, 2, 3, 4, 5]` for weekdays.",
								Required:            true,
								ElementType:         types.Int64Type,
								Validators: []validator.Set{
									setvalidator.ValueInt64sAre(int64validator.Between(0, 6)), // 0=Sunday through 6=Saturday (0-indexed)
									setvalidator.SizeBetween(1, 7),                             // At least 1 day required, max 7 days (e.g., all week = [0,1,2,3,4,5,6])
								},
							},
							"from_hour": schema.StringAttribute{
								MarkdownDescription: "Start time in HH:MM format (e.g., `09:00`).",
								Required:            true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										mustCompileRegex(`^([01]\d|2[0-3]):([0-5]\d)$`),
										"must be in HH:MM format (e.g., 09:00)",
									),
								},
							},
							"to_hour": schema.StringAttribute{
								MarkdownDescription: "End time in HH:MM format (e.g., `17:00`).",
								Required:            true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										mustCompileRegex(`^([01]\d|2[0-3]):([0-5]\d)$`),
										"must be in HH:MM format (e.g., 17:00)",
									),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *DatabasePolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabasePolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.DatabasePolicyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate at least 1 target database
	if len(data.TargetDatabase) == 0 {
		resp.Diagnostics.AddError(
			"Missing Target Databases",
			"At least one target_database block is required. Database access policies must have at least one target database.",
		)
	}

	// Validate at least 1 principal
	if len(data.Principal) == 0 {
		resp.Diagnostics.AddError(
			"Missing Principals",
			"At least one principal block is required. Database access policies must have at least one principal (user/group/role).",
		)
	}

	// Validate principal directory requirements (USER/GROUP need source_directory)
	for i, principal := range data.Principal {
		principalType := principal.PrincipalType.ValueString()
		if principalType == "USER" || principalType == "GROUP" {
			if principal.SourceDirectoryName.IsNull() || principal.SourceDirectoryName.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Missing Source Directory Name",
					fmt.Sprintf("principals[%d]: source_directory_name is required for principal_type %s", i, principalType),
				)
			}
			if principal.SourceDirectoryID.IsNull() || principal.SourceDirectoryID.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Missing Source Directory ID",
					fmt.Sprintf("principals[%d]: source_directory_id is required for principal_type %s", i, principalType),
				)
			}
		}
	}

	// Validate authentication method profiles match
	for i, targetDB := range data.TargetDatabase {
		authMethod := targetDB.AuthenticationMethod.ValueString()
		switch authMethod {
		case "db_auth":
			if targetDB.DBAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: db_auth_profile block is required when authentication_method is 'db_auth'", i),
				)
			}
		case "ldap_auth":
			if targetDB.LDAPAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: ldap_auth_profile block is required when authentication_method is 'ldap_auth'", i),
				)
			}
		case "oracle_auth":
			if targetDB.OracleAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: oracle_auth_profile block is required when authentication_method is 'oracle_auth'", i),
				)
			}
		case "mongo_auth":
			if targetDB.MongoAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: mongo_auth_profile block is required when authentication_method is 'mongo_auth'", i),
				)
			}
		case "sqlserver_auth":
			if targetDB.SQLServerAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: sqlserver_auth_profile block is required when authentication_method is 'sqlserver_auth'", i),
				)
			}
		case "rds_iam_user_auth":
			if targetDB.RDSIAMUserAuthProfile == nil {
				resp.Diagnostics.AddError(
					"Missing Authentication Profile",
					fmt.Sprintf("target_databases[%d]: rds_iam_user_auth_profile block is required when authentication_method is 'rds_iam_user_auth'", i),
				)
			}
		}
	}
}

func (r *DatabasePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.DatabasePolicyModel

	// Read Terraform plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform state to SDK policy (metadata only)
	policy := data.ToSDK()

	// Build inline target databases
	if len(data.TargetDatabase) > 0 {
		policy.Targets = make(map[string]uapsiadbmodels.ArkUAPSIADBTargets)

		for i, targetDB := range data.TargetDatabase {
			// Fetch database workspace to get instance details
			databaseID := targetDB.DatabaseWorkspaceID.ValueString()
			databaseIDInt, err := strconv.Atoi(databaseID)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Database ID",
					fmt.Sprintf("target_databases[%d]: database_workspace_id must be a valid integer: %s", i, err.Error()),
				)
				return
			}

			database, err := r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
				ID: databaseIDInt,
			})
			if err != nil {
				resp.Diagnostics.Append(client.MapError(err, fmt.Sprintf("fetch database workspace for target_databases[%d]", i)))
				return
			}

			// Determine workspace type (always "FQDN/IP" for all databases)
			workspaceType := "FQDN/IP"

			// Build instance target with authentication profile
			instanceTarget, buildErr := buildInstanceTarget(ctx, database, targetDB)
			if buildErr != nil {
				resp.Diagnostics.AddError(
					"Failed to Build Target",
					fmt.Sprintf("target_databases[%d]: %s", i, buildErr.Error()),
				)
				return
			}

			// Add to targets map
			targets := policy.Targets[workspaceType]
			targets.Instances = append(targets.Instances, *instanceTarget)
			policy.Targets[workspaceType] = targets

			tflog.Debug(ctx, "Added target database to policy", map[string]interface{}{
				"database_id":    databaseID,
				"workspace_type": workspaceType,
				"auth_method":    targetDB.AuthenticationMethod.ValueString(),
			})
		}
	}

	// Build inline principals
	if len(data.Principal) > 0 {
		policy.Principals = make([]uapcommonmodels.ArkUAPPrincipal, len(data.Principal))
		for i, principal := range data.Principal {
			policy.Principals[i] = uapcommonmodels.ArkUAPPrincipal{
				ID:                  principal.PrincipalID.ValueString(),
				Name:                principal.PrincipalName.ValueString(),
				Type:                principal.PrincipalType.ValueString(),
				SourceDirectoryName: principal.SourceDirectoryName.ValueString(),
				SourceDirectoryID:   principal.SourceDirectoryID.ValueString(),
			}

			tflog.Info(ctx, "SENDING PRINCIPAL TO API", map[string]interface{}{
				"principal_id":          principal.PrincipalID.ValueString(),
				"principal_name":        principal.PrincipalName.ValueString(),
				"principal_type":        principal.PrincipalType.ValueString(),
				"source_directory_name": principal.SourceDirectoryName.ValueString(),
				"source_directory_id":   principal.SourceDirectoryID.ValueString(),
			})
		}
	}

	// Create policy with retry logic
	var createdPolicy *uapsiadbmodels.ArkUAPSIADBAccessPolicy
	err := client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		var createErr error
		createdPolicy, createErr = r.providerData.UAPClient.Db().AddPolicy(policy)
		return createErr
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "create database policy"))
		return
	}

	// Log what API returned
	tflog.Info(ctx, "RECEIVED FROM API", map[string]interface{}{
		"principals_count": len(createdPolicy.Principals),
	})
	for i, p := range createdPolicy.Principals {
		tflog.Info(ctx, fmt.Sprintf("API RETURNED PRINCIPAL %d", i), map[string]interface{}{
			"id":   p.ID,
			"name": p.Name,
			"type": p.Type,
		})
	}

	// Set only the ID fields from API response
	// Don't call FromSDK() here - it tries to populate computed metadata fields
	// (created_by, updated_on) which causes "unknown value" errors during CREATE.
	// Terraform will automatically call Read() after Create() to populate all fields.
	data.ID = types.StringValue(createdPolicy.Metadata.PolicyID)
	data.PolicyID = types.StringValue(createdPolicy.Metadata.PolicyID)

	// Set last_modified to empty string (API doesn't return this field on create)
	data.LastModified = types.StringValue("")

	// Explicitly set computed metadata fields to null to avoid "unknown value" errors
	// These will be populated by the automatic Read() call after Create()
	changeInfoAttrTypes := map[string]attr.Type{
		"user":      types.StringType,
		"timestamp": types.StringType,
	}
	data.CreatedBy = types.ObjectNull(changeInfoAttrTypes)
	data.UpdatedOn = types.ObjectNull(changeInfoAttrTypes)

	tflog.Info(ctx, "Created database policy", map[string]interface{}{
		"policy_id":          data.PolicyID.ValueString(),
		"policy_name":        data.Name.ValueString(),
		"target_databases":   len(data.TargetDatabase),
		"principals":         len(data.Principal),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabasePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.DatabasePolicyModel

	// Read Terraform state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	// Fetch policy from API
	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})

	if err != nil {
		// If policy not found, remove from state
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(client.MapError(err, "read database policy"))
		return
	}

	// Update state with fetched policy
	if err := data.FromSDK(ctx, policy); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Policy Response",
			fmt.Sprintf("Failed to convert API response to state: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Read database policy", map[string]interface{}{
		"policy_id": data.PolicyID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabasePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.DatabasePolicyModel

	// Read Terraform plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	// Convert new state to SDK (metadata only)
	updatedPolicy := data.ToSDK()

	// Build inline target databases (if provided)
	if len(data.TargetDatabase) > 0 {
		updatedPolicy.Targets = make(map[string]uapsiadbmodels.ArkUAPSIADBTargets)

		for i, targetDB := range data.TargetDatabase {
			// Fetch database workspace to get instance details
			databaseID := targetDB.DatabaseWorkspaceID.ValueString()
			databaseIDInt, err := strconv.Atoi(databaseID)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Database ID",
					fmt.Sprintf("target_databases[%d]: database_workspace_id must be a valid integer: %s", i, err.Error()),
				)
				return
			}

			database, err := r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
				ID: databaseIDInt,
			})
			if err != nil {
				resp.Diagnostics.Append(client.MapError(err, fmt.Sprintf("fetch database workspace for target_databases[%d]", i)))
				return
			}

			// Determine workspace type (always "FQDN/IP" for all databases)
			workspaceType := "FQDN/IP"

			// Build instance target with authentication profile
			instanceTarget, buildErr := buildInstanceTarget(ctx, database, targetDB)
			if buildErr != nil {
				resp.Diagnostics.AddError(
					"Failed to Build Target",
					fmt.Sprintf("target_databases[%d]: %s", i, buildErr.Error()),
				)
				return
			}

			// Add to targets map
			targets := updatedPolicy.Targets[workspaceType]
			targets.Instances = append(targets.Instances, *instanceTarget)
			updatedPolicy.Targets[workspaceType] = targets

			tflog.Debug(ctx, "Updated target database in policy", map[string]interface{}{
				"database_id":    databaseID,
				"workspace_type": workspaceType,
				"auth_method":    targetDB.AuthenticationMethod.ValueString(),
			})
		}
	}

	// Build inline principals (if provided)
	if len(data.Principal) > 0 {
		updatedPolicy.Principals = make([]uapcommonmodels.ArkUAPPrincipal, len(data.Principal))
		for i, principal := range data.Principal {
			updatedPolicy.Principals[i] = uapcommonmodels.ArkUAPPrincipal{
				ID:                  principal.PrincipalID.ValueString(),
				Name:                principal.PrincipalName.ValueString(),
				Type:                principal.PrincipalType.ValueString(),
				SourceDirectoryName: principal.SourceDirectoryName.ValueString(),
				SourceDirectoryID:   principal.SourceDirectoryID.ValueString(),
			}

			tflog.Debug(ctx, "Updated principal in policy", map[string]interface{}{
				"principal_id":   principal.PrincipalID.ValueString(),
				"principal_type": principal.PrincipalType.ValueString(),
			})
		}
	}

	// Update policy with retry logic
	err := client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		_, err := r.providerData.UAPClient.Db().UpdatePolicy(updatedPolicy)
		return err
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "update database policy"))
		return
	}

	// Fetch updated policy to get computed fields
	refreshedPolicy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "refresh policy after update"))
		return
	}

	// Update state with refreshed policy
	if err := data.FromSDK(ctx, refreshedPolicy); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Policy Response",
			fmt.Sprintf("Failed to convert API response to state: %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Updated database policy", map[string]interface{}{
		"policy_id":        data.PolicyID.ValueString(),
		"target_databases": len(data.TargetDatabase),
		"principals":       len(data.Principal),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabasePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.DatabasePolicyModel

	// Read Terraform state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	// Delete policy with retry logic using workaround (ARK SDK v1.5.0 bug)
	// Note: API automatically cascades deletion to principals and targets
	err := client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		return client.DeleteDatabasePolicyDirect(ctx, r.providerData.AuthContext, policyID)
	})

	if err != nil {
		// If already deleted, treat as success
		if client.IsNotFoundError(err) {
			tflog.Info(ctx, "Policy already deleted", map[string]interface{}{
				"policy_id": policyID,
			})
			return
		}

		resp.Diagnostics.Append(client.MapError(err, "delete database policy"))
		return
	}

	tflog.Info(ctx, "Deleted database policy", map[string]interface{}{
		"policy_id": policyID,
	})
}

func (r *DatabasePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by policy ID
	policyID := req.ID

	// Fetch policy from API
	policy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err, "import database policy"))
		return
	}

	// Convert to state model
	var data models.DatabasePolicyModel
	if err := data.FromSDK(ctx, policy); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Policy Response",
			fmt.Sprintf("Failed to convert API response to state: %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Imported database policy", map[string]interface{}{
		"policy_id":   data.PolicyID.ValueString(),
		"policy_name": data.Name.ValueString(),
	})

	// Save imported state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mustCompileRegex compiles a regex pattern and panics if it fails (for use in validators)
func mustCompileRegex(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("failed to compile regex pattern %q: %v", pattern, err))
	}
	return re
}

// buildInstanceTarget creates an ArkUAPSIADBInstanceTarget from database workspace and assignment data
// This function handles all 6 authentication methods and their corresponding profiles
func buildInstanceTarget(ctx context.Context, database *dbmodels.ArkSIADBDatabase, targetDB models.InlineDatabaseAssignmentModel) (*uapsiadbmodels.ArkUAPSIADBInstanceTarget, error) {
	authMethod := targetDB.AuthenticationMethod.ValueString()

	instanceTarget := &uapsiadbmodels.ArkUAPSIADBInstanceTarget{
		InstanceName:         database.Name,
		InstanceType:         database.ProviderDetails.Family,
		InstanceID:           strconv.Itoa(database.ID),
		AuthenticationMethod: authMethod,
	}

	// Set the appropriate profile based on authentication method
	switch authMethod {
	case "db_auth":
		if targetDB.DBAuthProfile == nil {
			return nil, fmt.Errorf("db_auth_profile block is required when authentication_method is 'db_auth'")
		}
		var roles []string
		targetDB.DBAuthProfile.Roles.ElementsAs(ctx, &roles, false)
		instanceTarget.DBAuthProfile = &uapsiadbmodels.ArkUAPSIADBDBAuthProfile{Roles: roles}

	case "ldap_auth":
		if targetDB.LDAPAuthProfile == nil {
			return nil, fmt.Errorf("ldap_auth_profile block is required when authentication_method is 'ldap_auth'")
		}
		var assignGroups []string
		targetDB.LDAPAuthProfile.AssignGroups.ElementsAs(ctx, &assignGroups, false)
		instanceTarget.LDAPAuthProfile = &uapsiadbmodels.ArkUAPSIADBLDAPAuthProfile{AssignGroups: assignGroups}

	case "oracle_auth":
		if targetDB.OracleAuthProfile == nil {
			return nil, fmt.Errorf("oracle_auth_profile block is required when authentication_method is 'oracle_auth'")
		}
		var roles []string
		targetDB.OracleAuthProfile.Roles.ElementsAs(ctx, &roles, false)
		instanceTarget.OracleAuthProfile = &uapsiadbmodels.ArkUAPSIADBOracleAuthProfile{
			Roles:       roles,
			DbaRole:     targetDB.OracleAuthProfile.DbaRole.ValueBool(),
			SysdbaRole:  targetDB.OracleAuthProfile.SysdbaRole.ValueBool(),
			SysoperRole: targetDB.OracleAuthProfile.SysoperRole.ValueBool(),
		}

	case "mongo_auth":
		if targetDB.MongoAuthProfile == nil {
			return nil, fmt.Errorf("mongo_auth_profile block is required when authentication_method is 'mongo_auth'")
		}
		mongoProfile := &uapsiadbmodels.ArkUAPSIADBMongoAuthProfile{}

		// Global builtin roles
		if !targetDB.MongoAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalRoles []string
			targetDB.MongoAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalRoles, false)
			mongoProfile.GlobalBuiltinRoles = globalRoles
		}

		// Database builtin roles
		if !targetDB.MongoAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltinRoles := make(map[string][]string)
			targetDB.MongoAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltinRoles, false)
			mongoProfile.DatabaseBuiltinRoles = dbBuiltinRoles
		}

		// Database custom roles
		if !targetDB.MongoAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustomRoles := make(map[string][]string)
			targetDB.MongoAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustomRoles, false)
			mongoProfile.DatabaseCustomRoles = dbCustomRoles
		}

		instanceTarget.MongoAuthProfile = mongoProfile

	case "sqlserver_auth":
		if targetDB.SQLServerAuthProfile == nil {
			return nil, fmt.Errorf("sqlserver_auth_profile block is required when authentication_method is 'sqlserver_auth'")
		}
		sqlProfile := &uapsiadbmodels.ArkUAPSIADBSqlServerAuthProfile{}

		// Global builtin roles
		if !targetDB.SQLServerAuthProfile.GlobalBuiltinRoles.IsNull() {
			var globalBuiltin []string
			targetDB.SQLServerAuthProfile.GlobalBuiltinRoles.ElementsAs(ctx, &globalBuiltin, false)
			sqlProfile.GlobalBuiltinRoles = globalBuiltin
		}

		// Global custom roles
		if !targetDB.SQLServerAuthProfile.GlobalCustomRoles.IsNull() {
			var globalCustom []string
			targetDB.SQLServerAuthProfile.GlobalCustomRoles.ElementsAs(ctx, &globalCustom, false)
			sqlProfile.GlobalCustomRoles = globalCustom
		}

		// Database builtin roles
		if !targetDB.SQLServerAuthProfile.DatabaseBuiltinRoles.IsNull() {
			dbBuiltin := make(map[string][]string)
			targetDB.SQLServerAuthProfile.DatabaseBuiltinRoles.ElementsAs(ctx, &dbBuiltin, false)
			sqlProfile.DatabaseBuiltinRoles = dbBuiltin
		}

		// Database custom roles
		if !targetDB.SQLServerAuthProfile.DatabaseCustomRoles.IsNull() {
			dbCustom := make(map[string][]string)
			targetDB.SQLServerAuthProfile.DatabaseCustomRoles.ElementsAs(ctx, &dbCustom, false)
			sqlProfile.DatabaseCustomRoles = dbCustom
		}

		instanceTarget.SQLServerAuthProfile = sqlProfile

	case "rds_iam_user_auth":
		if targetDB.RDSIAMUserAuthProfile == nil {
			return nil, fmt.Errorf("rds_iam_user_auth_profile block is required when authentication_method is 'rds_iam_user_auth'")
		}
		instanceTarget.RDSIAMUserAuthProfile = &uapsiadbmodels.ArkUAPSIADBRDSIAMUserAuthProfile{
			DBUser: targetDB.RDSIAMUserAuthProfile.DBUser.ValueString(),
		}

	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", authMethod)
	}

	return instanceTarget, nil
}
