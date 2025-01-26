-- Drop federation-related indexes
DROP INDEX IF EXISTS idx_users_did;
DROP INDEX IF EXISTS idx_users_handle;

-- Remove federation columns from users table
ALTER TABLE users
DROP COLUMN IF EXISTS did,
DROP COLUMN IF EXISTS handle,
DROP COLUMN IF EXISTS federation_type,
DROP COLUMN IF EXISTS last_federation_sync;
