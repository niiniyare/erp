[<-- Back to Index](README.md)

## Authentication (AuthN) — Who Are You?

> **[IMPLEMENTED]** — describes the v1.0 runtime login, MFA, SSO, and API key flows as implemented.

---

### Identity Model

All user types share one `users` table. No separate identity stores per surface.

```
user_type values → ActorType mapping (see domain/authz.go ActorTypeFromUserType):
  "SYSADMIN" | "PLATFORM"          → ActorPlatform  → domain: "_platform_"
  "PORTAL"   | "CUSTOMER"          → ActorPortal    → domain: "<tenantID>:portal"
  "API"      | "SERVICE"           → ActorAPI       → domain: "<tenantID>:api"
  "INTERNAL" (and anything else)   → ActorTenant    → domain: "<tenantID>"
```

- `tenant_id = NULL` → platform scope (cross-tenant administrators).
- `entity_id` → the entity node. Used to compute `EntityScope` at login.
- `principal_id` → portal users only: the contact/party record they represent.

---

### Standard Login Flow [IMPLEMENTED]

```
POST /auth/login  { email, password }

Step 1: Credential Verification (UserService.Authenticate)
  Load user by email (or username if no "@")
  Check IsLocked() — locked_until > NOW()
  Retrieve password hash from users table (separate column)
  bcrypt.CompareHashAndPassword (constant-time)
  On failure:
    IncrementFailedAttempts()
    If attempts >= MaxFailedAttempts (default: 5):
      LockAccount(user.ID, NOW() + LockoutDuration (default: 15min))
  On success:
    ResetFailedAttempts()
    UpdateLastLogin()

Step 2: MFA Check (SessionService.Login)
  If user.MfaEnabled == true:
    Generate 32-byte random pending token
    StorePendingMFA(ctx, rawPendingToken, userID) → Redis "mfa:login:pending:{token}" TTL=5min
    Return (nil, rawPendingToken, ErrMFARequired)
    → Caller redirects to CompleteMFALogin

Step 3: Session Construction (SessionService.buildAndPersistSession)
  Generate 32-byte random session token → sha256hex(token) = hash
  ResolveEntityScope(ctx, user.EntityID)  — DB query, result stored in session
  LoadLoginConfig(ctx, userID, tenantID)  — resolves flags + settings + prefs
  TTL = iam.session_ttl_hours tenant setting, else 8h default
  INSERT user_sessions: { user_id, user_type, session_token=hash, entity_scope JSONB,
                          configuration JSONB, is_active=true, expires_at }
  CacheResolved(ctx, hash, resolvedSession, ttl) → Redis "session:{hash}"

Step 4: Response
  Return (ResolvedSession, rawToken, nil)
  Caller sets cookie: HttpOnly+Secure+SameSite=Lax
```

**Session carries context only — no permissions, no role list. Authorization is always live via Casbin.**

---

### MFA Flow (TOTP) [IMPLEMENTED]

TOTP (RFC 6238) only. SMS OTP is not supported.

**Enrollment:**
```
POST /auth/mfa/initiate
  → UserService.InitiateMFA(userID)
  → Generate TOTP secret (base32, 20 bytes)
  → Encrypt with AES-256-GCM (key from MFAEncryptionKey config)
  → StorePendingMFASetup(userID, encryptedSecret) → Redis TTL
  → Return { secret: plaintext, qr_uri: "otpauth://totp/..." }

POST /auth/mfa/confirm  { code }
  → UserService.ConfirmMFA(userID, code)
  → GetPendingMFASetup from Redis
  → Decrypt → verifyTOTP(secret, code, window=1)
  → If valid: SetMFASecret(userID, encryptedSecret) in DB
  → ClearPendingMFASetup from Redis
  → MFA is now enabled for the user
```

**Login with MFA:**
```
POST /auth/login  → ErrMFARequired + pendingToken

POST /auth/mfa/complete  { pending_token, code }
  → SessionService.CompleteMFALogin(pendingToken, mfaCode)
  → GetPendingMFA(pendingToken) via Redis GETDEL (atomic consume)
      → Returns userID, deletes key atomically (prevents replay)
  → UserService.ValidateMFACode(userID, code):
      → GetMFASecret from DB (encrypted)
      → Decrypt → verifyTOTP(secret, code, window=1)
      → CheckAndMarkMFAReplay(userID, window) → Redis (prevents code reuse)
  → If valid: buildAndPersistSession → full session created
```

**Disable MFA:**
```
POST /auth/mfa/disable  { password }
  → UserService.DisableMFA(userID, password)
  → Verify password first (re-authentication required)
  → ClearMFASecret(userID) in DB
```

---

### SSO / OAuth Flow [IMPLEMENTED]

Supported providers: **Google**, **Microsoft** (OAuth 2.0 + OIDC userinfo endpoint).
SAML 2.0 is **[PLANNED - NOT IN v1.0]**.

Each tenant configures its own SSO provider via `SSOProvider`. Client secrets are AES-256-GCM encrypted at rest.

