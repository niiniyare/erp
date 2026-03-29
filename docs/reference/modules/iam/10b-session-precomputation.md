[<-- Back to Index](README.md)

## Session — Pre-Computed Everything

### Why Pre-Compute

A single page load triggers 8–12 API calls. Each call could naively query the DB for permissions, flags, settings, and entity scope. At load, this becomes the dominant source of latency. The solution: compute everything once at login, store it in the session JSONB, read it as in-memory map lookups on every subsequent request.

---

### What Gets Built at Login

```go
// internal/platform/iam.go — called once during session creation

func (s *IAMService) buildSession(ctx context.Context,
    user *domain.User) (*domain.ResolvedSession, error) {

    // Five queries, run in parallel
    var (
        perms    map[string]bool
        scope    domain.EntityScope
        flags    map[string]bool
        settings map[string]string
        prefs    map[string]string
    )

    g, gctx := errgroup.WithContext(ctx)
    g.Go(func() error { var err error; perms,    err = s.computePermissions(gctx, user); return err })
    g.Go(func() error { var err error; scope,    err = s.resolveEntityScope(gctx, user.EntityID); return err })
    g.Go(func() error { var err error; flags,    err = s.flagRepo.ResolveForTenant(gctx, user.TenantID); return err })
    g.Go(func() error { var err error; settings, err = s.settingRepo.ResolveForTenant(gctx, user.TenantID); return err })
    g.Go(func() error { var err error; prefs,    err = s.prefRepo.GetForUser(gctx, user.ID); return err })
    if err := g.Wait(); err != nil { return nil, err }

    return &domain.ResolvedSession{
        UserID:      user.ID,
        UserType:    user.UserType,
        TenantID:    user.TenantID,
        PrincipalID: user.PrincipalID,
        EntityID:    user.EntityID,
        DisplayName: user.DisplayName,
        Permissions: perms,
        EntityScope: scope,
        Configuration: domain.SessionConfiguration{
            Flags:    flags,
            Settings: settings,
            Prefs:    prefs,
        },
    }, nil
}
```

| Operation | Cost (concurrent) |
|---|---|
| ComputePermissions | ~2ms (role→permission JOIN) |
| ResolveEntityScope | ~1ms (entity path lookup) |
| FlagService.Resolve | ~1ms (LEFT JOIN flag tables) |
| SettingService.Resolve | ~1ms (LEFT JOIN setting tables) |
| UserPreferences | ~0.5ms |
| **Total at login** | **~5–6ms** (all five run concurrently) |

No further DB hits for auth, flags, or settings for the 8h session lifetime (default; configurable via tenant setting `iam.session_ttl_hours`).

---

### The ResolvedSession Type

```go
type ResolvedSession struct {
    UserID        uuid.UUID
    UserType      UserType
    TenantID      uuid.UUID
    PrincipalID   uuid.UUID
    EntityID      uuid.UUID
    DisplayName   string
    Permissions   map[string]bool      // "finance.transactions.approve" → true/false
    EntityScope   EntityScope          // type + path prefix for subtree queries
    Configuration SessionConfiguration
}

type SessionConfiguration struct {
    Flags    map[string]bool   // "finance.transactions.approval_workflow" → true
    Settings map[string]string // "finance.transactions.approval_threshold" → "100000"
    Prefs    map[string]string // "finance.entry_mode" → "spreadsheet"
}

// All checks are O(1) map lookups — no DB
func (s *ResolvedSession) Can(permission string) bool {
    return s.Permissions[permission]
}
// CanDo is a convenience wrapper for callers that hold resource and action separately.
func (s *ResolvedSession) CanDo(resource, action string) bool {
    return s.Can(resource + "." + action)
}
func (s *ResolvedSession) FeatureEnabled(key string) bool {
    if !s.Configuration.Flags[key] { return false }
    if idx := strings.Index(key, "."); idx > 0 {
        if !s.Configuration.Flags[key[:idx]] { return false }
    }
    return true
}
func (s *ResolvedSession) IsPlatform() bool { return s.UserType == UserTypePlatform }
func (s *ResolvedSession) IsPortal()   bool { return s.UserType == UserTypePortal }
```

---

### Session Invalidation

When a flag or setting that affects security changes, existing sessions are stale. Services handle this automatically:

```go
func (s *FlagService) Set(ctx context.Context,
    params domain.SetFlagParams) error {

    def, _ := s.repo.GetDefinition(ctx, params.FlagKey)
    if err := s.repo.UpsertTenantFlag(ctx, params); err != nil { return err }

    // A flag that makes a feature appear or disappear requires session refresh
    if def.IsModuleOrResourceFlag() {
        s.sessionRepo.InvalidateByTenant(ctx, params.TenantID)
    }
    return nil
}
```

**Invalidation matrix:**

| Trigger | Scope | Method |
|---|---|---|
| Logout | Single session | `InvalidateBySession()` |
| User suspended/terminated | All user sessions | `InvalidateByUser()` |
| Sensitive role/permission change | All user sessions | `InvalidateByUser()` |
| Module/resource flag toggled | All tenant sessions | `InvalidateByTenant()` |
| Session TTL (8h default) | Expired rows | Background cleanup job |

---

### In-Handler Usage

Handlers only read from the session — zero DB hits:

```go
func TransactionFormSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)

        cfg := TransactionFormConfig{
            // Flags — from session, zero DB hit
            ShowCurrencyField:   session.FeatureEnabled("finance.multi_currency"),
            ShowProjectField:    session.FeatureEnabled("finance.project_tracking"),
            ShowApprovalSection: session.FeatureEnabled("finance.transactions.approval_workflow"),

            // Settings — from session, typed helpers
            RequireCostCenter:   session.SettingBool("finance.transactions.cost_center_required", false),
            ApprovalThreshold:   session.SettingDecimal("finance.transactions.approval_threshold", decimal.Zero),
            DecimalPlaces:       session.SettingInt("finance.decimal_places", 2),

            // Permissions — from session
            CanPost:             session.CanDo("finance.transactions", "post"),
            CanApprove:          session.CanDo("finance.transactions", "approve"),
            CanVoid:             session.CanDo("finance.transactions", "void"),

            // User preferences — from session
            EntryMode:           session.Configuration.Prefs["finance.entry_mode"],
            ShowAccountCodes:    session.Configuration.Prefs["finance.show_account_codes"] == "true",
        }

        return c.JSON(buildTransactionForm(cfg))
    }
}
```

---

Next: [UI Navigation](./11b-ui-navigation.md)
