[<-- Back to Index](README.md)

## IAM Service Interfaces

All IAM services are accessed through the `*platform.Platform` facade. Business modules never import individual service files.

---

### AuthService

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

### IdentityService

```go
type IdentityService interface {
    CreateUser(ctx, params domain.UserCreateParams)          (*domain.User, error)
    InviteUser(ctx, params domain.UserInviteParams)          (*domain.Invitation, error)
    AcceptInvitation(ctx, params domain.AcceptInviteParams)  (*domain.User, error)
    GetUser(ctx, params domain.GetUserParams)                (*domain.User, error)
    ListUsers(ctx, params domain.UserListParams)             ([]*domain.User, int, error)
    UpdateUser(ctx, params domain.UserUpdateParams)          (*domain.User, error)
    DeactivateUser(ctx, params domain.DeactivateParams)      error
    SuspendUser(ctx, params domain.SuspendParams)            error
    UnsuspendUser(ctx, params domain.UnsuspendParams)        error
}
```

| Action | Permission required |
|---|---|
| List | `iam.users.read` |
| Invite | `iam.users.create` |
| Update | `iam.users.update` |
| Deactivate | `iam.users.delete` |
| Suspend | `iam.users.suspend` |

---

### AccessService

```go
type AccessService interface {
    CreateRole(ctx, params domain.RoleCreateParams)           (*domain.Role, error)
    UpdateRole(ctx, params domain.RoleUpdateParams)           (*domain.Role, error)
    DeleteRole(ctx, params domain.RoleDeleteParams)           error
    ListRoles(ctx, params domain.RoleListParams)              ([]*domain.Role, error)
    SyncRolePermissions(ctx, params domain.SyncPermsParams)   error
    AssignRole(ctx, params domain.AssignRoleParams)           error
    RevokeRole(ctx, params domain.RevokeRoleParams)           error
    GetUserRoles(ctx, params domain.GetRolesParams)           ([]*domain.Role, error)
    ListAssignments(ctx, params domain.AssignmentListParams)  ([]*domain.Assignment, int, error)
    Can(ctx, params domain.CanParams)                         (bool, error)
    GetPermissionMap(ctx, params domain.PermMapParams)        (map[string]bool, error)
}
```

---

### FlagService

```go
type FlagService interface {
    ListDefinitions(ctx, params domain.ListFlagDefsParams)    ([]*domain.FlagDefinition, error)
    GetDefinition(ctx, flagKey string)                        (*domain.FlagDefinition, error)

    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]bool, error)
    ListForTenant(ctx, params domain.ListFlagsParams)         ([]*domain.TenantFlagWithDef, error)
    Set(ctx, params domain.SetFlagParams)                     error
    Reset(ctx, params domain.ResetFlagParams)                 error
}
```

---

### SettingService

```go
type SettingService interface {
    ListDefinitions(ctx, params domain.ListSettingDefsParams) ([]*domain.SettingDefinition, error)

    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]string, error)
    ListForModule(ctx, params domain.ListSettingsParams)      ([]*domain.SettingWithValue, error)
    UpdateModuleSettings(ctx, params domain.UpdateSettingsParams) error
    GetSetting(ctx, params domain.GetSettingParams)           (string, error)
    ResetToDefault(ctx, params domain.ResetSettingParams)     error

    GetUserPreferences(ctx, userID uuid.UUID)                 (map[string]string, error)
    SetUserPreference(ctx, params domain.SetPrefParams)       error
}
```

---

### BootService

```go
type BootService interface {
    BuildAppShell(ctx, session *domain.ResolvedSession) (map[string]any, error)
    LoginSchema(message string)                         map[string]any
}
```

---

### System Roles (Seeded at Provisioning)

```go
var SystemRoles = []RoleSeed{
    {Slug: "platform.superadmin", AssignableTo: []string{"platform"},
     Permissions: []string{"platform.*"}},
    {Slug: "platform.support",    AssignableTo: []string{"platform"},
     Permissions: []string{"platform.tenants.read","platform.users.read","platform.audit.read"}},
    {Slug: "finance.controller",  AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.*"}},
    {Slug: "finance.accountant",  AssignableTo: []string{"tenant"},
     Permissions: []string{
         "finance.accounts.read","finance.accounts.create",
         "finance.transactions.read","finance.transactions.create","finance.transactions.post"}},
    {Slug: "finance.viewer",      AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.accounts.read","finance.transactions.read"}},
    {Slug: "portal.customer",     AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.invoices.read","portal.self.statements.read"}},
    {Slug: "portal.employee",     AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.payslips.read","portal.self.leave.create"}},
}
```

`AssignableTo` is enforced by Guard 1 in `AssignRole()` — see [Role Management](./07-role-management.md).

---

Next: [Cross-Module Integration](./15b-cross-module-integration.md)
