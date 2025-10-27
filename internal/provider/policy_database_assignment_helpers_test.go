package provider

import (
	"testing"

	uapsiadbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db/models"
)

// Test composite ID building
func TestBuildCompositeID(t *testing.T) {
	tests := []struct {
		name       string
		policyID   string
		dbID       string
		want       string
	}{
		{
			name:     "standard UUIDs",
			policyID: "12345678-1234-1234-1234-123456789012",
			dbID:     "42",
			want:     "12345678-1234-1234-1234-123456789012:42",
		},
		{
			name:     "simple IDs",
			policyID: "policy1",
			dbID:     "123",
			want:     "policy1:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCompositeID(tt.policyID, tt.dbID)
			if got != tt.want {
				t.Errorf("buildCompositeID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test composite ID parsing
func TestParseCompositeID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantPolicyID  string
		wantDBID      string
		wantErr       bool
	}{
		{
			name:         "valid composite ID",
			id:           "policy-123:database-456",
			wantPolicyID: "policy-123",
			wantDBID:     "database-456",
			wantErr:      false,
		},
		{
			name:         "valid with UUID",
			id:           "12345678-1234-1234-1234-123456789012:42",
			wantPolicyID: "12345678-1234-1234-1234-123456789012",
			wantDBID:     "42",
			wantErr:      false,
		},
		{
			name:    "missing colon",
			id:      "policy-123-database-456",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "only colon",
			id:      ":",
			wantErr: true,
		},
		{
			name:         "multiple colons (takes first)",
			id:           "policy:123:extra",
			wantPolicyID: "policy",
			wantDBID:     "123:extra",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicyID, gotDBID, err := parseCompositeID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCompositeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotPolicyID != tt.wantPolicyID {
					t.Errorf("parseCompositeID() policyID = %v, want %v", gotPolicyID, tt.wantPolicyID)
				}
				if gotDBID != tt.wantDBID {
					t.Errorf("parseCompositeID() dbID = %v, want %v", gotDBID, tt.wantDBID)
				}
			}
		})
	}
}

// Test workspace type determination
func TestDetermineWorkspaceType(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "AWS platform",
			platform: "AWS",
			want:     "AWS",
		},
		{
			name:     "AWS lowercase",
			platform: "aws",
			want:     "AWS",
		},
		{
			name:     "AZURE platform",
			platform: "AZURE",
			want:     "AZURE",
		},
		{
			name:     "GCP platform",
			platform: "GCP",
			want:     "GCP",
		},
		{
			name:     "ATLAS platform",
			platform: "ATLAS",
			want:     "ATLAS",
		},
		{
			name:     "ON-PREMISE platform",
			platform: "ON-PREMISE",
			want:     "FQDN/IP",
		},
		{
			name:     "on-premise lowercase",
			platform: "on-premise",
			want:     "FQDN/IP",
		},
		{
			name:     "empty string defaults to FQDN/IP",
			platform: "",
			want:     "FQDN/IP",
		},
		{
			name:     "unknown platform defaults to FQDN/IP",
			platform: "UNKNOWN",
			want:     "FQDN/IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineWorkspaceType(tt.platform)
			if got != tt.want {
				t.Errorf("determineWorkspaceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test finding database in policy
func TestFindDatabaseInPolicy(t *testing.T) {
	// Setup test policy
	policy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
			"AWS": {
				Instances: []uapsiadbmodels.ArkUAPSIADBInstanceTarget{
					{InstanceID: "123", InstanceName: "db1"},
					{InstanceID: "456", InstanceName: "db2"},
				},
			},
			"FQDN_IP": {
				Instances: []uapsiadbmodels.ArkUAPSIADBInstanceTarget{
					{InstanceID: "789", InstanceName: "db3"},
				},
			},
		},
	}

	tests := []struct {
		name       string
		policy     *uapsiadbmodels.ArkUAPSIADBAccessPolicy
		databaseID string
		wantFound  bool
		wantName   string
	}{
		{
			name:       "found in AWS workspace",
			policy:     policy,
			databaseID: "123",
			wantFound:  true,
			wantName:   "db1",
		},
		{
			name:       "found in FQDN_IP workspace",
			policy:     policy,
			databaseID: "789",
			wantFound:  true,
			wantName:   "db3",
		},
		{
			name:       "not found",
			policy:     policy,
			databaseID: "999",
			wantFound:  false,
		},
		{
			name:       "nil targets",
			policy:     &uapsiadbmodels.ArkUAPSIADBAccessPolicy{Targets: nil},
			databaseID: "123",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDatabaseInPolicy(tt.policy, tt.databaseID)
			if (got != nil) != tt.wantFound {
				t.Errorf("findDatabaseInPolicy() found = %v, want %v", got != nil, tt.wantFound)
				return
			}
			if got != nil && got.InstanceName != tt.wantName {
				t.Errorf("findDatabaseInPolicy() name = %v, want %v", got.InstanceName, tt.wantName)
			}
		})
	}
}

// Test finding database with workspace type
func TestFindDatabaseInPolicyWithType(t *testing.T) {
	policy := &uapsiadbmodels.ArkUAPSIADBAccessPolicy{
		Targets: map[string]uapsiadbmodels.ArkUAPSIADBTargets{
			"AWS": {
				Instances: []uapsiadbmodels.ArkUAPSIADBInstanceTarget{
					{InstanceID: "123", InstanceName: "db1"},
				},
			},
			"AZURE": {
				Instances: []uapsiadbmodels.ArkUAPSIADBInstanceTarget{
					{InstanceID: "456", InstanceName: "db2"},
				},
			},
		},
	}

	tests := []struct {
		name              string
		policy            *uapsiadbmodels.ArkUAPSIADBAccessPolicy
		databaseID        string
		wantFound         bool
		wantName          string
		wantWorkspaceType string
	}{
		{
			name:              "found in AWS",
			policy:            policy,
			databaseID:        "123",
			wantFound:         true,
			wantName:          "db1",
			wantWorkspaceType: "AWS",
		},
		{
			name:              "found in AZURE",
			policy:            policy,
			databaseID:        "456",
			wantFound:         true,
			wantName:          "db2",
			wantWorkspaceType: "AZURE",
		},
		{
			name:       "not found",
			policy:     policy,
			databaseID: "999",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, workspaceType, found := findDatabaseInPolicyWithType(tt.policy, tt.databaseID)
			if found != tt.wantFound {
				t.Errorf("findDatabaseInPolicyWithType() found = %v, want %v", found, tt.wantFound)
				return
			}
			if found {
				if target.InstanceName != tt.wantName {
					t.Errorf("findDatabaseInPolicyWithType() name = %v, want %v", target.InstanceName, tt.wantName)
				}
				if workspaceType != tt.wantWorkspaceType {
					t.Errorf("findDatabaseInPolicyWithType() workspaceType = %v, want %v", workspaceType, tt.wantWorkspaceType)
				}
			}
		})
	}
}
