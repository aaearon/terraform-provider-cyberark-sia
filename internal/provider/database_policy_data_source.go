package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	uapcommonmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/common/models"
	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DatabasePolicyDataSource{}

func NewDatabasePolicyDataSource() datasource.DataSource {
	return &DatabasePolicyDataSource{}
}

// DatabasePolicyDataSource defines the data source implementation.
type DatabasePolicyDataSource struct {
	providerData *ProviderData
}

// DatabasePolicyDataSourceModel describes the data source data model.
type DatabasePolicyDataSourceModel struct {
	// Input (one required)
	PolicyID types.String `tfsdk:"policy_id"`
	Name     types.String `tfsdk:"name"`

	// Computed
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
}

func (d *DatabasePolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_policy"
}

func (d *DatabasePolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches an existing SIA access policy by ID or name. Use this data source to reference policies when creating policy database assignments.",

		Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the policy. Either `policy_id` or `name` must be specified.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the policy. Either `policy_id` or `name` must be specified.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The policy ID (same as `policy_id` when looking up by ID, or the resolved ID when looking up by name).",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the policy.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the policy (e.g., 'Active', 'Inactive').",
				Computed:            true,
			},
		},
	}
}

func (d *DatabasePolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerData = providerData
}

func (d *DatabasePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabasePolicyDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of policy_id or name is provided
	if (data.PolicyID.IsNull() && data.Name.IsNull()) || (!data.PolicyID.IsNull() && !data.Name.IsNull()) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Exactly one of 'policy_id' or 'name' must be specified",
		)
		return
	}

	// Check if UAP client is available
	if d.providerData.UAPClient == nil {
		resp.Diagnostics.AddError(
			"UAP Client Not Configured",
			"The UAP client is not available. This is a provider configuration issue.",
		)
		return
	}

	uapAPI := d.providerData.UAPClient
	var policyID string

	// Lookup by policy_id
	if !data.PolicyID.IsNull() {
		policyID = data.PolicyID.ValueString()
		tflog.Debug(ctx, "Looking up policy by ID", map[string]interface{}{
			"policy_id": policyID,
		})

		policy, err := uapAPI.Db().Policy(&uapcommonmodels.ArkUAPGetPolicyRequest{
			PolicyID: policyID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Policy",
				fmt.Sprintf("Could not read policy with ID %s: %s", policyID, err.Error()),
			)
			return
		}

		// Populate computed fields
		data.ID = types.StringValue(policy.Metadata.PolicyID)
		data.Description = types.StringValue(policy.Metadata.Description)
		data.Status = types.StringValue(policy.Metadata.Status.Status)

		tflog.Info(ctx, "Successfully read policy by ID", map[string]interface{}{
			"policy_id": policyID,
			"name":      policy.Metadata.Name,
		})
	} else {
		// Lookup by name
		policyName := data.Name.ValueString()
		tflog.Debug(ctx, "Looking up policy by name", map[string]interface{}{
			"name": policyName,
		})

		// List all policies and filter by name
		tflog.Debug(ctx, "Calling ListPolicies()", map[string]interface{}{
			"searching_for": policyName,
		})

		policyPages, err := uapAPI.Db().ListPolicies()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Policies",
				fmt.Sprintf("Could not list policies to find policy named '%s': %s", policyName, err.Error()),
			)
			return
		}

		tflog.Debug(ctx, "ListPolicies() returned channel, reading pages...")

		var foundPolicy *uapsiadbmodels.ArkUAPSIADBAccessPolicy
		var allPolicyNames []string
		pageCount := 0
		for page := range policyPages {
			pageCount++
			tflog.Debug(ctx, "Processing policy page", map[string]interface{}{
				"page_number": pageCount,
				"items_count": len(page.Items),
			})
			for _, policy := range page.Items {
				allPolicyNames = append(allPolicyNames, policy.Metadata.Name)
				tflog.Debug(ctx, "Found policy in list", map[string]interface{}{
					"policy_name": policy.Metadata.Name,
					"policy_id":   policy.Metadata.PolicyID,
				})
				if policy.Metadata.Name == policyName {
					foundPolicy = policy
					break
				}
			}
			if foundPolicy != nil {
				break
			}
		}

		tflog.Debug(ctx, "Policy lookup complete", map[string]interface{}{
			"searched_for":     policyName,
			"pages_processed":  pageCount,
			"policies_found":   len(allPolicyNames),
			"all_policy_names": allPolicyNames,
		})

		if foundPolicy == nil {
			resp.Diagnostics.AddError(
				"Policy Not Found",
				fmt.Sprintf("No policy found with name '%s'. Ensure the policy exists and you have permission to read it.", policyName),
			)
			return
		}

		// Populate computed fields
		data.ID = types.StringValue(foundPolicy.Metadata.PolicyID)
		data.PolicyID = types.StringValue(foundPolicy.Metadata.PolicyID)
		data.Description = types.StringValue(foundPolicy.Metadata.Description)
		data.Status = types.StringValue(foundPolicy.Metadata.Status.Status)

		tflog.Info(ctx, "Successfully read policy by name", map[string]interface{}{
			"name":      policyName,
			"policy_id": foundPolicy.Metadata.PolicyID,
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
