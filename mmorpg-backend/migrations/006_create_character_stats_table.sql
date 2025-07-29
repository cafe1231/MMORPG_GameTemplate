-- Create character stats table
CREATE TABLE IF NOT EXISTS character_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID UNIQUE NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    -- Primary stats (base values before equipment/buffs)
    strength INTEGER DEFAULT 10,
    dexterity INTEGER DEFAULT 10,
    intelligence INTEGER DEFAULT 10,
    wisdom INTEGER DEFAULT 10,
    constitution INTEGER DEFAULT 10,
    charisma INTEGER DEFAULT 10,
    -- Combat stats
    health_current INTEGER DEFAULT 100,
    health_max INTEGER DEFAULT 100,
    mana_current INTEGER DEFAULT 50,
    mana_max INTEGER DEFAULT 50,
    stamina_current INTEGER DEFAULT 100,
    stamina_max INTEGER DEFAULT 100,
    -- Derived stats (will be expanded in Phase 3)
    attack_power INTEGER DEFAULT 0,
    spell_power INTEGER DEFAULT 0,
    defense INTEGER DEFAULT 0,
    critical_chance DECIMAL(5,2) DEFAULT 5.00, -- Percentage
    critical_damage DECIMAL(5,2) DEFAULT 150.00, -- Percentage
    dodge_chance DECIMAL(5,2) DEFAULT 5.00, -- Percentage
    block_chance DECIMAL(5,2) DEFAULT 0.00, -- Percentage
    -- Movement and other stats
    movement_speed DECIMAL(5,2) DEFAULT 100.00, -- Base 100%
    attack_speed DECIMAL(5,2) DEFAULT 100.00, -- Base 100%
    cast_speed DECIMAL(5,2) DEFAULT 100.00, -- Base 100%
    -- Resource regeneration (per second)
    health_regen DECIMAL(5,2) DEFAULT 1.00,
    mana_regen DECIMAL(5,2) DEFAULT 1.00,
    stamina_regen DECIMAL(5,2) DEFAULT 5.00,
    -- Points for player allocation
    stat_points_available INTEGER DEFAULT 0,
    skill_points_available INTEGER DEFAULT 0,
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Constraints
    CONSTRAINT check_primary_stats CHECK (
        strength >= 1 AND strength <= 999 AND
        dexterity >= 1 AND dexterity <= 999 AND
        intelligence >= 1 AND intelligence <= 999 AND
        wisdom >= 1 AND wisdom <= 999 AND
        constitution >= 1 AND constitution <= 999 AND
        charisma >= 1 AND charisma <= 999
    ),
    CONSTRAINT check_health CHECK (
        health_current >= 0 AND 
        health_current <= health_max AND 
        health_max >= 1 AND health_max <= 999999
    ),
    CONSTRAINT check_mana CHECK (
        mana_current >= 0 AND 
        mana_current <= mana_max AND 
        mana_max >= 0 AND mana_max <= 999999
    ),
    CONSTRAINT check_stamina CHECK (
        stamina_current >= 0 AND 
        stamina_current <= stamina_max AND 
        stamina_max >= 0 AND stamina_max <= 999999
    ),
    CONSTRAINT check_percentages CHECK (
        critical_chance >= 0 AND critical_chance <= 100 AND
        dodge_chance >= 0 AND dodge_chance <= 100 AND
        block_chance >= 0 AND block_chance <= 100
    ),
    CONSTRAINT check_points CHECK (
        stat_points_available >= 0 AND stat_points_available <= 9999 AND
        skill_points_available >= 0 AND skill_points_available <= 9999
    )
);

-- Create indexes for performance
CREATE INDEX idx_character_stats_character_id ON character_stats(character_id);
CREATE INDEX idx_character_stats_health ON character_stats(health_current, health_max);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_character_stats_updated_at BEFORE UPDATE
    ON character_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to calculate derived stats based on primary stats and class
CREATE OR REPLACE FUNCTION calculate_derived_stats(char_id UUID)
RETURNS void AS $$
DECLARE
    char_class VARCHAR(50);
    str INTEGER;
    dex INTEGER;
    int INTEGER;
    wis INTEGER;
    con INTEGER;
    cha INTEGER;
BEGIN
    -- Get character class and primary stats
    SELECT c.class_type, cs.strength, cs.dexterity, cs.intelligence, 
           cs.wisdom, cs.constitution, cs.charisma
    INTO char_class, str, dex, int, wis, con, cha
    FROM characters c
    JOIN character_stats cs ON cs.character_id = c.id
    WHERE c.id = char_id;
    
    -- Update derived stats based on class formulas
    UPDATE character_stats
    SET
        health_max = CASE
            WHEN char_class IN ('warrior', 'paladin') THEN 100 + (con * 10) + (str * 2)
            WHEN char_class IN ('rogue', 'ranger') THEN 100 + (con * 8) + (dex * 2)
            WHEN char_class IN ('mage', 'warlock') THEN 100 + (con * 6) + (int * 1)
            WHEN char_class IN ('priest', 'druid') THEN 100 + (con * 7) + (wis * 2)
            ELSE 100 + (con * 8)
        END,
        mana_max = CASE
            WHEN char_class IN ('mage', 'warlock') THEN 50 + (int * 10) + (wis * 2)
            WHEN char_class IN ('priest', 'druid') THEN 50 + (wis * 10) + (int * 2)
            WHEN char_class IN ('paladin') THEN 50 + (wis * 5)
            ELSE 50 + (int * 2)
        END,
        stamina_max = 100 + (con * 5) + (str * 2),
        attack_power = CASE
            WHEN char_class IN ('warrior', 'paladin') THEN str * 2 + dex
            WHEN char_class IN ('rogue', 'ranger') THEN dex * 2 + str
            ELSE str + dex
        END,
        spell_power = CASE
            WHEN char_class IN ('mage', 'warlock') THEN int * 3
            WHEN char_class IN ('priest', 'druid') THEN wis * 3
            WHEN char_class IN ('paladin') THEN (wis * 2 + int) / 2
            ELSE 0
        END,
        defense = con * 2 + (str + dex) / 2,
        critical_chance = 5.00 + (dex * 0.1),
        dodge_chance = 5.00 + (dex * 0.2),
        health_regen = 1.00 + (con * 0.1),
        mana_regen = CASE
            WHEN char_class IN ('mage', 'warlock', 'priest', 'druid') THEN 1.00 + (wis * 0.2)
            ELSE 1.00 + (wis * 0.05)
        END,
        stamina_regen = 5.00 + (con * 0.2)
    WHERE character_id = char_id;
    
    -- Ensure current values don't exceed new maximums
    UPDATE character_stats
    SET
        health_current = LEAST(health_current, health_max),
        mana_current = LEAST(mana_current, mana_max),
        stamina_current = LEAST(stamina_current, stamina_max)
    WHERE character_id = char_id;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to recalculate derived stats when primary stats change
CREATE OR REPLACE FUNCTION trigger_calculate_derived_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR (
        OLD.strength != NEW.strength OR
        OLD.dexterity != NEW.dexterity OR
        OLD.intelligence != NEW.intelligence OR
        OLD.wisdom != NEW.wisdom OR
        OLD.constitution != NEW.constitution OR
        OLD.charisma != NEW.charisma
    ) THEN
        PERFORM calculate_derived_stats(NEW.character_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER recalculate_derived_stats
AFTER INSERT OR UPDATE ON character_stats
FOR EACH ROW EXECUTE FUNCTION trigger_calculate_derived_stats();