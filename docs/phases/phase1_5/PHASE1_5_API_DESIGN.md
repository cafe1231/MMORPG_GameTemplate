# 🔌 Phase 1.5: Character System - API Design Specification

## Overview

This document defines the RESTful API endpoints, request/response formats, error handling, and integration patterns for the character management system.

---

## API Endpoints

### Base URL
- Development: `http://localhost:8090/api/v1`
- Production: `https://api.mmorpg.com/api/v1`

### Authentication
All endpoints require JWT authentication via Bearer token:
```
Authorization: Bearer <jwt_access_token>
```

---

## Character Management Endpoints

### 1. Create Character
Creates a new character for the authenticated user.

**Endpoint**: `POST /characters`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
Content-Type: application/json
```

**Request Body**:
```json
{
    "name": "Gandalf",
    "class": 2,
    "race": 1,
    "gender": 1,
    "appearance": {
        "hair_style": 3,
        "hair_color": "#FFFFFF",
        "face_type": 2,
        "skin_tone": "#F5DEB3",
        "body_type": 1,
        "height": 1.0,
        "eye_color": "#0080FF",
        "facial_hair": 2,
        "scars": [1, 3],
        "tattoos": []
    },
    "starting_zone": "human_starting_area"
}
```

**Success Response** (201 Created):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "class": 2,
        "race": 1,
        "gender": 1,
        "level": 1,
        "experience": 0,
        "appearance": {
            "hair_style": 3,
            "hair_color": "#FFFFFF",
            "face_type": 2,
            "skin_tone": "#F5DEB3",
            "body_type": 1,
            "height": 1.0,
            "eye_color": "#0080FF",
            "facial_hair": 2,
            "scars": [1, 3],
            "tattoos": []
        },
        "stats": {
            "health": 100,
            "max_health": 100,
            "mana": 150,
            "max_mana": 150,
            "stamina": 100,
            "max_stamina": 100,
            "strength": 8,
            "intelligence": 15,
            "dexterity": 10,
            "vitality": 10,
            "wisdom": 12,
            "charisma": 10
        },
        "position": {
            "zone": "human_starting_area",
            "x": 100.0,
            "y": 200.0,
            "z": 50.0,
            "rotation": 0.0,
            "map_layer": 0
        },
        "created_at": "2025-07-29T10:00:00Z",
        "is_selected": false
    }
}
```

**Error Responses**:

400 Bad Request - Validation Error:
```json
{
    "success": false,
    "error_code": "VALIDATION_ERROR",
    "error_message": "Invalid character data",
    "errors": {
        "name": "Character name must be between 3 and 32 characters",
        "class": "Invalid character class"
    }
}
```

409 Conflict - Name Taken:
```json
{
    "success": false,
    "error_code": "CHARACTER_NAME_TAKEN",
    "error_message": "Character name 'Gandalf' is already taken",
    "suggestions": ["Gandalf123", "GandalfTheGrey", "Gandalf_01"]
}
```

403 Forbidden - Character Limit:
```json
{
    "success": false,
    "error_code": "CHARACTER_LIMIT_REACHED",
    "error_message": "Maximum character limit (5) reached",
    "character_slots": {
        "used": 5,
        "max": 5,
        "is_premium": false
    }
}
```

### 2. List Characters
Retrieves all characters for the authenticated user.

**Endpoint**: `GET /characters`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
```

**Query Parameters**:
- `include_deleted` (boolean, optional): Include soft-deleted characters
- `sort` (string, optional): Sort order (last_played, level, name)
- `order` (string, optional): asc or desc

**Success Response** (200 OK):
```json
{
    "success": true,
    "characters": [
        {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "name": "Gandalf",
            "class": 2,
            "race": 1,
            "gender": 1,
            "level": 60,
            "zone": "eastern_kingdoms",
            "last_played_at": "2025-07-28T15:30:00Z",
            "is_selected": true,
            "thumbnail_url": "https://cdn.mmorpg.com/characters/123e4567.jpg"
        },
        {
            "id": "456e7890-e89b-12d3-a456-426614174001",
            "name": "Legolas",
            "class": 3,
            "race": 2,
            "gender": 1,
            "level": 45,
            "zone": "kalimdor",
            "last_played_at": "2025-07-27T10:15:00Z",
            "is_selected": false,
            "thumbnail_url": "https://cdn.mmorpg.com/characters/456e7890.jpg"
        }
    ],
    "character_slots": {
        "used": 2,
        "max": 5,
        "is_premium": false
    }
}
```

### 3. Get Character Details
Retrieves detailed information about a specific character.

**Endpoint**: `GET /characters/{character_id}`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "class": 2,
        "race": 1,
        "gender": 1,
        "level": 60,
        "experience": 5234500,
        "appearance": {
            "hair_style": 3,
            "hair_color": "#FFFFFF",
            "face_type": 2,
            "skin_tone": "#F5DEB3",
            "body_type": 1,
            "height": 1.0,
            "eye_color": "#0080FF",
            "facial_hair": 2,
            "scars": [1, 3],
            "tattoos": [],
            "custom_features": {
                "beard_length": 0.8,
                "voice_pitch": 0.3
            }
        },
        "stats": {
            "health": 3500,
            "max_health": 3500,
            "mana": 2800,
            "max_mana": 2800,
            "stamina": 100,
            "max_stamina": 100,
            "strength": 25,
            "intelligence": 180,
            "dexterity": 45,
            "vitality": 60,
            "wisdom": 150,
            "charisma": 80,
            "attack_power": 50,
            "spell_power": 450,
            "defense": 120,
            "crit_chance": 0.25,
            "crit_damage": 2.0,
            "attack_speed": 1.2,
            "move_speed": 1.0,
            "physical_resist": 15,
            "magical_resist": 35,
            "fire_resist": 20,
            "ice_resist": 25,
            "lightning_resist": 15,
            "class_stats": {
                "mana_regen": 50,
                "spell_penetration": 15
            }
        },
        "position": {
            "zone": "eastern_kingdoms",
            "x": 1234.56,
            "y": 2345.67,
            "z": 123.45,
            "rotation": 1.57,
            "map_layer": 0
        },
        "equipment": {
            "head": "item_123",
            "shoulders": "item_456",
            "chest": "item_789",
            "main_hand": "item_abc",
            "off_hand": null
        },
        "inventory": {
            "slots": 80,
            "used_slots": 45,
            "gold": 12500
        },
        "created_at": "2025-01-15T10:00:00Z",
        "updated_at": "2025-07-28T15:30:00Z",
        "last_played_at": "2025-07-28T15:30:00Z",
        "is_selected": true
    }
}
```

