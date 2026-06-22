---
title: "Chapter 15: Authentication and Session Management"
part: "Part III — The API Layer"
chapter: 15
section: "15-auth-sessions"
related:
  - "[Chapter 14: Multi-Tenancy Middleware](14-multitenancy-middleware.md)"
  - "[Chapter 16: Role-Based Access Control](16-rbac.md)"
  - "[Chapter 41: Redis Usage](../part-07-multitenancy/41-redis-usage.md)"
---

# Chapter 15: Authentication and Session Management

Awo uses server-side sessions for browser clients and bearer tokens for API/mobile clients. This chapter covers the rationale for this hybrid, the complete login flow, session storage in Redis, invalidation mechanics, and password management.

---

## 15.1. Session-Based Authentication Rationale

### 15.1.1. Why Session Cookies Over JWT for Browser Clients

JWTs are stateless: the server cannot invalidate a token before it expires. In an ERP system where:
- An employee is terminated and their access must be revoked immediately
- A role is changed and the user should see new permissions without re-logging in
- A security incident requires invalidating all sessions across all users

...stateless JWTs are insufficient. Awo uses server-side sessions (stored in Redis) so that access can be revoked instantly by deleting the session record.

Additional browser security benefits of `HttpOnly` cookies vs `localStorage` JWTs:
- `HttpOnly` cookies are inaccessible to JavaScript — XSS attacks cannot steal them
- `SameSite=Lax` provides CSRF protection without requiring separate CSRF tokens
- Cookies are automatically included in requests — no client-side token management code

### 15.1.2. Where JWT Is Appropriate

JWTs are appropriate for:
- **Mobile clients** — cannot use `HttpOnly` cookies; JWTs stored in secure storage
- **Inter-service calls** — short-lived tokens for service-to-service authentication
- **Webhook delivery** — signed JWT in `Authorization` header proves webhook authenticity
- **Download links** — time-limited signed URL for secure file downloads

### 15.1.3. The Session/JWT Hybrid

Browser → session cookie
Mobile → JWT (issued on login, rotated on each request, short TTL of 15 minutes)
Service → JWT (signed with shared secret, 5-minute TTL, not rotated)

---

## 15.2. Login Flow

### 15.2.1. Credential Submission and Verification

```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "jane@acme.co.ke",
  "password": "correct-horse-battery-staple"
}
```

Server:
1. Looks up user by email within the resolved tenant schema
2. Checks `locked_until` — if account is locked, returns 401 with `ACCOUNT_LOCKED`
3. Compares submitted password against stored bcrypt hash
4. If invalid: increments `failed_login_count`, locks account if threshold exceeded
5. If valid: clears `failed_login_count`, proceeds to MFA check

### 15.2.2. MFA Challenge — TOTP and Backup Codes

If the user has MFA enabled:

```json
// Intermediate response — not yet authenticated
{
  "status": "mfa_required",
  "mfa_challenge_token": "eyJ...",   // short-lived (5 min) token for the MFA step
  "mfa_methods": ["totp", "backup_code"]
}
```

Client submits the TOTP code:
```
POST /api/v1/auth/mfa-verify
{ "challenge_token": "eyJ...", "code": "123456" }
```

### 15.2.3. Session Creation — Generating a Session ID, Storing in Redis

```go
func createSession(ctx context.Context, user *User, tenant *Tenant, ip string) (*Session, error) {
    sessionID := uuid.New().String()

    sessionData := SessionData{
        UserID:    user.ID,
        TenantID:  tenant.ID,
        Roles:     user.Roles,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(cfg.SessionTTL),
        IP:        ip,
        UserAgent: extractUserAgent(ctx),
    }

    data, _ := json.Marshal(sessionData)
    key := "session:" + sessionID

    err := redis.Set(ctx, key, data, cfg.SessionTTL)
    if err != nil {
        return nil, err
    }

    // Also index by user ID for bulk invalidation
    redis.SAdd(ctx, "user_sessions:"+user.ID.String(), sessionID)
    redis.Expire(ctx, "user_sessions:"+user.ID.String(), cfg.SessionTTL+time.Hour)

    return &Session{ID: sessionID, ExpiresAt: sessionData.ExpiresAt}, nil
}
```

### 15.2.4. Session Cookie Attributes

```go
c.Cookie(&fiber.Cookie{
    Name:     "awo_session",
    Value:    session.ID,
    HTTPOnly: true,            // inaccessible to JavaScript
    Secure:   true,            // HTTPS only
    SameSite: "Lax",           // CSRF protection
    Domain:   ".awo.app",      // shared across subdomains
    Path:     "/",
    MaxAge:   int(cfg.SessionTTL.Seconds()),
})
```

`SameSite=Lax` allows cross-site navigation (clicking a link) to include the cookie, but blocks cross-site form submissions — the right balance for an ERP where users may be linked to from external systems.

---

## 15.3. Session Validation Middleware

### 15.3.1. Cookie Extraction and Session ID Validation

```go
func Session(sessionService SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sessionID := c.Cookies("awo_session")
        if sessionID == "" {
            // Check Bearer token for mobile/API clients
            sessionID = extractBearerToken(c)
        }
        if sessionID == "" {
            return c.Status(401).JSON(UnauthorizedError("no session"))
        }

        session, err := sessionService.Get(c.UserContext(), sessionID)
        if err != nil {
            if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
                clearSessionCookie(c)
                return c.Status(401).JSON(UnauthorizedError("session expired"))
            }
            return err
        }

        // Store actor in context
        c.Locals("actor_id", session.UserID)
        c.Locals("actor_roles", session.Roles)
        ctx := actor.WithContext(c.UserContext(), &Actor{
            ID:    session.UserID,
            Roles: session.Roles,
        })
        c.SetUserContext(ctx)

        // Slide the expiry window
        sessionService.Touch(c.UserContext(), sessionID)

        return c.Next()
    }
}
```

