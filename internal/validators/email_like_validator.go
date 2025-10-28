package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// emailLikeValidator validates that a string looks like an email address
type emailLikeValidator struct{}

// Email-like pattern: local-part@domain (flexible to accommodate various formats)
// Accepts: letters, numbers, dots, hyphens, underscores, plus signs in local part
// Requires @ symbol and domain with at least one dot
// CyberArk Cloud Directory uses format: user@cyberark.cloud.XXXXX (numbers in TLD)
var emailLikePattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z0-9]{2,}$`)

// Description returns a plain text description of the validator's behavior
func (v emailLikeValidator) Description(ctx context.Context) string {
	return "Value must be in email format (e.g., 'user@example.com' or 'tim.schindler@cyberark.cloud.40562')"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior
func (v emailLikeValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be in email format (e.g., `user@example.com` or `tim.schindler@cyberark.cloud.40562`)"
}

// ValidateString validates the email-like format
func (v emailLikeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is unknown or null (during plan phase)
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	if !emailLikePattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Email Format",
			fmt.Sprintf("Value %q is not a valid email format. Expected format: 'local-part@domain' (e.g., 'user@example.com' or 'tim.schindler@cyberark.cloud.40562').", value),
		)
	}
}

// EmailLike returns a validator that ensures the string is in email format
func EmailLike() validator.String {
	return emailLikeValidator{}
}
