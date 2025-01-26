package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lukelittle/claroz/claroz-backend/internal/federation"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/repository"
)

// MockATProtoClient implements federation.ATProtoClientInterface for testing
type MockATProtoClient struct {
	handleProfiles map[string]*federation.FederatedProfile
	didProfiles    map[string]*federation.FederatedProfile
}

func NewMockATProtoClient() *MockATProtoClient {
	return &MockATProtoClient{
		handleProfiles: make(map[string]*federation.FederatedProfile),
		didProfiles:    make(map[string]*federation.FederatedProfile),
	}
}

// AddProfile adds a profile for testing
func (m *MockATProtoClient) AddProfile(profile *federation.FederatedProfile) {
	m.handleProfiles[profile.Handle] = profile
	m.didProfiles[profile.DID] = profile
}

func (m *MockATProtoClient) ResolveHandle(ctx context.Context, handle string) (*federation.FederatedProfile, error) {
	if profile, exists := m.handleProfiles[handle]; exists {
		return profile, nil
	}
	return nil, federation.ErrProfileNotFound
}

func (m *MockATProtoClient) GetProfile(ctx context.Context, did string) (*federation.FederatedProfile, error) {
	if profile, exists := m.didProfiles[did]; exists {
		return profile, nil
	}
	return nil, federation.ErrProfileNotFound
}

func setupFederationTestRouter() (*gin.Engine, repository.UserRepositoryInterface, federation.ATProtoClientInterface) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := NewMockUserRepository()
	mockClient := NewMockATProtoClient()

	handler := &FederationHandler{
		userRepo:  mockRepo,
		atpClient: mockClient,
	}

	router.GET("/federation/resolve/*handle", handler.ResolveRemoteProfile)
	router.POST("/federation/sync/*did", handler.SyncRemoteProfile)

	return router, mockRepo, mockClient
}

func TestFederationHandler_ResolveRemoteProfile(t *testing.T) {
	router, userRepo, atpClient := setupFederationTestRouter()

	// Set up test data
	testProfile := &federation.FederatedProfile{
		Handle:      "test.bsky.social",
		DID:         "did:plc:test123",
		DisplayName: "Test User",
		Description: "Test bio",
		Avatar:      "https://example.com/avatar.jpg",
	}
	atpClient.(*MockATProtoClient).AddProfile(testProfile)

	tests := []struct {
		name         string
		handle       string
		setupRepo    bool
		expectedCode int
	}{
		{
			name:         "resolve new remote profile",
			handle:       "test.bsky.social",
			setupRepo:    false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "resolve existing profile",
			handle:       "test.bsky.social",
			setupRepo:    true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid handle",
			handle:       "invalid.handle",
			setupRepo:    false,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "empty handle",
			handle:       "",
			setupRepo:    false,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRepo {
				user := &models.User{
					Handle:         testProfile.Handle,
					DID:            testProfile.DID,
					FullName:       testProfile.DisplayName,
					Bio:            testProfile.Description,
					Avatar:         testProfile.Avatar,
					FederationType: "remote",
				}
				userRepo.Create(user)
			}

			url := "/federation/resolve/" + tt.handle
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if w.Code == http.StatusOK {
				var response models.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Handle != testProfile.Handle {
					t.Errorf("Expected handle %s, got %s", testProfile.Handle, response.Handle)
				}
				if response.DID != testProfile.DID {
					t.Errorf("Expected DID %s, got %s", testProfile.DID, response.DID)
				}
			}
		})
	}
}

func TestFederationHandler_SyncRemoteProfile(t *testing.T) {
	router, userRepo, atpClient := setupFederationTestRouter()

	// Set up test data
	testProfile := &federation.FederatedProfile{
		Handle:      "test.bsky.social",
		DID:         "did:plc:test123",
		DisplayName: "Test User",
		Description: "Test bio",
		Avatar:      "https://example.com/avatar.jpg",
	}
	atpClient.(*MockATProtoClient).AddProfile(testProfile)

	// Create initial user
	user := &models.User{
		Handle:         testProfile.Handle,
		DID:            testProfile.DID,
		FullName:       "Old Name",
		Bio:            "Old bio",
		Avatar:         "old-avatar.jpg",
		FederationType: "remote",
	}
	userRepo.Create(user)

	tests := []struct {
		name         string
		did          string
		expectedCode int
	}{
		{
			name:         "sync existing profile",
			did:          testProfile.DID,
			expectedCode: http.StatusOK,
		},
		{
			name:         "non-existent profile",
			did:          "did:plc:nonexistent",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "empty did",
			did:          "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/federation/sync/" + tt.did
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if w.Code == http.StatusOK {
				var response models.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Verify profile was updated
				if response.FullName != testProfile.DisplayName {
					t.Errorf("Expected name %s, got %s", testProfile.DisplayName, response.FullName)
				}
				if response.Bio != testProfile.Description {
					t.Errorf("Expected bio %s, got %s", testProfile.Description, response.Bio)
				}
				if response.Avatar != testProfile.Avatar {
					t.Errorf("Expected avatar %s, got %s", testProfile.Avatar, response.Avatar)
				}
			}
		})
	}
}
