# Authentication API

## Overview

The Authentication API provides secure authentication and session management for AWO ERP. It implements JWT-based authentication with refresh token support for secure, stateless authentication across all system components.

## Key Features

- **JWT Token Authentication**: Industry-standard JSON Web Tokens
- **Refresh Token Support**: Long-term session management with automatic renewal
- **Multi-tenant Authentication**: Tenant-scoped authentication and authorization
- **Session Management**: Login, logout, and token validation endpoints
- **Security Headers**: CSRF protection and secure cookie handling

## Endpoints Overview

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/v1/auth/login` | Authenticate user and obtain JWT tokens |
| POST | `/api/v1/auth/refresh` | Refresh expired access token |
| POST | `/api/v1/auth/logout` | Invalidate current session |
| POST | `/api/v1/auth/validate` | Validate token and get user context |

## Quick Start

### User Login
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "developer@company.com",
    "password": "your-secure-password",
    "tenant_id": "your-tenant-uuid"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "user-uuid",
    "username": "developer@company.com",
    "email": "developer@company.com",
    "roles": ["developer", "api_user"],
    "permissions": ["read:users", "write:entities"],
    "tenant_id": "tenant-uuid"
  }
}
```

### Token Refresh
```bash
curl -X POST "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Token Validation
```bash
curl -X POST "http://localhost:8080/api/v1/auth/validate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

### User Logout
```bash
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

## Authentication Flow

### Standard Login Flow
1. **Submit Credentials**: POST to `/auth/login` with username, password, and tenant_id
2. **Receive Tokens**: Get access_token (short-lived) and refresh_token (long-lived)
3. **Use Access Token**: Include `Authorization: Bearer <token>` header in API requests
4. **Handle Expiry**: Use refresh_token to get new access_token when expired

### Token Refresh Flow
1. **Monitor Expiry**: Track access_token expiration time
2. **Proactive Refresh**: Refresh tokens 1-2 minutes before expiry
3. **Handle Refresh Failures**: Re-authenticate if refresh fails
4. **Update Headers**: Use new access_token in subsequent requests

## Security Features

### Multi-Factor Authentication (MFA)
When MFA is enabled, the login response includes additional requirements:

```json
{
  "mfa_required": true,
  "mfa_methods": ["totp", "sms"],
  "temp_token": "temp-mfa-token-123",
  "expires_in": 300
}
```

Complete MFA with:
```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_token": "temp-mfa-token-123",
    "method": "totp",
    "code": "123456"
  }'
```

### Session Management
- **Access Tokens**: 15-minute expiry for security
- **Refresh Tokens**: 30-day expiry for user convenience
- **Automatic Cleanup**: Expired tokens are automatically purged
- **Session Tracking**: All active sessions are tracked per user

## Error Handling

### Common Authentication Errors

#### Invalid Credentials
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Username or password is incorrect",
    "details": {"field": "credentials"}
  }
}
```

#### Token Expired
```json
{
  "error": {
    "code": "TOKEN_EXPIRED", 
    "message": "Access token has expired",
    "details": {"expires_at": "2025-01-01T12:15:00Z"}
  }
}
```

#### Invalid Tenant
```json
{
  "error": {
    "code": "INVALID_TENANT",
    "message": "User does not have access to specified tenant",
    "details": {"tenant_id": "tenant-uuid"}
  }
}
```

#### Rate Limit Exceeded
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many authentication attempts",
    "details": {"retry_after": 300}
  }
}
```

## Best Practices

### Token Storage
- **Secure Storage**: Store refresh tokens in httpOnly cookies or secure storage
- **Memory-Only Access Tokens**: Keep access tokens in memory, not persistent storage
- **Cross-Site Protection**: Use SameSite cookie attributes for web applications

### Error Recovery
- **Automatic Retry**: Implement automatic token refresh on 401 responses
- **Graceful Degradation**: Handle authentication failures gracefully
- **User Feedback**: Provide clear feedback for authentication issues

### Security Considerations
- **HTTPS Only**: Always use HTTPS in production
- **Token Rotation**: Implement token rotation for security
- **Audit Logging**: Log all authentication events for security monitoring
- **IP Restrictions**: Consider IP-based access restrictions for sensitive accounts

## Integration Examples

### JavaScript/Node.js Client
```javascript
class AuthClient {
  constructor(baseUrl, tenantId) {
    this.baseUrl = baseUrl;
    this.tenantId = tenantId;
    this.accessToken = null;
    this.refreshToken = null;
    this.tokenExpiry = null;
  }

