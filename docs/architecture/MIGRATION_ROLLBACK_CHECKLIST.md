# Migration Rollback Safety Checklist

## Pre-Rollback Assessment

### 1. Identify the Issue
- [ ] Document the specific error or issue requiring rollback
- [ ] Identify which migration version caused the problem
- [ ] Determine the impact on users and services
- [ ] Check if the issue can be fixed without rollback

### 2. Assess Data Changes
- [ ] Review what data was modified by the migration
- [ ] Check for any data that would be lost in rollback
- [ ] Identify dependent data in other tables
- [ ] Determine if data needs to be exported first

### 3. Service Dependencies
- [ ] List all services using the affected schema
- [ ] Check for cached data that needs invalidation
- [ ] Identify API endpoints that will be affected
- [ ] Review any scheduled jobs or background tasks

## Rollback Execution

### 4. Preparation Steps
```bash
# Create a backup before rollback
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > pre_rollback_backup_$(date +%Y%m%d_%H%M%S).sql

# Export any critical data that might be lost
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\copy (SELECT * FROM affected_table) TO 'affected_data.csv' CSV HEADER"

# Check current migration version
migrate -path ./migrations -database $DATABASE_URL version
```

### 5. Service Coordination
- [ ] Notify team members of impending rollback
- [ ] Put services in maintenance mode if needed
- [ ] Stop background workers and scheduled jobs
- [ ] Scale down services to prevent new connections

```bash
# Example: Scale down services
kubectl scale deployment auth-service --replicas=0
kubectl scale deployment gateway-service --replicas=0
```

### 6. Execute Rollback
```bash
# Dry run (if your migration tool supports it)
migrate -path ./migrations -database $DATABASE_URL down 1 -dry-run

# Actual rollback
migrate -path ./migrations -database $DATABASE_URL down 1

# Verify rollback
migrate -path ./migrations -database $DATABASE_URL version
```

### 7. Data Recovery (if needed)
```sql
-- Example: Restore specific data
\copy temp_table FROM 'affected_data.csv' CSV HEADER
INSERT INTO original_table SELECT * FROM temp_table ON CONFLICT DO NOTHING;
```

## Post-Rollback Verification

### 8. Database Verification
```sql
-- Run validation script
\i scripts/db/validate-migrations.sql

-- Check for orphaned data
SELECT COUNT(*) FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE '%_backup%';

-- Verify constraints
SELECT conname, conrelid::regclass
FROM pg_constraint
WHERE NOT convalidated;
```

### 9. Application Testing
- [ ] Deploy previous version of application code
- [ ] Run health checks on all services
- [ ] Execute smoke tests for critical paths
- [ ] Monitor error logs for issues

### 10. Performance Check
```sql
-- Check for slow queries
SELECT query, calls, mean_time
FROM pg_stat_statements
WHERE mean_time > 1000  -- queries over 1 second
ORDER BY mean_time DESC
LIMIT 10;

-- Verify indexes are being used
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan;
```

## Recovery Actions

### 11. If Rollback Fails
1. **Check for Dirty State**
   ```sql
   SELECT version, dirty FROM schema_migrations;
   ```

2. **Manual Recovery Options**
   - Force clean state: `migrate force <version>`
   - Restore from backup
   - Apply manual SQL fixes

3. **Emergency Contacts**
   - Database Administrator: ___________
   - Platform Team Lead: ___________
   - On-call Engineer: ___________

### 12. Post-Incident Actions
- [ ] Document what went wrong
- [ ] Update migration tests to catch similar issues
- [ ] Review and improve rollback procedures
- [ ] Schedule post-mortem meeting
- [ ] Update monitoring to detect similar issues

## Common Rollback Scenarios

### Scenario 1: Column Rename
**Issue**: Column renamed, old code expects old name
**Safe Rollback**: Yes, if no data written to new column
**Action**: 
1. Rollback migration
2. Use dual-write pattern in next attempt

### Scenario 2: NOT NULL Constraint Added
**Issue**: Existing code tries to insert NULL
**Safe Rollback**: Yes
**Action**:
1. Rollback migration
2. Add default value first
3. Backfill data
4. Then add constraint

### Scenario 3: Table Dropped
**Issue**: Old code still references table
**Safe Rollback**: Only if table was renamed (not dropped)
**Action**:
1. Restore from backup if truly dropped
2. Otherwise rollback rename

### Scenario 4: Data Type Change
**Issue**: Data corruption or precision loss
**Safe Rollback**: Depends on conversion
**Action**:
1. Check if original data is preserved
2. May need to restore from backup
3. Consider using a new column instead

## Rollback Prevention Best Practices

1. **Always Test Rollbacks**
   - Include rollback in CI/CD pipeline
   - Test with production-like data volumes
   - Verify no data loss occurs

2. **Use Feature Flags**
   - Deploy code that works with both schemas
   - Gradually enable new schema usage
   - Easy rollback without database changes

3. **Staged Rollouts**
   - Deploy to staging first
   - Monitor for 24-48 hours
   - Use canary deployments

4. **Comprehensive Testing**
   - Unit tests for migrations
   - Integration tests with real database
   - Load tests for performance impact

5. **Migration Reviews**
   - Require peer review for all migrations
   - DBA review for complex changes
   - Include rollback plan in PR description