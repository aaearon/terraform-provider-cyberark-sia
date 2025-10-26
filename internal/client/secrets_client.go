// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
)

// SecretsClient wraps RestClient for database secret operations
// Thin wrapper (~50 lines) that delegates to generic RestClient
type SecretsClient struct {
	RestClient *RestClient
}

// NewSecretsClient creates a new secrets client using the generic RestClient
func NewSecretsClient(restClient *RestClient) *SecretsClient {
	return &SecretsClient{RestClient: restClient}
}

// CreateSecret creates a new database secret
func (c *SecretsClient) CreateSecret(ctx context.Context, secret *models.SecretAPI) (*models.SecretAPI, error) {
	var response models.SecretAPI
	err := c.RestClient.DoRequest(ctx, "POST", "/api/secrets/db", secret, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}
	return &response, nil
}

// GetSecret retrieves a database secret by ID
func (c *SecretsClient) GetSecret(ctx context.Context, id string) (*models.SecretAPI, error) {
	var response models.SecretAPI
	path := fmt.Sprintf("/api/secrets/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "GET", path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return &response, nil
}

// UpdateSecret updates an existing database secret
func (c *SecretsClient) UpdateSecret(ctx context.Context, id string, secret *models.SecretAPI) (*models.SecretAPI, error) {
	var response models.SecretAPI
	path := fmt.Sprintf("/api/secrets/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "PUT", path, secret, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}
	return &response, nil
}

// DeleteSecret deletes a database secret
func (c *SecretsClient) DeleteSecret(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/secrets/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// ListSecrets lists all database secrets (for import support)
func (c *SecretsClient) ListSecrets(ctx context.Context) ([]*models.SecretAPI, error) {
	var response []*models.SecretAPI
	err := c.RestClient.DoRequest(ctx, "GET", "/api/secrets/db", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return response, nil
}