**Error Response** (404 Not Found):
```json
{
    "success": false,
    "error_code": "CHARACTER_NOT_FOUND",
    "error_message": "Character not found"
}
```

### 4. Update Character
Updates character appearance or custom features.

**Endpoint**: `PUT /characters/{character_id}`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
Content-Type: application/json
```

**Request Body**:
```json
{
    "appearance": {
        "hair_style": 5,
        "hair_color": "#808080",
        "custom_features": {
            "beard_length": 1.0
        }
    }
}
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "appearance": {
            "hair_style": 5,
            "hair_color": "#808080",
            "custom_features": {
                "beard_length": 1.0,
                "voice_pitch": 0.3
            }
        },
        "updated_at": "2025-07-29T11:00:00Z"
    }
}
```

### 5. Delete Character
Soft deletes a character (can be recovered within 30 days).

**Endpoint**: `DELETE /characters/{character_id}`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "message": "Character deleted successfully",
    "recovery_token": "abc123def456ghi789",
    "recovery_expires_at": "2025-08-28T10:00:00Z"
}
```

### 6. Select Character
Selects a character for gameplay.

**Endpoint**: `POST /characters/{character_id}/select`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "level": 60,
        "zone": "eastern_kingdoms",
        "is_selected": true
    },
    "session": {
        "session_id": "sess_123456",
        "world_server": "world-1.mmorpg.com",
        "world_port": 7777
    }
}
```

### 7. Get Selected Character
Retrieves the currently selected character.

**Endpoint**: `GET /characters/selected`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "level": 60,
        "class": 2,
        "race": 1,
        "zone": "eastern_kingdoms"
    }
}
```

**Error Response** (404 Not Found):
```json
{
    "success": false,
    "error_code": "NO_CHARACTER_SELECTED",
    "error_message": "No character selected"
}
```

### 8. Validate Character Name
Checks if a character name is available.

**Endpoint**: `POST /characters/validate-name`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
Content-Type: application/json
```

**Request Body**:
```json
{
    "name": "Gandalf"
}
```

**Success Response** (200 OK) - Available:
```json
{
    "available": true,
    "name": "Gandalf"
}
```

**Success Response** (200 OK) - Not Available:
```json
{
    "available": false,
    "reason": "Name already taken",
    "suggestions": ["Gandalf123", "GandalfTheGrey", "Gandalf_01"]
}
```

### 9. Recover Deleted Character
Recovers a soft-deleted character.

**Endpoint**: `POST /characters/recover`

**Request Headers**:
```
Authorization: Bearer <jwt_access_token>
Content-Type: application/json
```

**Request Body**:
```json
{
    "recovery_token": "abc123def456ghi789"
}
```

**Success Response** (200 OK):
```json
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Gandalf",
        "level": 60,
        "recovered_at": "2025-07-29T12:00:00Z"
    }
}
```

---

## Error Codes

### Character-Specific Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `CHARACTER_NAME_TAKEN` | 409 | Character name already exists |
| `CHARACTER_LIMIT_REACHED` | 403 | User has reached max character limit |
| `INVALID_CHARACTER_NAME` | 400 | Name contains invalid characters or length |
| `CHARACTER_NOT_FOUND` | 404 | Character doesn't exist |
| `CHARACTER_NOT_OWNED` | 403 | Character belongs to another user |
| `INVALID_CLASS_RACE_COMBO` | 400 | Invalid class/race combination |
| `CHARACTER_ALREADY_SELECTED` | 409 | Character is already selected |
| `NO_CHARACTER_SELECTED` | 404 | No character currently selected |
| `CHARACTER_DELETED` | 410 | Character has been deleted |
| `INVALID_RECOVERY_TOKEN` | 400 | Recovery token is invalid or expired |
| `CHARACTER_IN_USE` | 409 | Character is currently in game |

### Generic Error Response Format

```json
{
    "success": false,
    "error_code": "ERROR_CODE",
    "error_message": "Human-readable error message",
    "details": {
        "field": "Additional error context"
    },
    "request_id": "req_123456789"
}
```

---

## Rate Limiting

### Endpoint-Specific Limits

| Endpoint | Rate Limit | Window |
|----------|------------|---------|
| Create Character | 5 requests | 1 hour |
| List Characters | 60 requests | 1 minute |
| Get Character | 120 requests | 1 minute |
| Update Character | 30 requests | 1 minute |
| Delete Character | 3 requests | 1 hour |
| Validate Name | 20 requests | 1 minute |

### Rate Limit Headers

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1627890123
```

