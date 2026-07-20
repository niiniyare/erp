> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Authentication (AuthN) — Who Are You?

### Unified Identity Model

All user types share one `users` table. No separate identity stores per surface.

```sql
CREATE TABLE users (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text        UNIQUE NOT NULL,
  user_type     text        NOT NULL,     -- 'platform' | 'tenant' | 'portal' | 'third_party'
  tenant_id     uuid        REFERENCES tenants(id) NULL,  -- NULL for platform users
  principal_id  uuid        NULL,         -- portal users: contact_id or employee_id
  entity_id     uuid        NULL,         -- organisational node (see Entity Hierarchy)
  display_name  text        NOT NULL,
  password_hash text        NOT NULL,     -- bcrypt cost 12
  mfa_secret    text        NULL,         -- TOTP secret, AES-256-GCM encrypted
  mfa_enabled   bool        NOT NULL DEFAULT false,
  status        text        NOT NULL DEFAULT 'active', -- 'active'|'suspended'|'invited'
  failed_login_attempts  int  NOT NULL DEFAULT 0,
  locked_until           timestamptz NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_login_at timestamptz NULL
);
```

**Scopes:**
- `tenant_id = NULL` → platform scope: global admins, Awo operators. Uses `admin_role` DB pool.
- `tenant_id = <uuid>` → all other types: employees, portal contacts, API integrations. All queries RLS-scoped.

**`principal_id`** — portal users' identity is their business record (a contact or employee). Portal handlers always read `principal_id` from session — never from query params.

**`entity_id`** — the organisational node. Determines which subtree of the hierarchy the user can access. Resolved into `EntityScope` at login.

---

### Login Flow

```
Step 1: Credential Verification
  Load user by email → check status != 'suspended' → check locked_until
  bcrypt.CompareHashAndPassword (constant-time)
  On failure: increment failed_login_attempts
    ≥5 failures → locked_until = NOW() + escalating backoff (15m→30m→1h→2h)
  Always return generic "Invalid email or password" — never reveal which field failed

Step 2: MFA (if mfa_enabled = true)
  No code in request → return { mfa_required: true }; UI shows MFA field
  Code present → TOTP validate ±1 window → replay check

Step 3: Session Construction (the expensive step, runs once)
  Generate 32-byte random token; store SHA-256 hash
  ComputePermissions()      → map[string]bool (roles → permissions + deny policies)
  ResolveEntityScope()      → EntityScope (entity_id → subtree path)
  FlagService.Resolve()     → map[string]bool (flag definitions + tenant overrides)
  SettingService.Resolve()  → map[string]string (setting definitions + tenant overrides)
  PrefService.GetForUser()  → map[string]string (user preferences)
  INSERT sessions: token_hash, permissions JSONB, entity_scope JSONB,
                   configuration JSONB {flags, settings, prefs},
                   expires_at = NOW() + 24h

Step 4: Response
  Set HttpOnly+Secure+SameSite=Lax cookie: awo_session = plaintext token
  Return { status: 0, data: { reload_schema: true } }
  amis reloads /schema/boot → BootService builds app shell → full UI
```

The session construction runs five queries at login. After that, every auth, flag, setting, and preference check for the next 24 hours is an in-memory lookup.

---

### Multi-Factor Authentication

TOTP (RFC 6238) only. SMS OTP not supported — SIM-swapping is unacceptable for a financial system.

```go
func (s *IAMService) ValidateMFA(ctx context.Context,
    params domain.ValidateMFAParams) (bool, error) {
    user, _ := s.repo.GetUser(ctx, domain.GetUserParams{UserID: params.UserID})
    secret, _ := s.crypto.Decrypt(user.MFASecret)
    if !gotp.NewDefaultTOTP(secret).VerifyWithWindow(params.Code, time.Now().Unix(), 1) {
        return false, nil
    }
    // Replay prevention: cache used code for 3 windows (90s)
    key := fmt.Sprintf("mfa_used:%s:%d", params.UserID, time.Now().Unix()/30)
    if s.cache.Exists(ctx, key).Val() > 0 { return false, domain.ErrMFACodeReplayed }
    s.cache.Set(ctx, key, "1", 90*time.Second)
    return true, nil
}
```

