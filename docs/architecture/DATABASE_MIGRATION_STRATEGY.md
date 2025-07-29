# Database Migration Strategy

## Executive Summary

This document outlines a comprehensive database migration strategy for the MMORPG Game Template, spanning all development phases from Phase 1 (Authentication) through Phase 4 (Production). The strategy ensures backward compatibility, zero-downtime deployments, and graceful schema evolution with robust rollback procedures.

## Migration Tool Selection

We will use **golang-migrate** as our primary migration tool for the following reasons:
- Native Go integration
- Support for up/down migrations
- Transaction support
- Version tracking
- Multiple database support
- CLI and programmatic interfaces

## Migration Principles

### 1. Backward Compatibility
- All migrations must be backward compatible for at least one version
- New columns should have defaults or be nullable initially
- Column/table renames use a two-phase approach
- Breaking changes require feature flags

### 2. Zero-Downtime Deployments
- Use online schema changes when possible
- Avoid exclusive locks on large tables
- Split large migrations into smaller chunks
- Use background jobs for data migrations

### 3. Schema Evolution
- Follow semantic versioning for schema changes
- Document all migrations thoroughly
- Test migrations on production-like data volumes
- Monitor migration performance

### 4. Rollback Procedures
- Every migration must have a corresponding rollback
- Test rollbacks in staging environment
- Document data loss implications
- Maintain rollback windows (24-48 hours)

## Phase-Specific Migration Plans

### Phase 1.5: Character System

#### New Tables

```sql
-- 004_create_characters_table.sql
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) UNIQUE NOT NULL,
    class VARCHAR(50) NOT NULL,
    race VARCHAR(50) NOT NULL,
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,
    health INTEGER DEFAULT 100,
    max_health INTEGER DEFAULT 100,
    mana INTEGER DEFAULT 100,
    max_mana INTEGER DEFAULT 100,
    stamina INTEGER DEFAULT 100,
    max_stamina INTEGER DEFAULT 100,
    position_x FLOAT DEFAULT 0,
    position_y FLOAT DEFAULT 0,
    position_z FLOAT DEFAULT 0,
    rotation_yaw FLOAT DEFAULT 0,
    zone_id INTEGER DEFAULT 1,
    is_online BOOLEAN DEFAULT FALSE,
    last_played TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    play_time INTERVAL DEFAULT '0 seconds',
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_level CHECK (level >= 1 AND level <= 100),
    CONSTRAINT check_health CHECK (health >= 0 AND health <= max_health),
    CONSTRAINT check_mana CHECK (mana >= 0 AND mana <= max_mana),
    CONSTRAINT check_stamina CHECK (stamina >= 0 AND stamina <= max_stamina)
);

-- Indexes
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_characters_name_lower ON characters(LOWER(name));
CREATE INDEX idx_characters_is_online ON characters(is_online) WHERE is_deleted = FALSE;
CREATE INDEX idx_characters_zone_id ON characters(zone_id) WHERE is_online = TRUE;
CREATE INDEX idx_characters_last_played ON characters(last_played DESC);

-- 005_create_character_stats_table.sql
CREATE TABLE character_stats (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    strength INTEGER DEFAULT 10,
    dexterity INTEGER DEFAULT 10,
    intelligence INTEGER DEFAULT 10,
    wisdom INTEGER DEFAULT 10,
    constitution INTEGER DEFAULT 10,
    charisma INTEGER DEFAULT 10,
    armor_class INTEGER DEFAULT 10,
    attack_power INTEGER DEFAULT 10,
    spell_power INTEGER DEFAULT 10,
    critical_chance FLOAT DEFAULT 5.0,
    critical_damage FLOAT DEFAULT 150.0,
    dodge_chance FLOAT DEFAULT 5.0,
    block_chance FLOAT DEFAULT 5.0,
    movement_speed FLOAT DEFAULT 100.0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 006_create_character_inventory_table.sql
CREATE TABLE character_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot_type VARCHAR(50) NOT NULL, -- 'bag', 'equipment', 'bank'
    slot_number INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    quantity INTEGER DEFAULT 1,
    durability INTEGER,
    enchantments JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(character_id, slot_type, slot_number),
    CONSTRAINT check_quantity CHECK (quantity > 0),
    CONSTRAINT check_slot_number CHECK (slot_number >= 0)
);

-- Indexes
CREATE INDEX idx_inventory_character_id ON character_inventory(character_id);
CREATE INDEX idx_inventory_item_id ON character_inventory(item_id);
```

