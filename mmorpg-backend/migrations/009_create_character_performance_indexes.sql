-- Additional performance indexes and optimizations for character system

-- Composite indexes for common queries
CREATE INDEX idx_characters_user_active ON characters(user_id, is_deleted) 
    WHERE is_deleted = FALSE;

CREATE INDEX idx_characters_user_slot_active ON characters(user_id, slot_number) 
    WHERE is_deleted = FALSE;

CREATE INDEX idx_characters_class_level ON characters(class_type, level) 
    WHERE is_deleted = FALSE;

-- Index for leaderboard queries
CREATE INDEX idx_characters_level_experience ON characters(level DESC, experience DESC) 
    WHERE is_deleted = FALSE;

-- Index for character restoration queries
CREATE INDEX idx_characters_deletion_scheduled ON characters(deletion_scheduled_at) 
    WHERE is_deleted = TRUE AND deletion_scheduled_at IS NOT NULL;

-- Partial index for online character tracking (will be used in Phase 2)
CREATE INDEX idx_character_position_recent_movement ON character_position(last_movement) 
    WHERE last_movement > NOW() - INTERVAL '5 minutes';

-- Create materialized view for character statistics (refresh periodically)
CREATE MATERIALIZED VIEW mv_character_statistics AS
SELECT 
    c.class_type,
    c.race,
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE c.level >= 10) as level_10_plus,
    COUNT(*) FILTER (WHERE c.level >= 50) as level_50_plus,
    COUNT(*) FILTER (WHERE c.level = 100) as max_level,
    AVG(c.level) as avg_level,
    AVG(EXTRACT(EPOCH FROM c.total_play_time)/3600)::DECIMAL(10,2) as avg_play_hours
FROM characters c
WHERE c.is_deleted = FALSE
GROUP BY c.class_type, c.race;

-- Create index on materialized view
CREATE INDEX idx_mv_character_statistics_class_race ON mv_character_statistics(class_type, race);

-- Create function to refresh character statistics
CREATE OR REPLACE FUNCTION refresh_character_statistics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_character_statistics;
END;
$$ LANGUAGE plpgsql;

-- Create table for tracking character name history (for moderation)
CREATE TABLE IF NOT EXISTS character_name_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    old_name VARCHAR(30) NOT NULL,
    new_name VARCHAR(30) NOT NULL,
    changed_by UUID REFERENCES users(id),
    change_reason VARCHAR(255),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_name_history_character ON character_name_history(character_id);
CREATE INDEX idx_character_name_history_old_name ON character_name_history(LOWER(old_name));
CREATE INDEX idx_character_name_history_changed_at ON character_name_history(changed_at);

-- Create function to log character name changes
CREATE OR REPLACE FUNCTION log_character_name_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.name != NEW.name THEN
        INSERT INTO character_name_history (
            character_id, old_name, new_name
        ) VALUES (
            NEW.id, OLD.name, NEW.name
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for name change logging
CREATE TRIGGER log_name_changes
AFTER UPDATE ON characters
FOR EACH ROW
WHEN (OLD.name IS DISTINCT FROM NEW.name)
EXECUTE FUNCTION log_character_name_change();

-- Create function for efficient character listing with pagination
CREATE OR REPLACE FUNCTION get_user_characters(
    p_user_id UUID,
    p_include_deleted BOOLEAN DEFAULT FALSE,
    p_limit INTEGER DEFAULT 10,
    p_offset INTEGER DEFAULT 0
)
RETURNS TABLE(
    character_id UUID,
    name VARCHAR(30),
    slot_number INTEGER,
    level INTEGER,
    experience BIGINT,
    class_type VARCHAR(50),
    race VARCHAR(50),
    gender VARCHAR(20),
    last_played_at TIMESTAMP WITH TIME ZONE,
    world_id VARCHAR(100),
    zone_id VARCHAR(100),
    is_deleted BOOLEAN,
    deletion_scheduled_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id,
        c.name,
        c.slot_number,
        c.level,
        c.experience,
        c.class_type,
        c.race,
        c.gender,
        c.last_played_at,
        cp.world_id,
        cp.zone_id,
        c.is_deleted,
        c.deletion_scheduled_at
    FROM characters c
    LEFT JOIN character_position cp ON cp.character_id = c.id
    WHERE c.user_id = p_user_id
        AND (p_include_deleted OR c.is_deleted = FALSE)
    ORDER BY c.slot_number
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

-- Create function to check character name availability
CREATE OR REPLACE FUNCTION is_character_name_available(
    p_name VARCHAR(30)
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN NOT EXISTS(
        SELECT 1 FROM characters 
        WHERE LOWER(name) = LOWER(p_name)
    );
END;
$$ LANGUAGE plpgsql;

-- Add comment documentation
COMMENT ON TABLE characters IS 'Main character table storing core character information';
COMMENT ON TABLE character_appearance IS 'Character visual customization data';
COMMENT ON TABLE character_stats IS 'Character attributes and combat statistics';
COMMENT ON TABLE character_position IS 'Character world position and movement data';
COMMENT ON TABLE character_name_history IS 'Audit log of character name changes';

COMMENT ON FUNCTION soft_delete_character IS 'Marks a character as deleted with 30-day recovery window';
COMMENT ON FUNCTION restore_character IS 'Restores a soft-deleted character within recovery window';
COMMENT ON FUNCTION find_nearby_characters IS 'Finds characters within specified distance for networking';
COMMENT ON FUNCTION calculate_derived_stats IS 'Recalculates combat stats based on primary attributes and class';