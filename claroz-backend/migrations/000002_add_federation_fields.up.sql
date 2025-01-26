-- Add federation fields to users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS did TEXT,
ADD COLUMN IF NOT EXISTS handle TEXT,
ADD COLUMN IF NOT EXISTS federation_type TEXT DEFAULT 'local',
ADD COLUMN IF NOT EXISTS last_federation_sync TIMESTAMP WITH TIME ZONE;

-- Create unique index for DID
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_did ON users(did) WHERE did IS NOT NULL;
-- Create unique index for handle
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_handle ON users(handle) WHERE handle IS NOT NULL;