#### Data Migration
```sql
-- Update users table character_count after characters are created
UPDATE users u
SET character_count = (
    SELECT COUNT(*) 
    FROM characters c 
    WHERE c.user_id = u.id AND c.is_deleted = FALSE
);
```

### Phase 2A: Real-time Networking Foundation

#### New Tables

```sql
-- 007_create_player_presence_table.sql
CREATE TABLE player_presence (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    server_id VARCHAR(100) NOT NULL,
    instance_id VARCHAR(100) NOT NULL,
    connection_id VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET NOT NULL,
    connected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    latency_ms INTEGER,
    status VARCHAR(50) DEFAULT 'online', -- online, away, busy, offline
    CONSTRAINT check_latency CHECK (latency_ms >= 0)
);

-- Indexes
CREATE INDEX idx_presence_server_id ON player_presence(server_id);
CREATE INDEX idx_presence_instance_id ON player_presence(instance_id);
CREATE INDEX idx_presence_last_heartbeat ON player_presence(last_heartbeat);

-- 008_create_world_state_table.sql
CREATE TABLE world_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id INTEGER NOT NULL,
    instance_id VARCHAR(100) NOT NULL,
    state_type VARCHAR(50) NOT NULL, -- 'weather', 'event', 'spawn'
    state_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_world_state_zone_instance ON world_state(zone_id, instance_id);
CREATE INDEX idx_world_state_type ON world_state(state_type);
CREATE INDEX idx_world_state_expires ON world_state(expires_at) WHERE expires_at IS NOT NULL;

-- Partitioning for high-volume position updates
-- 009_create_character_positions_table.sql
CREATE TABLE character_positions (
    character_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    position_x FLOAT NOT NULL,
    position_y FLOAT NOT NULL,
    position_z FLOAT NOT NULL,
    rotation_yaw FLOAT NOT NULL,
    velocity_x FLOAT DEFAULT 0,
    velocity_y FLOAT DEFAULT 0,
    velocity_z FLOAT DEFAULT 0,
    zone_id INTEGER NOT NULL,
    instance_id VARCHAR(100) NOT NULL,
    PRIMARY KEY (character_id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create partitions for the next 30 days
CREATE TABLE character_positions_y2025m01 PARTITION OF character_positions
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Auto-partition function (to be run monthly)
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_date := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    partition_name := 'character_positions_y' || TO_CHAR(partition_date, 'YYYY') || 'm' || TO_CHAR(partition_date, 'MM');
    start_date := partition_date;
    end_date := partition_date + INTERVAL '1 month';
    
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF character_positions FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;
```

### Phase 2B: Combat & Interaction Systems

#### New Tables

```sql
-- 010_create_combat_log_table.sql
CREATE TABLE combat_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source_id UUID, -- character_id or npc_id
    target_id UUID, -- character_id or npc_id
    action_type VARCHAR(50) NOT NULL, -- 'damage', 'heal', 'buff', 'debuff'
    action_subtype VARCHAR(100), -- spell name, ability name
    amount INTEGER,
    is_critical BOOLEAN DEFAULT FALSE,
    is_blocked BOOLEAN DEFAULT FALSE,
    is_dodged BOOLEAN DEFAULT FALSE,
    zone_id INTEGER NOT NULL,
    instance_id VARCHAR(100) NOT NULL
) PARTITION BY RANGE (timestamp);

-- Create initial partition
CREATE TABLE combat_log_y2025m01 PARTITION OF combat_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Indexes on partition parent
CREATE INDEX idx_combat_log_source ON combat_log(source_id, timestamp DESC);
CREATE INDEX idx_combat_log_target ON combat_log(target_id, timestamp DESC);

-- 011_create_character_abilities_table.sql
CREATE TABLE character_abilities (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    ability_id INTEGER NOT NULL,
    level INTEGER DEFAULT 1,
    experience INTEGER DEFAULT 0,
    last_used TIMESTAMP WITH TIME ZONE,
    cooldown_until TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    keybind VARCHAR(50),
    PRIMARY KEY (character_id, ability_id)
);

-- 012_create_buffs_debuffs_table.sql
CREATE TABLE character_effects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    effect_id INTEGER NOT NULL,
    source_id UUID, -- who applied the effect
    stacks INTEGER DEFAULT 1,
    duration_seconds INTEGER,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_debuff BOOLEAN DEFAULT FALSE,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_effects_character_expires ON character_effects(character_id, expires_at);
CREATE INDEX idx_effects_expires ON character_effects(expires_at);
```

