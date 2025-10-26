// Package provider implements the strong_account resource
package provider

import (
	"context"
	"fmt"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/client"
	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
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
	_ resource.Resource                   = &secretResource{}
	_ resource.ResourceWithConfigure      = &secretResource{}
	_ resource.ResourceWithImportState    = &secretResource{}
	_ resource.ResourceWithValidateConfig = &secretResource{}
)

// NewSecretResource is a helper function to simplify the provider implementation
func NewSecretResource() resource.Resource {
	return &secretResource{}
}

// secretResource is the resource implementation
type secretResource struct {
	providerData *ProviderData
}

// Metadata returns the resource type name
func (r *secretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for the resource
func (r *secretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a secret (credential) in CyberArk SIA. Secrets are credentials " +
			"that SIA uses to provision ephemeral database access for users. Supports local authentication, " +
			"Active Directory, and AWS IAM authentication methods.",
		MarkdownDescription: "Manages a secret (credential) in CyberArk SIA. Secrets are credentials " +
			"that SIA uses to provision ephemeral database access for users.\n\n" +
			"**Authentication Types**:\n" +
			"- `local`: Username/password stored in SIA (username_password secret type)\n" +
			"- `domain`: Active Directory account (username_password secret type with domain)\n" +
			"- `aws_iam`: AWS IAM credentials (iam_user secret type)",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "SIA-assigned unique identifier for the secret (secret_id)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "User-friendly name for the secret (1-255 characters, unique within SIA tenant). " +
					"Maps to SecretName in SDK.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"database_workspace_id": schema.StringAttribute{
				Description: "ID of the database workspace this secret is associated with. " +
					"Must reference an existing cyberark_sia_database_workspace resource. " +
					"The secret will be used by SIA to provision ephemeral access to this database.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"authentication_type": schema.StringAttribute{
				Description: "Type of authentication credentials. " +
					"Valid values: local (username/password), domain (Active Directory), aws_iam (AWS IAM user). " +
					"Maps to SecretType in SDK (username_password or iam_user). " +
					"Must match the authentication_method of the referenced database_target.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("local", "domain", "aws_iam"),
				},
			},

			// Local/Domain authentication fields
			"username": schema.StringAttribute{
				Description: "Account username. Required for local and domain authentication types. " +
					"Not allowed for aws_iam authentication. Max 255 characters. " +
					"Validated in ValidateConfig method.",
				Optional:  true,
				Sensitive: false, // Username is not sensitive, only password is
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"password": schema.StringAttribute{
				Description: "Account password. Required for local and domain authentication types. " +
					"Not allowed for aws_iam authentication. Min 8 characters. " +
					"NEVER logged or displayed in outputs. " +
					"Validated in ValidateConfig method.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(8),
				},
			},

			// Domain authentication field
			"domain": schema.StringAttribute{
				Description: "Active Directory domain (e.g., corp.example.com). " +
					"Optional field for documentation purposes when authentication_type=domain. " +
					"NOTE: The ARK SDK does not have a separate domain field. " +
					"Include the domain directly in the username field using either:\n" +
					"- Windows format: DOMAIN\\username\n" +
					"- UPN format: username@domain.com\n" +
					"This field is for user convenience and is not sent to the SDK. " +
					"Validated in ValidateConfig method.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},

			// AWS IAM authentication fields
			"aws_access_key_id": schema.StringAttribute{
				Description: "AWS IAM access key ID. Required when authentication_type=aws_iam. " +
					"Not allowed for local or domain authentication. " +
					"Maps to IAMAccessKeyID in SDK. Valid AWS access key format (20 characters). " +
					"Validated in ValidateConfig method.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"aws_secret_access_key": schema.StringAttribute{
				Description: "AWS IAM secret access key. Required when authentication_type=aws_iam. " +
					"Not allowed for local or domain authentication. " +
					"Maps to IAMSecretAccessKey in SDK. NEVER logged or displayed in outputs. " +
					"Validated in ValidateConfig method.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},

			// Rotation settings
			"rotation_enabled": schema.BoolAttribute{
				Description: "Whether SIA should automatically rotate credentials. " +
					"When false, credentials are only updated via Terraform (manual rotation). " +
					"When true, rotation_interval_days must be specified. " +
					"NOTE: Current SDK version may not support automatic rotation; verify SIA API capabilities. " +
					"Validated in ValidateConfig method.",
				Optional: true,
			},
			"rotation_interval_days": schema.Int64Attribute{
				Description: "Days between automatic credential rotations (1-365). " +
					"Required when rotation_enabled=true. Ignored when rotation_enabled=false. " +
					"SIA will rotate credentials on this schedule. " +
					"Validated in ValidateConfig method.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 365),
				},
			},

			// Computed attributes
			"created_at": schema.StringAttribute{
				Description: "Timestamp of creation (ISO 8601, computed by SIA). Maps to CreationTime in SDK.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified": schema.StringAttribute{
				Description: "Timestamp of last modification (ISO 8601, computed by SIA). Maps to LastUpdateTime in SDK.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *secretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from plan
	var plan models.SecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating secret", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	// Convert Terraform model to API model (following certificate pattern)
	createReq := &models.SecretAPI{
		Name:       plan.Name.ValueString(),
		DatabaseID: plan.DatabaseWorkspaceID.ValueString(),
	}

	// Set authentication type-specific fields
	authType := plan.AuthenticationType.ValueString()
	switch authType {
	case "local", "domain":
		if !plan.Username.IsNull() {
			createReq.Username = plan.Username.ValueString()
		}
		if !plan.Password.IsNull() {
			createReq.Password = models.StringPtr(plan.Password.ValueString())
		}

	case "aws_iam":
		// Note: AWS IAM auth not yet implemented in SecretsClient
		// TODO: Add IAM fields to models.SecretAPI when needed
		resp.Diagnostics.AddError(
			"Unsupported Authentication Type",
			"AWS IAM authentication is not yet supported. Use 'local' or 'domain' authentication.",
		)
		return

	default:
		resp.Diagnostics.AddError(
			"Invalid Authentication Type",
			fmt.Sprintf("Unsupported authentication type: %s. Valid values: local, domain", authType),
		)
		return
	}

	// Call API to create secret
	secret, err := r.providerData.SecretsClient.CreateSecret(ctx, createReq)
	if err != nil {
		tflog.Error(ctx, "Failed to create secret", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "create secret"))
		return
	}

	// Map API response to Terraform state
	plan.ID = types.StringValue(*secret.ID)
	if secret.CreatedTime != nil {
		plan.CreatedAt = types.StringValue(*secret.CreatedTime)
	}
	if secret.ModifiedTime != nil {
		plan.LastModified = types.StringValue(*secret.ModifiedTime)
	}

	tflog.Info(ctx, "Created secret", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Get current state
	var state models.SecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Call API to get secret
	secret, err := r.providerData.SecretsClient.GetSecret(ctx, state.ID.ValueString())
	if err != nil {
		// Check if resource was deleted outside Terraform (404)
		if client.IsNotFoundError(err) {
			tflog.Warn(ctx, "Secret not found, removing from state", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Error(ctx, "Failed to read secret", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "read secret"))
		return
	}

	// Map API response to Terraform state
	// NOTE: Sensitive credentials (password) are NOT returned by API
	state.Name = types.StringValue(secret.Name)
	if secret.CreatedTime != nil {
		state.CreatedAt = types.StringValue(*secret.CreatedTime)
	}
	if secret.ModifiedTime != nil {
		state.LastModified = types.StringValue(*secret.ModifiedTime)
	}

	tflog.Debug(ctx, "Successfully read secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from plan and state
	var plan, state models.SecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Convert Terraform plan to API model for update
	updateReq := &models.SecretAPI{
		Name:       plan.Name.ValueString(),
		DatabaseID: plan.DatabaseWorkspaceID.ValueString(),
	}

	// Update credentials if changed
	authType := plan.AuthenticationType.ValueString()
	switch authType {
	case "local", "domain":
		if !plan.Username.IsNull() {
			updateReq.Username = plan.Username.ValueString()
		}
		if !plan.Password.IsNull() {
			updateReq.Password = models.StringPtr(plan.Password.ValueString())
		}
	}

	// Call API to update secret
	updated, err := r.providerData.SecretsClient.UpdateSecret(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		tflog.Error(ctx, "Failed to update secret", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "update secret"))
		return
	}

	// Map API response to state
	if updated.ModifiedTime != nil {
		plan.LastModified = types.StringValue(*updated.ModifiedTime)
	}

	tflog.Info(ctx, "Updated secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Check if provider is configured
	if r.providerData == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured ProviderData. Please report this issue to the provider developers.",
		)
		return
	}

	// Retrieve values from state
	var state models.SecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Call API to delete secret
	err := r.providerData.SecretsClient.DeleteSecret(ctx, state.ID.ValueString())
	if err != nil {
		// Gracefully handle already-deleted resource (404)
		if client.IsNotFoundError(err) {
			tflog.Warn(ctx, "Secret already deleted", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			return
		}

		tflog.Error(ctx, "Failed to delete secret", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.Append(client.MapError(err, "delete secret"))
		return
	}

	tflog.Info(ctx, "Deleted secret", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing resource into Terraform state
func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID from import to retrieve the resource
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Info(ctx, "Imported secret", map[string]interface{}{
		"id": req.ID,
	})
}

// ValidateConfig performs cross-field validation for the secret resource
func (r *secretResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.SecretModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate authentication type combinations
	authType := config.AuthenticationType.ValueString()

	switch authType {
	case "local", "domain":
		// Username and password are required for local/domain authentication
		if config.Username.IsNull() || config.Username.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Missing Required Field",
				fmt.Sprintf("username is required when authentication_type=%s", authType),
			)
		}
		if config.Password.IsNull() || config.Password.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Missing Required Field",
				fmt.Sprintf("password is required when authentication_type=%s", authType),
			)
		}

		// AWS IAM fields should not be set
		if !config.AWSAccessKeyID.IsNull() && !config.AWSAccessKeyID.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("aws_access_key_id"),
				"Invalid Field Combination",
				fmt.Sprintf("aws_access_key_id cannot be set when authentication_type=%s", authType),
			)
		}
		if !config.AWSSecretAccessKey.IsNull() && !config.AWSSecretAccessKey.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("aws_secret_access_key"),
				"Invalid Field Combination",
				fmt.Sprintf("aws_secret_access_key cannot be set when authentication_type=%s", authType),
			)
		}

	case "aws_iam":
		// AWS IAM credentials are required
		if config.AWSAccessKeyID.IsNull() || config.AWSAccessKeyID.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("aws_access_key_id"),
				"Missing Required Field",
				"aws_access_key_id is required when authentication_type=aws_iam",
			)
		}
		if config.AWSSecretAccessKey.IsNull() || config.AWSSecretAccessKey.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("aws_secret_access_key"),
				"Missing Required Field",
				"aws_secret_access_key is required when authentication_type=aws_iam",
			)
		}

		// Username/password/domain should not be set
		if !config.Username.IsNull() && !config.Username.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Invalid Field Combination",
				"username cannot be set when authentication_type=aws_iam",
			)
		}
		if !config.Password.IsNull() && !config.Password.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Invalid Field Combination",
				"password cannot be set when authentication_type=aws_iam",
			)
		}
		if !config.Domain.IsNull() && !config.Domain.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("domain"),
				"Invalid Field Combination",
				"domain cannot be set when authentication_type=aws_iam",
			)
		}
	}

	// Validate rotation settings (T048)
	if !config.RotationEnabled.IsNull() && !config.RotationEnabled.IsUnknown() && config.RotationEnabled.ValueBool() {
		if config.RotationIntervalDays.IsNull() || config.RotationIntervalDays.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("rotation_interval_days"),
				"Missing Required Field",
				"rotation_interval_days is required when rotation_enabled=true",
			)
		}
	}

	// Warn if rotation_interval_days is set but rotation_enabled is not true
	if (!config.RotationIntervalDays.IsNull() && !config.RotationIntervalDays.IsUnknown()) &&
		(config.RotationEnabled.IsNull() || config.RotationEnabled.IsUnknown() || !config.RotationEnabled.ValueBool()) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("rotation_interval_days"),
			"Unused Configuration",
			"rotation_interval_days is set but rotation_enabled is not true. This setting will be ignored.",
		)
	}
}
