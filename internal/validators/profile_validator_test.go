package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAuthenticationMethodValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		{
			name:      "valid db_auth",
			value:     types.StringValue("db_auth"),
			expectErr: false,
		},
		{
			name:      "valid ldap_auth",
			value:     types.StringValue("ldap_auth"),
			expectErr: false,
		},
		{
			name:      "valid oracle_auth",
			value:     types.StringValue("oracle_auth"),
			expectErr: false,
		},
		{
			name:      "valid mongo_auth",
			value:     types.StringValue("mongo_auth"),
			expectErr: false,
		},
		{
			name:      "valid sqlserver_auth",
			value:     types.StringValue("sqlserver_auth"),
			expectErr: false,
		},
		{
			name:      "valid rds_iam_user_auth",
			value:     types.StringValue("rds_iam_user_auth"),
			expectErr: false,
		},
		{
			name:      "invalid method",
			value:     types.StringValue("invalid_auth"),
			expectErr: true,
		},
		{
			name:      "empty string",
			value:     types.StringValue(""),
			expectErr: true,
		},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := AuthenticationMethod()
			req := validator.StringRequest{
				Path:        path.Root("authentication_method"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("AuthenticationMethod() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestAuthenticationMethodValidator_Description(t *testing.T) {
	v := AuthenticationMethod()
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
