package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPrincipalTypeValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		// Valid types (uppercase only)
		{
			name:      "valid USER",
			value:     types.StringValue("USER"),
			expectErr: false,
		},
		{
			name:      "valid GROUP",
			value:     types.StringValue("GROUP"),
			expectErr: false,
		},
		{
			name:      "valid ROLE",
			value:     types.StringValue("ROLE"),
			expectErr: false,
		},
		// Invalid lowercase versions (case-sensitive)
		{
			name:      "invalid lowercase user",
			value:     types.StringValue("user"),
			expectErr: true,
		},
		{
			name:      "invalid lowercase group",
			value:     types.StringValue("group"),
			expectErr: true,
		},
		{
			name:      "invalid lowercase role",
			value:     types.StringValue("role"),
			expectErr: true,
		},
		// Invalid mixed case
		{
			name:      "invalid mixed case User",
			value:     types.StringValue("User"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case Group",
			value:     types.StringValue("Group"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case Role",
			value:     types.StringValue("Role"),
			expectErr: true,
		},
		// Invalid values
		{
			name:      "invalid ADMIN",
			value:     types.StringValue("ADMIN"),
			expectErr: true,
		},
		{
			name:      "invalid SERVICE",
			value:     types.StringValue("SERVICE"),
			expectErr: true,
		},
		{
			name:      "invalid random string",
			value:     types.StringValue("invalid"),
			expectErr: true,
		},
		{
			name:      "invalid arbitrary text",
			value:     types.StringValue("not_a_principal_type"),
			expectErr: true,
		},
		// Empty string
		{
			name:      "empty string",
			value:     types.StringValue(""),
			expectErr: true,
		},
		// Null/unknown values (should pass - skip validation)
		{
			name:      "null value (allowed)",
			value:     types.StringNull(),
			expectErr: false,
		},
		{
			name:      "unknown value (allowed)",
			value:     types.StringUnknown(),
			expectErr: false,
		},
		// Near-matches (plural forms)
		{
			name:      "invalid USERS (plural)",
			value:     types.StringValue("USERS"),
			expectErr: true,
		},
		{
			name:      "invalid GROUPS (plural)",
			value:     types.StringValue("GROUPS"),
			expectErr: true,
		},
		{
			name:      "invalid ROLES (plural)",
			value:     types.StringValue("ROLES"),
			expectErr: true,
		},
		// Whitespace variations
		{
			name:      "invalid leading whitespace",
			value:     types.StringValue(" USER"),
			expectErr: true,
		},
		{
			name:      "invalid trailing whitespace",
			value:     types.StringValue("USER "),
			expectErr: true,
		},
		{
			name:      "invalid surrounding whitespace",
			value:     types.StringValue(" USER "),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PrincipalType()
			req := validator.StringRequest{
				Path:        path.Root("principal_type"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("PrincipalType() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestPrincipalTypeValidator_Description(t *testing.T) {
	v := PrincipalType()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	expectedDesc := "Value must be 'USER', 'GROUP', or 'ROLE'"
	if desc != expectedDesc {
		t.Errorf("Description() = %q, want %q", desc, expectedDesc)
	}

	markdownDesc := v.MarkdownDescription(ctx)
	if markdownDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}
	expectedMarkdown := "Value must be `USER`, `GROUP`, or `ROLE`"
	if markdownDesc != expectedMarkdown {
		t.Errorf("MarkdownDescription() = %q, want %q", markdownDesc, expectedMarkdown)
	}
}
