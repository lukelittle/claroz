package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/testutils"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	t.Run("create valid user", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			FullName: "Test User",
			Handle:   "@testuser1",
			DID:      "did:plc:testuser1",
		}

		err := repo.Create(user)
		if err != nil {
			t.Errorf("Failed to create user: %v", err)
		}

		if user.ID == uuid.Nil {
			t.Error("Expected user ID to be set")
		}

		// Verify user was created
		created, err := repo.GetByID(user.ID)
		if err != nil {
			t.Errorf("Failed to get created user: %v", err)
		}
		if created.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, created.Username)
		}
	})

	t.Run("create duplicate email", func(t *testing.T) {
		user1 := &models.User{
			Username: "user1",
			Email:    "duplicate@example.com",
			Password: "password123",
			Handle:   "@user1",
			DID:      "did:plc:user1",
		}
		err := repo.Create(user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		user2 := &models.User{
			Username: "user2",
			Email:    "duplicate@example.com",
			Password: "password123",
			Handle:   "@user2",
			DID:      "did:plc:user2",
		}
		err = repo.Create(user2)
		if err == nil {
			t.Error("Expected error when creating user with duplicate email")
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	t.Run("get existing user", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Handle:   "@testuser3",
			DID:      "did:plc:testuser3",
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		found, err := repo.GetByID(user.ID)
		if err != nil {
			t.Errorf("Failed to get user by ID: %v", err)
		}
		if found.ID != user.ID {
			t.Errorf("Expected user ID %v, got %v", user.ID, found.ID)
		}
	})

	t.Run("get non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(uuid.New())
		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	t.Run("get existing user by email", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Handle:   "@testuser4",
			DID:      "did:plc:testuser4",
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		found, err := repo.GetByEmail(user.Email)
		if err != nil {
			t.Errorf("Failed to get user by email: %v", err)
		}
		if found.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, found.Email)
		}
	})

	t.Run("get non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail("nonexistent@example.com")
		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		FullName: "Original Name",
		Handle:   "@testuser5",
		DID:      "did:plc:testuser5",
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("update user details", func(t *testing.T) {
		user.FullName = "Updated Name"
		user.Bio = "New bio"

		err := repo.Update(user)
		if err != nil {
			t.Errorf("Failed to update user: %v", err)
		}

		updated, err := repo.GetByID(user.ID)
		if err != nil {
			t.Errorf("Failed to get updated user: %v", err)
		}
		if updated.FullName != "Updated Name" {
			t.Errorf("Expected updated name 'Updated Name', got %s", updated.FullName)
		}
		if updated.Bio != "New bio" {
			t.Errorf("Expected updated bio 'New bio', got %s", updated.Bio)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Handle:   "@testuser6",
		DID:      "did:plc:testuser6",
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("delete existing user", func(t *testing.T) {
		err := repo.Delete(user.ID)
		if err != nil {
			t.Errorf("Failed to delete user: %v", err)
		}

		// Verify user was deleted
		_, err = repo.GetByID(user.ID)
		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}

func TestUserRepository_Federation(t *testing.T) {
	db := testutils.NewTestDB(t)
	defer func() {
		if err := db.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}()

	repo := NewUserRepository(db.DB)

	// Clean up any existing data
	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}

	t.Run("find by handle", func(t *testing.T) {
		user := &models.User{
			Username:       "testuser",
			Email:          "test@example.com",
			Password:       "password123",
			Handle:         "@testuser7.bsky.social",
			DID:            "did:plc:testuser7",
			FederationType: "remote",
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		found, err := repo.FindByHandle(user.Handle)
		if err != nil {
			t.Errorf("Failed to find user by handle: %v", err)
		}
		if found.Handle != user.Handle {
			t.Errorf("Expected handle %s, got %s", user.Handle, found.Handle)
		}
	})

	// Clean up data between tests
	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}

	t.Run("find by DID", func(t *testing.T) {
		user := &models.User{
			Username:       "testuser8",
			Email:          "test8@example.com",
			Password:       "password123",
			Handle:         "@testuser8.bsky.social",
			DID:            "did:plc:testuser8",
			FederationType: "remote",
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		found, err := repo.FindByDID(user.DID)
		if err != nil {
			t.Errorf("Failed to find user by DID: %v", err)
		}
		if found.DID != user.DID {
			t.Errorf("Expected DID %s, got %s", user.DID, found.DID)
		}
	})

	// Clean up data between tests
	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}

	t.Run("get remote users", func(t *testing.T) {
		// Create some local and remote users
		users := []*models.User{
			{
				Username:       "local1",
				Email:          "local1@example.com",
				Password:       "password123",
				Handle:         "@local1",
				DID:            "did:plc:local1",
				FederationType: "local",
			},
			{
				Username:           "remote1",
				Email:              "remote1@example.com",
				Password:           "password123",
				Handle:             "@remote1",
				DID:                "did:plc:remote1",
				FederationType:     "remote",
				LastFederationSync: time.Now(),
			},
			{
				Username:           "remote2",
				Email:              "remote2@example.com",
				Password:           "password123",
				Handle:             "@remote2",
				DID:                "did:plc:remote2",
				FederationType:     "remote",
				LastFederationSync: time.Now(),
			},
		}

		for _, u := range users {
			err := repo.Create(u)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}
		}

		remoteUsers, err := repo.GetRemoteUsers()
		if err != nil {
			t.Errorf("Failed to get remote users: %v", err)
		}

		if len(remoteUsers) != 2 {
			t.Errorf("Expected 2 remote users, got %d", len(remoteUsers))
		}

		for _, u := range remoteUsers {
			if u.FederationType != "remote" {
				t.Errorf("Expected federation type 'remote', got %s", u.FederationType)
			}
		}
	})

	if err := db.CleanupData(); err != nil {
		t.Errorf("Failed to cleanup test data: %v", err)
	}
}
