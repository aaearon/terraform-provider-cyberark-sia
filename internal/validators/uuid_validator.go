package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// uuidValidator validates that a string is a valid UUID (v4 format)
type uuidValidator struct{}

// UUID pattern: 8-4-4-4-12 hex digits (case-insensitive)
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Description returns a plain text description of the validator's behavior
func (v uuidValidator) Description(ctx context.Context) string {
	return "Value must be a valid UUID format (e.g., 'c2c7bcc6-9560-44e0-8dff-5be221cd37ee')"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior
func (v uuidValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be a valid UUID format (e.g., `c2c7bcc6-9560-44e0-8dff-5be221cd37ee`)"
}

// ValidateString validates the UUID format
func (v uuidValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is unknown or null (during plan phase)
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	if !uuidPattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID Format",
			fmt.Sprintf("Value %q is not a valid UUID. Expected format: 8-4-4-4-12 hexadecimal digits (e.g., 'c2c7bcc6-9560-44e0-8dff-5be221cd37ee').", value),
		)
	}
}

// UUID returns a validator that ensures the string is a valid UUID
func UUID() validator.String {
	return uuidValidator{}
}
