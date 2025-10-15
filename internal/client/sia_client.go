// Package client provides CyberArk SIA API client wrappers
package client

import (
	"fmt"

	"github.com/cyberark/ark-sdk-golang/pkg/auth"
	"github.com/cyberark/ark-sdk-golang/pkg/services/sia"
)

// NewSIAClient creates a new SIA API client using authenticated ARK SDK
func NewSIAClient(ispAuth *auth.ArkISPAuth) (*sia.ArkSIAAPI, error) {
	if ispAuth == nil {
		return nil, fmt.Errorf("ispAuth cannot be nil")
	}

	// Initialize SIA API client with authenticated ISP Auth
	siaAPI, err := sia.NewArkSIAAPI(ispAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SIA API client: %w", err)
	}

	return siaAPI, nil
}
