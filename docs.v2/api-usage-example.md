# User API Usage Examples

This document provides examples of how to use the User API endpoints following the data flow pattern.

## API Endpoints

### Authentication

#### Authenticate User
```bash
POST /api/v1/users/auth
Content-Type: application/json

{
  "identifier": "john.doe@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "user": {
    "id": "uuid",
    "entity_id": "uuid",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "user_type": "INTERNAL",
    "is_active": true,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "token": "jwt-token-here",
  "expires_at": "2023-01-01T08:00:00Z"
}
```

### User Management

#### Create User
```bash
POST /api/v1/users
Content-Type: application/json

{
  "entity_id": "uuid",
  "username": "jane.doe",
  "email": "jane.doe@example.com",
  "password": "password123",
  "user_type": "INTERNAL",
  "account_status": "ACTIVE",
  "mfa_enabled": false
}
```

#### Get User
```bash
GET /api/v1/users/{user_id}
```

#### List Users
```bash
GET /api/v1/users?limit=20&offset=0&user_type=INTERNAL&account_status=ACTIVE
```

#### Update User
```bash
PUT /api/v1/users/{user_id}
Content-Type: application/json

{
  "username": "new.username",
  "email": "new.email@example.com",
  "account_status": "ACTIVE",
  "mfa_enabled": true
}
```

#### Delete User (Soft Delete)
```bash
DELETE /api/v1/users/{user_id}
```

#### Update Password
```bash
PUT /api/v1/users/{user_id}/password
Content-Type: application/json

{
  "current_password": "oldpassword123",
  "new_password": "newpassword123"
}
```

#### Search Users
```bash
GET /api/v1/users/search?q=john&limit=10&offset=0
```

#### Get User Roles
```bash
GET /api/v1/users/{user_id}/roles
```

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

### 400 Bad Request
```json
{
  "error": "Invalid request payload"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid credentials"
}
```

### 404 Not Found
```json
{
  "error": "User not found"
}
```

### 409 Conflict
```json
{
  "error": "Email already exists"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

## Integration Example

Here's how to wire up the complete user API in your main application:

```go
package main

import (
    "github.com/niiniyare/erp/internal/api/handlers"
    "github.com/niiniyare/erp/internal/core/user"
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/metrics"
    "github.com/niiniyare/erp/internal/shared/tracing"
    "github.com/niiniyare/erp/db/sqlc"
)

func main() {
    // Initialize dependencies
    store := sqlc.NewStore(db) // your database connection
    tracingService := tracing.NewTracingService(tracingConfig)
    metricsService := metrics.NewMetricsService(metricsConfig)
    cacheService := cache.NewRedisService(redisConfig)
    
    // Create user repository and service
    userRepo := user.NewRepository(store, tracingService, metricsService)
    userService := user.NewService(userRepo, cacheService, tracingService, metricsService)
    
    // Create router with all services
    router := handlers.NewRouter(tenantService, entityService, userService, tracingService, metricsService)
    
    // Start server
    router.Run(":8080")
}
```

## Observability Features

The User API includes  observability:

### Distributed Tracing
- Every request creates a root span
- Service and repository operations create child spans
- Full request correlation across layers
- OpenTelemetry compatible

### Metrics Collection
- HTTP request metrics (duration, count, status codes)
- Business operation metrics (user creation, authentication)
- Database operation metrics (query duration, error rates)
- Cache operation metrics (hit/miss rates)

### Structured Logging
- Context-aware logging throughout the request flow
- Consistent log fields and formats
- Error correlation with trace IDs
- Performance tracking

### Caching Strategy
- Multi-key caching (ID, email, username)
- TTL-based expiration
- Cache invalidation on updates
- Cache miss handling with fallback

This API implementation follows the Clean Architecture data flow pattern with proper separation of concerns,  observability, and production-ready error handling.