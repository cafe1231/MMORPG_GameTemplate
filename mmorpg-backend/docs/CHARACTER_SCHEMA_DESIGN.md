# Character Database Schema Design

## Overview

This document describes the character system database schema for Phase 1.5 of the MMORPG Game Template. The schema extends the existing authentication system with comprehensive character management capabilities.

## Database Tables

### 1. Characters Table (`characters`)
Main table storing core character information.

**Key Features:**
- Multiple characters per user (5 default, 10 premium)
- Soft deletion with 30-day recovery window
- Unique character names (case-insensitive)
- Slot-based organization (1-100)
- Play time tracking

**Key Fields:**
- `id`: UUID primary key
- `user_id`: Foreign key to users table
- `name`: Unique character name (3-30 chars)
- `slot_number`: Character slot (1-100)
- `level`: Current level (1-100)
- `experience`: Total experience points
- `class_type`: Character class (warrior, mage, rogue, etc.)
- `race`: Character race (human, elf, dwarf, orc, etc.)
- `gender`: Character gender
- `is_deleted`: Soft deletion flag
- `deletion_scheduled_at`: Auto-deletion timestamp

### 2. Character Appearance Table (`character_appearance`)
Stores visual customization data.

**Key Features:**
- Comprehensive appearance customization
- Hex color support for skin, eyes, hair
- Body proportions with JSON storage
- Support for scars, tattoos, accessories

**Key Fields:**
- `character_id`: One-to-one with characters
- `face_type`, `skin_color`, `eye_color`
- `hair_style`, `hair_color`
- `body_type`, `height` (0.80-1.20 multiplier)
- `body_proportions`: JSONB for detailed adjustments

### 3. Character Stats Table (`character_stats`)
Manages character attributes and combat statistics.

**Key Features:**
- Primary stats (STR, DEX, INT, WIS, CON, CHA)
- Derived stats auto-calculation based on class
- Resource management (health, mana, stamina)
- Combat modifiers (crit, dodge, block)
- Stat/skill point allocation

**Key Fields:**
- Primary attributes (1-999 range)
- Current/max health, mana, stamina
- Combat stats (attack_power, spell_power, defense)
- Percentage modifiers (critical_chance, dodge_chance)
- Available stat/skill points

### 4. Character Position Table (`character_position`)
Tracks character location and movement data.

**Key Features:**
- 3D position and rotation tracking
- Multi-world/zone support
- Instance tracking for dungeons/raids
- Safe position for respawn/unstuck
- Spatial indexing for proximity queries

**Key Fields:**
- World, zone, and map identifiers
- 3D coordinates (x, y, z)
- Rotation (pitch, yaw, roll)
- Velocity for movement prediction
- Instance information
- Safe position coordinates

## Key Database Features

### 1. Automatic Initialization
When a character is created:
- Appearance defaults are set
- Class-based stats are assigned
- Starting position based on race
- All related tables are populated

### 2. Constraint Validation
- Character limit per user enforced
- Name uniqueness (case-insensitive)
- Name format validation (alphanumeric)
- Slot number uniqueness per user
- Stat range validation

### 3. Performance Optimization
- Comprehensive indexing strategy
- Spatial index for position queries
- Materialized view for statistics
- Partial indexes for common queries

### 4. Audit and History
- Character name change history
- Automatic timestamp updates
- Soft deletion tracking

## Key Functions

### Character Management
- `soft_delete_character(character_id)`: Mark character for deletion
- `restore_character(character_id)`: Restore within 30 days
- `cleanup_deleted_characters()`: Remove expired characters

### Position and Movement
- `save_safe_position(character_id)`: Save current as safe position
- `teleport_to_safe_position(character_id)`: Unstuck feature
- `find_nearby_characters(character_id, distance)`: Proximity search

### Stats and Progression
- `calculate_derived_stats(character_id)`: Update combat stats
- `update_character_play_time(character_id, duration)`: Track playtime

### Utility Functions
- `get_user_characters(user_id)`: Paginated character list
- `is_character_name_available(name)`: Name availability check
- `refresh_character_statistics()`: Update statistics view

## Migration Files

1. `004_create_characters_table.sql` - Main character table
2. `005_create_character_appearance_table.sql` - Appearance customization
3. `006_create_character_stats_table.sql` - Stats and attributes
4. `007_create_character_position_table.sql` - Position tracking
5. `008_create_character_initialization_triggers.sql` - Auto-initialization
6. `009_create_character_performance_indexes.sql` - Performance optimization

## Usage Examples

### Create a Character
```sql
INSERT INTO characters (user_id, name, slot_number, class_type, race, gender)
VALUES ('user-uuid', 'Aragorn', 1, 'warrior', 'human', 'male');
-- Triggers automatically create appearance, stats, and position records
```

### Find Nearby Characters
```sql
SELECT * FROM find_nearby_characters('character-uuid', 500.0);
```

### Soft Delete and Restore
```sql
-- Delete
SELECT soft_delete_character('character-uuid');

-- Restore (within 30 days)
SELECT restore_character('character-uuid');
```

### Get User's Characters
```sql
SELECT * FROM get_user_characters('user-uuid', false, 10, 0);
```

## Performance Considerations

1. **Spatial Queries**: Use the spatial index for efficient proximity searches
2. **Character Lists**: Use the composite indexes for user character queries
3. **Statistics**: Refresh the materialized view periodically (e.g., hourly)
4. **Cleanup Jobs**: Schedule regular cleanup of expired deletions and sessions

## Future Extensibility

The schema is designed to support future phases:
- **Phase 2**: Position table ready for real-time networking
- **Phase 3**: Stats table can be extended with new attributes
- **Phase 4**: Appearance system supports additional customization
- **Items/Inventory**: Can reference character_id as foreign key
- **Skills/Abilities**: Can extend the stats system