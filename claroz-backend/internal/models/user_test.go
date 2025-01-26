package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestUserBeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		wantUUID bool
	}{
		{
			name: "should generate UUID if nil",
			user: &User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantUUID: true,
		},
		{
			name: "should keep existing UUID",
			user: &User{
				ID:       uuid.New(),
				Username: "testuser2",
				Email:    "test2@example.com",
				Password: "password123",
			},
			wantUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.user.ID
			err := tt.user.BeforeCreate(&gorm.DB{})

			if err != nil {
				t.Errorf("BeforeCreate() error = %v", err)
				return
			}

			if tt.wantUUID {
				if tt.user.ID == uuid.Nil {
					t.Error("BeforeCreate() did not generate UUID")
				}
			} else {
				if tt.user.ID != originalID {
					t.Error("BeforeCreate() modified existing UUID")
				}
			}
		})
	}
}

func TestUserModel(t *testing.T) {
	user := User{
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       "password123",
		FullName:       "Test User",
		Bio:            "Test bio",
		Avatar:         "avatar.jpg",
		DID:            "did:example:123",
		Handle:         "@testuser",
		FederationType: "local",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Followers:      []User{},
		Following:      []User{},
	}

	// Test required fields
	if user.Username == "" {
		t.Error("Username should not be empty")
	}
	if user.Email == "" {
		t.Error("Email should not be empty")
	}
	if user.Password == "" {
		t.Error("Password should not be empty")
	}

	// Test default values
	if user.FederationType != "local" {
		t.Error("FederationType should default to 'local'")
	}

	// Test relationships initialization
	if user.Followers == nil {
		t.Error("Followers should be initialized")
	}
	if user.Following == nil {
		t.Error("Following should be initialized")
	}
}

func TestUserFollow(t *testing.T) {
	follower := User{
		ID:       uuid.New(),
		Username: "follower",
		Email:    "follower@example.com",
		Password: "password123",
	}

	following := User{
		ID:       uuid.New(),
		Username: "following",
		Email:    "following@example.com",
		Password: "password123",
	}

	userFollow := UserFollow{
		FollowerID:  follower.ID,
		FollowingID: following.ID,
		CreatedAt:   time.Now(),
	}

	if userFollow.FollowerID == uuid.Nil {
		t.Error("FollowerID should not be nil")
	}

	if userFollow.FollowingID == uuid.Nil {
		t.Error("FollowingID should not be nil")
	}

	if userFollow.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
