// Package client provides CyberArk SIA API client wrappers
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CertificatesClientOAuth2 handles certificate CRUD operations using direct OAuth2 access tokens.
// Unlike the ARK SDK-based CertificatesClient, this implementation uses the access token
// directly without exchanging it for an ID token.
//
// Authentication Flow:
//   1. Obtain access token via GetPlatformAccessToken() (OAuth2 client credentials)
//   2. Construct SIA API URL from JWT token claims
//   3. Use access token as Bearer token for all API calls
//
// This resolves the 401 Unauthorized issue caused by using ID tokens for API authorization.
type CertificatesClientOAuth2 struct {
	baseURL     string      // SIA API base URL (e.g., https://abc123.dpa.cyberark.cloud)
	accessToken string      // OAuth2 access token (NOT ID token)
	httpClient  *http.Client
}

// NewCertificatesClientOAuth2 creates a new OAuth2-based certificates client.
//
// Parameters:
//   - baseURL: SIA API base URL (e.g., https://abc123.dpa.cyberark.cloud)
//   - accessToken: OAuth2 access token from GetPlatformAccessToken()
//
// Returns:
//   - *CertificatesClientOAuth2: Initialized client ready for CRUD operations
//   - error: If validation fails
func NewCertificatesClientOAuth2(baseURL, accessToken string) (*CertificatesClientOAuth2, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken cannot be empty")
	}

	return &CertificatesClientOAuth2{
		baseURL:     baseURL,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// doRequest executes an HTTP request with proper authentication headers
func (c *CertificatesClientOAuth2) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header (Bearer access token)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// CreateCertificate creates a new certificate in SIA.
// API Endpoint: POST /api/certificates
func (c *CertificatesClientOAuth2) CreateCertificate(ctx context.Context, req *CertificateCreateRequest) (*Certificate, error) {
	if req.CertBody == "" {
		return nil, fmt.Errorf("cert_body is required for certificate creation")
	}

	var cert Certificate
	err := RetryWithBackoff(ctx, &RetryConfig{
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  BaseDelay,
		MaxDelay:   MaxDelay,
	}, func() error {
		resp, err := c.doRequest(ctx, http.MethodPost, certificatesURL, req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("failed to create certificate - [%d] - [%s]", resp.StatusCode, string(bodyBytes))
		}

		if err := json.Unmarshal(bodyBytes, &cert); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// GetCertificate retrieves a certificate by ID from SIA.
// API Endpoint: GET /api/certificates/{id}
func (c *CertificatesClientOAuth2) GetCertificate(ctx context.Context, id string) (*Certificate, error) {
	if id == "" {
		return nil, fmt.Errorf("certificate ID cannot be empty")
	}

	path := fmt.Sprintf(certificateURL, id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate %s: %w", id, err)
	}
	defer resp.Body.Close()

	// Handle 404 for drift detection
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get certificate %s - [%d] - [%s]", id, resp.StatusCode, string(bodyBytes))
	}

	var cert Certificate
	if err := json.Unmarshal(bodyBytes, &cert); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cert, nil
}

// UpdateCertificate updates an existing certificate in SIA.
// API Endpoint: PUT /api/certificates/{id}
func (c *CertificatesClientOAuth2) UpdateCertificate(ctx context.Context, id string, req *CertificateUpdateRequest) (*Certificate, error) {
	if req.CertBody == "" {
		return nil, fmt.Errorf("cert_body is required for all certificate updates (API requirement)")
	}

	path := fmt.Sprintf(certificateURL, id)

	var cert Certificate
	err := RetryWithBackoff(ctx, &RetryConfig{
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  BaseDelay,
		MaxDelay:   MaxDelay,
	}, func() error {
		resp, err := c.doRequest(ctx, http.MethodPut, path, req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to update certificate - [%d] - [%s]", resp.StatusCode, string(bodyBytes))
		}

		if err := json.Unmarshal(bodyBytes, &cert); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// DeleteCertificate deletes a certificate from SIA by ID.
// API Endpoint: DELETE /api/certificates/{id}
func (c *CertificatesClientOAuth2) DeleteCertificate(ctx context.Context, certificateID string) error {
	if certificateID == "" {
		return fmt.Errorf("certificate ID cannot be empty")
	}

	path := fmt.Sprintf(certificateURL, certificateID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete certificate %s: %w", certificateID, err)
	}
	defer resp.Body.Close()

	// Handle 404 as success (already deleted)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete certificate %s - [%d] - [%s]", certificateID, resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ListCertificates retrieves all certificates from SIA.
// API Endpoint: GET /api/certificates
func (c *CertificatesClientOAuth2) ListCertificates(ctx context.Context) ([]CertificateListItem, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, certificatesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list certificates - [%d] - [%s]", resp.StatusCode, string(bodyBytes))
	}

	var response CertificateListResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Certificates == nil || response.Certificates.Items == nil {
		return []CertificateListItem{}, nil
	}

	return response.Certificates.Items, nil
}
