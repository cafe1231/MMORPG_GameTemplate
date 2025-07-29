-- Create character appearance table
CREATE TABLE IF NOT EXISTS character_appearance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID UNIQUE NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    -- Face customization
    face_type INTEGER DEFAULT 1,
    skin_color VARCHAR(7) DEFAULT '#FFD4B2', -- Hex color
    eye_color VARCHAR(7) DEFAULT '#4B8BF5',
    hair_style INTEGER DEFAULT 1,
    hair_color VARCHAR(7) DEFAULT '#3B2F2F',
    facial_hair_style INTEGER DEFAULT 0, -- 0 = none
    facial_hair_color VARCHAR(7) DEFAULT '#3B2F2F',
    -- Body customization
    body_type INTEGER DEFAULT 1, -- 1 = athletic, 2 = muscular, 3 = slim, etc.
    height DECIMAL(3,2) DEFAULT 1.0, -- Multiplier: 0.80 to 1.20
    body_proportions JSONB DEFAULT '{}', -- Detailed body adjustments
    -- Additional features
    scars INTEGER[] DEFAULT '{}',
    tattoos INTEGER[] DEFAULT '{}',
    accessories INTEGER[] DEFAULT '{}',
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_height CHECK (height >= 0.80 AND height <= 1.20),
    CONSTRAINT check_face_type CHECK (face_type >= 1 AND face_type <= 20),
    CONSTRAINT check_hair_style CHECK (hair_style >= 0 AND hair_style <= 50),
    CONSTRAINT check_body_type CHECK (body_type >= 1 AND body_type <= 10)
);

-- Create indexes for performance
CREATE INDEX idx_character_appearance_character_id ON character_appearance(character_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_character_appearance_updated_at BEFORE UPDATE
    ON character_appearance FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to validate hex color format
CREATE OR REPLACE FUNCTION is_valid_hex_color(color VARCHAR(7))
RETURNS BOOLEAN AS $$
BEGIN
    RETURN color ~ '^#[0-9A-Fa-f]{6}$';
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Add constraints to validate hex colors
ALTER TABLE character_appearance
ADD CONSTRAINT check_skin_color CHECK (is_valid_hex_color(skin_color)),
ADD CONSTRAINT check_eye_color CHECK (is_valid_hex_color(eye_color)),
ADD CONSTRAINT check_hair_color CHECK (is_valid_hex_color(hair_color)),
ADD CONSTRAINT check_facial_hair_color CHECK (is_valid_hex_color(facial_hair_color));