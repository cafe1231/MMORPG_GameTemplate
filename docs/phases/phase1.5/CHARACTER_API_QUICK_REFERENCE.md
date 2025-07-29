# Character API Quick Reference

## Base URL
```
https://api.mmorpg.com/api/v1/characters
```

## Authentication
All endpoints require JWT token except `/check-name`:
```
Authorization: Bearer <token>
```

## Endpoints

### Character Management
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/characters` | List all characters | ✓ |
| POST | `/characters` | Create new character | ✓ |
| GET | `/characters/check-name?name={name}` | Check name availability | ✗ |
| GET | `/characters/deleted` | List deleted characters | ✓ |
| GET | `/characters/{id}` | Get character details | ✓ |
| DELETE | `/characters/{id}` | Soft delete character | ✓ |
| POST | `/characters/{id}/restore` | Restore deleted character | ✓ |
| POST | `/characters/{id}/select` | Select for gameplay | ✓ |
| DELETE | `/characters/{id}/permanent` | Permanently delete | ✓ |

### Character Customization
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/characters/{id}/appearance` | Get appearance | ✓ |
| PUT | `/characters/{id}/appearance` | Update appearance | ✓ |
| GET | `/characters/{id}/stats` | Get stats | ✓ |
| POST | `/characters/{id}/stats/allocate` | Allocate stat points | ✓ |
| GET | `/characters/{id}/position` | Get position | ✓ |
| PUT | `/characters/{id}/position` | Update position | ✓ |

## Common Request Bodies

### Create Character
```json
{
  "name": "string",
  "slot_number": 1,
  "class_type": "warrior",
  "race": "human",
  "gender": "male",
  "appearance": {
    "face_type": 1,
    "skin_color": "#F5DEB3",
    "eye_color": "#4B0082",
    "hair_style": 5,
    "hair_color": "#2F4F4F",
    "body_type": 2,
    "height": 1.0
  }
}
```

### Allocate Stat Points
```json
{
  "stat": "strength",
  "points": 3
}
```

### Permanent Delete
```json
{
  "confirmation_code": "DELETE-CharName-2025",
  "password": "account_password"
}
```

## Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `CHARACTER_NOT_FOUND` | Character doesn't exist | 404 |
| `CHARACTER_NAME_TAKEN` | Name already in use | 409 |
| `CHARACTER_LIMIT_REACHED` | Max characters reached | 403 |
| `INVALID_CHARACTER_NAME` | Name violates rules | 400 |
| `CHARACTER_DELETED` | Character is deleted | 410 |
| `CHARACTER_CANNOT_RESTORE` | Past recovery deadline | 403 |
| `CHARACTER_BELONGS_TO_OTHER` | Not your character | 403 |
| `CHARACTER_ONLINE` | Can't modify online char | 409 |
| `NO_STAT_POINTS` | No points available | 400 |
| `RATE_LIMITED` | Too many requests | 429 |

## Rate Limits

| Operation | Limit | Window |
|-----------|-------|--------|
| List/Get | 60/min | 1 minute |
| Create | 5/hour | 1 hour |
| Check Name | 30/min | 1 minute |
| Delete | 3/day | 24 hours |
| Update Appearance | 10/hour | 1 hour |
| Position Update | 100/sec | 1 second |

## Valid Enums

### Class Types
`warrior`, `mage`, `rogue`, `cleric`, `ranger`, `paladin`, `warlock`, `druid`

### Races
`human`, `elf`, `dwarf`, `orc`, `undead`, `gnome`, `troll`, `halfling`

### Gender
`male`, `female`, `other`

### Stats
`strength`, `dexterity`, `intelligence`, `wisdom`, `constitution`, `charisma`