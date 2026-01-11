# IAM Service (Identity and Access Management)

The IAM Service is a core security component of the CSIC Platform responsible for managing user identities, authentication, authorization, and access control.

## Features

### Authentication
- **JWT-based Authentication**: Secure token-based authentication with access and refresh tokens
- **Password Management**: Secure password hashing using bcrypt
- **Multi-Factor Authentication (MFA)**: TOTP-based MFA support
- **Session Management**: Session tracking and revocation
- **Account Lockout**: Automatic lockout after failed login attempts

### User Management
- **User Registration**: Self-service user registration
- **User Profile Management**: Update user details and preferences
- **Password Operations**: Change and reset passwords
- **Account Status**: Enable/disable user accounts

### Role-Based Access Control
- **Role Management**: Create and manage roles with permissions
- **Permission Assignment**: Assign permissions to roles
- **Role Hierarchy**: Support for role inheritance
- **Default Roles**: Admin, Analyst, Auditor, User roles

### Security Features
- **Password Policies**: Enforce password complexity requirements
- **Audit Logging**: Log all authentication and authorization events
- **API Key Management**: Generate and manage API keys for programmatic access
- **Session Security**: Track sessions with IP and user agent validation

## Architecture

This service follows Clean Architecture principles:

```
services/security/iam/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entity.go          # Domain entities (User, Role, Session)
│   │   │   └── errors.go          # Domain error definitions
│   │   ├── ports/
│   │   │   ├── repository.go      # Repository interface definitions
│   │   │   ├── service.go         # Service interface definitions
│   │   │   └── infrastructure.go  # Infrastructure interface definitions
│   │   └── service/
│   │       ├── authentication_service.go  # Authentication logic
│   │       └── user_service.go            # User management logic
│   ├── adapter/
│   │   └── infrastructure/
│   │       └── jwt.go             # JWT token implementation
│   └── handler/
│       └── http_handler.go        # HTTP API handlers
├── migrations/
│   └── 001_create_tables.sql      # Database schema
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout (revoke session) |
| POST | `/api/v1/auth/logout-all` | Logout from all devices |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user profile |
| PUT | `/api/v1/users/me/password` | Change password |
| POST | `/api/v1/users` | Create user (admin) |
| GET | `/api/v1/users` | List all users |
| GET | `/api/v1/users/{id}` | Get user by ID |
| PUT | `/api/v1/users/{id}` | Update user |
| DELETE | `/api/v1/users/{id}` | Delete user |
| POST | `/api/v1/users/{id}/enable` | Enable user |
| POST | `/api/v1/users/{id}/disable` | Disable user |
| POST | `/api/v1/users/{id}/reset-password` | Reset password (admin) |

### Roles

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/roles` | List all roles |
| POST | `/api/v1/roles` | Create role |
| GET | `/api/v1/roles/{id}` | Get role by ID |
| PUT | `/api/v1/roles/{id}` | Update role |
| DELETE | `/api/v1/roles/{id}` | Delete role |

### Sessions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sessions` | Get current user sessions |
| DELETE | `/api/v1/sessions/{id}` | Revoke specific session |
| DELETE | `/api/v1/sessions` | Revoke all sessions |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

## Authentication Flow

### Login
```bash
curl -X POST http://localhost:8083/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "securepassword"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "role": "Admin"
  }
}
```

### Using the Access Token
```bash
curl -X GET http://localhost:8083/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

### Refreshing the Token
```bash
curl -X POST http://localhost:8083/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

## Default Roles and Permissions

### Admin
- Full system access
- Can manage users, roles, and permissions
- Access to all modules and features

### Analyst
- Read access to compliance data
- Can view reports and analytics
- Cannot modify system configuration

### Auditor
- Read-only access
- Can view all audit logs
- Cannot modify any data

### User
- Basic authenticated access
- Can view their own profile
- Limited to own data

## Security Configuration

### Password Requirements
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

### Session Configuration
- Access token expiry: 15 minutes
- Refresh token expiry: 7 days
- Maximum login attempts: 5
- Lockout duration: 30 minutes

### JWT Claims
```json
{
  "sub": "user_id",
  "username": "admin",
  "email": "admin@example.com",
  "role": "Admin",
  "permissions": ["users:read", "users:write"],
  "iat": 1234567890,
  "exp": 1234568790,
  "type": "access"
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_HOST` | HTTP server host | `0.0.0.0` |
| `APP_PORT` | HTTP server port | `8083` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `csic` |
| `DB_PASSWORD` | PostgreSQL password | `csic_secret` |
| `DB_NAME` | PostgreSQL database | `csic_platform` |
| `JWT_ACCESS_SECRET` | JWT access token secret | Required |
| `JWT_REFRESH_SECRET` | JWT refresh token secret | Required |
| `ACCESS_EXPIRY` | Access token expiry | `15m` |
| `REFRESH_EXPIRY` | Refresh token expiry | `168h` |
| `LOG_LEVEL` | Logging level | `info` |

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    api_key VARCHAR(255),
    api_key_enabled BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP,
    login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Roles Table
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Permissions Table
```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description VARCHAR(255)
);
```

### Sessions Table
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP
);
```

## Installation

### Prerequisites
- Go 1.21+
- PostgreSQL 14+

### Building
```bash
cd services/security/iam
go build -o bin/iam ./cmd
```

### Running
```bash
./bin/iam
```

### Database Migration
```bash
psql -h localhost -U csic -d csic_platform -f migrations/001_create_tables.sql
```

## Integration with Other Services

### Token Validation Middleware
```go
func AuthMiddleware(authService ports.AuthenticationService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            authContext, err := authService.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            // Add auth context to request
            ctx := context.WithValue(r.Context(), "auth", authContext)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Testing

```bash
# Unit tests
go test ./... -v

# Integration tests
go test ./... -tags=integration -v

# Load testing
hey -n 1000 -c 10 -m POST -d '{"username":"test","password":"test"}' http://localhost:8083/api/v1/auth/login
```

## Monitoring

The service exposes the following metrics:
- `iam_login_attempts_total`: Total login attempts by status
- `iam_users_created_total`: Total users created
- `iam_sessions_created_total`: Total sessions created
- `iam_sessions_revoked_total`: Total sessions revoked
- `iam_auth_latency_seconds`: Authentication latency

## License

This project is part of the CSIC Platform and is licensed under the MIT License.
