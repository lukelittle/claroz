package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/utils"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	return router
}

func TestAuthMiddleware(t *testing.T) {
	router := setupTestRouter()
	testUserID := uuid.New()

	// Generate a valid token for testing
	validToken, err := utils.GenerateToken(testUserID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
		checkUserID  bool
	}{
		{
			name:         "valid token",
			authHeader:   "Bearer " + validToken,
			expectedCode: http.StatusOK,
			checkUserID:  true,
		},
		{
			name:         "missing authorization header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid header format - no bearer",
			authHeader:   validToken,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid header format - wrong prefix",
			authHeader:   "NotBearer " + validToken,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalidtoken",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "expired token",
			authHeader:   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDU2Nzg5MCIsImV4cCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.checkUserID && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Check if userID in response matches the test user ID
				if responseUserID, ok := response["userID"].(string); !ok {
					t.Error("userID not found in response or not a string")
				} else {
					if uuid.MustParse(responseUserID) != testUserID {
						t.Errorf("Expected userID %v, got %v", testUserID, responseUserID)
					}
				}
			}
		})
	}
}

func TestAuthMiddleware_UserIDInContext(t *testing.T) {
	router := gin.New()
	testUserID := uuid.New()

	// Generate a valid token
	validToken, err := utils.GenerateToken(testUserID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Add middleware and test handler
	router.Use(AuthMiddleware())
	router.GET("/test-context", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			t.Error("userID not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}

		// Check if the userID is of the correct type and value
		if id, ok := userID.(uuid.UUID); !ok {
			t.Error("userID is not of type uuid.UUID")
		} else if id != testUserID {
			t.Errorf("Expected userID %v, got %v", testUserID, id)
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test-context", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_MultipleRequests(t *testing.T) {
	router := setupTestRouter()
	testUserID := uuid.New()

	// Generate a valid token
	validToken, err := utils.GenerateToken(testUserID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Make multiple requests with the same token
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status code %d, got %d", i+1, http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Request %d: Failed to decode response: %v", i+1, err)
		}

		if responseUserID, ok := response["userID"].(string); !ok {
			t.Errorf("Request %d: userID not found in response or not a string", i+1)
		} else if uuid.MustParse(responseUserID) != testUserID {
			t.Errorf("Request %d: Expected userID %v, got %v", i+1, testUserID, responseUserID)
		}
	}
}
