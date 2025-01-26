package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"gorm.io/gorm"
)

// PostRepository implements PostRepositoryInterface
type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepositoryInterface {
	return &PostRepository{db: db}
}

// CreatePost creates a new post
func (r *PostRepository) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}

// GetPostByID retrieves a post by ID with associated user, comments, and likes
func (r *PostRepository) GetPostByID(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").
		Preload("Comments.User").
		Preload("Likes.User").
		First(&post, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPosts retrieves posts with pagination, ordered by creation date
func (r *PostRepository) GetPosts(page, pageSize int) ([]models.Post, error) {
	var posts []models.Post
	offset := (page - 1) * pageSize

	err := r.db.
		Preload("User").
		Preload("Comments.User").
		Preload("Likes.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

// DeletePost deletes a post and its associated comments and likes
func (r *PostRepository) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verify post exists and belongs to user
		var post models.Post
		if err := tx.First(&post, "id = ? AND user_id = ?", id, userID).Error; err != nil {
			return err
		}

		// Delete associated likes
		if err := tx.Delete(&models.Like{}, "post_id = ?", id).Error; err != nil {
			return err
		}

		// Delete associated comments
		if err := tx.Delete(&models.Comment{}, "post_id = ?", id).Error; err != nil {
			return err
		}

		// Delete the post
		return tx.Delete(&post).Error
	})
}

// AddComment adds a comment to a post
func (r *PostRepository) AddComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// DeleteComment deletes a comment
func (r *PostRepository) DeleteComment(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ? AND user_id = ?", id, userID).Error
}

// LikePost creates a new like for a post
func (r *PostRepository) LikePost(like *models.Like) error {
	return r.db.Create(like).Error
}

// UnlikePost removes a like from a post
func (r *PostRepository) UnlikePost(postID, userID uuid.UUID) error {
	return r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.Like{}).Error
}

// HasUserLikedPost checks if a user has already liked a post
func (r *PostRepository) HasUserLikedPost(postID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetPostLikes gets the total number of likes for a post
func (r *PostRepository) GetPostLikes(postID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}

// GetUserPosts retrieves all posts for a specific user
func (r *PostRepository) GetUserPosts(userID uuid.UUID) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.
		Preload("User").
		Preload("Comments.User").
		Preload("Likes.User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

// FollowUser creates a new follow relationship
func (r *PostRepository) FollowUser(followerID, followingID uuid.UUID) error {
	follow := models.UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   time.Now(),
	}
	return r.db.Create(&follow).Error
}

// UnfollowUser removes a follow relationship
func (r *PostRepository) UnfollowUser(followerID, followingID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.UserFollow{}).Error
}

// IsFollowing checks if a user is following another user
func (r *PostRepository) IsFollowing(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowersCount gets the number of followers for a user
func (r *PostRepository) GetFollowersCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Where("following_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetFollowingCount gets the number of users a user is following
func (r *PostRepository) GetFollowingCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}
