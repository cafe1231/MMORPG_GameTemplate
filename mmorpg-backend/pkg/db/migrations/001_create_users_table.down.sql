-- Rollback: create_users_table
-- Created: 2025-01-29

BEGIN;

-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_account_status;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop table
DROP TABLE IF EXISTS users;

COMMIT;