package repository

import (
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPostByID(id uuid.UUID) (*models.Post, error)
	GetPosts(page, pageSize int) ([]models.Post, error)
	DeletePost(id uuid.UUID, userID uuid.UUID) error
	AddComment(comment *models.Comment) error
	DeleteComment(id uuid.UUID, userID uuid.UUID) error
	LikePost(like *models.Like) error
	UnlikePost(postID, userID uuid.UUID) error
	HasUserLikedPost(postID, userID uuid.UUID) (bool, error)
	GetPostLikes(postID uuid.UUID) (int64, error)
	GetUserPosts(userID uuid.UUID) ([]models.Post, error)
	FollowUser(followerID, followingID uuid.UUID) error
	UnfollowUser(followerID, followingID uuid.UUID) error
	IsFollowing(followerID, followingID uuid.UUID) (bool, error)
	GetFollowersCount(userID uuid.UUID) (int64, error)
	GetFollowingCount(userID uuid.UUID) (int64, error)
}
