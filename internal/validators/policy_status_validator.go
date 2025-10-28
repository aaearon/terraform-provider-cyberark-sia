package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/slices"
)

// policyStatusValidator validates that policy status is "active" or "suspended" only
type policyStatusValidator struct{}

// Description returns a plain text description of the validator's behavior
func (v policyStatusValidator) Description(ctx context.Context) string {
	return "Value must be 'active' or 'suspended'"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior
func (v policyStatusValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be `active` or `suspended`"
}

// ValidateString validates the status value
func (v policyStatusValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is unknown or null (during plan phase)
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	validStatuses := []string{"active", "suspended"}

	if !slices.Contains(validStatuses, value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Status",
			fmt.Sprintf("Value %q is not valid. Must be 'active' or 'suspended'. "+
				"Note: 'expired', 'validating', and 'error' are server-managed statuses "+
				"and cannot be set by users.", value),
		)
	}
}

// PolicyStatus returns a validator that ensures policy status is "active" or "suspended"
func PolicyStatus() validator.String {
	return policyStatusValidator{}
}