### Phase 3: Core Gameplay Systems

#### New Tables

```sql
-- 013_create_guilds_table.sql
CREATE TABLE guilds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    tag VARCHAR(10) UNIQUE NOT NULL,
    description TEXT,
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,
    member_count INTEGER DEFAULT 0,
    max_members INTEGER DEFAULT 50,
    bank_gold BIGINT DEFAULT 0,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_member_count CHECK (member_count >= 0 AND member_count <= max_members)
);

-- 014_create_guild_members_table.sql
CREATE TABLE guild_members (
    guild_id UUID NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    rank VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    contribution_points BIGINT DEFAULT 0,
    note TEXT,
    PRIMARY KEY (guild_id, character_id)
);

-- 015_create_quests_progress_table.sql
CREATE TABLE character_quests (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    quest_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, completed, failed, abandoned
    current_stage INTEGER DEFAULT 0,
    objectives_data JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (character_id, quest_id)
);

-- 016_create_achievements_table.sql
CREATE TABLE character_achievements (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    achievement_id INTEGER NOT NULL,
    progress INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (character_id, achievement_id)
);

-- 017_create_friends_table.sql
CREATE TABLE character_relationships (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    related_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL, -- 'friend', 'blocked', 'ignored'
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (character_id, related_character_id),
    CONSTRAINT check_not_self CHECK (character_id != related_character_id)
);

-- 018_create_mail_table.sql
CREATE TABLE character_mail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    sender_name VARCHAR(50) NOT NULL,
    subject VARCHAR(200) NOT NULL,
    body TEXT,
    attachments JSONB DEFAULT '[]',
    gold_attached BIGINT DEFAULT 0,
    is_read BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_gold CHECK (gold_attached >= 0)
);

-- Indexes
CREATE INDEX idx_mail_recipient ON character_mail(recipient_id, is_deleted, sent_at DESC);
CREATE INDEX idx_mail_expires ON character_mail(expires_at) WHERE expires_at IS NOT NULL;
```

### Phase 4: Production Optimizations

#### Performance Tables

```sql
-- 019_create_analytics_events_table.sql
CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    user_id UUID,
    character_id UUID,
    session_id UUID,
    event_data JSONB NOT NULL,
    ip_address INET,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- 020_create_audit_log_table.sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID,
    character_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT
) PARTITION BY RANGE (timestamp);

-- 021_create_performance_metrics_table.sql
CREATE TABLE performance_metrics (
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    server_id VARCHAR(100) NOT NULL,
    metric_type VARCHAR(100) NOT NULL,
    metric_value NUMERIC,
    tags JSONB DEFAULT '{}',
    PRIMARY KEY (timestamp, server_id, metric_type)
) PARTITION BY RANGE (timestamp);
```

## Migration Best Practices

### Version Control Strategy

1. **File Naming Convention**
   ```
   XXX_description.up.sql    -- Forward migration
   XXX_description.down.sql  -- Rollback migration
   ```
   Where XXX is a sequential number (001, 002, etc.)

2. **Git Workflow**
   - All migrations must be reviewed via PR
   - Include migration dry-run output in PR
   - Tag releases with schema version

3. **Migration Metadata**
   ```sql
   -- Create schema_migrations table (managed by golang-migrate)
   CREATE TABLE schema_migrations (
       version BIGINT NOT NULL PRIMARY KEY,
       dirty BOOLEAN NOT NULL DEFAULT FALSE
   );
   ```

### Testing Procedures

1. **Unit Tests**
   ```go
   func TestMigration_XXX(t *testing.T) {
       // Test forward migration
       // Test rollback
       // Test idempotency
   }
   ```

2. **Integration Tests**
   - Run against empty database
   - Run against populated database
   - Test with production-like data volumes

3. **Performance Tests**
   ```sql
   -- Before migration
   EXPLAIN ANALYZE <query>;
   
   -- After migration
   EXPLAIN ANALYZE <query>;
   ```

4. **Rollback Tests**
   - Apply migration
   - Insert test data
   - Rollback migration
   - Verify data integrity

### Deployment Process

1. **Pre-deployment Checklist**
   - [ ] Backup production database
   - [ ] Test migrations in staging
   - [ ] Review query performance impact
   - [ ] Prepare rollback plan
   - [ ] Schedule maintenance window (if needed)