MFA mandatory for all users with `finance.*` or `platform.*` permissions. Configurable for others via `iam.mfa.required` flag.

---

### Session Architecture

DB sessions — not JWT. JWTs cannot be revoked; a session row is deleted instantly on logout or suspension.

```sql
CREATE TABLE sessions (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_type     text        NOT NULL,
  tenant_id     uuid        NULL,
  token_hash    text        UNIQUE NOT NULL,   -- SHA-256 of the plaintext token
  permissions   jsonb       NOT NULL DEFAULT '{}',
  entity_scope  jsonb       NOT NULL DEFAULT '{}',
  configuration jsonb       NOT NULL DEFAULT '{}',  -- {flags:{}, settings:{}, prefs:{}}
  created_at    timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  ip_address    text,
  user_agent    text
);
CREATE UNIQUE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at) WHERE expires_at < NOW();
```

Session validation atomically validates + touches `last_seen_at` in one query. No separate read then update.

**Invalidation triggers:**

| Event | Action |
|---|---|
| Logout | DELETE session row |
| User suspended | DELETE all user's sessions |
| Sensitive permission revoked | DELETE all user's sessions |
| Significant tenant flag changed | DELETE all tenant sessions |
| Session TTL (24h) | Cleanup job |

For non-urgent role additions, staleness up to 24h is acceptable. For immediate-effect changes (termination, suspension), call `InvalidateByUser()` which deletes sessions, forcing fresh permission computation at next login.

---

### Password Management

bcrypt cost 12 (~250ms/hash, ~4 guesses/sec). Requirements: 12+ chars, mixed case + digit + special, not in HIBP top-10k, not in last-5 hashes. Password reset: 32-byte token, 1-hour expiry, stored hashed, one-time-use, invalidates all sessions on success.

---

### OAuth & SSO

Protocols: OpenID Connect 1.0, SAML 2.0. JIT provisioning (auto-create user on first SSO login) controlled by `iam.sso.auto_provision` flag.

---

### API Keys

```sql
CREATE TABLE api_keys (
  id         uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  uuid    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name       text    NOT NULL,
  key_hash   text    UNIQUE NOT NULL,  -- SHA-256
  scopes     text[]  NOT NULL,         -- subset of permissions
  expires_at timestamptz NULL,
  created_by uuid    NOT NULL REFERENCES users(id),
  revoked_at timestamptz NULL
);
```

API key sessions are built at request time — not stored in `sessions` to avoid DB bloat from high-frequency API calls. Scopes are a ceiling: even if the owning user has broader permissions, the key can only exercise what is in `scopes`.

---

### AuthService Interface

```go
type AuthService interface {
    Login(ctx, params domain.LoginParams)                    (*domain.ResolvedSession, string, error)
    ValidateMFA(ctx, params domain.ValidateMFAParams)         (bool, error)
    ValidateSession(ctx, params domain.ValidateSessionParams) (*domain.ResolvedSession, error)
    Logout(ctx, params domain.LogoutParams)                  error
    RequestPasswordReset(ctx, params domain.PasswordResetRequestParams) error
    ResetPassword(ctx, params domain.PasswordResetParams)    error
    OAuthCallback(ctx, params domain.OAuthCallbackParams)    (*domain.ResolvedSession, string, error)
    InitiateMFA(ctx, params domain.InitiateMFAParams)        (*domain.MFAChallenge, error)
    DisableMFA(ctx, params domain.DisableMFAParams)          error
    ComputePermissions(ctx, params domain.ComputePermissionsParams) (map[string]bool, error)
}
```

---

Next: [Session — Pre-Computed Everything](./10b-session-precomputation.md)
