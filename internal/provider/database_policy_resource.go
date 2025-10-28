package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/client"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/validators"
	uapcommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/common/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DatabasePolicyResource{}
var _ resource.ResourceWithImportState = &DatabasePolicyResource{}

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
				MarkdownDescription: "Policy status. Valid values: `Active` (enabled), `Suspended` (disabled). **Note**: `Expired`, `Validating`, and `Error` are server-managed statuses and cannot be set by users.",
				Required:            true,
				Validators: []validator.String{
					validators.PolicyStatus(),
				},
			},
			"delegation_classification": schema.StringAttribute{
				MarkdownDescription: "Delegation classification. Valid values: `Restricted`, `Unrestricted`. Default: `Unrestricted`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Unrestricted"),
				Validators: []validator.String{
					stringvalidator.OneOf("Restricted", "Unrestricted"),
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
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{
			"time_frame": schema.SingleNestedBlock{
				MarkdownDescription: "Policy validity period. If not specified, policy is valid indefinitely.",
				Attributes: map[string]schema.Attribute{
					"from_time": schema.StringAttribute{
						MarkdownDescription: "Start time (ISO 8601 format, e.g., `2024-01-01T00:00:00Z`).",
						Required:            true,
					},
					"to_time": schema.StringAttribute{
						MarkdownDescription: "End time (ISO 8601 format, e.g., `2024-12-31T23:59:59Z`).",
						Required:            true,
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
							"days_of_the_week": schema.ListAttribute{
								MarkdownDescription: "Days access is allowed (0=Sunday through 6=Saturday). Example: `[1, 2, 3, 4, 5]` for weekdays.",
								Required:            true,
								ElementType:         types.Int64Type,
								Validators: []validator.List{
									listvalidator.ValueInt64sAre(int64validator.Between(0, 6)),
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
			"created_by": schema.SingleNestedBlock{
				MarkdownDescription: "User who created the policy (computed).",
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						MarkdownDescription: "Username.",
						Computed:            true,
					},
					"timestamp": schema.StringAttribute{
						MarkdownDescription: "Creation timestamp (ISO 8601).",
						Computed:            true,
					},
				},
			},
			"updated_on": schema.SingleNestedBlock{
				MarkdownDescription: "Last user who updated the policy (computed).",
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						MarkdownDescription: "Username.",
						Computed:            true,
					},
					"timestamp": schema.StringAttribute{
						MarkdownDescription: "Update timestamp (ISO 8601).",
						Computed:            true,
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

func (r *DatabasePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.DatabasePolicyModel

	// Read Terraform plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform state to SDK policy
	policy := data.ToSDK()

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
		resp.Diagnostics.Append(client.MapError(err)...)
		return
	}

	// Update state with created policy
	if err := data.FromSDK(ctx, createdPolicy); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Policy Response",
			fmt.Sprintf("Failed to convert API response to state: %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Created database policy", map[string]interface{}{
		"policy_id":   data.PolicyID.ValueString(),
		"policy_name": data.Name.ValueString(),
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

		resp.Diagnostics.Append(client.MapError(err)...)
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

	// Read-modify-write pattern: fetch full policy first
	existingPolicy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err)...)
		return
	}

	// Convert new state to SDK, preserving principals and targets
	updatedPolicy := data.ToSDK()
	updatedPolicy.Principals = existingPolicy.Principals       // Preserve principals
	updatedPolicy.Targets = existingPolicy.Targets             // Preserve targets

	// Update policy with retry logic
	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		return r.providerData.UAPClient.Db().UpdatePolicy(updatedPolicy)
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err)...)
		return
	}

	// Fetch updated policy to get computed fields
	refreshedPolicy, err := r.providerData.UAPClient.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
		PolicyID: policyID,
	})

	if err != nil {
		resp.Diagnostics.Append(client.MapError(err)...)
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
		"policy_id": data.PolicyID.ValueString(),
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

	// Delete policy with retry logic
	// Note: API automatically cascades deletion to principals and targets
	err := client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: client.DefaultMaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		return r.providerData.UAPClient.Db().DeletePolicy(&uapcommonmodels.ArkUAPDeletePolicyRequest{
			PolicyID: policyID,
		})
	})

	if err != nil {
		// If already deleted, treat as success
		if client.IsNotFoundError(err) {
			tflog.Info(ctx, "Policy already deleted", map[string]interface{}{
				"policy_id": policyID,
			})
			return
		}

		resp.Diagnostics.Append(client.MapError(err)...)
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
		resp.Diagnostics.Append(client.MapError(err)...)
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
