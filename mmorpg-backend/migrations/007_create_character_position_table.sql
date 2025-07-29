-- Create character position table for tracking location and movement
CREATE TABLE IF NOT EXISTS character_position (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID UNIQUE NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    -- World location
    world_id VARCHAR(100) NOT NULL DEFAULT 'starter_zone',
    zone_id VARCHAR(100) NOT NULL DEFAULT 'tutorial_area',
    map_id VARCHAR(100) NOT NULL DEFAULT 'main',
    -- 3D position coordinates
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    position_z DOUBLE PRECISION NOT NULL DEFAULT 100.0, -- Start slightly above ground
    -- Rotation (Unreal Engine uses Pitch, Yaw, Roll in degrees)
    rotation_pitch REAL DEFAULT 0.0,
    rotation_yaw REAL DEFAULT 0.0,
    rotation_roll REAL DEFAULT 0.0,
    -- Velocity for movement prediction
    velocity_x REAL DEFAULT 0.0,
    velocity_y REAL DEFAULT 0.0,
    velocity_z REAL DEFAULT 0.0,
    -- Instance information for dungeons/raids
    instance_id UUID,
    instance_type VARCHAR(50), -- dungeon, raid, battleground, etc.
    -- Safe position for respawn/unstuck
    safe_position_x DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    safe_position_y DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    safe_position_z DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    safe_world_id VARCHAR(100) NOT NULL DEFAULT 'starter_zone',
    safe_zone_id VARCHAR(100) NOT NULL DEFAULT 'tutorial_area',
    -- Metadata
    last_movement TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Constraints for rotation values (degrees)
    CONSTRAINT check_rotation CHECK (
        rotation_pitch >= -90 AND rotation_pitch <= 90 AND
        rotation_yaw >= -180 AND rotation_yaw <= 180 AND
        rotation_roll >= -180 AND rotation_roll <= 180
    ),
    -- Reasonable position bounds (can be adjusted based on world size)
    CONSTRAINT check_position CHECK (
        position_x >= -1000000 AND position_x <= 1000000 AND
        position_y >= -1000000 AND position_y <= 1000000 AND
        position_z >= -10000 AND position_z <= 50000
    )
);

-- Create indexes for performance
CREATE INDEX idx_character_position_character_id ON character_position(character_id);
CREATE INDEX idx_character_position_world_zone ON character_position(world_id, zone_id);
CREATE INDEX idx_character_position_instance ON character_position(instance_id) WHERE instance_id IS NOT NULL;
CREATE INDEX idx_character_position_last_movement ON character_position(last_movement);

-- Spatial index for position queries (finding nearby players)
CREATE INDEX idx_character_position_spatial ON character_position 
    USING GIST (
        point(position_x, position_y)
    );

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_character_position_updated_at BEFORE UPDATE
    ON character_position FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to update last movement timestamp
CREATE OR REPLACE FUNCTION update_last_movement()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.position_x != NEW.position_x OR 
       OLD.position_y != NEW.position_y OR 
       OLD.position_z != NEW.position_z OR
       OLD.rotation_pitch != NEW.rotation_pitch OR
       OLD.rotation_yaw != NEW.rotation_yaw OR
       OLD.rotation_roll != NEW.rotation_roll THEN
        NEW.last_movement = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to track movement
CREATE TRIGGER track_character_movement BEFORE UPDATE
    ON character_position FOR EACH ROW EXECUTE FUNCTION update_last_movement();

-- Create function to save current position as safe position
CREATE OR REPLACE FUNCTION save_safe_position(char_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE character_position
    SET
        safe_position_x = position_x,
        safe_position_y = position_y,
        safe_position_z = position_z,
        safe_world_id = world_id,
        safe_zone_id = zone_id
    WHERE character_id = char_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to teleport to safe position (unstuck feature)
CREATE OR REPLACE FUNCTION teleport_to_safe_position(char_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE character_position
    SET
        position_x = safe_position_x,
        position_y = safe_position_y,
        position_z = safe_position_z,
        world_id = safe_world_id,
        zone_id = safe_zone_id,
        velocity_x = 0,
        velocity_y = 0,
        velocity_z = 0,
        instance_id = NULL,
        instance_type = NULL
    WHERE character_id = char_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to find nearby characters (for Phase 2 networking)
CREATE OR REPLACE FUNCTION find_nearby_characters(
    char_id UUID,
    max_distance DOUBLE PRECISION DEFAULT 1000.0
)
RETURNS TABLE(
    character_id UUID,
    character_name VARCHAR(30),
    distance DOUBLE PRECISION,
    position_x DOUBLE PRECISION,
    position_y DOUBLE PRECISION,
    position_z DOUBLE PRECISION
) AS $$
DECLARE
    my_world_id VARCHAR(100);
    my_zone_id VARCHAR(100);
    my_instance_id UUID;
    my_x DOUBLE PRECISION;
    my_y DOUBLE PRECISION;
    my_z DOUBLE PRECISION;
BEGIN
    -- Get current character's position
    SELECT world_id, zone_id, instance_id, position_x, position_y, position_z
    INTO my_world_id, my_zone_id, my_instance_id, my_x, my_y, my_z
    FROM character_position
    WHERE character_id = char_id;
    
    -- Find nearby characters in the same world/zone/instance
    RETURN QUERY
    SELECT 
        c.id,
        c.name,
        SQRT(
            POWER(cp.position_x - my_x, 2) + 
            POWER(cp.position_y - my_y, 2) + 
            POWER(cp.position_z - my_z, 2)
        ) AS distance,
        cp.position_x,
        cp.position_y,
        cp.position_z
    FROM character_position cp
    JOIN characters c ON c.id = cp.character_id
    WHERE cp.character_id != char_id
        AND cp.world_id = my_world_id
        AND cp.zone_id = my_zone_id
        AND (cp.instance_id = my_instance_id OR (cp.instance_id IS NULL AND my_instance_id IS NULL))
        AND c.is_deleted = FALSE
        AND SQRT(
            POWER(cp.position_x - my_x, 2) + 
            POWER(cp.position_y - my_y, 2) + 
            POWER(cp.position_z - my_z, 2)
        ) <= max_distance
    ORDER BY distance;
END;
$$ LANGUAGE plpgsql;