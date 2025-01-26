package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
)

// MockPostRepository implements necessary methods for testing
type MockPostRepository struct {
	posts    map[uuid.UUID]*models.Post
	likes    map[uuid.UUID]map[uuid.UUID]bool // postID -> userID -> liked
	follows  map[uuid.UUID]map[uuid.UUID]bool // followerID -> followingID -> following
	comments map[uuid.UUID][]*models.Comment  // postID -> comments
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts:    make(map[uuid.UUID]*models.Post),
		likes:    make(map[uuid.UUID]map[uuid.UUID]bool),
		follows:  make(map[uuid.UUID]map[uuid.UUID]bool),
		comments: make(map[uuid.UUID][]*models.Comment),
	}
}

func (m *MockPostRepository) CreatePost(post *models.Post) error {
	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}
	m.posts[post.ID] = post
	return nil
}

func (m *MockPostRepository) GetPostByID(id uuid.UUID) (*models.Post, error) {
	if post, exists := m.posts[id]; exists {
		return post, nil
	}
	return nil, nil
}

func (m *MockPostRepository) GetPosts(page, pageSize int) ([]models.Post, error) {
	var posts []models.Post
	for _, post := range m.posts {
		posts = append(posts, *post)
	}
	return posts, nil
}

func (m *MockPostRepository) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	if post, exists := m.posts[id]; exists && post.UserID == userID {
		delete(m.posts, id)
		return nil
	}
	return nil
}

func (m *MockPostRepository) AddComment(comment *models.Comment) error {
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	m.comments[comment.PostID] = append(m.comments[comment.PostID], comment)
	return nil
}

func (m *MockPostRepository) DeleteComment(id uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (m *MockPostRepository) LikePost(like *models.Like) error {
	if _, exists := m.likes[like.PostID]; !exists {
		m.likes[like.PostID] = make(map[uuid.UUID]bool)
	}
	m.likes[like.PostID][like.UserID] = true
	return nil
}

func (m *MockPostRepository) UnlikePost(postID, userID uuid.UUID) error {
	if likes, exists := m.likes[postID]; exists {
		delete(likes, userID)
		if len(likes) == 0 {
			delete(m.likes, postID)
		}
	}
	return nil
}

func (m *MockPostRepository) HasUserLikedPost(postID, userID uuid.UUID) (bool, error) {
	if likes, exists := m.likes[postID]; exists {
		return likes[userID], nil
	}
	return false, nil
}

func (m *MockPostRepository) GetPostLikes(postID uuid.UUID) (int64, error) {
	if likes, exists := m.likes[postID]; exists {
		return int64(len(likes)), nil
	}
	return 0, nil
}

func (m *MockPostRepository) GetUserPosts(userID uuid.UUID) ([]models.Post, error) {
	var posts []models.Post
	for _, post := range m.posts {
		if post.UserID == userID {
			posts = append(posts, *post)
		}
	}
	return posts, nil
}

func (m *MockPostRepository) FollowUser(followerID, followingID uuid.UUID) error {
	if _, exists := m.follows[followerID]; !exists {
		m.follows[followerID] = make(map[uuid.UUID]bool)
	}
	m.follows[followerID][followingID] = true
	return nil
}

func (m *MockPostRepository) UnfollowUser(followerID, followingID uuid.UUID) error {
	if follows, exists := m.follows[followerID]; exists {
		delete(follows, followingID)
	}
	return nil
}

func (m *MockPostRepository) IsFollowing(followerID, followingID uuid.UUID) (bool, error) {
	if follows, exists := m.follows[followerID]; exists {
		return follows[followingID], nil
	}
	return false, nil
}

func (m *MockPostRepository) GetFollowersCount(userID uuid.UUID) (int64, error) {
	count := int64(0)
	for _, follows := range m.follows {
		if follows[userID] {
			count++
		}
	}
	return count, nil
}

func (m *MockPostRepository) GetFollowingCount(userID uuid.UUID) (int64, error) {
	if follows, exists := m.follows[userID]; exists {
		return int64(len(follows)), nil
	}
	return 0, nil
}

// MockFileStorage implements necessary methods for testing
type MockFileStorage struct {
	files map[string][]byte
}

func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{
		files: make(map[string][]byte),
	}
}

func (m *MockFileStorage) SaveFile(file *multipart.FileHeader) (string, error) {
	return "test-image-url.jpg", nil
}

