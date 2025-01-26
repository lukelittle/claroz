package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/utils"
	"gorm.io/gorm"
)

// MockUserRepository implements UserRepositoryInterface for testing
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) Update(user *models.User) error {
	if _, exists := m.users[user.Email]; !exists {
		return gorm.ErrRecordNotFound
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *MockUserRepository) FindByHandle(handle string) (*models.User, error) {
	for _, user := range m.users {
		if user.Handle == handle {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) FindByDID(did string) (*models.User, error) {
	for _, user := range m.users {
		if user.DID == did {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) GetRemoteUsers() ([]*models.User, error) {
	var remoteUsers []*models.User
	for _, user := range m.users {
		if user.FederationType == "remote" {
			remoteUsers = append(remoteUsers, user)
		}
	}
	return remoteUsers, nil
}

func setupTestRouter() (*gin.Engine, *MockUserRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockRepo := NewMockUserRepository()
	authHandler := NewAuthHandler(mockRepo)

	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	return router, mockRepo
}

func TestAuthHandler_Register(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name         string
		request      RegisterRequest
		expectedCode int
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid email",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "short password",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "short",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "missing username",
			request: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if w.Code == http.StatusCreated {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}
				if response.User.Username != tt.request.Username {
					t.Errorf("Expected username %s, got %s", tt.request.Username, response.User.Username)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Create a test user
	hashedPassword, _ := utils.HashPassword("password123")
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.Create(testUser)

	tests := []struct {
		name         string
		request      LoginRequest
		expectedCode int
	}{
		{
			name: "valid login",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid email",
			request: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "wrong password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "invalid email format",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if w.Code == http.StatusOK {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				// Verify token is valid
				claims, err := utils.ValidateToken(response.Token)
				if err != nil {
					t.Errorf("Invalid token: %v", err)
				}
				if claims.UserID != testUser.ID {
					t.Error("Token contains incorrect user ID")
				}
			}
		})
	}
}

func TestAuthHandler_DuplicateEmail(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Create initial user
	mockRepo.Create(&models.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	})

	// Try to register with same email
	req := RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Email already registered" {
		t.Errorf("Expected error message 'Email already registered', got %v", response["error"])
	}
}
