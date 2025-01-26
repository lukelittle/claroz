package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestPostBeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		post     *Post
		wantUUID bool
	}{
		{
			name: "should generate UUID if nil",
			post: &Post{
				UserID:   uuid.New(),
				Caption:  "Test post",
				ImageURL: "test.jpg",
			},
			wantUUID: true,
		},
		{
			name: "should keep existing UUID",
			post: &Post{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				Caption:  "Test post",
				ImageURL: "test.jpg",
			},
			wantUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.post.ID
			err := tt.post.BeforeCreate(&gorm.DB{})

			if err != nil {
				t.Errorf("BeforeCreate() error = %v", err)
				return
			}

			if tt.wantUUID {
				if tt.post.ID == uuid.Nil {
					t.Error("BeforeCreate() did not generate UUID")
				}
			} else {
				if tt.post.ID != originalID {
					t.Error("BeforeCreate() modified existing UUID")
				}
			}
		})
	}
}

func TestCommentBeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		comment  *Comment
		wantUUID bool
	}{
		{
			name: "should generate UUID if nil",
			comment: &Comment{
				PostID:  uuid.New(),
				UserID:  uuid.New(),
				Content: "Test comment",
			},
			wantUUID: true,
		},
		{
			name: "should keep existing UUID",
			comment: &Comment{
				ID:      uuid.New(),
				PostID:  uuid.New(),
				UserID:  uuid.New(),
				Content: "Test comment",
			},
			wantUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.comment.ID
			err := tt.comment.BeforeCreate(&gorm.DB{})

			if err != nil {
				t.Errorf("BeforeCreate() error = %v", err)
				return
			}

			if tt.wantUUID {
				if tt.comment.ID == uuid.Nil {
					t.Error("BeforeCreate() did not generate UUID")
				}
			} else {
				if tt.comment.ID != originalID {
					t.Error("BeforeCreate() modified existing UUID")
				}
			}
		})
	}
}

func TestLikeBeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		like     *Like
		wantUUID bool
	}{
		{
			name: "should generate UUID if nil",
			like: &Like{
				PostID: uuid.New(),
				UserID: uuid.New(),
			},
			wantUUID: true,
		},
		{
			name: "should keep existing UUID",
			like: &Like{
				ID:     uuid.New(),
				PostID: uuid.New(),
				UserID: uuid.New(),
			},
			wantUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.like.ID
			err := tt.like.BeforeCreate(&gorm.DB{})

			if err != nil {
				t.Errorf("BeforeCreate() error = %v", err)
				return
			}

			if tt.wantUUID {
				if tt.like.ID == uuid.Nil {
					t.Error("BeforeCreate() did not generate UUID")
				}
			} else {
				if tt.like.ID != originalID {
					t.Error("BeforeCreate() modified existing UUID")
				}
			}
		})
	}
}

func TestPostModel(t *testing.T) {
	userID := uuid.New()
	post := Post{
		UserID:    userID,
		Caption:   "Test caption",
		ImageURL:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Likes:     []Like{},
		Comments:  []Comment{},
	}

	// Test required fields
	if post.UserID == uuid.Nil {
		t.Error("UserID should not be nil")
	}
	if post.ImageURL == "" {
		t.Error("ImageURL should not be empty")
	}

	// Test timestamps
	if post.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if post.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Test relationships initialization
	if post.Likes == nil {
		t.Error("Likes should be initialized")
	}
	if post.Comments == nil {
		t.Error("Comments should be initialized")
	}
}

func TestCommentModel(t *testing.T) {
	comment := Comment{
		PostID:    uuid.New(),
		UserID:    uuid.New(),
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test required fields
	if comment.PostID == uuid.Nil {
		t.Error("PostID should not be nil")
	}
	if comment.UserID == uuid.Nil {
		t.Error("UserID should not be nil")
	}
	if comment.Content == "" {
		t.Error("Content should not be empty")
	}

	// Test timestamps
	if comment.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if comment.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestLikeModel(t *testing.T) {
	like := Like{
		PostID:    uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Now(),
	}

	// Test required fields
	if like.PostID == uuid.Nil {
		t.Error("PostID should not be nil")
	}
	if like.UserID == uuid.Nil {
		t.Error("UserID should not be nil")
	}

	// Test timestamp
	if like.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