2. **Deployment Steps**
   ```bash
   # 1. Create backup
   pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql
   
   # 2. Run migrations in transaction
   migrate -path ./migrations -database $DATABASE_URL up
   
   # 3. Verify migration
   migrate -path ./migrations -database $DATABASE_URL version
   
   # 4. Run validation queries
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f validate_migration.sql
   ```

3. **Post-deployment Monitoring**
   - Check application logs
   - Monitor database performance
   - Verify data integrity
   - Watch for deadlocks/timeouts

### Monitoring and Validation

1. **Health Checks**
   ```sql
   -- Check migration status
   SELECT version, dirty FROM schema_migrations;
   
   -- Check table sizes
   SELECT 
       schemaname,
       tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
   FROM pg_tables
   WHERE schemaname = 'public'
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
   ```

2. **Performance Monitoring**
   ```sql
   -- Slow query log
   SELECT 
       query,
       calls,
       total_time,
       mean_time,
       max_time
   FROM pg_stat_statements
   ORDER BY mean_time DESC
   LIMIT 20;
   ```

3. **Data Validation**
   ```sql
   -- Check for orphaned records
   SELECT c.* FROM characters c
   LEFT JOIN users u ON c.user_id = u.id
   WHERE u.id IS NULL;
   
   -- Check constraint violations
   SELECT conname, conrelid::regclass
   FROM pg_constraint
   WHERE NOT convalidated;
   ```

## Scaling Strategies

### Partitioning

1. **Time-based Partitioning**
   - Combat logs (monthly)
   - Analytics events (monthly)
   - Character positions (daily for active periods)
   - Audit logs (monthly)

2. **Partition Maintenance**
   ```sql
   -- Automated partition creation
   CREATE OR REPLACE FUNCTION create_partitions()
   RETURNS void AS $$
   BEGIN
       -- Create next month's partitions
       PERFORM create_monthly_partition();
       
       -- Drop old partitions (keep 6 months)
       PERFORM drop_old_partitions('combat_log', INTERVAL '6 months');
       PERFORM drop_old_partitions('analytics_events', INTERVAL '12 months');
   END;
   $$ LANGUAGE plpgsql;
   
   -- Schedule via pg_cron
   SELECT cron.schedule('create-partitions', '0 0 1 * *', 'SELECT create_partitions()');
   ```

### Archival Procedures

1. **Cold Storage Strategy**
   ```sql
   -- Archive old character data
   CREATE TABLE characters_archive (LIKE characters INCLUDING ALL);
   
   -- Move deleted characters older than 90 days
   INSERT INTO characters_archive
   SELECT * FROM characters
   WHERE is_deleted = TRUE 
   AND deleted_at < NOW() - INTERVAL '90 days';
   
   DELETE FROM characters
   WHERE is_deleted = TRUE 
   AND deleted_at < NOW() - INTERVAL '90 days';
   ```

2. **Data Retention Policies**
   - Combat logs: 6 months
   - Analytics: 12 months
   - Audit logs: 7 years
   - Deleted characters: 90 days active, then archive

### Performance Optimization

1. **Index Strategy**
   ```sql
   -- Covering indexes for common queries
   CREATE INDEX idx_characters_user_lookup 
   ON characters(user_id) 
   INCLUDE (name, level, class, is_online)
   WHERE is_deleted = FALSE;
   
   -- Partial indexes for active data
   CREATE INDEX idx_mail_unread 
   ON character_mail(recipient_id, sent_at DESC)
   WHERE is_read = FALSE AND is_deleted = FALSE;
   ```

2. **Materialized Views**
   ```sql
   -- Guild rankings
   CREATE MATERIALIZED VIEW guild_rankings AS
   SELECT 
       g.id,
       g.name,
       g.level,
       g.member_count,
       RANK() OVER (ORDER BY g.level DESC, g.experience DESC) as rank
   FROM guilds g
   WHERE g.member_count > 0;
   
   -- Refresh strategy
   CREATE OR REPLACE FUNCTION refresh_rankings()
   RETURNS void AS $$
   BEGIN
       REFRESH MATERIALIZED VIEW CONCURRENTLY guild_rankings;
   END;
   $$ LANGUAGE plpgsql;
   ```

3. **Connection Pooling**
   - Use PgBouncer for connection pooling
   - Configure pool sizes per service
   - Monitor connection usage

## Rollback Procedures

### Standard Rollback Process

