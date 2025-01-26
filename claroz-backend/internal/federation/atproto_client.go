package federation

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ATProtoClient struct {
	client  *http.Client
	pdsHost string
}

type FederatedProfile struct {
	DID         string
	Handle      string
	DisplayName string
	Description string
	Avatar      string
}

func NewATProtoClient(pdsHost string) (*ATProtoClient, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	return &ATProtoClient{
		client:  client,
		pdsHost: pdsHost,
	}, nil
}

func (c *ATProtoClient) ResolveHandle(ctx context.Context, handle string) (*FederatedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/xrpc/com.atproto.identity.resolveHandle?handle=%s", c.pdsHost, handle), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve handle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resolve handle, status: %d", resp.StatusCode)
	}

	// Parse response and extract DID
	// For now return a mock profile
	return &FederatedProfile{
		DID:         "did:plc:" + handle,
		Handle:      handle,
		DisplayName: handle,
	}, nil
}

func (c *ATProtoClient) GetProfile(ctx context.Context, did string) (*FederatedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/xrpc/app.bsky.actor.getProfile?actor=%s", c.pdsHost, did), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile, status: %d", resp.StatusCode)
	}

	// Parse response and create profile
	// For now return a mock profile
	return &FederatedProfile{
		DID:         did,
		Handle:      did[8:], // Remove "did:plc:" prefix
		DisplayName: did[8:],
	}, nil
}
