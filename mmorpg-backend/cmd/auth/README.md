# Auth Service

The authentication service handles user registration, login, session management, and JWT token generation for the MMORPG template.

## Features

- User registration with email and username
- JWT-based authentication (access + refresh tokens)
- Session management with Redis caching
- Rate limiting for login attempts
- Password hashing with bcrypt
- Account status management
- NATS integration for service communication

## Running the Service

### Local Development

```bash
# Start infrastructure services
docker-compose up -d

# Run the auth service
go run cmd/auth/main.go
```

### Docker Development

```bash
# Start all services with hot reload
docker-compose -f docker-compose.dev.yml up auth
```

## Configuration

The service uses environment variables for configuration:

- `MMORPG_AUTH_PORT` - Service port (default: 8081)
- `MMORPG_DATABASE_URL` - PostgreSQL connection string
- `MMORPG_REDIS_URL` - Redis connection string
- `MMORPG_NATS_URL` - NATS connection string
- `MMORPG_AUTH_JWTACCESSSECRET` - JWT access token secret
- `MMORPG_AUTH_JWTREFRESHSECRET` - JWT refresh token secret

## API Endpoints

All endpoints are proxied through the gateway at `http://localhost:8080/api/v1/auth/`

### Register
```
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "StrongPass123!",
  "username": "player123",
  "accept_terms": true
}
```

### Login
```
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "StrongPass123!"
}
```

### Logout
```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
{
  "logout_all_devices": false
}
```

### Refresh Token
```
POST /api/v1/auth/refresh
{
  "refresh_token": "<refresh_token>"
}
```

### Verify Token
```
GET /api/v1/auth/verify
Authorization: Bearer <access_token>
```

## Testing

```bash
# Run unit tests
go test ./internal/domain/auth/...

# Run integration tests
go test ./internal/adapters/auth/...
```

## Security Considerations

- Passwords must be at least 8 characters with 3 of: uppercase, lowercase, numbers, special characters
- Login rate limiting: 5 attempts per 15 minutes per IP
- JWT access tokens expire in 15 minutes
- JWT refresh tokens expire in 7 days
- Session limits: 10 concurrent sessions per user

## NATS Events

The auth service publishes/subscribes to:

- `auth.validate` - Token validation requests
- `auth.user.get` - User info requests
- `auth.session.created` - New session events
- `auth.session.destroyed` - Logout events