```
Step 1: BeginOAuth(tenantID, provider)
  → Load SSOProvider from DB for this tenant
  → Generate 24-byte random CSRF state token
  → Store in Redis: "sso:state:{token}" → {tenantID, provider}  TTL=10min
  → Return authorization URL with state parameter

Step 2: Provider redirects back: GET /auth/sso/callback?code=...&state=...

Step 3: ResolveUser(provider, code, state)
  → Validate CSRF state from Redis (GETDEL — single use)
  → Inject tenantID into context
  → Load SSOProvider config
  → Exchange code for access token (POST to provider token endpoint)
  → Fetch user info (GET provider userinfo endpoint)
  → Look up user by email in DB
  → If found: return existing user
  → If not found and AutoProvision=false: return ErrForbidden
  → If not found and AutoProvision=true:
      Require DefaultEntityID to be configured
      JIT-provision user with UserType="CUSTOMER", random placeholder password
      → RegisterNewUser()

Step 4: Caller passes user to SessionService.LoginWithSSO(user)
  → MFA is intentionally SKIPPED (IdP is the second factor)
  → buildAndPersistSession → full session
```

**Scopes**: Google requests `openid email profile`. Microsoft uses Graph userinfo endpoint.

---

### Password Management [IMPLEMENTED]

- Hashing: bcrypt with `DefaultCost` (~10-12 cost factor).
- Minimum requirements: 12+ characters, at least one uppercase, lowercase, digit, and special character.
- Password history: last 5 hashes stored; reuse check on change and reset.

**Forgot password flow:**
```
POST /auth/password/forgot  { email }
  → UserService.ForgotPassword(email)
  → Load user by email (if not found: return "", uuid.Nil, nil — always 200 to caller)
  → Generate 32-byte random token → sha256hex = hash
  → CreatePasswordResetToken(userID, hash, expiresAt=NOW()+1h)
  → Return rawToken (caller emails it; not stored in plaintext)

POST /auth/password/reset  { token, new_password }
  → UserService.ResetPassword(rawToken, newPassword)
  → sha256hex(rawToken) → lookup token row
  → Check: not expired, not used
  → Check: password strength policy
  → Check: not reused (last 5 hashes)
  → UpdatePasswordAndHistory(userID, newHash, history)
  → MarkPasswordResetTokenUsed(hash)
```

---

### API Keys [IMPLEMENTED]

API keys authenticate machine-to-machine requests. They produce a minimal `ResolvedSession` without a user session.

```
Raw token format: "eak_{64 hex chars}"  (prefix makes keys identifiable in logs)
Storage: only SHA-256 hash stored in DB (api_keys.key_hash)
```

**Creation:**
```
APIKeyService.CreateAPIKey(req)
  → Generate 32 random bytes → "eak_" + hex.EncodeToString(bytes)
  → hash = sha256(rawToken)
  → INSERT api_keys: {tenant_id, name, scopes, created_by, expires_at, key_hash=hash}
  → Return (APIKey, rawToken)  ← rawToken returned ONCE, never stored
```

**Validation:**
```
APIKeyService.ValidateAPIKey(rawToken)
  → hash = sha256(rawToken)
  → Redis cache-aside: "apikey:{hash}" TTL=5min
  → DB lookup: GetByHash(hash)
  → If not found, revoked, or expired: return nil, nil
  → Build ResolvedSession: UserType="API", EntityScope=EntityScopeAll,
      UserID=key.CreatedBy, TenantID=key.TenantID
  → Cache and return
```

**Note on API key authorization**: The `ResolvedSession` from an API key has `UserType="API"`, which maps to `ActorAPI`. For Casbin enforcement to work, the API key's tenant domain (`APIDomain(tenantID)`) must have policies assigned. The `Scopes` field on `APIKey` is stored but the current `buildAPIKeySession` does not enforce scopes via Casbin — it sets up the session for the Casbin path. Scope enforcement against Casbin policies is the caller's responsibility.

**Revocation:**
```
APIKeyService.RevokeAPIKey(keyID)
  → repo.Revoke(keyID) — sets revoked_at in DB
  → Cache eviction: best-effort (key hash not known from keyID; cached entries
    expire naturally after 5min TTL)
```

---

### Session Invalidation [IMPLEMENTED]

| Trigger | Method | Redis behavior |
|---------|--------|----------------|
| Logout | `repo.Invalidate(hash)` | DELETE "session:{hash}" immediately |
| Suspend user / role revoke | `repo.InvalidateByUser(userID)` | DB: all sessions inactive; evict tracked hashes from Redis |
| Tenant-wide logout | `repo.InvalidateByTenant(tenantID)` | DB: all tenant sessions inactive; Redis: natural TTL expiry |
| Token TTL expires | DB row `expires_at < NOW()` | TouchAndGetSession returns nil for expired rows |

`InvalidateByUser` evicts Redis using the `user_sessions:{userID}` index (list of hashes stored at login). `InvalidateByTenant` does not proactively evict Redis (tenant-wide logouts are rare; entries expire naturally).

---

### What Is Not Implemented in v1.0

- SAML 2.0 **[PLANNED]**
- SMS/email OTP MFA (only TOTP) **[PLANNED]**
- Mandatory MFA enforcement by permission scope (e.g. "require MFA for finance.*") **[PLANNED]**
- API key scope enforcement via Casbin (scopes stored but not automatically enforced at middleware layer) **[PARTIAL]**

---

Next: [Session Model](./10b-session-precomputation.md)