  async login(username, password) {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        username,
        password, 
        tenant_id: this.tenantId
      })
    });

    const data = await response.json();
    
    if (response.ok) {
      this.accessToken = data.access_token;
      this.refreshToken = data.refresh_token;
      this.tokenExpiry = Date.now() + (data.expires_in * 1000);
      return data;
    }
    
    throw new Error(`Login failed: ${data.error?.message}`);
  }

  async refreshAccessToken() {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({refresh_token: this.refreshToken})
    });

    const data = await response.json();
    
    if (response.ok) {
      this.accessToken = data.access_token;
      this.tokenExpiry = Date.now() + (data.expires_in * 1000);
      return data;
    }
    
    throw new Error(`Token refresh failed: ${data.error?.message}`);
  }

  async logout() {
    if (!this.accessToken) return;
    
    await fetch(`${this.baseUrl}/api/v1/auth/logout`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.accessToken}`,
        'Content-Type': 'application/json'
      }
    });

    this.accessToken = null;
    this.refreshToken = null; 
    this.tokenExpiry = null;
  }

  isTokenExpired() {
    return !this.accessToken || Date.now() >= this.tokenExpiry - 60000;
  }

  getAuthHeader() {
    return this.accessToken ? `Bearer ${this.accessToken}` : null;
  }
}
```

### Python Client
```python
import requests
from datetime import datetime, timedelta
import jwt

class AuthClient:
    def __init__(self, base_url, tenant_id):
        self.base_url = base_url
        self.tenant_id = tenant_id
        self.access_token = None
        self.refresh_token = None
        self.token_expiry = None
        self.user_info = None
    
    def login(self, username, password):
        response = requests.post(
            f"{self.base_url}/api/v1/auth/login",
            json={
                'username': username,
                'password': password,
                'tenant_id': self.tenant_id
            }
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data['access_token']
        self.refresh_token = data['refresh_token']
        self.token_expiry = datetime.now() + timedelta(seconds=data['expires_in'])
        self.user_info = data['user']
        
        return data
    
    def refresh_access_token(self):
        response = requests.post(
            f"{self.base_url}/api/v1/auth/refresh",
            json={'refresh_token': self.refresh_token}
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data['access_token']
        self.token_expiry = datetime.now() + timedelta(seconds=data['expires_in'])
        
        return data
    
    def logout(self):
        if self.access_token:
            requests.post(
                f"{self.base_url}/api/v1/auth/logout",
                headers={'Authorization': f"Bearer {self.access_token}"}
            )
        
        self.access_token = None
        self.refresh_token = None
        self.token_expiry = None
        self.user_info = None
    
    def is_token_expired(self):
        return (not self.access_token or 
                datetime.now() >= self.token_expiry - timedelta(minutes=1))
    
    def get_auth_headers(self):
        if self.is_token_expired():
            self.refresh_access_token()
        
        return {'Authorization': f"Bearer {self.access_token}"}
```

## Monitoring & Analytics

### Authentication Metrics
- **Login Success Rate**: Track successful vs failed login attempts
- **Token Usage**: Monitor access token usage patterns
- **Session Duration**: Track average session lengths
- **MFA Adoption**: Monitor MFA usage across users

### Security Monitoring
- **Failed Login Attempts**: Track brute force attempts
- **Anomalous Access**: Detect unusual login patterns
- **Token Abuse**: Monitor for token reuse or sharing
- **Geographic Patterns**: Track login locations for anomaly detection

---

**Next**: [Users API](../users/index.md) | **Up**: [API Reference](../index.md)