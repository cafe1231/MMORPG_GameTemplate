-- Rollback: create_sessions_table
-- Created: 2025-01-29

BEGIN;

-- Drop functions first
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop trigger
DROP TRIGGER IF EXISTS update_sessions_last_active ON sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_last_active_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_device_id;

-- Drop table
DROP TABLE IF EXISTS sessions;

COMMIT;