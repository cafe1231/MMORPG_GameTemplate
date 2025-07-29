# Character Management API Design

## Table of Contents
1. [Overview](#overview)
2. [API Endpoints](#api-endpoints)
3. [Request/Response Schemas](#requestresponse-schemas)
4. [Error Codes](#error-codes)
5. [Authentication](#authentication)
6. [Rate Limiting](#rate-limiting)
7. [Gateway Configuration](#gateway-configuration)
8. [Examples](#examples)

## Overview

The Character Management API provides comprehensive endpoints for creating, managing, and customizing player characters in the MMORPG. All endpoints use JSON for request/response payloads and require JWT authentication.

### Base URL
```
https://api.mmorpg.com/api/v1/characters
```

### Content Type
All requests and responses use `application/json`.

## API Endpoints

### Character Operations

#### 1. List Characters
**GET** `/api/v1/characters`

Lists all characters for the authenticated user.

**Response:** 200 OK
```json
{
  "characters": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Shadowblade",
      "slot_number": 1,
      "level": 45,
      "experience": 1250000,
      "class_type": "rogue",
      "race": "elf",
      "gender": "female",
      "created_at": "2025-01-15T10:30:00Z",
      "last_played_at": "2025-01-28T15:45:30Z",
      "total_play_time": 432000
    }
  ],
  "total_slots": 10,
  "used_slots": 3,
  "max_slots": 10
}
```

#### 2. Get Character Details
**GET** `/api/v1/characters/{character_id}`

Retrieves detailed information about a specific character.

**Response:** 200 OK
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Shadowblade",
  "slot_number": 1,
  "level": 45,
  "experience": 1250000,
  "class_type": "rogue",
  "race": "elf",
  "gender": "female",
  "created_at": "2025-01-15T10:30:00Z",
  "last_played_at": "2025-01-28T15:45:30Z",
  "total_play_time": 432000,
  "is_deleted": false,
  "is_online": false,
  "current_world": "realm_01",
  "guild_id": null,
  "guild_name": null
}
```

#### 3. Create Character
**POST** `/api/v1/characters`

Creates a new character.

**Request:**
```json
{
  "name": "Shadowblade",
  "slot_number": 1,
  "class_type": "rogue",
  "race": "elf",
  "gender": "female",
  "appearance": {
    "face_type": 3,
    "skin_color": "#F5DEB3",
    "eye_color": "#4B0082",
    "hair_style": 7,
    "hair_color": "#2F4F4F",
    "facial_hair_style": 0,
    "facial_hair_color": "#000000",
    "body_type": 2,
    "height": 1.75
  }
}
```

**Response:** 201 Created
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Shadowblade",
  "slot_number": 1,
  "level": 1,
  "experience": 0,
  "class_type": "rogue",
  "race": "elf",
  "gender": "female",
  "created_at": "2025-01-29T10:30:00Z",
  "last_played_at": "2025-01-29T10:30:00Z",
  "total_play_time": 0
}
```

#### 4. Check Name Availability
**GET** `/api/v1/characters/check-name?name={character_name}`

Checks if a character name is available.

**Response:** 200 OK
```json
{
  "available": true,
  "name": "Shadowblade",
  "suggestions": []
}
```

**Response (Name Taken):** 200 OK
```json
{
  "available": false,
  "name": "Shadowblade",
  "suggestions": [
    "Shadowblade123",
    "ShadowbladeX",
    "Shadowblade_01"
  ]
}
```

#### 5. Delete Character (Soft Delete)
**DELETE** `/api/v1/characters/{character_id}`

Marks a character for deletion (30-day recovery period).

**Response:** 200 OK
```json
{
  "message": "Character marked for deletion",
  "deletion_date": "2025-02-28T10:30:00Z",
  "recovery_deadline": "2025-03-30T10:30:00Z"
}
```

#### 6. Restore Character
**POST** `/api/v1/characters/{character_id}/restore`

Restores a soft-deleted character.

**Response:** 200 OK
```json
{
  "message": "Character restored successfully",
  "character": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Shadowblade",
    "slot_number": 1
  }
}
```

#### 7. Select Character for Gameplay
**POST** `/api/v1/characters/{character_id}/select`

Selects a character to enter the game world.

**Response:** 200 OK
```json
{
  "success": true,
  "character_id": "550e8400-e29b-41d4-a716-446655440000",
  "world_server": "realm_01.game.mmorpg.com",
  "world_port": 7777,
  "session_token": "eyJhbGciOiJIUzI1NiIs...",
  "spawn_location": {
    "world_id": "realm_01",
    "zone_id": "elven_forest",
    "map_id": "starting_area",
    "position": {
      "x": 1250.5,
      "y": 2048.0,
      "z": 156.75
    }
  }
}
```

### Character Appearance

#### 8. Get Character Appearance
**GET** `/api/v1/characters/{character_id}/appearance`

**Response:** 200 OK
```json
{
  "face_type": 3,
  "skin_color": "#F5DEB3",
  "eye_color": "#4B0082",
  "hair_style": 7,
  "hair_color": "#2F4F4F",
  "facial_hair_style": 0,
  "facial_hair_color": "#000000",
  "body_type": 2,
  "height": 1.75,
  "body_proportions": {
    "shoulder_width": 1.0,
    "chest_size": 1.0,
    "waist_size": 1.0,
    "hip_size": 1.0,
    "arm_length": 1.0,
    "leg_length": 1.0
  },
  "scars": [1, 3],
  "tattoos": [5, 12],
  "accessories": [101, 205]
}
```

#### 9. Update Character Appearance
**PUT** `/api/v1/characters/{character_id}/appearance`

Updates character appearance (may require in-game currency or items).

**Request:**
```json
{
  "hair_style": 12,
  "hair_color": "#8B4513",
  "tattoos": [5, 12, 18]
}
```

**Response:** 200 OK
```json
{
  "success": true,
  "cost": {
    "currency": "gold",
    "amount": 500
  },
  "appearance": {
    "face_type": 3,
    "skin_color": "#F5DEB3",
    "eye_color": "#4B0082",
    "hair_style": 12,
    "hair_color": "#8B4513",
    "facial_hair_style": 0,
    "facial_hair_color": "#000000",
    "body_type": 2,
    "height": 1.75,
    "body_proportions": { /* ... */ },
    "scars": [1, 3],
    "tattoos": [5, 12, 18],
    "accessories": [101, 205]
  }
}
```

### Character Stats

#### 10. Get Character Stats
**GET** `/api/v1/characters/{character_id}/stats`

**Response:** 200 OK
```json
{
  "primary_stats": {
    "strength": 25,
    "dexterity": 40,
    "intelligence": 15,
    "wisdom": 18,
    "constitution": 30,
    "charisma": 20
  },
  "combat_stats": {
    "health_current": 2150,
    "health_max": 2500,
    "mana_current": 800,
    "mana_max": 1000,
    "stamina_current": 90,
    "stamina_max": 100
  },
  "derived_stats": {
    "attack_power": 185,
    "spell_power": 65,
    "defense": 245,
    "critical_chance": 22.5,
    "critical_damage": 175.0,
    "dodge_chance": 18.5,
    "block_chance": 12.0,
    "movement_speed": 115.0,
    "attack_speed": 1.85,
    "cast_speed": 1.0,
    "health_regen": 15.5,
    "mana_regen": 8.5,
    "stamina_regen": 5.0
  },
  "points": {
    "stat_points_available": 5,
    "skill_points_available": 3
  }
}
```

#### 11. Allocate Stat Points
**POST** `/api/v1/characters/{character_id}/stats/allocate`

**Request:**
```json
{
  "stat": "strength",
  "points": 3
}
```

**Response:** 200 OK
```json
{
  "success": true,
  "stat": "strength",
  "new_value": 28,
  "points_remaining": 2,
  "affected_stats": {
    "attack_power": 194,
    "carrying_capacity": 280
  }
}
```

### Character Position

#### 12. Get Character Position
**GET** `/api/v1/characters/{character_id}/position`

**Response:** 200 OK
```json
{
  "world_id": "realm_01",
  "zone_id": "elven_forest",
  "map_id": "silverleaf_grove",
  "position": {
    "x": 1250.5,
    "y": 2048.0,
    "z": 156.75
  },
  "rotation": {
    "pitch": 0.0,
    "yaw": 135.5,
    "roll": 0.0
  },
  "velocity": {
    "x": 0.0,
    "y": 0.0,
    "z": 0.0
  },
  "is_safe_zone": true,
  "last_updated": "2025-01-28T15:45:30Z"
}
```

#### 13. Update Character Position
**PUT** `/api/v1/characters/{character_id}/position`

Updates character position (usually called by game server).

**Request:**
```json
{
  "world_id": "realm_01",
  "zone_id": "elven_forest",
  "map_id": "silverleaf_grove",
  "position_x": 1255.5,
  "position_y": 2050.0,
  "position_z": 156.75,
  "rotation_pitch": 0.0,
  "rotation_yaw": 145.5,
  "rotation_roll": 0.0,
  "velocity_x": 5.0,
  "velocity_y": 2.0,
  "velocity_z": 0.0
}
```

### Character Management

#### 14. List Deleted Characters
**GET** `/api/v1/characters/deleted`

Lists soft-deleted characters that can be restored.

**Response:** 200 OK
```json
{
  "deleted_characters": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "name": "OldWarrior",
      "slot_number": 5,
      "level": 20,
      "class_type": "warrior",
      "race": "human",
      "gender": "male",
      "deleted_at": "2025-01-15T10:30:00Z",
      "can_restore_until": "2025-02-14T10:30:00Z",
      "days_remaining": 16
    }
  ]
}
```

#### 15. Permanently Delete Character
**DELETE** `/api/v1/characters/{character_id}/permanent`

Permanently deletes a character (requires confirmation).

**Request:**
```json
{
  "confirmation_code": "DELETE-Shadowblade-2025",
  "password": "current_account_password"
}
```

**Response:** 200 OK
```json
{
  "message": "Character permanently deleted",
  "character_name": "Shadowblade"
}
```

## Request/Response Schemas

### Common Types

#### Character Class Types
- `warrior`
- `mage`
- `rogue`
- `cleric`
- `ranger`
- `paladin`
- `warlock`
- `druid`

#### Character Races
- `human`
- `elf`
- `dwarf`
- `orc`
- `undead`
- `gnome`
- `troll`
- `halfling`

#### Gender Types
- `male`
- `female`
- `other`

#### Body Types
- `0` - Slim
- `1` - Athletic
- `2` - Average
- `3` - Muscular
- `4` - Heavy

### Validation Rules

#### Character Name
- Length: 3-16 characters
- Allowed: Letters, numbers (not at start)
- Not allowed: Special characters, spaces, offensive words
- Case-insensitive uniqueness
- Reserved names blocked

#### Appearance Values
- Face Type: 0-20 (race-dependent)
- Hair Style: 0-30 (race/gender-dependent)
- Height: 0.8-1.2 (multiplier from base race height)
- Colors: Hex format (#RRGGBB)
- Body Proportions: 0.8-1.2 (multipliers)

## Error Codes

### Character-Specific Error Codes

| Code | HTTP Status | Message | Description |
|------|-------------|---------|-------------|
| `CHARACTER_NOT_FOUND` | 404 | Character not found | The specified character does not exist |
| `CHARACTER_NAME_TAKEN` | 409 | Character name is already taken | Name already in use |
| `CHARACTER_LIMIT_REACHED` | 403 | Character limit reached | Account has max characters |
| `INVALID_CHARACTER_NAME` | 400 | Invalid character name | Name violates naming rules |
| `CHARACTER_DELETED` | 410 | Character is deleted | Character is soft-deleted |
| `CHARACTER_CANNOT_RESTORE` | 403 | Character cannot be restored | Past recovery deadline |
| `INVALID_SLOT_NUMBER` | 400 | Invalid slot number | Slot out of range |
| `SLOT_OCCUPIED` | 409 | Character slot occupied | Slot already has character |
| `CHARACTER_BELONGS_TO_OTHER` | 403 | Access denied | Character owned by another user |
| `INVALID_CLASS` | 400 | Invalid character class | Class type not recognized |
| `INVALID_RACE` | 400 | Invalid character race | Race type not recognized |
| `INVALID_GENDER` | 400 | Invalid gender | Gender type not recognized |
| `INVALID_APPEARANCE` | 400 | Invalid appearance option | Appearance value out of range |
| `NO_STAT_POINTS` | 400 | No stat points available | Cannot allocate points |
| `STAT_MAX_REACHED` | 400 | Stat maximum reached | Stat at maximum value |
| `CHARACTER_IN_COMBAT` | 409 | Character in combat | Cannot perform action in combat |
| `CHARACTER_ONLINE` | 409 | Character is online | Cannot modify online character |

### Error Response Format

```json
{
  "error": {
    "code": "CHARACTER_NAME_TAKEN",
    "message": "Character name is already taken",
    "details": {
      "name": "Shadowblade",
      "suggestions": ["Shadowblade123", "ShadowbladeX"]
    }
  },
  "timestamp": "2025-01-29T10:30:00Z",
  "request_id": "req_123456789"
}
```

## Authentication

All character endpoints require JWT authentication via the Authorization header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Required Claims
- `user_id`: User UUID
- `session_id`: Session identifier
- `roles`: User roles array
- `exp`: Token expiration

### Token Validation
The character service validates tokens by calling the auth service via NATS:
```
auth.validate -> { token: "...", required_roles: [] }
```

## Rate Limiting

### Endpoint-Specific Limits

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| List Characters | 60/min | 1 minute |
| Create Character | 5/hour | 1 hour |
| Check Name | 30/min | 1 minute |
| Delete Character | 3/day | 24 hours |
| Update Appearance | 10/hour | 1 hour |
| Allocate Stats | 100/hour | 1 hour |
| Update Position | 100/sec | 1 second |

### Rate Limit Headers

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1706527800
```

### Rate Limit Response

**429 Too Many Requests**
```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests",
    "retry_after": 30
  }
}
```

## Gateway Configuration

### Routing Rules

```yaml
# Gateway routes for character service
routes:
  - path: /api/v1/characters
    service: character-service
    methods: [GET, POST, PUT, DELETE]
    auth_required: true
    rate_limit:
      requests_per_minute: 60
      burst: 10

  - path: /api/v1/characters/check-name
    service: character-service
    methods: [GET]
    auth_required: false  # Allow checking before login
    rate_limit:
      requests_per_minute: 30
      burst: 5

  - path: /api/v1/characters/{id}/position
    service: character-service
    methods: [GET, PUT]
    auth_required: true
    rate_limit:
      requests_per_second: 100
      burst: 200
```

### Service Discovery

```go
// Gateway proxy configuration
services := map[string]string{
    "character-service": "http://character:8081",
    "auth-service":      "http://auth:8080",
}

// Health check endpoints
healthChecks := map[string]string{
    "character-service": "/health",
    "auth-service":      "/health",
}
```

### Middleware Chain

1. **CORS Handler** - Handle preflight requests
2. **Request ID** - Generate unique request ID
3. **Logger** - Log request details
4. **Rate Limiter** - Apply rate limits
5. **Auth Validator** - Validate JWT token
6. **Request Router** - Route to service
7. **Response Handler** - Format response

## Examples

### Example 1: Creating a New Character

**Request:**
```bash
curl -X POST https://api.mmorpg.com/api/v1/characters \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Shadowblade",
    "slot_number": 1,
    "class_type": "rogue",
    "race": "elf",
    "gender": "female",
    "appearance": {
      "face_type": 3,
      "skin_color": "#F5DEB3",
      "eye_color": "#4B0082",
      "hair_style": 7,
      "hair_color": "#2F4F4F",
      "body_type": 2,
      "height": 1.75
    }
  }'
```

**Success Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Shadowblade",
  "slot_number": 1,
  "level": 1,
  "experience": 0,
  "class_type": "rogue",
  "race": "elf",
  "gender": "female",
  "created_at": "2025-01-29T10:30:00Z",
  "last_played_at": "2025-01-29T10:30:00Z",
  "total_play_time": 0
}
```

**Error Response (Name Taken):**
```json
{
  "error": {
    "code": "CHARACTER_NAME_TAKEN",
    "message": "Character name is already taken",
    "details": {
      "name": "Shadowblade",
      "suggestions": ["Shadowblade123", "ShadowbladeX", "Shadowblade_01"]
    }
  },
  "timestamp": "2025-01-29T10:30:00Z",
  "request_id": "req_123456789"
}
```

### Example 2: Selecting Character for Gameplay

**Request:**
```bash
curl -X POST https://api.mmorpg.com/api/v1/characters/550e8400-e29b-41d4-a716-446655440000/select \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "success": true,
  "character_id": "550e8400-e29b-41d4-a716-446655440000",
  "world_server": "realm_01.game.mmorpg.com",
  "world_port": 7777,
  "session_token": "eyJhbGciOiJIUzI1NiIs...",
  "spawn_location": {
    "world_id": "realm_01",
    "zone_id": "elven_forest",
    "map_id": "starting_area",
    "position": {
      "x": 1250.5,
      "y": 2048.0,
      "z": 156.75
    }
  }
}
```

### Example 3: Checking Name Availability

**Request:**
```bash
curl -X GET "https://api.mmorpg.com/api/v1/characters/check-name?name=Gandalf" \
  -H "Content-Type: application/json"
```

**Response (Available):**
```json
{
  "available": true,
  "name": "Gandalf",
  "suggestions": []
}
```

**Response (Taken):**
```json
{
  "available": false,
  "name": "Gandalf",
  "suggestions": [
    "GandalfTheGrey",
    "Gandalf2025",
    "Gandalf_01"
  ]
}
```

### Example 4: Allocating Stat Points

**Request:**
```bash
curl -X POST https://api.mmorpg.com/api/v1/characters/550e8400-e29b-41d4-a716-446655440000/stats/allocate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "stat": "strength",
    "points": 3
  }'
```

**Response:**
```json
{
  "success": true,
  "stat": "strength",
  "new_value": 28,
  "points_remaining": 2,
  "affected_stats": {
    "attack_power": 194,
    "carrying_capacity": 280,
    "melee_damage_bonus": 14
  }
}
```

## Implementation Notes

### Performance Considerations

1. **Caching Strategy**
   - Cache character lists per user (5 min TTL)
   - Cache character details (1 min TTL)
   - Cache name availability checks (30 sec TTL)
   - Invalidate on updates

2. **Database Optimization**
   - Index on user_id + deleted_at for list queries
   - Index on name (case-insensitive) for uniqueness
   - Composite index on user_id + slot_number
   - Partition by user_id for large scale

3. **Batch Operations**
   - Support batch stat allocation
   - Batch appearance updates
   - Bulk character data fetch for UI

### Security Considerations

1. **Input Validation**
   - Sanitize character names
   - Validate appearance ranges
   - Check slot boundaries
   - Verify ownership on all operations

2. **Rate Limiting**
   - Per-user limits
   - IP-based limits for name checks
   - Exponential backoff for repeated failures

3. **Audit Logging**
   - Log all character creations
   - Log deletions with reason
   - Track name changes
   - Monitor suspicious patterns

### Future Enhancements

1. **Character Transfers**
   - Server transfers
   - Account transfers
   - Name change service

2. **Advanced Features**
   - Character templates
   - Appearance presets
   - Stat builds library
   - Character comparison

3. **Social Features**
   - Character profiles
   - Achievement showcase
   - Leaderboards integration
   - Friend character viewing