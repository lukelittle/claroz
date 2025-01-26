package testutils

import (
	"fmt"
	"testing"

	"github.com/lukelittle/claroz/claroz-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDB provides a test database instance and cleanup function
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "claroz_test",
			SSLMode:  "disable",
		},
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Drop all tables and recreate them
	err = db.Exec(`DROP TABLE IF EXISTS likes, comments, posts, user_follows, users CASCADE`).Error
	if err != nil {
		t.Fatalf("Failed to drop tables: %v", err)
	}

	// Run migrations
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			full_name TEXT,
			bio TEXT,
			avatar TEXT,
			d_id TEXT UNIQUE,
			handle TEXT UNIQUE,
			federation_type TEXT DEFAULT 'local',
			last_federation_sync TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);

		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			caption TEXT,
			image_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);

		CREATE TABLE IF NOT EXISTS comments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS likes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(post_id, user_id)
		);

		CREATE TABLE IF NOT EXISTS user_follows (
			follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id)
		);

		CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id);
		CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_follows_follower_id ON user_follows(follower_id);
		CREATE INDEX IF NOT EXISTS idx_user_follows_following_id ON user_follows(following_id);
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up any existing data
	tdb := &TestDB{DB: db}
	if err := tdb.CleanupData(); err != nil {
		t.Fatalf("Failed to clean up test data: %v", err)
	}

	return tdb
}

// Cleanup cleans up the test database
func (tdb *TestDB) Cleanup() error {
	// Get the underlying SQL database
	sqlDB, err := tdb.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Close the connection
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

// CleanupData removes all data from the test tables
func (tdb *TestDB) CleanupData() error {
	// Delete all records from tables in reverse order of dependencies
	err := tdb.DB.Exec("DELETE FROM likes").Error
	if err != nil {
		return err
	}

	err = tdb.DB.Exec("DELETE FROM comments").Error
	if err != nil {
		return err
	}

	err = tdb.DB.Exec("DELETE FROM posts").Error
	if err != nil {
		return err
	}

	err = tdb.DB.Exec("DELETE FROM user_follows").Error
	if err != nil {
		return err
	}

	err = tdb.DB.Exec("DELETE FROM users").Error
	if err != nil {
		return err
	}

	return nil
}
