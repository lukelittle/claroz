package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserFollow struct {
	FollowerID  uuid.UUID `gorm:"type:uuid;not null"`
	FollowingID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
}

// User represents a user in the system
type User struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username           string         `json:"username" gorm:"uniqueIndex;not null" example:"johndoe"`
	Email              string         `json:"email" gorm:"uniqueIndex;not null" example:"john@example.com"`
	Password           string         `json:"-" gorm:"not null"` // "-" excludes from JSON
	FullName           string         `json:"full_name" example:"John Doe"`
	Bio                string         `json:"bio" example:"Software engineer and tech enthusiast"`
	Avatar             string         `json:"avatar" example:"https://example.com/avatar.jpg"`
	DID                string         `json:"did" gorm:"uniqueIndex" example:"did:web:example.com"`
	Handle             string         `json:"handle" gorm:"uniqueIndex" example:"@johndoe"`
	FederationType     string         `json:"federation_type" gorm:"default:local" example:"local"`
	LastFederationSync time.Time      `json:"last_federation_sync" example:"2024-01-26T00:35:27Z"`
	CreatedAt          time.Time      `json:"created_at" example:"2024-01-26T00:35:27Z"`
	UpdatedAt          time.Time      `json:"updated_at" example:"2024-01-26T00:35:27Z"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`

	// Self-referential many-to-many relationships for followers/following
	Followers []User `json:"followers,omitempty" gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowingID;References:ID;joinReferences:FollowerID"`
	Following []User `json:"following,omitempty" gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowerID;References:ID;joinReferences:FollowingID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Followers == nil {
		u.Followers = []User{}
	}
	if u.Following == nil {
		u.Following = []User{}
	}
	return nil
}
