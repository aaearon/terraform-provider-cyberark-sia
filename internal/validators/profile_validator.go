// Package validators provides custom validators for Terraform resources
package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ validator.String = authenticationMethodValidator{}

// authenticationMethodValidator validates that authentication_method is one of the supported values.
type authenticationMethodValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (v authenticationMethodValidator) Description(ctx context.Context) string {
	return "value must be one of the supported authentication methods: db_auth, ldap_auth, oracle_auth, mongo_auth, sqlserver_auth, rds_iam_user_auth"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v authenticationMethodValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be one of: `db_auth`, `ldap_auth`, `oracle_auth`, `mongo_auth`, `sqlserver_auth`, `rds_iam_user_auth`"
}

// ValidateString performs the validation.
func (v authenticationMethodValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there's nothing to validate
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	tflog.Trace(ctx, "Validating authentication method", map[string]interface{}{
		"method": value,
	})

	// List of valid authentication methods
	validMethods := []string{
		"db_auth",
		"ldap_auth",
		"oracle_auth",
		"mongo_auth",
		"sqlserver_auth",
		"rds_iam_user_auth",
	}

	// Check if value is valid
	isValid := false
	for _, method := range validMethods {
		if value == method {
			isValid = true
			break
		}
	}

	if !isValid {
		tflog.Warn(ctx, "Authentication method validation failed", map[string]interface{}{
			"method":        value,
			"valid_methods": validMethods,
		})
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Authentication Method",
			fmt.Sprintf("Value %q is not a valid authentication method. "+
				"Must be one of: db_auth, ldap_auth, oracle_auth, mongo_auth, sqlserver_auth, rds_iam_user_auth",
				value,
			),
		)
	}
}

// AuthenticationMethod returns a validator that checks if a string is a valid authentication method.
func AuthenticationMethod() validator.String {
	return authenticationMethodValidator{}
}