func (m *MockFileStorage) DeleteFile(path string) error {
	delete(m.files, path)
	return nil
}

func setupPostTestRouter() (*gin.Engine, *MockPostRepository, *MockFileStorage) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockRepo := NewMockPostRepository()
	mockStorage := NewMockFileStorage()
	postHandler := NewPostHandler(mockRepo, mockStorage)

	// Add middleware to set test user ID
	router.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})

	router.POST("/posts", postHandler.CreatePost)
	router.GET("/posts/:id", postHandler.GetPost)
	router.GET("/posts", postHandler.GetPosts)
	router.DELETE("/posts/:id", postHandler.DeletePost)
	router.POST("/posts/:id/comments", postHandler.AddComment)
	router.POST("/posts/:id/like", postHandler.LikePost)
	router.DELETE("/posts/:id/like", postHandler.UnlikePost)
	router.POST("/users/:id/follow", postHandler.FollowUser)
	router.DELETE("/users/:id/follow", postHandler.UnfollowUser)
	router.GET("/users/:userId/posts", postHandler.GetUserPosts)

	return router, mockRepo, mockStorage
}

func TestPostHandler_CreatePost(t *testing.T) {
	router, _, _ := setupPostTestRouter()

	t.Run("create post with valid data", func(t *testing.T) {
		// Create multipart form data
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("image", "test.jpg")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		part.Write([]byte("test image content"))
		writer.WriteField("caption", "Test caption")
		writer.Close()

		req := httptest.NewRequest("POST", "/posts", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Post
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Caption != "Test caption" {
			t.Errorf("Expected caption %s, got %s", "Test caption", response.Caption)
		}
		if response.ImageURL == "" {
			t.Error("Expected image URL to be set")
		}
	})

	t.Run("create post without image", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("caption", "Test caption")
		writer.Close()

		req := httptest.NewRequest("POST", "/posts", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestPostHandler_GetPost(t *testing.T) {
	router, mockRepo, _ := setupPostTestRouter()

	// Create a test post
	testPost := &models.Post{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Caption:  "Test post",
		ImageURL: "test.jpg",
	}
	mockRepo.CreatePost(testPost)

	t.Run("get existing post", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/posts/"+testPost.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Post
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != testPost.ID {
			t.Errorf("Expected post ID %v, got %v", testPost.ID, response.ID)
		}
	})

	t.Run("get non-existent post", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/posts/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestPostHandler_LikeUnlike(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := NewMockPostRepository()
	mockStorage := NewMockFileStorage()
	postHandler := NewPostHandler(mockRepo, mockStorage)

	testPost := &models.Post{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Caption:  "Test post",
		ImageURL: "test.jpg",
	}
	mockRepo.CreatePost(testPost)

	currentUserID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", currentUserID)
		c.Next()
	})

	router.POST("/posts/:id/like", postHandler.LikePost)
	router.DELETE("/posts/:id/like", postHandler.UnlikePost)

	t.Run("like post", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/"+testPost.ID.String()+"/like", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify like was added
		likes, _ := mockRepo.GetPostLikes(testPost.ID)
		if likes != 1 {
			t.Errorf("Expected 1 like, got %d", likes)
		}
	})

	t.Run("unlike post", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/"+testPost.ID.String()+"/like", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify like was removed
		likes, _ := mockRepo.GetPostLikes(testPost.ID)
		if likes != 0 {
			t.Errorf("Expected 0 likes, got %d", likes)
		}
	})
}

func TestPostHandler_FollowUnfollow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := NewMockPostRepository()
	mockStorage := NewMockFileStorage()
	postHandler := NewPostHandler(mockRepo, mockStorage)

	testUserID := uuid.New()
	currentUserID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", currentUserID)
		c.Next()
	})

	router.POST("/users/:id/follow", postHandler.FollowUser)
	router.DELETE("/users/:id/follow", postHandler.UnfollowUser)

	t.Run("follow user", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users/"+testUserID.String()+"/follow", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify follow relationship
		isFollowing, _ := mockRepo.IsFollowing(currentUserID, testUserID)
		if !isFollowing {
			t.Error("Expected user to be following")
		}
	})

	t.Run("unfollow user", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/"+testUserID.String()+"/follow", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify follow relationship was removed
		isFollowing, _ := mockRepo.IsFollowing(currentUserID, testUserID)
		if isFollowing {
			t.Error("Expected user to not be following")
		}
	})
}
