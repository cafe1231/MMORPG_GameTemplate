# Phase 1.5 Character Service Testing Guide

## Overview

This guide provides comprehensive testing procedures for the Phase 1.5 character management system. It covers unit tests, integration tests, API testing, and manual verification procedures.

---

## Prerequisites

### Environment Setup

1. **Backend Services Running**:
```bash
cd mmorpg-backend
docker-compose -f docker-compose.dev.yml up -d
```

2. **Verify Services**:
```bash
docker-compose -f docker-compose.dev.yml ps
```

Expected output:
- PostgreSQL (5432)
- Redis (6379)
- NATS (4222)
- Auth Service (8080)
- Gateway Service (8090)
- Character Service (8082)

3. **Database Migrations**:
```bash
cd mmorpg-backend
make migrate-up
```

---

## Unit Testing

### Run All Tests

```bash
cd mmorpg-backend
make test
```

### Run Character Service Tests Only

```bash
cd mmorpg-backend/internal/application/character
go test -v ./...

cd ../../adapters/character
go test -v ./...
```

### Test Coverage

Generate coverage report:
```bash
cd mmorpg-backend
make coverage

# View HTML report
open coverage.html
```

Target coverage: >90%

### Key Test Files

1. **Service Layer Tests**:
   - `internal/application/character/service_test.go`
   - Tests business logic, validation, transactions

2. **Repository Tests**:
   - `internal/adapters/character/*_repository_postgres_test.go`
   - Tests database operations, error handling

3. **Cache Tests**:
   - `internal/adapters/character/redis/cache_test.go`
   - Tests caching logic, TTL, invalidation

4. **Event Tests**:
   - `internal/adapters/character/nats/publisher_test.go`
   - Tests event publishing, serialization

5. **Middleware Tests**:
   - `internal/adapters/character/jwt_middleware_test.go`
   - Tests authentication, token parsing

---

## Integration Testing

### Database Integration

1. **Test Database Connection**:
```bash
# Connect to test database
docker exec -it mmorpg-postgres psql -U dev -d mmorpg_test
```

2. **Verify Schema**:
```sql
-- List tables
\dt

-- Check character table structure
\d characters

-- Verify triggers
SELECT trigger_name FROM information_schema.triggers 
WHERE trigger_schema = 'public';
```

3. **Test Data Integrity**:
```sql
-- Test character creation trigger
INSERT INTO characters (user_id, name, slot, class, level) 
VALUES ('test-user', 'TestChar', 1, 'warrior', 1);

-- Verify related tables populated
SELECT * FROM character_stats WHERE character_id = (
    SELECT id FROM characters WHERE name = 'TestChar'
);
```

### Redis Integration

1. **Test Cache Operations**:
```bash
# Connect to Redis
docker exec -it mmorpg-redis redis-cli

# Check keys
KEYS character:*
KEYS user_characters:*

# Inspect cache content
GET character:{character-id}
```

2. **Test Cache Expiration**:
```bash
# Check TTL
TTL character:{character-id}
TTL user_characters:{user-id}
```

### NATS Integration

1. **Subscribe to Events**:
```bash
# Install NATS CLI if needed
go install github.com/nats-io/natscli/nats@latest

# Subscribe to character events
nats sub "character.>"
```

2. **Trigger Events**:
Create/update/delete characters via API and observe events

---

## API Testing

### Authentication Setup

1. **Get JWT Token**:
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Password123!",
    "accept_terms": true
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Password123!"
  }' | jq -r '.access_token'
```

2. **Export Token**:
```bash
export TOKEN="your-jwt-token-here"
```

### Character CRUD Operations

#### 1. Create Character

```bash
curl -X POST http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Aragorn",
    "class": "warrior",
    "appearance": {
      "face": 1,
      "hair": 3,
      "skinColor": "#FFE0BD"
    }
  }' | jq
```

Expected Response:
```json
{
  "id": "character-uuid",
  "user_id": "user-uuid",
  "name": "Aragorn",
  "class": "warrior",
  "level": 1,
  "slot": 1,
  "is_selected": false,
  "created_at": "2025-07-29T10:00:00Z",
  "updated_at": "2025-07-29T10:00:00Z"
}
```

#### 2. List Characters

```bash
curl -X GET http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected Response:
```json
{
  "characters": [
    {
      "id": "character-uuid",
      "name": "Aragorn",
      "class": "warrior",
      "level": 1,
      "slot": 1,
      "is_selected": false
    }
  ],
  "total": 1,
  "max_slots": 5
}
```

#### 3. Get Character Details

```bash
curl -X GET http://localhost:8090/api/v1/characters/{character-id} \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected Response includes full character data with stats, appearance, and position.

#### 4. Update Character

```bash
curl -X PUT http://localhost:8090/api/v1/characters/{character-id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "level": 2
  }' | jq
```

#### 5. Select Character

```bash
curl -X POST http://localhost:8090/api/v1/characters/{character-id}/select \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 6. Delete Character

```bash
curl -X DELETE http://localhost:8090/api/v1/characters/{character-id} \
  -H "Authorization: Bearer $TOKEN"
```

### Error Testing

#### 1. Duplicate Name

```bash
curl -X POST http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Aragorn",
    "class": "warrior"
  }'
```

Expected: 409 Conflict - "character name already exists"

#### 2. Invalid Class

```bash
curl -X POST http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gandalf",
    "class": "invalid-class"
  }'
```

Expected: 400 Bad Request - "invalid character class"

#### 3. Slot Limit

