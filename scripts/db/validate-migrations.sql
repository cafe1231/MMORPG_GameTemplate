-- Migration Validation Script
-- Run this after migrations to ensure database integrity

-- Check migration status
SELECT 
    'Current Migration Version' as check_name,
    version::text as result,
    CASE WHEN dirty THEN 'DIRTY - Manual intervention required!' ELSE 'Clean' END as status
FROM schema_migrations;

-- Check table existence
WITH expected_tables AS (
    SELECT unnest(ARRAY[
        'users',
        'sessions'
    ]) as table_name
),
actual_tables AS (
    SELECT tablename as table_name
    FROM pg_tables
    WHERE schemaname = 'public'
)
SELECT 
    'Missing Tables' as check_name,
    string_agg(e.table_name, ', ') as result,
    CASE 
        WHEN COUNT(e.table_name) = 0 THEN 'OK - All tables exist'
        ELSE 'ERROR - Missing tables!'
    END as status
FROM expected_tables e
LEFT JOIN actual_tables a ON e.table_name = a.table_name
WHERE a.table_name IS NULL;

-- Check indexes
WITH expected_indexes AS (
    SELECT unnest(ARRAY[
        'idx_users_email',
        'idx_users_username',
        'idx_users_account_status',
        'idx_users_created_at',
        'idx_sessions_user_id',
        'idx_sessions_token_hash',
        'idx_sessions_expires_at',
        'idx_sessions_device_id'
    ]) as index_name
),
actual_indexes AS (
    SELECT indexname as index_name
    FROM pg_indexes
    WHERE schemaname = 'public'
)
SELECT 
    'Missing Indexes' as check_name,
    COALESCE(string_agg(e.index_name, ', '), 'None') as result,
    CASE 
        WHEN COUNT(e.index_name) = 0 THEN 'OK - All indexes exist'
        ELSE 'WARNING - Missing indexes (may impact performance)'
    END as status
FROM expected_indexes e
LEFT JOIN actual_indexes a ON e.index_name = a.index_name
WHERE a.index_name IS NULL;

-- Check triggers
WITH expected_triggers AS (
    SELECT 
        'users' as table_name,
        'update_users_updated_at' as trigger_name
    UNION ALL
    SELECT 
        'sessions' as table_name,
        'update_sessions_last_active' as trigger_name
),
actual_triggers AS (
    SELECT 
        event_object_table as table_name,
        trigger_name
    FROM information_schema.triggers
    WHERE trigger_schema = 'public'
)
SELECT 
    'Missing Triggers' as check_name,
    COALESCE(string_agg(e.table_name || '.' || e.trigger_name, ', '), 'None') as result,
    CASE 
        WHEN COUNT(*) = 0 THEN 'OK - All triggers exist'
        ELSE 'ERROR - Missing triggers!'
    END as status
FROM expected_triggers e
LEFT JOIN actual_triggers a 
    ON e.table_name = a.table_name 
    AND e.trigger_name = a.trigger_name
WHERE a.trigger_name IS NULL;

-- Check constraints
SELECT 
    'Constraint Violations' as check_name,
    COALESCE(string_agg(conname || ' on ' || conrelid::regclass::text, ', '), 'None') as result,
    CASE 
        WHEN COUNT(*) = 0 THEN 'OK - All constraints valid'
        ELSE 'ERROR - Invalid constraints!'
    END as status
FROM pg_constraint
WHERE NOT convalidated;

-- Check for orphaned records
SELECT 
    'Orphaned Sessions' as check_name,
    COUNT(*)::text as result,
    CASE 
        WHEN COUNT(*) = 0 THEN 'OK - No orphaned records'
        ELSE 'WARNING - Found orphaned sessions'
    END as status
FROM sessions s
LEFT JOIN users u ON s.user_id = u.id
WHERE u.id IS NULL;

-- Check table sizes
SELECT 
    'Large Tables' as check_name,
    string_agg(
        tablename || ': ' || pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)),
        ', '
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
    ) as result,
    'INFO' as status
FROM pg_tables
WHERE schemaname = 'public'
AND pg_total_relation_size(schemaname||'.'||tablename) > 1024 * 1024; -- Tables over 1MB

-- Database statistics
SELECT 
    'Database Size' as check_name,
    pg_size_pretty(pg_database_size(current_database())) as result,
    'INFO' as status;