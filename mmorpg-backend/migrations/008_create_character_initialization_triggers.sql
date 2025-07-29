-- Create function to initialize character data when a new character is created
CREATE OR REPLACE FUNCTION initialize_character_data()
RETURNS TRIGGER AS $$
DECLARE
    default_stats RECORD;
BEGIN
    -- Initialize character appearance with defaults based on race/gender
    INSERT INTO character_appearance (character_id)
    VALUES (NEW.id);
    
    -- Initialize character stats with class-based defaults
    INSERT INTO character_stats (
        character_id,
        strength, dexterity, intelligence, wisdom, constitution, charisma
    )
    SELECT
        NEW.id,
        CASE 
            WHEN NEW.class_type IN ('warrior', 'paladin') THEN 15
            WHEN NEW.class_type IN ('rogue', 'ranger') THEN 12
            ELSE 10
        END AS strength,
        CASE 
            WHEN NEW.class_type IN ('rogue', 'ranger') THEN 15
            WHEN NEW.class_type IN ('warrior', 'paladin') THEN 12
            ELSE 10
        END AS dexterity,
        CASE 
            WHEN NEW.class_type IN ('mage', 'warlock') THEN 15
            WHEN NEW.class_type IN ('priest', 'druid') THEN 12
            ELSE 10
        END AS intelligence,
        CASE 
            WHEN NEW.class_type IN ('priest', 'druid') THEN 15
            WHEN NEW.class_type IN ('mage', 'warlock', 'paladin') THEN 12
            ELSE 10
        END AS wisdom,
        CASE 
            WHEN NEW.class_type IN ('warrior', 'paladin') THEN 14
            WHEN NEW.class_type IN ('rogue', 'ranger') THEN 12
            ELSE 11
        END AS constitution,
        CASE 
            WHEN NEW.class_type IN ('paladin', 'priest') THEN 12
            WHEN NEW.class_type = 'warlock' THEN 8
            ELSE 10
        END AS charisma;
    
    -- Initialize character position based on race starting zone
    INSERT INTO character_position (
        character_id,
        world_id, zone_id, map_id,
        position_x, position_y, position_z,
        safe_position_x, safe_position_y, safe_position_z,
        safe_world_id, safe_zone_id
    )
    SELECT
        NEW.id,
        CASE NEW.race
            WHEN 'human' THEN 'eastern_kingdoms'
            WHEN 'elf' THEN 'kalimdor'
            WHEN 'dwarf' THEN 'eastern_kingdoms'
            WHEN 'orc' THEN 'kalimdor'
            ELSE 'starter_zone'
        END AS world_id,
        CASE NEW.race
            WHEN 'human' THEN 'elwynn_forest'
            WHEN 'elf' THEN 'teldrassil'
            WHEN 'dwarf' THEN 'dun_morogh'
            WHEN 'orc' THEN 'durotar'
            ELSE 'tutorial_area'
        END AS zone_id,
        'main' AS map_id,
        CASE NEW.race
            WHEN 'human' THEN -8949.95
            WHEN 'elf' THEN 10311.3
            WHEN 'dwarf' THEN -6240.32
            WHEN 'orc' THEN -618.518
            ELSE 0.0
        END AS position_x,
        CASE NEW.race
            WHEN 'human' THEN -132.493
            WHEN 'elf' THEN 831.267
            WHEN 'dwarf' THEN 331.031
            WHEN 'orc' THEN -4147.39
            ELSE 0.0
        END AS position_y,
        CASE NEW.race
            WHEN 'human' THEN 83.8126
            WHEN 'elf' THEN 1326.41
            WHEN 'dwarf' THEN 384.248
            WHEN 'orc' THEN 49.921
            ELSE 100.0
        END AS position_z,
        -- Safe positions are the same as starting positions
        CASE NEW.race
            WHEN 'human' THEN -8949.95
            WHEN 'elf' THEN 10311.3
            WHEN 'dwarf' THEN -6240.32
            WHEN 'orc' THEN -618.518
            ELSE 0.0
        END AS safe_position_x,
        CASE NEW.race
            WHEN 'human' THEN -132.493
            WHEN 'elf' THEN 831.267
            WHEN 'dwarf' THEN 331.031
            WHEN 'orc' THEN -4147.39
            ELSE 0.0
        END AS safe_position_y,
        CASE NEW.race
            WHEN 'human' THEN 83.8126
            WHEN 'elf' THEN 1326.41
            WHEN 'dwarf' THEN 384.248
            WHEN 'orc' THEN 49.921
            ELSE 100.0
        END AS safe_position_z,
        CASE NEW.race
            WHEN 'human' THEN 'eastern_kingdoms'
            WHEN 'elf' THEN 'kalimdor'
            WHEN 'dwarf' THEN 'eastern_kingdoms'
            WHEN 'orc' THEN 'kalimdor'
            ELSE 'starter_zone'
        END AS safe_world_id,
        CASE NEW.race
            WHEN 'human' THEN 'elwynn_forest'
            WHEN 'elf' THEN 'teldrassil'
            WHEN 'dwarf' THEN 'dun_morogh'
            WHEN 'orc' THEN 'durotar'
            ELSE 'tutorial_area'
        END AS safe_zone_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to initialize character data after character creation
