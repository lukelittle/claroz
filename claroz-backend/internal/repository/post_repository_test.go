package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/testutils"
)

var userCounter = 0

func createTestUser(t *testing.T, repo *UserRepository) *models.User {
	userCounter++
	user := &models.User{
		Username:       fmt.Sprintf("testuser%d", userCounter),
		Email:          fmt.Sprintf("test%d@example.com", userCounter),
		Password:       "password123",
		Handle:         fmt.Sprintf("@testuser%d", userCounter),
		DID:            fmt.Sprintf("did:plc:testuser%d", userCounter),
		FederationType: "local",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestPostRepository_CreatePost(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	userRepo := NewUserRepository(db.DB)
	postRepo := NewPostRepository(db.DB)
	user := createTestUser(t, userRepo)

	t.Run("create valid post", func(t *testing.T) {
		post := &models.Post{
			UserID:   user.ID,
			Caption:  "Test post",
			ImageURL: "test.jpg",
		}

		err := postRepo.CreatePost(post)
		if err != nil {
			t.Errorf("Failed to create post: %v", err)
		}

		if post.ID == uuid.Nil {
			t.Error("Expected post ID to be set")
		}

		// Verify post was created with relationships
		created, err := postRepo.GetPostByID(post.ID)
		if err != nil {
			t.Errorf("Failed to get created post: %v", err)
		}
		if created.Caption != post.Caption {
			t.Errorf("Expected caption %s, got %s", post.Caption, created.Caption)
		}
		if created.User.ID != user.ID {
			t.Error("User relationship not properly set")
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestPostRepository_GetPosts(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	userRepo := NewUserRepository(db.DB)
	postRepo := NewPostRepository(db.DB)
	user := createTestUser(t, userRepo)

	// Create multiple posts
	posts := make([]*models.Post, 5)
	for i := 0; i < 5; i++ {
		post := &models.Post{
			UserID:    user.ID,
			Caption:   "Test post",
			ImageURL:  "test.jpg",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour), // Different creation times
		}
		err := postRepo.CreatePost(post)
		if err != nil {
			t.Fatalf("Failed to create test post: %v", err)
		}
		posts[i] = post
	}

	t.Run("get posts with pagination", func(t *testing.T) {
		fetchedPosts, err := postRepo.GetPosts(1, 3)
		if err != nil {
			t.Errorf("Failed to get posts: %v", err)
		}

		if len(fetchedPosts) != 3 {
			t.Errorf("Expected 3 posts, got %d", len(fetchedPosts))
		}

		// Verify ordering (newest first)
		for i := 1; i < len(fetchedPosts); i++ {
			if fetchedPosts[i-1].CreatedAt.Before(fetchedPosts[i].CreatedAt) {
				t.Error("Posts not properly ordered by creation date")
			}
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestPostRepository_Comments(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	userRepo := NewUserRepository(db.DB)
	postRepo := NewPostRepository(db.DB)
	user := createTestUser(t, userRepo)

	post := &models.Post{
		UserID:   user.ID,
		Caption:  "Test post",
		ImageURL: "test.jpg",
	}
	err := postRepo.CreatePost(post)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	t.Run("add and delete comment", func(t *testing.T) {
		comment := &models.Comment{
			PostID:  post.ID,
			UserID:  user.ID,
			Content: "Test comment",
		}

		err := postRepo.AddComment(comment)
		if err != nil {
			t.Errorf("Failed to add comment: %v", err)
		}

		// Verify comment was added
		postWithComments, err := postRepo.GetPostByID(post.ID)
		if err != nil {
			t.Errorf("Failed to get post with comments: %v", err)
		}
		if len(postWithComments.Comments) != 1 {
			t.Errorf("Expected 1 comment, got %d", len(postWithComments.Comments))
		}

		// Delete comment
		err = postRepo.DeleteComment(comment.ID, user.ID)
		if err != nil {
			t.Errorf("Failed to delete comment: %v", err)
		}

		// Verify comment was deleted
		postWithComments, err = postRepo.GetPostByID(post.ID)
		if err != nil {
			t.Errorf("Failed to get post after comment deletion: %v", err)
		}
		if len(postWithComments.Comments) != 0 {
			t.Error("Expected no comments after deletion")
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestPostRepository_Likes(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	userRepo := NewUserRepository(db.DB)
	postRepo := NewPostRepository(db.DB)
	user := createTestUser(t, userRepo)

	post := &models.Post{
		UserID:   user.ID,
		Caption:  "Test post",
		ImageURL: "test.jpg",
	}
	err := postRepo.CreatePost(post)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	t.Run("like and unlike post", func(t *testing.T) {
		like := &models.Like{
			PostID: post.ID,
			UserID: user.ID,
		}

		// Like post
		err := postRepo.LikePost(like)
		if err != nil {
			t.Errorf("Failed to like post: %v", err)
		}

		// Verify like was added
		hasLiked, err := postRepo.HasUserLikedPost(post.ID, user.ID)
		if err != nil {
			t.Errorf("Failed to check if user liked post: %v", err)
		}
		if !hasLiked {
			t.Error("Expected user to have liked post")
		}

		// Unlike post
		err = postRepo.UnlikePost(post.ID, user.ID)
		if err != nil {
			t.Errorf("Failed to unlike post: %v", err)
		}

		// Verify like was removed
		hasLiked, err = postRepo.HasUserLikedPost(post.ID, user.ID)
		if err != nil {
			t.Errorf("Failed to check if user liked post: %v", err)
		}
		if hasLiked {
			t.Error("Expected user to have unliked post")
		}
	})

	t.Run("get post likes count", func(t *testing.T) {
		// Add multiple likes
		for i := 0; i < 3; i++ {
			otherUser := createTestUser(t, userRepo)
			like := &models.Like{
				PostID: post.ID,
				UserID: otherUser.ID,
			}
			err := postRepo.LikePost(like)
			if err != nil {
				t.Fatalf("Failed to create like: %v", err)
			}
		}

		count, err := postRepo.GetPostLikes(post.ID)
		if err != nil {
			t.Errorf("Failed to get post likes count: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected 3 likes, got %d", count)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestPostRepository_Follow(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	userRepo := NewUserRepository(db.DB)
	postRepo := NewPostRepository(db.DB)
	user1 := createTestUser(t, userRepo)
	user2 := createTestUser(t, userRepo)

	t.Run("follow and unfollow user", func(t *testing.T) {
		// Follow user
		err := postRepo.FollowUser(user1.ID, user2.ID)
		if err != nil {
			t.Errorf("Failed to follow user: %v", err)
		}

		// Verify following status
		isFollowing, err := postRepo.IsFollowing(user1.ID, user2.ID)
		if err != nil {
			t.Errorf("Failed to check following status: %v", err)
		}
		if !isFollowing {
			t.Error("Expected user1 to be following user2")
		}

		// Unfollow user
		err = postRepo.UnfollowUser(user1.ID, user2.ID)
		if err != nil {
			t.Errorf("Failed to unfollow user: %v", err)
		}

		// Verify following status
		isFollowing, err = postRepo.IsFollowing(user1.ID, user2.ID)
		if err != nil {
			t.Errorf("Failed to check following status: %v", err)
		}
		if isFollowing {
			t.Error("Expected user1 to not be following user2")
		}
	})

	t.Run("get followers and following count", func(t *testing.T) {
		// Create multiple followers
		for i := 0; i < 3; i++ {
			follower := createTestUser(t, userRepo)
			err := postRepo.FollowUser(follower.ID, user1.ID)
			if err != nil {
				t.Fatalf("Failed to create follow relationship: %v", err)
			}
		}

		followersCount, err := postRepo.GetFollowersCount(user1.ID)
		if err != nil {
			t.Errorf("Failed to get followers count: %v", err)
		}
		if followersCount != 3 {
			t.Errorf("Expected 3 followers, got %d", followersCount)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}
