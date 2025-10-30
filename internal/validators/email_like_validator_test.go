package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEmailLikeValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		// Valid standard emails
		{
			name:      "valid standard email",
			value:     types.StringValue("user@example.com"),
			expectErr: false,
		},
		{
			name:      "valid email with dots",
			value:     types.StringValue("john.doe@company.com"),
			expectErr: false,
		},
		// Valid CyberArk cloud directory format
		{
			name:      "valid CyberArk cloud directory format",
			value:     types.StringValue("tim@cyberark.cloud.12345"),
			expectErr: false,
		},
		{
			name:      "valid CyberArk with dots and numbers",
			value:     types.StringValue("tim.schindler@cyberark.cloud.40562"),
			expectErr: false,
		},
		// Valid Active Directory format
		{
			name:      "valid Active Directory format",
			value:     types.StringValue("SchindlerT@domain.com"),
			expectErr: false,
		},
		{
			name:      "valid AD with mixed case",
			value:     types.StringValue("JohnDoe@CORP.EXAMPLE.COM"),
			expectErr: false,
		},
		// Valid with numbers and dots
		{
			name:      "valid with numbers in local part",
			value:     types.StringValue("user123@test.com"),
			expectErr: false,
		},
		{
			name:      "valid with multiple dots in domain",
			value:     types.StringValue("user.123@test.domain.co.uk"),
			expectErr: false,
		},
		// Valid with hyphens
		{
			name:      "valid with hyphens in local part",
			value:     types.StringValue("first-last@company.com"),
			expectErr: false,
		},
		{
			name:      "valid with hyphens in domain",
			value:     types.StringValue("user@my-company.com"),
			expectErr: false,
		},
		{
			name:      "valid with hyphens in both parts",
			value:     types.StringValue("first-last@my-company.com"),
			expectErr: false,
		},
		// Valid with plus signs
		{
			name:      "valid with plus sign",
			value:     types.StringValue("user+tag@example.com"),
			expectErr: false,
		},
		{
			name:      "valid with multiple plus signs",
			value:     types.StringValue("user+tag+test@example.com"),
			expectErr: false,
		},
		// Valid with underscores
		{
			name:      "valid with underscores",
			value:     types.StringValue("user_name@example.com"),
			expectErr: false,
		},
		{
			name:      "valid with percent",
			value:     types.StringValue("user%test@example.com"),
			expectErr: false,
		},
		// Valid complex examples
		{
			name:      "valid complex local part",
			value:     types.StringValue("user.name+tag-test_123@example.com"),
			expectErr: false,
		},
		{
			name:      "valid long TLD",
			value:     types.StringValue("user@example.museum"),
			expectErr: false,
		},
		{
			name:      "valid numeric TLD (CyberArk cloud)",
			value:     types.StringValue("service-account@cyberark.cloud.99999"),
			expectErr: false,
		},
		// Invalid: no @ symbol
		{
			name:      "invalid no @ symbol",
			value:     types.StringValue("username"),
			expectErr: true,
		},
		{
			name:      "invalid no @ with domain-like string",
			value:     types.StringValue("username.example.com"),
			expectErr: true,
		},
		// Invalid: multiple @ symbols
		{
			name:      "invalid multiple @ symbols",
			value:     types.StringValue("user@@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid @ in domain",
			value:     types.StringValue("user@domain@com"),
			expectErr: true,
		},
		// Invalid: missing domain
		{
			name:      "invalid missing domain",
			value:     types.StringValue("user@"),
			expectErr: true,
		},
		{
			name:      "invalid missing TLD",
			value:     types.StringValue("user@domain"),
			expectErr: true,
		},
		// Invalid: missing username
		{
			name:      "invalid missing username",
			value:     types.StringValue("@example.com"),
			expectErr: true,
		},
		// Invalid: spaces
		{
			name:      "invalid space in local part",
			value:     types.StringValue("user name@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid space in domain",
			value:     types.StringValue("user@example .com"),
			expectErr: true,
		},
		{
			name:      "invalid leading space",
			value:     types.StringValue(" user@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid trailing space",
			value:     types.StringValue("user@example.com "),
			expectErr: true,
		},
		// Invalid: special characters not allowed
		{
			name:      "invalid special char in local part (parentheses)",
			value:     types.StringValue("user(name)@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid special char in local part (brackets)",
			value:     types.StringValue("user[name]@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid special char in domain",
			value:     types.StringValue("user@exa!mple.com"),
			expectErr: true,
		},
		{
			name:      "invalid special char (asterisk)",
			value:     types.StringValue("user*@example.com"),
			expectErr: true,
		},
		{
			name:      "invalid special char (equals)",
			value:     types.StringValue("user=test@example.com"),
			expectErr: true,
		},
		// Invalid: empty string
		{
			name:      "invalid empty string",
			value:     types.StringValue(""),
			expectErr: true,
		},
		// Invalid: edge cases
		{
			name:      "invalid only @ symbol",
			value:     types.StringValue("@"),
			expectErr: true,
		},
		// Note: The following patterns are technically allowed by the lenient regex
		// to accommodate various email formats including CyberArk cloud directory
		{
			name:      "edge case: dot before @ (allowed by lenient pattern)",
			value:     types.StringValue("user.@example.com"),
			expectErr: false, // Lenient pattern allows this
		},
		{
			name:      "edge case: dot after @ (allowed by lenient pattern)",
			value:     types.StringValue("user@.example.com"),
			expectErr: false, // Lenient pattern allows this
		},
		{
			name:      "edge case: consecutive dots in local part (allowed by lenient pattern)",
			value:     types.StringValue("user..name@example.com"),
			expectErr: false, // Lenient pattern allows this
		},
		{
			name:      "edge case: consecutive dots in domain (allowed by lenient pattern)",
			value:     types.StringValue("user@example..com"),
			expectErr: false, // Lenient pattern allows this
		},
		{
			name:      "invalid TLD too short",
			value:     types.StringValue("user@example.c"),
			expectErr: true,
		},
		// Valid: null/unknown values (should pass - skip validation)
		{
			name:      "null value (allowed - skip validation)",
			value:     types.StringNull(),
			expectErr: false,
		},
		{
			name:      "unknown value (allowed - skip validation)",
			value:     types.StringUnknown(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := EmailLike()
			req := validator.StringRequest{
				Path:        path.Root("principal_name"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("EmailLike() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestEmailLikeValidator_Description(t *testing.T) {
	v := EmailLike()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	markdownDesc := v.MarkdownDescription(ctx)
	if markdownDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}
}
