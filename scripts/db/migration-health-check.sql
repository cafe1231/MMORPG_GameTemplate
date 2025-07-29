-- Migration Health Check Dashboard
-- Run this periodically to monitor database health

-- Summary Dashboard
SELECT '=== DATABASE MIGRATION HEALTH CHECK ===' as section;
SELECT 'Timestamp: ' || NOW()::text as info;
SELECT 'Database: ' || current_database() as info;
SELECT '';

-- 1. Migration Status
SELECT '--- MIGRATION STATUS ---' as section;
SELECT 
    version as current_version,
    CASE WHEN dirty THEN 'DIRTY STATE - MANUAL FIX REQUIRED!' ELSE 'Clean' END as state,
    (SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public') as table_count,
    (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public') as index_count
FROM schema_migrations;

-- 2. Table Health
SELECT '';
SELECT '--- TABLE HEALTH ---' as section;
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as indexes_size,
    n_live_tup as live_rows,
    n_dead_tup as dead_rows,
    CASE 
        WHEN n_live_tup > 0 THEN ROUND(100.0 * n_dead_tup / n_live_tup, 2)
        ELSE 0
    END as dead_row_percent,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||relname) DESC;

-- 3. Index Usage
SELECT '';
SELECT '--- INDEX USAGE ---' as section;
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(schemaname||'.'||indexname)) as index_size,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 100 THEN 'RARELY USED'
        ELSE 'ACTIVE'
    END as usage_status
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan;

-- 4. Slow Queries (requires pg_stat_statements)
SELECT '';
SELECT '--- SLOW QUERIES (if pg_stat_statements enabled) ---' as section;
SELECT 
    CASE WHEN EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements'
    ) THEN 'pg_stat_statements is enabled'
    ELSE 'pg_stat_statements is NOT enabled - consider enabling for query monitoring'
    END as status;

-- 5. Connection Stats
SELECT '';
SELECT '--- CONNECTION STATS ---' as section;
SELECT 
    datname,
    numbackends as active_connections,
    xact_commit as committed_transactions,
    xact_rollback as rolled_back_transactions,
    CASE 
        WHEN xact_commit + xact_rollback > 0 
        THEN ROUND(100.0 * xact_rollback / (xact_commit + xact_rollback), 2)
        ELSE 0
    END as rollback_ratio,
    tup_returned as rows_returned,
    tup_fetched as rows_fetched,
    tup_inserted as rows_inserted,
    tup_updated as rows_updated,
    tup_deleted as rows_deleted
FROM pg_stat_database
WHERE datname = current_database();

-- 6. Lock Analysis
SELECT '';
SELECT '--- CURRENT LOCKS ---' as section;
SELECT 
    locktype,
    relation::regclass as table_name,
    mode,
    granted,
    pid,
    query_start,
    NOW() - query_start as duration,
    state,
    LEFT(query, 100) as query_snippet
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE relation IS NOT NULL
AND a.datname = current_database()
ORDER BY query_start;

-- 7. Foreign Key Constraints
SELECT '';
SELECT '--- FOREIGN KEY CONSTRAINTS ---' as section;
SELECT
    conname as constraint_name,
    conrelid::regclass as table_name,
    a.attname as column_name,
    confrelid::regclass as references_table,
    af.attname as references_column
FROM pg_constraint c
JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
JOIN pg_attribute af ON af.attrelid = c.confrelid AND af.attnum = ANY(c.confkey)
WHERE c.contype = 'f'
AND c.connamespace = 'public'::regnamespace
ORDER BY conrelid::regclass::text, conname;

-- 8. Missing Indexes Suggestions
SELECT '';
SELECT '--- MISSING INDEX SUGGESTIONS ---' as section;
WITH table_scans as (
    SELECT 
        schemaname,
        tablename,
        n_live_tup,
        seq_scan,
        idx_scan,
        CASE WHEN seq_scan + idx_scan > 0 
             THEN ROUND(100.0 * seq_scan / (seq_scan + idx_scan), 2)
             ELSE 0 
        END as seq_scan_pct
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
    AND n_live_tup > 1000  -- Only consider tables with >1000 rows
)
SELECT 
    schemaname || '.' || tablename as table_name,
    n_live_tup as row_count,
    seq_scan as sequential_scans,
    idx_scan as index_scans,
    seq_scan_pct || '%' as seq_scan_percentage,
    CASE 
        WHEN seq_scan_pct > 90 THEN 'HIGH - Consider adding indexes'
        WHEN seq_scan_pct > 50 THEN 'MEDIUM - Monitor query patterns'
        ELSE 'LOW - Indexes are being used effectively'
    END as recommendation
FROM table_scans
WHERE seq_scan > 100  -- Ignore tables with very few scans
ORDER BY seq_scan_pct DESC;

-- 9. Data Integrity Checks
SELECT '';
SELECT '--- DATA INTEGRITY ---' as section;

-- Check for users without any sessions (might be normal)
SELECT 
    'Users without sessions' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'OK'
        ELSE 'INFO - ' || COUNT(*) || ' users have never logged in'
    END as status
FROM users u
LEFT JOIN sessions s ON u.id = s.user_id
WHERE s.id IS NULL;

-- Check for expired sessions still in database
SELECT 
    'Expired sessions' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'OK'
        ELSE 'WARNING - Run cleanup_expired_sessions() to remove'
    END as status
FROM sessions
WHERE expires_at < NOW();

-- 10. Performance Recommendations
SELECT '';
SELECT '--- RECOMMENDATIONS ---' as section;
SELECT recommendation FROM (
    -- Vacuum recommendations
    SELECT 1 as ord, 
        'Run VACUUM on ' || tablename || ' (dead rows: ' || n_dead_tup || ')' as recommendation
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
    AND n_dead_tup > 1000
    AND (last_vacuum IS NULL OR last_vacuum < NOW() - INTERVAL '7 days')
    AND (last_autovacuum IS NULL OR last_autovacuum < NOW() - INTERVAL '1 day')
    
    UNION ALL
    
    -- Analyze recommendations
    SELECT 2 as ord,
        'Run ANALYZE on ' || tablename || ' (last analyzed: ' || 
        COALESCE(last_analyze::text, 'never') || ')' as recommendation
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
    AND (last_analyze IS NULL OR last_analyze < NOW() - INTERVAL '7 days')
    AND (last_autoanalyze IS NULL OR last_autoanalyze < NOW() - INTERVAL '1 day')
    
    UNION ALL
    
    -- Unused index recommendations
    SELECT 3 as ord,
        'Consider dropping unused index: ' || indexname || ' on ' || tablename as recommendation
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public'
    AND idx_scan = 0
    AND indexname NOT LIKE '%_pkey'  -- Don't suggest dropping primary keys
    
) recommendations
ORDER BY ord;