1. **Immediate Rollback (< 1 hour)**
   ```bash
   # Stop application servers
   kubectl scale deployment auth-service --replicas=0
   
   # Rollback migration
   migrate -path ./migrations -database $DATABASE_URL down 1
   
   # Deploy previous version
   kubectl set image deployment/auth-service auth-service=auth-service:v1.2.3
   
   # Scale back up
   kubectl scale deployment auth-service --replicas=3
   ```

2. **Delayed Rollback (> 1 hour)**
   - Assess data changes
   - Create data migration script
   - Test in staging
   - Schedule maintenance window

### Rollback Decision Matrix

| Scenario | Impact | Rollback Strategy |
|----------|--------|------------------|
| Schema change only | Low | Direct rollback |
| Data migration | Medium | Reverse migration |
| Data loss possible | High | Restore from backup |
| Cross-service change | Critical | Coordinated rollback |

## Migration Tooling

### Setup golang-migrate

1. **Installation**
   ```bash
   # Install CLI
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Add to project
   go get -u github.com/golang-migrate/migrate/v4
   go get -u github.com/golang-migrate/migrate/v4/database/postgres
   go get -u github.com/golang-migrate/migrate/v4/source/file
   ```

2. **Integration Code**
   ```go
   package db
   
   import (
       "database/sql"
       "embed"
       
       "github.com/golang-migrate/migrate/v4"
       "github.com/golang-migrate/migrate/v4/database/postgres"
       "github.com/golang-migrate/migrate/v4/source/iofs"
   )
   
   //go:embed migrations/*.sql
   var migrationsFS embed.FS
   
   func RunMigrations(db *sql.DB) error {
       driver, err := postgres.WithInstance(db, &postgres.Config{})
       if err != nil {
           return err
       }
       
       source, err := iofs.New(migrationsFS, "migrations")
       if err != nil {
           return err
       }
       
       m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
       if err != nil {
           return err
       }
       
       return m.Up()
   }
   ```

### Helper Scripts

1. **Migration Generator**
   ```bash
   #!/bin/bash
   # scripts/create-migration.sh
   
   if [ -z "$1" ]; then
       echo "Usage: ./create-migration.sh <migration_name>"
       exit 1
   fi
   
   TIMESTAMP=$(date +%Y%m%d%H%M%S)
   MIGRATION_NAME=$1
   
   # Create up migration
   cat > "migrations/${TIMESTAMP}_${MIGRATION_NAME}.up.sql" << EOF
   -- Migration: ${MIGRATION_NAME}
   -- Created: $(date)
   
   BEGIN;
   
   -- Your migration SQL here
   
   COMMIT;
   EOF
   
   # Create down migration
   cat > "migrations/${TIMESTAMP}_${MIGRATION_NAME}.down.sql" << EOF
   -- Rollback: ${MIGRATION_NAME}
   -- Created: $(date)
   
   BEGIN;
   
   -- Your rollback SQL here
   
   COMMIT;
   EOF
   
   echo "Created migration: ${TIMESTAMP}_${MIGRATION_NAME}"
   ```

2. **Migration Validator**
   ```go
   package migrations
   
   import (
       "testing"
       "database/sql"
   )
   
   func TestMigrations(t *testing.T) {
       // Test each migration can be applied and rolled back
       db := setupTestDB(t)
       defer db.Close()
       
       // Apply all migrations
       if err := RunMigrations(db); err != nil {
           t.Fatalf("Failed to run migrations: %v", err)
       }
       
       // Verify schema
       tables := []string{
           "users", "sessions", "characters", "character_stats",
           "character_inventory", "player_presence", "world_state",
       }
       
       for _, table := range tables {
           var exists bool
           err := db.QueryRow(`
               SELECT EXISTS (
                   SELECT FROM information_schema.tables 
                   WHERE table_schema = 'public' 
                   AND table_name = $1
               )`, table).Scan(&exists)
           
           if err != nil || !exists {
               t.Errorf("Table %s does not exist", table)
           }
       }
   }
   ```

## Conclusion

This migration strategy provides a robust framework for evolving the MMORPG database schema through all development phases. Key benefits include:

1. **Predictability**: Clear migration paths for each phase
2. **Safety**: Comprehensive rollback procedures
3. **Performance**: Built-in optimization strategies
4. **Maintainability**: Consistent tooling and practices

Regular reviews and updates of this strategy ensure it remains aligned with project needs and industry best practices.