### 15.3.2. Redis Lookup — Session Data Structure

```
Key: session:{session_uuid}
Value: JSON { user_id, tenant_id, roles, created_at, expires_at, ip, user_agent }
TTL: configured session TTL (default 8 hours, slid on each request)
```

The Redis lookup is O(1). Session validation adds ~1ms to each request (single Redis GET).

### 15.3.3. Session Expiry — Absolute vs Sliding

Awo supports both modes configured per-tenant:
- **Sliding expiry** (default): TTL is extended on each request. Session stays alive as long as the user is active. Kicked out only after `idle_timeout` (default: 8 hours) of inactivity.
- **Absolute expiry**: Session has a hard maximum duration regardless of activity (e.g. 8 hours from login). Required for some compliance regimes (PCI-DSS, ISO 27001).

```go
// Sliding: Touch extends TTL
sessionService.Touch(ctx, sessionID)

// Absolute: Touch does nothing
// Configured via: tenant.Config.SessionMode = "absolute"
```

### 15.3.4. The 401 Interception Flow in amis

When a browser session expires mid-workflow:
1. The next API request returns 401
2. amis detects 401 on any API call via a global response interceptor
3. amis shows a re-authentication overlay (not a full page reload)
4. User enters credentials in the overlay
5. New session is created, the original request is retried
6. Workflow resumes where the user left off

The amis global error handler:
```javascript
// In amis configuration
fetchConfig: {
  onError: (response) => {
    if (response.status === 401) {
      showReauthOverlay()
      return true // signal that error is handled
    }
    return false
  }
}
```

---

## 15.4. Session Invalidation

### 15.4.1. Logout — Session Deletion from Redis

```go
func (s *sessionService) Logout(ctx context.Context, sessionID string) error {
    // Delete session
    if err := s.redis.Del(ctx, "session:"+sessionID); err != nil {
        return err
    }
    // Clear cookie on response
    clearSessionCookie(c)
    return nil
}
```

### 15.4.2. Forced Invalidation — All Sessions for a User

Required when: password changed, role changed, account locked, security incident.

```go
func (s *sessionService) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
    // Get all session IDs for the user
    sessionIDs, err := s.redis.SMembers(ctx, "user_sessions:"+userID.String())
    if err != nil {
        return err
    }

    // Delete all sessions in a pipeline
    pipe := s.redis.Pipeline()
    for _, sid := range sessionIDs {
        pipe.Del(ctx, "session:"+sid)
    }
    pipe.Del(ctx, "user_sessions:"+userID.String())
    return pipe.Exec(ctx)
}
```

This is called automatically by the IAM service when a user's roles are changed (see Chapter 16).

### 15.4.3. Concurrent Session Limits

Configurable per tenant: `max_concurrent_sessions` (default: 5 for ERP, 1 for high-security tenants).

When a new session is created and the user already has `max_concurrent_sessions` active sessions, the oldest session is invalidated:

```go
if sessionCount >= tenant.Config.MaxConcurrentSessions {
    oldest := getOldestSession(userSessionIDs)
    sessionService.Invalidate(ctx, oldest)
}
```

---

## 15.5. Password Management

### 15.5.1. Password Hashing — bcrypt

```go
import "golang.org/x/crypto/bcrypt"

// On registration/password change:
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// bcrypt.DefaultCost = 10; increase to 12 for high-security tenants

// On login verification:
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(submitted))
```

bcrypt's `DefaultCost` of 10 means ~100ms per hash operation — fast enough for login, slow enough to make brute-force impractical.

### 15.5.2. Password Reset Flow

```
1. POST /api/v1/auth/forgot-password { email }
   → Generate time-limited token (HMAC-signed, 1-hour TTL)
   → Store token hash in Redis
   → Send email with reset link (async via Temporal)
   → Always return 200 (do not confirm whether email exists)

2. POST /api/v1/auth/reset-password { token, new_password }
   → Validate token signature
   → Validate token not expired (Redis TTL)
   → Validate token not already used (Redis delete is atomic)
   → Update password hash
   → Invalidate all existing sessions
   → Return 200
```

### 15.5.3. Password Policy Enforcement

```go
type PasswordPolicy struct {
    MinLength        int
    RequireUppercase bool
    RequireDigit     bool
    RequireSpecial   bool
    DisallowReuse    int  // cannot reuse last N passwords
}
```

The password policy is configured per-tenant in `TenantConfig`. Previous password hashes (last N) are stored in a `user_password_history` table for reuse prevention.

### 15.5.4. Credential Stuffing Protection

```go
const (
    maxFailedAttempts = 5
    lockoutDuration   = 15 * time.Minute
)

// Track failures per user
func (s *authService) recordFailedAttempt(ctx context.Context, userID uuid.UUID) {
    key := fmt.Sprintf("login_failures:%s", userID)
    count, _ := s.redis.Incr(ctx, key)
    s.redis.Expire(ctx, key, lockoutDuration)

    if count >= maxFailedAttempts {
        // Lock the account
        s.userRepo.Update(ctx, userID, user.Update{
            LockedUntil: ptr.Time(time.Now().Add(lockoutDuration)),
        })
        // Alert security team via event
        s.events.Emit(ctx, SecurityEvent{
            Type:   "ACCOUNT_LOCKED",
            UserID: userID,
        })
    }
}
```

IP-level rate limiting (Chapter 13) provides the outer defence against credential stuffing. Per-account lockout is the inner defence for targeted attacks on specific accounts.
