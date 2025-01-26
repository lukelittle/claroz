-- Drop indexes
DROP INDEX IF EXISTS idx_likes_post_id;
DROP INDEX IF EXISTS idx_likes_user_id;
DROP INDEX IF EXISTS idx_user_follows_follower_id;
DROP INDEX IF EXISTS idx_user_follows_following_id;

-- Drop tables
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS user_follows;

-- Add back likes column to posts table
ALTER TABLE posts ADD COLUMN IF NOT EXISTS likes INTEGER DEFAULT 0;
