package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/repository"
	"github.com/lukelittle/claroz/claroz-backend/internal/utils"
)

type PostHandler struct {
	postRepo repository.PostRepositoryInterface
	storage  utils.FileStorageInterface
}

// CommentRequest represents a comment creation request
type CommentRequest struct {
	Content string `json:"content" binding:"required" example:"Great post!"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

func NewPostHandler(postRepo repository.PostRepositoryInterface, storage utils.FileStorageInterface) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		storage:  storage,
	}
}

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post with an image and caption
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param image formData file true "Image file"
// @Param caption formData string false "Post caption"
// @Success 201 {object} models.Post
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}

	imageURL, err := h.storage.SaveFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
		return
	}

	post := &models.Post{
		UserID:   userID.(uuid.UUID),
		Caption:  c.PostForm("caption"),
		ImageURL: imageURL,
	}

	if err := h.postRepo.CreatePost(post); err != nil {
		_ = h.storage.DeleteFile(imageURL)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetPost godoc
// @Summary Get a post by ID
// @Description Retrieve a single post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.Post
// @Failure 400 {object} object{error=string} "Invalid post ID"
// @Failure 404 {object} object{error=string} "Post not found"
// @Router /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := h.postRepo.GetPostByID(id)
	if err != nil || post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetPosts godoc
// @Summary Get posts with pagination
// @Description Retrieve a list of posts with pagination support
// @Tags posts
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param pageSize query int false "Page size (default: 10, max: 50)" minimum(1) maximum(50)
// @Success 200 {array} models.Post
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	posts, err := h.postRepo.GetPosts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// DeletePost godoc
// @Summary Delete a post
// @Description Delete a post by ID (only by post owner)
// @Tags posts
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Post ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} object{error=string} "Invalid post ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "Post not found"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := h.postRepo.GetPostByID(postID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	if err := h.postRepo.DeletePost(postID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}

	_ = h.storage.DeleteFile(post.ImageURL)

	c.JSON(http.StatusOK, MessageResponse{Message: "post deleted successfully"})
}

// AddComment godoc
// @Summary Add a comment to a post
// @Description Add a new comment to a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Post ID"
// @Param comment body CommentRequest true "Comment content"
// @Success 201 {object} models.Comment
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts/{id}/comments [post]
func (h *PostHandler) AddComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var input CommentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	comment := &models.Comment{
		PostID:  postID,
		UserID:  userID.(uuid.UUID),
		Content: input.Content,
	}

	if err := h.postRepo.AddComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// LikePost godoc
// @Summary Like a post
// @Description Add a like to a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Post ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} object{error=string} "Invalid post ID or already liked"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts/{id}/like [post]
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	hasLiked, err := h.postRepo.HasUserLikedPost(postID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check like status"})
		return
	}

	if hasLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post already liked"})
		return
	}

	like := &models.Like{
		PostID: postID,
		UserID: userID.(uuid.UUID),
	}

	if err := h.postRepo.LikePost(like); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like post"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "post liked successfully"})
}

// UnlikePost godoc
// @Summary Unlike a post
// @Description Remove a like from a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Post ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} object{error=string} "Invalid post ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /posts/{id}/like [delete]
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	if err := h.postRepo.UnlikePost(postID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike post"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "post unliked successfully"})
}

// FollowUser godoc
// @Summary Follow a user
// @Description Follow another user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID to follow"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} object{error=string} "Invalid user ID or cannot follow yourself"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{id}/follow [post]
func (h *PostHandler) FollowUser(c *gin.Context) {
	followerID, _ := c.Get("userID")
	followingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if followerID.(uuid.UUID) == followingID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	if err := h.postRepo.FollowUser(followerID.(uuid.UUID), followingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "user followed successfully"})
}

// UnfollowUser godoc
// @Summary Unfollow a user
// @Description Unfollow a previously followed user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID to unfollow"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} object{error=string} "Invalid user ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{id}/follow [delete]
func (h *PostHandler) UnfollowUser(c *gin.Context) {
	followerID, _ := c.Get("userID")
	followingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.postRepo.UnfollowUser(followerID.(uuid.UUID), followingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "user unfollowed successfully"})
}

// GetUserPosts godoc
// @Summary Get user's posts
// @Description Retrieve all posts for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} models.Post
// @Failure 400 {object} object{error=string} "Invalid user ID"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{userId}/posts [get]
func (h *PostHandler) GetUserPosts(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	posts, err := h.postRepo.GetUserPosts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user posts"})
		return
	}

	c.JSON(http.StatusOK, posts)
}
