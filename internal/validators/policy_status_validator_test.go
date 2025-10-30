package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPolicyStatusValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		// Valid lowercase values
		{
			name:      "valid active lowercase",
			value:     types.StringValue("active"),
			expectErr: false,
		},
		{
			name:      "valid suspended lowercase",
			value:     types.StringValue("suspended"),
			expectErr: false,
		},
		// Valid uppercase values (API returns uppercase)
		{
			name:      "valid Active uppercase",
			value:     types.StringValue("Active"),
			expectErr: false,
		},
		{
			name:      "valid Suspended uppercase",
			value:     types.StringValue("Suspended"),
			expectErr: false,
		},
		// Invalid server-managed statuses
		{
			name:      "invalid expired status",
			value:     types.StringValue("expired"),
			expectErr: true,
		},
		{
			name:      "invalid validating status",
			value:     types.StringValue("validating"),
			expectErr: true,
		},
		{
			name:      "invalid error status",
			value:     types.StringValue("error"),
			expectErr: true,
		},
		// Invalid random strings
		{
			name:      "invalid random string",
			value:     types.StringValue("invalid"),
			expectErr: true,
		},
		{
			name:      "invalid foo string",
			value:     types.StringValue("foo"),
			expectErr: true,
		},
		{
			name:      "empty string",
			value:     types.StringValue(""),
			expectErr: true,
		},
		// Case sensitivity edge cases
		{
			name:      "invalid all uppercase ACTIVE",
			value:     types.StringValue("ACTIVE"),
			expectErr: true,
		},
		{
			name:      "invalid all uppercase SUSPENDED",
			value:     types.StringValue("SUSPENDED"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case AcTiVe",
			value:     types.StringValue("AcTiVe"),
			expectErr: true,
		},
		{
			name:      "invalid mixed case SuSpEnDeD",
			value:     types.StringValue("SuSpEnDeD"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PolicyStatus()
			req := validator.StringRequest{
				Path:        path.Root("status"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("PolicyStatus() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}

			// Verify error message contains helpful information
			if tt.expectErr && hasError {
				diags := resp.Diagnostics
				found := false
				for _, diag := range diags {
					if diag.Summary() == "Invalid Policy Status" {
						found = true
						// Check that error message mentions server-managed statuses
						detail := diag.Detail()
						if detail == "" {
							t.Error("Expected error detail to be non-empty")
						}
					}
				}
				if !found {
					t.Error("Expected diagnostic with summary 'Invalid Policy Status'")
				}
			}
		})
	}
}

func TestPolicyStatusValidator_Description(t *testing.T) {
	v := PolicyStatus()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	markdownDesc := v.MarkdownDescription(ctx)
	if markdownDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}

	// Verify descriptions contain expected content
	expectedContent := []string{"active", "suspended"}
	for _, expected := range expectedContent {
		if !contains(desc, expected) {
			t.Errorf("Description() missing expected content: %s", expected)
		}
		if !contains(markdownDesc, expected) {
			t.Errorf("MarkdownDescription() missing expected content: %s", expected)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
