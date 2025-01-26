package federation

import (
	"context"
	"fmt"
)

// ATProtoClientInterface defines the interface for AT Protocol client operations
type ATProtoClientInterface interface {
	ResolveHandle(ctx context.Context, handle string) (*FederatedProfile, error)
	GetProfile(ctx context.Context, did string) (*FederatedProfile, error)
}

// ErrProfileNotFound is returned when a profile cannot be found
var ErrProfileNotFound = fmt.Errorf("profile not found")
