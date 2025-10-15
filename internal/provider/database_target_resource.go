// Package provider implements the database_target resource
package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/client"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
	dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &databaseTargetResource{}
	_ resource.ResourceWithConfigure   = &databaseTargetResource{}
	_ resource.ResourceWithImportState = &databaseTargetResource{}
)

// NewDatabaseTargetResource is a helper function to simplify the provider implementation
func NewDatabaseTargetResource() resource.Resource {
	return &databaseTargetResource{}
}

// databaseTargetResource is the resource implementation
type databaseTargetResource struct {
	providerData *ProviderData
}

// Metadata returns the resource type name
func (r *databaseTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_target"
}

// Schema defines the schema for the resource
func (r *databaseTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database target in CyberArk SIA. Database targets represent existing databases " +
			"(AWS RDS, Azure SQL, on-premise) registered with SIA for secure access management.",
		MarkdownDescription: "Manages a database target in CyberArk SIA. Database targets represent existing databases " +
			"(AWS RDS, Azure SQL, on-premise) registered with SIA for secure access management.\n\n" +
			"**Note**: This resource does NOT create databases. It only registers existing databases with SIA.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "SIA-assigned unique identifier for the database target",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "User-friendly name for the database target (1-255 characters, unique within SIA tenant)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"database_type": schema.StringAttribute{
				Description: "Type of database system. User must ensure compatibility with SIA. " +
					"Common values: postgresql, mysql, mariadb, mongodb, oracle, sqlserver, db2",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"database_version": schema.StringAttribute{
				Description: "Database version (semver format). User must ensure minimum version requirements for SIA.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"address": schema.StringAttribute{
				Description: "Hostname, IP address, or FQDN of the database server",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				Description: "TCP port for database connections (1-65535)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"database_name": schema.StringAttribute{
				Description: "Specific database/schema name (optional, database-dependent, max 255 characters)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"authentication_method": schema.StringAttribute{
				Description: "How SIA authenticates to the database. User must ensure compatibility with database_type. " +
					"Common values: local, domain, aws_iam",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cloud_provider": schema.StringAttribute{
				Description: "Cloud provider hosting the database (aws, azure, on_premise). Defaults to on_premise.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure", "on_premise"),
				},
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region (required if cloud_provider=aws)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("aws_account_id"),
					),
				},
			},
			"aws_account_id": schema.StringAttribute{
				Description: "AWS account ID as 12-digit number (required if cloud_provider=aws)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("aws_region"),
					),
					stringvalidator.LengthBetween(12, 12),
				},
			},
			"azure_tenant_id": schema.StringAttribute{
				Description: "Azure tenant ID as UUID (required if cloud_provider=azure)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("azure_subscription_id"),
					),
				},
			},
			"azure_subscription_id": schema.StringAttribute{
				Description: "Azure subscription ID as UUID (required if cloud_provider=azure)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("azure_tenant_id"),
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "User-provided description (max 1024 characters)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Key-value metadata tags (max 50 tags, key/value max 255 characters each)",
				ElementType: types.StringType,
				Optional:    true,
			},
			"last_modified": schema.StringAttribute{
				Description: "Timestamp of last modification (ISO 8601, computed by SIA)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *databaseTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	// Type assertion with error handling
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

// Create creates the resource and sets the initial Terraform state
func (r *databaseTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from plan
	var plan models.DatabaseTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating database target", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	// Convert tags from Terraform map to Go map[string]string
	var tags map[string]string
	if !plan.Tags.IsNull() {
		tags = make(map[string]string)
		diag := plan.Tags.ElementsAs(ctx, &tags, false)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}

	// Build ARK SDK request model
	// Per docs/sdk-integration.md: Use siaAPI.WorkspacesDB().AddDatabase()
	addDatabaseReq := &dbmodels.ArkSIADBAddDatabase{
		Name:                     plan.Name.ValueString(),
		ProviderEngine:           plan.DatabaseType.ValueString(), // Maps to database_type
		ReadWriteEndpoint:        plan.Address.ValueString(),
		Port:                     int(plan.Port.ValueInt64()),
		Platform:                 plan.CloudProvider.ValueString(),
		Region:                   plan.AWSRegion.ValueString(),
		ConfiguredAuthMethodType: plan.AuthenticationMethod.ValueString(),
		Tags:                     tags,
	}

	// Wrap SDK call with retry logic per docs/sdk-integration.md
	var database *dbmodels.ArkSIADBDatabase
	err := client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: r.providerData.MaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		var apiErr error
		database, apiErr = r.providerData.SIAAPI.WorkspacesDB().AddDatabase(addDatabaseReq)
		return apiErr
	})

	if err != nil {
		tflog.Error(ctx, "Failed to create database target", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "create database target"))
		return
	}

	// Map response to state
	plan.ID = types.StringValue(strconv.Itoa(database.ID))
	plan.LastModified = types.StringValue("") // TODO: Extract from response if available

	tflog.Info(ctx, "Created database target", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *databaseTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Get current state
	var state models.DatabaseTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Convert string ID to int for SDK
	databaseID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Database ID",
			fmt.Sprintf("Unable to convert ID to integer: %s", err.Error()),
		)
		return
	}

	// Per docs/sdk-integration.md: Use siaAPI.WorkspacesDB().Database()
	// Note: SDK method is "Database", not "GetDatabase"
	// Handle 404 as resource deleted (drift detection)
	var database *dbmodels.ArkSIADBDatabase
	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: r.providerData.MaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		var apiErr error
		database, apiErr = r.providerData.SIAAPI.WorkspacesDB().Database(&dbmodels.ArkSIADBGetDatabase{
			ID: databaseID,
		})
		return apiErr
	})

	if err != nil {
		// Check if resource was deleted outside Terraform (404)
		// Per sdk-integration.md: Handle 404 as resource deleted
		if client.IsNotFoundError(err) {
			tflog.Warn(ctx, "Database target not found, removing from state", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Error(ctx, "Failed to read database target", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "read database target"))
		return
	}

	// Map response to state - update fields from API response
	state.Name = types.StringValue(database.Name)
	state.Address = types.StringValue(database.ReadWriteEndpoint)
	state.Port = types.Int64Value(int64(database.Port))
	state.CloudProvider = types.StringValue(database.Platform)
	state.AWSRegion = types.StringValue(database.Region)
	// TODO: Map remaining fields from database response

	tflog.Debug(ctx, "Successfully read database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *databaseTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from plan and state
	var plan, state models.DatabaseTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Convert string ID to int for SDK
	databaseID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Database ID",
			fmt.Sprintf("Unable to convert ID to integer: %s", err.Error()),
		)
		return
	}

	// Convert tags from Terraform map to Go map[string]string
	var tags map[string]string
	if !plan.Tags.IsNull() {
		tags = make(map[string]string)
		diag := plan.Tags.ElementsAs(ctx, &tags, false)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}

	// Build update request with only changed fields
	// Per docs/sdk-integration.md: Use siaAPI.WorkspacesDB().UpdateDatabase()
	// SDK signature: UpdateDatabase(*ArkSIADBUpdateDatabase) (*ArkSIADBDatabase, error)
	updateReq := &dbmodels.ArkSIADBUpdateDatabase{
		ID:                       databaseID,
		NewName:                  plan.Name.ValueString(),
		ProviderEngine:           plan.DatabaseType.ValueString(),
		ReadWriteEndpoint:        plan.Address.ValueString(),
		Port:                     int(plan.Port.ValueInt64()),
		Platform:                 plan.CloudProvider.ValueString(),
		Region:                   plan.AWSRegion.ValueString(),
		ConfiguredAuthMethodType: plan.AuthenticationMethod.ValueString(),
		Tags:                     tags,
	}

	// Wrap SDK call with retry logic
	var updated *dbmodels.ArkSIADBDatabase
	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: r.providerData.MaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		var apiErr error
		updated, apiErr = r.providerData.SIAAPI.WorkspacesDB().UpdateDatabase(updateReq)
		return apiErr
	})

	if err != nil {
		tflog.Error(ctx, "Failed to update database target", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "update database target"))
		return
	}

	// Map response to state
	plan.ID = types.StringValue(strconv.Itoa(updated.ID))
	plan.LastModified = types.StringValue("") // TODO: Extract from response if available

	tflog.Info(ctx, "Updated database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *databaseTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from state
	var state models.DatabaseTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Convert string ID to int for SDK
	databaseID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Database ID",
			fmt.Sprintf("Unable to convert ID to integer: %s", err.Error()),
		)
		return
	}

	// Per docs/sdk-integration.md: Use siaAPI.WorkspacesDB().DeleteDatabase()
	// SDK signature: DeleteDatabase(*ArkSIADBDeleteDatabase) error
	// Gracefully handle already-deleted resources
	err = client.RetryWithBackoff(ctx, &client.RetryConfig{
		MaxRetries: r.providerData.MaxRetries,
		BaseDelay:  client.BaseDelay,
		MaxDelay:   client.MaxDelay,
	}, func() error {
		return r.providerData.SIAAPI.WorkspacesDB().DeleteDatabase(&dbmodels.ArkSIADBDeleteDatabase{
			ID: databaseID,
		})
	})

	if err != nil {
		// Gracefully handle already-deleted resource (404)
		if client.IsNotFoundError(err) {
			tflog.Warn(ctx, "Database target already deleted", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			return
		}

		tflog.Error(ctx, "Failed to delete database target", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "delete database target"))
		return
	}

	tflog.Info(ctx, "Deleted database target", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing resource into Terraform state
func (r *databaseTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID from import to retrieve the resource
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Info(ctx, "Imported database target", map[string]interface{}{
		"id": req.ID,
	})
}