### Rate Limit Exceeded Response (429):
```json
{
    "success": false,
    "error_code": "RATE_LIMIT_EXCEEDED",
    "error_message": "Too many requests",
    "retry_after": 30
}
```

---

## Webhook Events

The character service can send webhooks for significant events:

### Character Created
```json
{
    "event": "character.created",
    "timestamp": "2025-07-29T10:00:00Z",
    "data": {
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "character_id": "123e4567-e89b-12d3-a456-426614174000",
        "character_name": "Gandalf",
        "class": 2,
        "race": 1,
        "level": 1
    }
}
```

### Character Deleted
```json
{
    "event": "character.deleted",
    "timestamp": "2025-07-29T11:00:00Z",
    "data": {
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "character_id": "123e4567-e89b-12d3-a456-426614174000",
        "character_name": "Gandalf",
        "permanent_delete_at": "2025-08-28T11:00:00Z"
    }
}
```

---

## Integration Examples

### Unreal Engine C++ HTTP Request

```cpp
void UMMORPGCharacterSubsystem::CreateCharacter(const FCharacterCreateRequest& Request)
{
    // Get auth token
    FString AuthToken = GetAuthSubsystem()->GetAccessToken();
    
    // Create HTTP request
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = 
        FHttpModule::Get().CreateRequest();
    
    HttpRequest->SetURL(TEXT("http://localhost:8090/api/v1/characters"));
    HttpRequest->SetVerb(TEXT("POST"));
    HttpRequest->SetHeader(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthToken));
    HttpRequest->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
    
    // Serialize request
    FString JsonString;
    FJsonObjectConverter::UStructToJsonObjectString(Request, JsonString);
    HttpRequest->SetContentAsString(JsonString);
    
    // Bind response handler
    HttpRequest->OnProcessRequestComplete().BindUObject(
        this, &UMMORPGCharacterSubsystem::HandleCreateCharacterResponse);
    
    // Send request
    HttpRequest->ProcessRequest();
}
```

### Go Backend Handler

```go
func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from JWT
    userID, err := auth.GetUserIDFromContext(r.Context())
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Parse request
    var req CreateCharacterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := h.validator.Struct(req); err != nil {
        renderValidationError(w, err)
        return
    }
    
    // Create character
    character, err := h.service.CreateCharacter(r.Context(), userID, req)
    if err != nil {
        switch {
        case errors.Is(err, ErrCharacterNameTaken):
            renderError(w, "CHARACTER_NAME_TAKEN", "Name already taken", http.StatusConflict)
        case errors.Is(err, ErrCharacterLimitReached):
            renderError(w, "CHARACTER_LIMIT_REACHED", "Character limit reached", http.StatusForbidden)
        default:
            renderError(w, "INTERNAL_ERROR", "Failed to create character", http.StatusInternalServerError)
        }
        return
    }
    
    // Return success response
    renderJSON(w, CharacterResponse{
        Success:   true,
        Character: character,
    }, http.StatusCreated)
}
```

---

## API Versioning

The API uses URL versioning: `/api/v1/characters`

Future versions will be available at `/api/v2/characters` with backward compatibility maintained for deprecated endpoints.

### Version Migration Headers

```
X-API-Version: 1
X-API-Deprecated: true
X-API-Sunset-Date: 2026-01-01
```

---

## Testing

### Example cURL Commands

Create Character:
```bash
curl -X POST http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gandalf",
    "class": 2,
    "race": 1,
    "gender": 1,
    "appearance": {
      "hair_style": 3,
      "hair_color": "#FFFFFF"
    }
  }'
```

List Characters:
```bash
curl -X GET http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Select Character:
```bash
curl -X POST http://localhost:8090/api/v1/characters/123e4567-e89b-12d3-a456-426614174000/select \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

This API design provides a comprehensive interface for character management while maintaining consistency with the existing authentication system and preparing for future networking features.