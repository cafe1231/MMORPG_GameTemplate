-- Create characters table
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) UNIQUE NOT NULL,
    slot_number INTEGER NOT NULL,
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,
    class_type VARCHAR(50) NOT NULL, -- warrior, mage, rogue, etc.
    race VARCHAR(50) NOT NULL, -- human, elf, dwarf, etc.
    gender VARCHAR(20) NOT NULL, -- male, female, other
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deletion_scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_played_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total_play_time INTERVAL DEFAULT '0 seconds',
    CONSTRAINT check_slot_number CHECK (slot_number >= 1 AND slot_number <= 100),
    CONSTRAINT check_level CHECK (level >= 1 AND level <= 100),
    CONSTRAINT check_experience CHECK (experience >= 0),
    CONSTRAINT check_character_name CHECK (LENGTH(name) >= 3 AND LENGTH(name) <= 30),
    CONSTRAINT unique_user_slot UNIQUE (user_id, slot_number)
);

-- Create indexes for performance
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_characters_name ON characters(LOWER(name));
CREATE INDEX idx_characters_is_deleted ON characters(is_deleted);
CREATE INDEX idx_characters_deleted_at ON characters(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_characters_level ON characters(level);
CREATE INDEX idx_characters_class_type ON characters(class_type);
CREATE INDEX idx_characters_race ON characters(race);
CREATE INDEX idx_characters_last_played_at ON characters(last_played_at);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_characters_updated_at BEFORE UPDATE
    ON characters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to handle character soft deletion
CREATE OR REPLACE FUNCTION soft_delete_character(character_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE characters 
    SET 
        is_deleted = TRUE,
        deleted_at = NOW(),
        deletion_scheduled_at = NOW() + INTERVAL '30 days'
    WHERE id = character_id AND is_deleted = FALSE;
END;
$$ LANGUAGE plpgsql;

-- Create function to restore deleted character
CREATE OR REPLACE FUNCTION restore_character(character_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE characters 
    SET 
        is_deleted = FALSE,
        deleted_at = NULL,
        deletion_scheduled_at = NULL
    WHERE id = character_id 
        AND is_deleted = TRUE 
        AND deletion_scheduled_at > NOW();
END;
$$ LANGUAGE plpgsql;

-- Create function to permanently delete characters past recovery period
CREATE OR REPLACE FUNCTION cleanup_deleted_characters()
RETURNS void AS $$
BEGIN
    DELETE FROM characters 
    WHERE is_deleted = TRUE 
        AND deletion_scheduled_at IS NOT NULL 
        AND deletion_scheduled_at <= NOW();
END;
$$ LANGUAGE plpgsql;

-- Create function to update character count on user
CREATE OR REPLACE FUNCTION update_user_character_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users 
        SET character_count = (
            SELECT COUNT(*) 
            FROM characters 
            WHERE user_id = NEW.user_id AND is_deleted = FALSE
        )
        WHERE id = NEW.user_id;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.is_deleted != NEW.is_deleted THEN
            UPDATE users 
            SET character_count = (
                SELECT COUNT(*) 
                FROM characters 
                WHERE user_id = NEW.user_id AND is_deleted = FALSE
            )
            WHERE id = NEW.user_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users 
        SET character_count = (
            SELECT COUNT(*) 
            FROM characters 
            WHERE user_id = OLD.user_id AND is_deleted = FALSE
        )
        WHERE id = OLD.user_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to maintain user character count
CREATE TRIGGER maintain_user_character_count
AFTER INSERT OR UPDATE OR DELETE ON characters
FOR EACH ROW EXECUTE FUNCTION update_user_character_count();