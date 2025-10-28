package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// locationTypeValidator validates that location type is "FQDN/IP" only
// Note: Database policies ONLY support "FQDN/IP" regardless of cloud provider
type locationTypeValidator struct{}

// Description returns a plain text description of the validator's behavior
func (v locationTypeValidator) Description(ctx context.Context) string {
	return "Value must be 'FQDN/IP' (the only supported location type for database policies)"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior
func (v locationTypeValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be `FQDN/IP` (the only supported location type for database policies)"
}

// ValidateString validates the location type value
func (v locationTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is unknown or null (during plan phase)
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Database policies only support "FQDN/IP" location type
	// This is true regardless of cloud provider (AWS, Azure, GCP, on-premise)
	if value != "FQDN/IP" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Location Type",
			"Value must be 'FQDN/IP'. Database policies only support the 'FQDN/IP' location type, "+
				"regardless of the cloud provider (AWS, Azure, GCP, on-premise, or Atlas).",
		)
	}
}

// LocationType returns a validator that ensures location type is "FQDN/IP"
func LocationType() validator.String {
	return locationTypeValidator{}
}
