package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/slices"
)

// principalTypeValidator validates that principal type is "USER", "GROUP", or "ROLE"
type principalTypeValidator struct{}

// Description returns a plain text description of the validator's behavior
func (v principalTypeValidator) Description(ctx context.Context) string {
	return "Value must be 'USER', 'GROUP', or 'ROLE'"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior
func (v principalTypeValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be `USER`, `GROUP`, or `ROLE`"
}

// ValidateString validates the principal type value
func (v principalTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is unknown or null (during plan phase)
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	validTypes := []string{"USER", "GROUP", "ROLE"}

	if !slices.Contains(validTypes, value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Principal Type",
			fmt.Sprintf("Value %q is not valid. Must be 'USER', 'GROUP', or 'ROLE'.", value),
		)
	}
}

// PrincipalType returns a validator that ensures principal type is valid
func PrincipalType() validator.String {
	return principalTypeValidator{}
}
