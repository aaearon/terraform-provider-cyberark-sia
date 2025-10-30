package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUUIDValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		{
			name:      "valid UUID v4 lowercase",
			value:     types.StringValue("550e8400-e29b-41d4-a716-446655440000"),
			expectErr: false,
		},
		{
			name:      "valid UUID v4 uppercase",
			value:     types.StringValue("550E8400-E29B-41D4-A716-446655440000"),
			expectErr: false,
		},
		{
			name:      "valid UUID v4 mixed case",
			value:     types.StringValue("550e8400-E29B-41d4-A716-446655440000"),
			expectErr: false,
		},
		{
			name:      "valid UUID from docs example",
			value:     types.StringValue("c2c7bcc6-9560-44e0-8dff-5be221cd37ee"),
			expectErr: false,
		},
		{
			name:      "valid UUID all zeros",
			value:     types.StringValue("00000000-0000-0000-0000-000000000000"),
			expectErr: false,
		},
		{
			name:      "valid UUID all f's",
			value:     types.StringValue("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			expectErr: false,
		},
		{
			name:      "invalid missing hyphens",
			value:     types.StringValue("550e8400e29b41d4a716446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid wrong hyphen positions",
			value:     types.StringValue("550e8400e-29b-41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid too short",
			value:     types.StringValue("550e8400-e29b-41d4-a716"),
			expectErr: true,
		},
		{
			name:      "invalid too long",
			value:     types.StringValue("550e8400-e29b-41d4-a716-446655440000-extra"),
			expectErr: true,
		},
		{
			name:      "invalid non-hex character g",
			value:     types.StringValue("550g8400-e29b-41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid non-hex character z",
			value:     types.StringValue("550e8400-e29b-41d4-a716-44665544000z"),
			expectErr: true,
		},
		{
			name:      "invalid special character @",
			value:     types.StringValue("550e8400@e29b-41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid special character #",
			value:     types.StringValue("550e8400-e29b-41d4-a716-446655440#00"),
			expectErr: true,
		},
		{
			name:      "invalid space in UUID",
			value:     types.StringValue("550e8400 e29b-41d4-a716-446655440000"),
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
			name:      "invalid correct length wrong format",
			value:     types.StringValue("12345678-1234-1234-1234-12345678901g"),
			expectErr: true,
		},
		{
			name:      "invalid missing segment",
			value:     types.StringValue("550e8400--41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid extra hyphen",
			value:     types.StringValue("550e8400-e29b--41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid leading hyphen",
			value:     types.StringValue("-550e8400-e29b-41d4-a716-446655440000"),
			expectErr: true,
		},
		{
			name:      "invalid trailing hyphen",
			value:     types.StringValue("550e8400-e29b-41d4-a716-446655440000-"),
			expectErr: true,
		},
		{
			name:      "invalid curly braces",
			value:     types.StringValue("{550e8400-e29b-41d4-a716-446655440000}"),
			expectErr: true,
		},
		{
			name:      "invalid urn prefix",
			value:     types.StringValue("urn:uuid:550e8400-e29b-41d4-a716-446655440000"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := UUID()
			req := validator.StringRequest{
				Path:        path.Root("uuid"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("UUID() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestUUIDValidator_Description(t *testing.T) {
	v := UUID()
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
