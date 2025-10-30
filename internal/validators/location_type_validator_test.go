package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocationTypeValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		{
			name:      "valid FQDN/IP",
			value:     types.StringValue("FQDN/IP"),
			expectErr: false,
		},
		{
			name:      "invalid lowercase fqdn/ip",
			value:     types.StringValue("fqdn/ip"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case Fqdn/Ip",
			value:     types.StringValue("Fqdn/Ip"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case fqdn/IP",
			value:     types.StringValue("fqdn/IP"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case FQDN/ip",
			value:     types.StringValue("FQDN/ip"),
			expectErr: true,
		},
		{
			name:      "invalid partial FQDN",
			value:     types.StringValue("FQDN"),
			expectErr: true,
		},
		{
			name:      "invalid partial IP",
			value:     types.StringValue("IP"),
			expectErr: true,
		},
		{
			name:      "invalid cloud-specific AWS-VPC",
			value:     types.StringValue("AWS-VPC"),
			expectErr: true,
		},
		{
			name:      "invalid cloud-specific AZURE-VNET",
			value:     types.StringValue("AZURE-VNET"),
			expectErr: true,
		},
		{
			name:      "invalid cloud-specific GCP-VPC",
			value:     types.StringValue("GCP-VPC"),
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
		{
			name:      "invalid near-match FQDN/IP/CIDR",
			value:     types.StringValue("FQDN/IP/CIDR"),
			expectErr: true,
		},
		{
			name:      "invalid near-match FQDN-IP",
			value:     types.StringValue("FQDN-IP"),
			expectErr: true,
		},
		{
			name:      "invalid near-match FQDN_IP",
			value:     types.StringValue("FQDN_IP"),
			expectErr: true,
		},
		{
			name:      "invalid with leading space",
			value:     types.StringValue(" FQDN/IP"),
			expectErr: true,
		},
		{
			name:      "invalid with trailing space",
			value:     types.StringValue("FQDN/IP "),
			expectErr: true,
		},
		{
			name:      "invalid with extra slashes",
			value:     types.StringValue("FQDN//IP"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := LocationType()
			req := validator.StringRequest{
				Path:        path.Root("location_type"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("LocationType() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestLocationTypeValidator_Description(t *testing.T) {
	v := LocationType()
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
