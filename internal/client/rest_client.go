// Package client provides CyberArk SIA API client wrappers
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RestClient provides a generic HTTP client for SIA API operations
// with OAuth2 authentication, retry logic, and error mapping
type RestClient struct {
	HTTPClient  *http.Client
	BaseURL     string
	AccessToken string
}

// NewRestClient creates a new generic REST client with OAuth2 access token
func NewRestClient(baseURL, accessToken string) (*RestClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	return &RestClient{
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		AccessToken: accessToken,
	}, nil
}

// DoRequest performs a generic HTTP request with retry logic and error handling
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - method: HTTP method (POST, GET, PUT, DELETE)
//   - path: API path (e.g., "/api/workspaces/db", "/api/secrets/db/{id}")
//   - requestBody: Request body to be marshaled to JSON (nil for GET/DELETE)
//   - responseData: Pointer to struct for unmarshaling response JSON (nil if no response expected)
//
// Returns:
//   - error: HTTP error, JSON parsing error, or nil on success
func (c *RestClient) DoRequest(
	ctx context.Context,
	method string,
	path string,
	requestBody interface{},
	responseData interface{},
) error {
	// Construct full URL
	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Execute request with retry logic for transient errors
	return RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
		// Marshal request body to JSON if provided
		var bodyReader io.Reader
		if requestBody != nil {
			bodyBytes, err := json.Marshal(requestBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)

			// Log request (sanitized - no sensitive data)
			tflog.Debug(ctx, "REST API request", map[string]interface{}{
				"method": method,
				"path":   path,
			})
		}

		// Create HTTP request with context
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set Authorization header (Bearer token)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

		// Set content type for requests with body
		if requestBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")

		// Execute HTTP request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response body
		respBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		// Check for non-2xx status codes and map to errors
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Create error message with status code and response body
			errorMsg := fmt.Sprintf("HTTP %d - %s", resp.StatusCode, string(respBodyBytes))
			return fmt.Errorf("%s", errorMsg)
		}

		// Log successful response
		tflog.Debug(ctx, "REST API response", map[string]interface{}{
			"status_code": resp.StatusCode,
			"method":      method,
			"path":        path,
		})

		// Unmarshal response JSON if responseData provided
		if responseData != nil && len(respBodyBytes) > 0 {
			if err := json.Unmarshal(respBodyBytes, responseData); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	})
}
