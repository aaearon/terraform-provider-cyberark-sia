// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/models"
)

// DatabaseWorkspaceClient wraps RestClient for database workspace operations
// Thin wrapper (~50 lines) that delegates to generic RestClient
type DatabaseWorkspaceClient struct {
	RestClient *RestClient
}

// NewDatabaseWorkspaceClient creates a new database workspace client using the generic RestClient
func NewDatabaseWorkspaceClient(restClient *RestClient) *DatabaseWorkspaceClient {
	return &DatabaseWorkspaceClient{RestClient: restClient}
}

// CreateDatabaseWorkspace creates a new database workspace
func (c *DatabaseWorkspaceClient) CreateDatabaseWorkspace(ctx context.Context, workspace *models.DatabaseWorkspaceAPI) (*models.DatabaseWorkspaceAPI, error) {
	var response models.DatabaseWorkspaceAPI
	err := c.RestClient.DoRequest(ctx, "POST", "/api/workspaces/db", workspace, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create database workspace: %w", err)
	}
	return &response, nil
}

// GetDatabaseWorkspace retrieves a database workspace by ID
func (c *DatabaseWorkspaceClient) GetDatabaseWorkspace(ctx context.Context, id string) (*models.DatabaseWorkspaceAPI, error) {
	var response models.DatabaseWorkspaceAPI
	path := fmt.Sprintf("/api/workspaces/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "GET", path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get database workspace: %w", err)
	}
	return &response, nil
}

// UpdateDatabaseWorkspace updates an existing database workspace
func (c *DatabaseWorkspaceClient) UpdateDatabaseWorkspace(ctx context.Context, id string, workspace *models.DatabaseWorkspaceAPI) (*models.DatabaseWorkspaceAPI, error) {
	var response models.DatabaseWorkspaceAPI
	path := fmt.Sprintf("/api/workspaces/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "PUT", path, workspace, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to update database workspace: %w", err)
	}
	return &response, nil
}

// DeleteDatabaseWorkspace deletes a database workspace
func (c *DatabaseWorkspaceClient) DeleteDatabaseWorkspace(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/workspaces/db/%s", id)
	err := c.RestClient.DoRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete database workspace: %w", err)
	}
	return nil
}

// ListDatabaseWorkspaces lists all database workspaces (for import support)
func (c *DatabaseWorkspaceClient) ListDatabaseWorkspaces(ctx context.Context) ([]*models.DatabaseWorkspaceAPI, error) {
	var response []*models.DatabaseWorkspaceAPI
	err := c.RestClient.DoRequest(ctx, "GET", "/api/workspaces/db", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list database workspaces: %w", err)
	}
	return response, nil
}