CREATE TRIGGER initialize_character_after_insert
AFTER INSERT ON characters
FOR EACH ROW EXECUTE FUNCTION initialize_character_data();

-- Create function to validate character creation constraints
CREATE OR REPLACE FUNCTION validate_character_creation()
RETURNS TRIGGER AS $$
DECLARE
    user_char_count INTEGER;
    user_max_chars INTEGER;
    name_exists BOOLEAN;
BEGIN
    -- Check if user hasn't exceeded character limit
    SELECT character_count, max_characters 
    INTO user_char_count, user_max_chars
    FROM users 
    WHERE id = NEW.user_id;
    
    IF user_char_count >= user_max_chars THEN
        RAISE EXCEPTION 'User has reached maximum character limit (% of %)', 
            user_char_count, user_max_chars;
    END IF;
    
    -- Check if character name is already taken (case-insensitive)
    SELECT EXISTS(
        SELECT 1 FROM characters 
        WHERE LOWER(name) = LOWER(NEW.name) 
        AND id != COALESCE(NEW.id, gen_random_uuid())
    ) INTO name_exists;
    
    IF name_exists THEN
        RAISE EXCEPTION 'Character name "%" is already taken', NEW.name;
    END IF;
    
    -- Validate character name format (alphanumeric only, no spaces)
    IF NEW.name !~ '^[A-Za-z][A-Za-z0-9]{2,29}$' THEN
        RAISE EXCEPTION 'Character name must be 3-30 characters, start with a letter, and contain only letters and numbers';
    END IF;
    
    -- Ensure slot number is unique for this user
    IF EXISTS(
        SELECT 1 FROM characters 
        WHERE user_id = NEW.user_id 
        AND slot_number = NEW.slot_number 
        AND id != COALESCE(NEW.id, gen_random_uuid())
        AND is_deleted = FALSE
    ) THEN
        RAISE EXCEPTION 'Slot number % is already occupied', NEW.slot_number;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to validate character creation
CREATE TRIGGER validate_character_before_insert
BEFORE INSERT ON characters
FOR EACH ROW EXECUTE FUNCTION validate_character_creation();

-- Create view for active characters with all related data
CREATE OR REPLACE VIEW v_character_details AS
SELECT 
    c.id,
    c.user_id,
    c.name,
    c.slot_number,
    c.level,
    c.experience,
    c.class_type,
    c.race,
    c.gender,
    c.created_at,
    c.last_played_at,
    c.total_play_time,
    -- Stats
    cs.health_current,
    cs.health_max,
    cs.mana_current,
    cs.mana_max,
    cs.stamina_current,
    cs.stamina_max,
    cs.strength,
    cs.dexterity,
    cs.intelligence,
    cs.wisdom,
    cs.constitution,
    cs.charisma,
    cs.stat_points_available,
    cs.skill_points_available,
    -- Position
    cp.world_id,
    cp.zone_id,
    cp.position_x,
    cp.position_y,
    cp.position_z,
    cp.rotation_yaw,
    -- Appearance (selected fields)
    ca.face_type,
    ca.skin_color,
    ca.hair_style,
    ca.hair_color,
    ca.body_type
FROM characters c
LEFT JOIN character_stats cs ON cs.character_id = c.id
LEFT JOIN character_position cp ON cp.character_id = c.id
LEFT JOIN character_appearance ca ON ca.character_id = c.id
WHERE c.is_deleted = FALSE;

-- Create function to update character play time on logout
CREATE OR REPLACE FUNCTION update_character_play_time(
    char_id UUID,
    session_duration INTERVAL
)
RETURNS void AS $$
BEGIN
    UPDATE characters
    SET 
        total_play_time = total_play_time + session_duration,
        last_played_at = NOW()
    WHERE id = char_id;
END;
$$ LANGUAGE plpgsql;