Create 5 characters, then try to create a 6th:
```bash
# After creating 5 characters
curl -X POST http://localhost:8090/api/v1/characters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Overflow",
    "class": "mage"
  }'
```

Expected: 400 Bad Request - "maximum character slots reached"

#### 4. Unauthorized Access

```bash
# Try to access another user's character
curl -X GET http://localhost:8090/api/v1/characters/{other-user-character-id} \
  -H "Authorization: Bearer $TOKEN"
```

Expected: 404 Not Found

---

## Performance Testing

### Load Testing with k6

1. **Install k6**:
```bash
# macOS
brew install k6

# Windows
choco install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

2. **Create Test Script** (`character-load-test.js`):
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
};

const BASE_URL = 'http://localhost:8090';
const TOKEN = __ENV.TOKEN;

export default function() {
  // List characters
  let res = http.get(`${BASE_URL}/api/v1/characters`, {
    headers: { 'Authorization': `Bearer ${TOKEN}` },
  });
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  sleep(1);
}
```

3. **Run Load Test**:
```bash
TOKEN=your-jwt-token k6 run character-load-test.js
```

### Database Performance

1. **Query Performance**:
```sql
-- Check query execution plans
EXPLAIN ANALYZE 
SELECT c.*, cs.*, ca.*, cp.*
FROM characters c
LEFT JOIN character_stats cs ON c.id = cs.character_id
LEFT JOIN character_appearance ca ON c.id = ca.character_id
LEFT JOIN character_position cp ON c.id = cp.character_id
WHERE c.user_id = 'user-id' AND c.deleted_at IS NULL;
```

2. **Index Usage**:
```sql
-- Verify indexes are being used
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;
```

---

## Manual Testing Checklist

### Character Creation Flow

- [ ] Can create character with valid data
- [ ] Name validation works (3-20 chars, alphanumeric)
- [ ] Class validation enforced
- [ ] Slot assignment correct
- [ ] Default stats initialized
- [ ] Default appearance created
- [ ] Starting position set
- [ ] Creation event published

### Character Management

- [ ] Can list all user characters
- [ ] Can view character details
- [ ] Can update character level
- [ ] Can select active character
- [ ] Only one character selected at a time
- [ ] Can soft delete character
- [ ] Deleted characters excluded from list
- [ ] Cannot exceed slot limit

### Error Handling

- [ ] Duplicate name rejected
- [ ] Invalid class rejected
- [ ] Missing required fields handled
- [ ] Unauthorized access blocked
- [ ] Database errors handled gracefully
- [ ] Cache errors don't break functionality

### Performance

- [ ] Character list loads quickly (<100ms)
- [ ] Individual character loads fast (<100ms)
- [ ] Updates complete promptly (<200ms)
- [ ] No memory leaks during extended use
- [ ] Cache improves performance

### Security

- [ ] JWT required for all endpoints
- [ ] Users isolated from each other
- [ ] SQL injection prevented
- [ ] XSS in responses prevented
- [ ] Rate limiting works (when implemented)

---

## Debugging Common Issues

### Character Not Found

1. Check if soft deleted:
```sql
SELECT * FROM characters WHERE id = 'character-id';
```

2. Verify user ownership:
```sql
SELECT user_id FROM characters WHERE id = 'character-id';
```

### Cache Inconsistency

1. Clear specific cache:
```bash
docker exec -it mmorpg-redis redis-cli
DEL character:{character-id}
DEL user_characters:{user-id}
```

2. Monitor cache operations:
```bash
docker exec -it mmorpg-redis redis-cli
MONITOR
```

### Event Not Publishing

1. Check NATS connection:
```bash
docker logs mmorpg-nats
```

2. Verify service configuration:
```bash
docker exec mmorpg-character env | grep NATS
```

### Database Connection Issues

1. Check connection pool:
```sql
SELECT count(*) FROM pg_stat_activity 
WHERE datname = 'mmorpg';
```

2. Review service logs:
```bash
docker logs mmorpg-character -f
```

---

## Test Data Management

### Create Test Data

```sql
-- Create test user
INSERT INTO users (id, email, username, password_hash) 
VALUES ('test-user-1', 'test1@example.com', 'testuser1', 'hash');

-- Create test characters
INSERT INTO characters (user_id, name, slot, class, level) 
VALUES 
  ('test-user-1', 'TestWarrior', 1, 'warrior', 10),
  ('test-user-1', 'TestMage', 2, 'mage', 5),
  ('test-user-1', 'TestRogue', 3, 'rogue', 15);
```

### Clean Test Data

```sql
-- Soft delete test characters
UPDATE characters SET deleted_at = NOW() 
WHERE user_id LIKE 'test-%';

-- Hard delete (careful!)
DELETE FROM characters WHERE user_id LIKE 'test-%';
DELETE FROM users WHERE id LIKE 'test-%';
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Character Service Tests

on:
  push:
    paths:
      - 'mmorpg-backend/internal/application/character/**'
      - 'mmorpg-backend/internal/adapters/character/**'
      - 'mmorpg-backend/internal/domain/character/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd mmorpg-backend
          make test-character
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

---

## Summary

This testing guide covers:

1. **Unit Tests**: >90% coverage of business logic
2. **Integration Tests**: Database, cache, and event system
3. **API Tests**: Full CRUD operations and error cases
4. **Performance Tests**: Load testing and optimization
5. **Manual Tests**: Comprehensive checklist
6. **Debugging**: Common issues and solutions

Regular testing ensures the character service remains reliable, performant, and secure as development continues.

---

*Last Updated: 2025-07-29*
*Version: 1.0*