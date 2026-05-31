# Authorization Module — Test Cases

[<-- Back to Index](README.md)

## Table of Contents
- [Types & Helpers Tests](#types--helpers-tests)
- [Adapter Layer Tests](#adapter-layer-tests)
- [Service Layer Tests](#service-layer-tests)
- [Role Management Tests](#role-management-tests)
- [Policy Management Tests](#policy-management-tests)
- [Temporal Role Tests](#temporal-role-tests)
- [Middleware Tests](#middleware-tests)
- [Domain Isolation Tests](#domain-isolation-tests)
- [Integration Tests](#integration-tests)
- [Performance Tests](#performance-tests)
- [Security Tests](#security-tests)

---

## Types & Helpers Tests

### Subject Builder Tests

#### Test Case: PlatformSubject
```
Test ID: AZ-TYP-001
Description: Verify PlatformSubject returns correct prefix
Given: userID = "usr_ops_001"
When: Calling authz.PlatformSubject("usr_ops_001")
Then:
  - Returns "platform:usr_ops_001"
  - Prefix is exactly "platform:"
  - Input is not modified or trimmed
```
- [ ] **Status:** Pending — `types_test.go:TestSubjectBuilders`

#### Test Case: TenantSubject
```
Test ID: AZ-TYP-002
Description: Verify TenantSubject returns correct prefix
Given: userID = "usr_amina_hassan"
When: Calling authz.TenantSubject("usr_amina_hassan")
Then:
  - Returns "tenant:usr_amina_hassan"
  - Prefix is exactly "tenant:"
```
- [ ] **Status:** Pending — `types_test.go:TestSubjectBuilders`

#### Test Case: PortalSubject
```
Test ID: AZ-TYP-003
Description: Verify PortalSubject returns correct prefix
Given: userID = "cust_nairobi_001"
When: Calling authz.PortalSubject("cust_nairobi_001")
Then:
  - Returns "portal:cust_nairobi_001"
  - Prefix is exactly "portal:"
```
- [ ] **Status:** Pending — `types_test.go:TestSubjectBuilders`

#### Test Case: APISubject
```
Test ID: AZ-TYP-004
Description: Verify APISubject returns correct prefix
Given: clientID = "cli_quickbooks_sync"
When: Calling authz.APISubject("cli_quickbooks_sync")
Then:
  - Returns "api:cli_quickbooks_sync"
  - Prefix is exactly "api:"
```
- [ ] **Status:** Pending — `types_test.go:TestSubjectBuilders`

### Domain Builder Tests

#### Test Case: TenantDomain
```
Test ID: AZ-TYP-010
Description: Verify TenantDomain returns the raw UUID
Given: tenantID = "a1b2c3d4-5678-90ab-cdef-0123456789ab"
When: Calling authz.TenantDomain("a1b2c3d4-5678-90ab-cdef-0123456789ab")
Then:
  - Returns "a1b2c3d4-5678-90ab-cdef-0123456789ab" (unchanged)
  - No suffix is appended
```
- [ ] **Status:** Pending — `types_test.go:TestDomainBuilders`

#### Test Case: PortalDomain
```
Test ID: AZ-TYP-011
Description: Verify PortalDomain appends :portal suffix
Given: tenantID = "a1b2c3d4-5678-90ab-cdef-0123456789ab"
When: Calling authz.PortalDomain("a1b2c3d4-5678-90ab-cdef-0123456789ab")
Then:
  - Returns "a1b2c3d4-5678-90ab-cdef-0123456789ab:portal"
  - Suffix is exactly ":portal"
```
- [ ] **Status:** Pending — `types_test.go:TestDomainBuilders`

#### Test Case: APIDomain
```
Test ID: AZ-TYP-012
Description: Verify APIDomain appends :api suffix
Given: tenantID = "a1b2c3d4-5678-90ab-cdef-0123456789ab"
When: Calling authz.APIDomain("a1b2c3d4-5678-90ab-cdef-0123456789ab")
Then:
  - Returns "a1b2c3d4-5678-90ab-cdef-0123456789ab:api"
  - Suffix is exactly ":api"
```
- [ ] **Status:** Pending — `types_test.go:TestDomainBuilders`

#### Test Case: DomainPlatform constant
```
Test ID: AZ-TYP-013
Description: Verify the platform domain constant
Then:
  - authz.DomainPlatform == "_platform_"
  - Leading and trailing underscores present
  - Cannot be confused with a tenant UUID
```
- [ ] **Status:** Pending — `types_test.go:TestConstants`

### AssignOpt Tests

#### Test Case: WithExpiry option
```
Test ID: AZ-TYP-020
Description: Verify WithExpiry sets expires_at on assignOpts
Given: expiry := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
When: Creating an AssignOpt via authz.WithExpiry(expiry)
And: Applying it to an assignOpts struct
Then:
  - opts.expiresAt is non-nil
  - *opts.expiresAt == expiry (exact equality)
```
- [ ] **Status:** Pending — `types_test.go:TestAssignOpts`

#### Test Case: WithAssignedBy option
```
Test ID: AZ-TYP-021
Description: Verify WithAssignedBy sets assignedBy field
Given: assignerID = "tenant:usr_ceo_001"
When: Applying authz.WithAssignedBy("tenant:usr_ceo_001")
Then: opts.assignedBy == "tenant:usr_ceo_001"
```
- [ ] **Status:** Pending — `types_test.go:TestAssignOpts`

#### Test Case: WithDelegatedBy option
```
Test ID: AZ-TYP-022
Description: Verify WithDelegatedBy sets delegatedBy field
Given: delegatorID = "tenant:usr_cfo"
When: Applying authz.WithDelegatedBy("tenant:usr_cfo")
Then: opts.delegatedBy == "tenant:usr_cfo"
```
- [ ] **Status:** Pending — `types_test.go:TestAssignOpts`

#### Test Case: Multiple options applied in order
```
Test ID: AZ-TYP-023
Description: Verify multiple opts are all applied
Given: Multiple AssignOpts (expiry + assignedBy + delegatedBy)
When: All three opts are applied to a single assignOpts struct
Then:
  - All three fields are set correctly
  - Options do not interfere with each other
  - Last write wins if same field is set twice
```
- [ ] **Status:** Pending — `types_test.go:TestAssignOpts_MultipleOptions`

### Error Type Tests

#### Test Case: Error string format
```
Test ID: AZ-TYP-030
Description: Verify Error.Error() produces expected format
Given: ErrForbidden = &Error{Code:"AUTHZ_FORBIDDEN", Message:"access denied", HTTPStatus:403}
When: Calling ErrForbidden.Error()
Then: Returns "[authz] AUTHZ_FORBIDDEN: access denied"
```
- [ ] **Status:** Pending — `errors_test.go:TestErrorString`

#### Test Case: Sentinel error HTTP status codes
```
Test ID: AZ-TYP-031
Description: Verify each sentinel has correct HTTP status
Test Data:
  - ErrForbidden.HTTPStatus      == 403
  - ErrUnauthorized.HTTPStatus   == 401
  - ErrInvalidRequest.HTTPStatus == 400
  - ErrPolicyConflict.HTTPStatus == 409
When: Inspecting sentinel error values
Then: Each HTTPStatus matches expected HTTP semantics
```
- [ ] **Status:** Pending — `errors_test.go:TestSentinelErrors`

---

## Adapter Layer Tests

### LoadPolicy Tests

#### Test Case: LoadPolicy empty database
```
Test ID: AZ-ADP-001
Description: Verify LoadPolicy succeeds with zero rules
Given: Empty casbin_rule table
When: Calling adapter.LoadPolicy(model)
Then:
  - Returns nil error
  - Model contains no policies or role assignments
  - No panic on empty result set
```
- [ ] **Status:** Pending — `adapter_test.go:TestLoadPolicy_Empty`

#### Test Case: LoadPolicy with p-rules
```
Test ID: AZ-ADP-002
Description: Verify p-rules are loaded into model correctly
Given: casbin_rule rows:
  | ptype | v0                  | v1              | v2       | v3    | v4    | v5 |
  | p     | role:finance-manager| tenant-uuid-1   | invoice/*| *     | allow |    |
  | p     | role:sales-rep      | tenant-uuid-1   | order/*  | read  | allow |    |
When: Calling adapter.LoadPolicy(model)
Then:
  - Both p-rules appear in model["p"]["p"]
  - v0..v4 are correctly mapped to sub,dom,obj,act,eft
  - Empty v5 is handled without error
```
- [ ] **Status:** Pending — `adapter_test.go:TestLoadPolicy_PRules`

#### Test Case: LoadPolicy with g-rules (role assignments)
```
Test ID: AZ-ADP-003
Description: Verify g-rules (role memberships) are loaded correctly
Given: casbin_rule rows:
  | ptype | v0                    | v1                  | v2          |
  | g     | tenant:usr_amina      | role:finance-manager| tenant-uuid |
  | g     | tenant:usr_ceo        | role:tenant-admin   | tenant-uuid |
When: Calling adapter.LoadPolicy(model)
Then:
  - Both g-rules appear in model["g"]["g"]
  - Role hierarchy is correctly established in memory
  - Casbin enforcer reflects the roles via GetRolesForUserInDomain
```
- [ ] **Status:** Pending — `adapter_test.go:TestLoadPolicy_GRules`

### SavePolicy Tests

#### Test Case: SavePolicy truncates and rewrites
```
Test ID: AZ-ADP-010
Description: Verify SavePolicy does full replace (truncate + insert)
Given: casbin_rule table with 5 existing rules
And: Model containing 3 new rules (no overlap)
When: Calling adapter.SavePolicy(model)
Then:
  - All 5 old rules are gone
  - Exactly 3 new rules are present
  - Operation is atomic (no partial state visible)
```
- [ ] **Status:** Pending — `adapter_test.go:TestSavePolicy_FullReplace`

### AddPolicy / RemovePolicy Tests

#### Test Case: AddPolicy inserts single rule
```
Test ID: AZ-ADP-020
Description: Verify single rule insert via AddPolicy
Given: Empty casbin_rule table
When: Calling adapter.AddPolicy("p", "p", []string{"role:cfo","dom","invoice/*","*","allow"})
Then:
  - Exactly 1 row exists in casbin_rule
  - ptype="p", v0="role:cfo", v1="dom", v2="invoice/*", v3="*", v4="allow", v5=""
```
- [ ] **Status:** Pending — `adapter_test.go:TestAddPolicy_SingleInsert`

#### Test Case: AddPolicy is idempotent (ON CONFLICT DO NOTHING)
```
Test ID: AZ-ADP-021
Description: Verify duplicate insert does not error
Given: Existing rule (ptype, v0..v5) already in casbin_rule
When: Calling adapter.AddPolicy with the identical rule a second time
Then:
  - Returns nil error (no unique constraint violation propagated)
  - Still exactly 1 row (no duplicate)
```
- [ ] **Status:** Pending — `adapter_test.go:TestAddPolicy_Idempotent`

#### Test Case: AddPolicies batch insert
```
Test ID: AZ-ADP-022
Description: Verify bulk insert via AddPolicies
Given: 10 distinct rules to insert
When: Calling adapter.AddPolicies("p", "p", rules)
Then:
  - All 10 rules are present in casbin_rule
  - Single DB round-trip (batch, not N individual queries)
```
- [ ] **Status:** Pending — `adapter_test.go:TestAddPolicies_BatchInsert`

#### Test Case: RemovePolicy deletes matching rule
```
Test ID: AZ-ADP-030
Description: Verify rule deletion by exact match
Given: 3 rules in casbin_rule; target is the second
When: Calling adapter.RemovePolicy("p","p", rule2)
Then:
  - rule2 is gone
  - rule1 and rule3 are untouched
  - Returns nil error
```
- [ ] **Status:** Pending — `adapter_test.go:TestRemovePolicy_ExactMatch`

#### Test Case: RemovePolicy for non-existent rule
```
Test ID: AZ-ADP-031
Description: Verify RemovePolicy does not error on missing rule
Given: casbin_rule does not contain the target rule
When: Calling adapter.RemovePolicy for the absent rule
Then:
  - Returns nil (not an error — rule is already gone)
  - Table state is unchanged
```
- [ ] **Status:** Pending — `adapter_test.go:TestRemovePolicy_NotFound`

#### Test Case: RemoveFilteredPolicy by domain
```
Test ID: AZ-ADP-040
Description: Verify filtered deletion removes all rules for a domain
Given: Rules for domain-A (3 rules) and domain-B (2 rules) in casbin_rule
When: Calling adapter.RemoveFilteredPolicy("p","p", 1, "domain-A")
Then:
  - All 3 domain-A rules are gone
  - Both domain-B rules remain
```
- [ ] **Status:** Pending — `adapter_test.go:TestRemoveFilteredPolicy_ByDomain`

#### Test Case: ruleToValues pads short rule to 6 slots
```
Test ID: AZ-ADP-050
Description: Verify rule shorter than 6 values is padded with empty strings
Given: rule = []string{"role:cfo", "dom", "invoice/*"}  (3 values)
When: Calling ruleToValues(rule)
Then:
  - v0="role:cfo", v1="dom", v2="invoice/*"
  - v3="", v4="", v5="" (padded)
  - No index-out-of-bounds panic
```
- [ ] **Status:** Pending — `adapter_test.go:TestRuleToValues`

---

## Service Layer Tests

### Constructor Tests

#### Test Case: New() with nil Pool
```
Test ID: AZ-SVC-001
Description: Verify New() rejects nil DB pool
Given: Config with Pool=nil
When: Calling authz.New(cfg)
Then:
  - Returns non-nil error
  - Error contains "pool is required"
  - Returned Service is nil
```
- [ ] **Status:** Pending — `service_test.go:TestNew_NilPool`

#### Test Case: New() with nil Logger
```
Test ID: AZ-SVC-002
Description: Verify New() rejects nil logger
Given: Config with valid Pool but Logger=nil
When: Calling authz.New(cfg)
Then:
  - Returns error containing "logger is required"
  - Service is not initialized
```
- [ ] **Status:** Pending — `service_test.go:TestNew_NilLogger`

#### Test Case: New() success
```
Test ID: AZ-SVC-003
Description: Verify New() initializes enforcer and loads policies from DB
Given: Valid Config with connected Pool and Logger
When: Calling authz.New(cfg)
Then:
  - Returns non-nil Service and nil error
  - Casbin in-memory model is populated (LoadPolicy called)
  - EnableAutoSave is active (writes go to DB immediately)
```
- [ ] **Status:** Pending — `service_test.go:TestNew_Success`

### Enforce Tests

#### Test Case: Enforce empty Subject
```
Test ID: AZ-SVC-010
Description: Verify Enforce rejects empty Subject
Given: Request{Subject:"", Domain:"dom", Object:"obj", Action:"act"}
When: Calling svc.Enforce(ctx, req)
Then:
  - Returns (false, ErrInvalidRequest)
  - Casbin enforcer is NOT called
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_InvalidRequest`

#### Test Case: Enforce empty Domain
```
Test ID: AZ-SVC-011
Description: Verify Enforce rejects empty Domain
Given: Request{Subject:"tenant:usr_001", Domain:"", Object:"obj", Action:"act"}
When: Calling svc.Enforce(ctx, req)
Then: Returns (false, ErrInvalidRequest)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_InvalidRequest`

#### Test Case: Enforce empty Object
```
Test ID: AZ-SVC-012
Description: Verify Enforce rejects empty Object
Given: Request{Subject:"tenant:usr_001", Domain:"dom", Object:"", Action:"act"}
When: Calling svc.Enforce(ctx, req)
Then: Returns (false, ErrInvalidRequest)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_InvalidRequest`

#### Test Case: Enforce empty Action
```
Test ID: AZ-SVC-013
Description: Verify Enforce rejects empty Action
Given: Request with empty Action=""
When: Calling svc.Enforce(ctx, req)
Then: Returns (false, ErrInvalidRequest)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_InvalidRequest`

#### Test Case: Enforce allow — exact match
```
Test ID: AZ-SVC-020
Description: Verify Enforce returns true when matching allow policy exists
Given: Policy: {sub:"role:finance-manager", dom:"dom-1", obj:"invoice/inv_001", act:"read", eft:"allow"}
And:   Role: tenant:usr_amina → role:finance-manager in dom-1
When: Enforcing {Subject:"tenant:usr_amina", Domain:"dom-1", Object:"invoice/inv_001", Action:"read"}
Then: Returns (true, nil)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_Allow`

#### Test Case: Enforce allow — wildcard object
```
Test ID: AZ-SVC-021
Description: Verify keyMatch2 wildcard matching on object
Given: Policy with Object="invoice/*" and Effect="allow"
When: Enforcing with Object="invoice/inv_999"
Then: Returns (true, nil) — wildcard matches
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_WildcardObject`

#### Test Case: Enforce allow — wildcard action
```
Test ID: AZ-SVC-022
Description: Verify wildcard action "*" covers all actions
Given: Policy with Action="*" and Effect="allow"
When: Enforcing with Action="create" or "delete" or "approve"
Then: Returns (true, nil) for all — wildcard matches
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_WildcardAction`

#### Test Case: Enforce deny — explicit deny overrides allow
```
Test ID: AZ-SVC-030
Description: Verify deny-override: one deny beats any number of allows
Given: Allow policy: role:finance-manager | dom | invoice/* | * | allow
And:   Deny policy:  tenant:usr_terminated | dom | * | * | deny
And:   Role: tenant:usr_terminated → role:finance-manager in dom
When: Enforcing {Subject:"tenant:usr_terminated", Domain:"dom", Object:"invoice/123", Action:"read"}
Then:
  - Returns (false, nil) — deny wins
  - (false, nil) is NOT an error; it is a valid authorization result
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_DenyOverride`

#### Test Case: Enforce deny — no matching rule
```
Test ID: AZ-SVC-031
Description: Verify Enforce returns false when no rule applies
Given: No policies exist for the subject+domain+object+action combination
When: Enforcing that combination
Then:
  - Returns (false, nil)
  - Default deny (Casbin: no allow → deny)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_DefaultDeny`

#### Test Case: Enforce calls revokeExpiredRoles on hot path
```
Test ID: AZ-SVC-040
Description: Verify expired role is revoked before evaluation
Given: Role assigned with expires_at = 1 hour ago (past)
And:   Role has an allow policy
When: Calling svc.Enforce(ctx, req)
Then:
  - revokeExpiredRoles fires and removes g-rule from Casbin
  - Enforce returns (false, nil) because role is now gone
  - role_assignments.is_active is set to FALSE in DB
```
- [ ] **Status:** Pending — `service_test.go:TestEnforce_TriggersExpiry`

### EnforceBatch Tests

#### Test Case: EnforceBatch empty input
```
Test ID: AZ-SVC-050
Description: Verify EnforceBatch returns nil for empty slice
Given: reqs = []authz.Request{}
When: Calling svc.EnforceBatch(ctx, reqs)
Then: Returns (nil, nil) — empty input, no error
```
- [ ] **Status:** Pending — `service_test.go:TestEnforceBatch_Empty`

#### Test Case: EnforceBatch with invalid request in batch
```
Test ID: AZ-SVC-051
Description: Verify EnforceBatch rejects batch with any invalid request
Given: Batch of 3 requests; second has empty Subject
When: Calling svc.EnforceBatch(ctx, reqs)
Then: Returns (nil, ErrInvalidRequest)
```
- [ ] **Status:** Pending — `service_test.go:TestEnforceBatch_InvalidRequest`

#### Test Case: EnforceBatch mixed allow/deny results
```
Test ID: AZ-SVC-052
Description: Verify batch returns parallel bool slice
Given: 3 requests:
  - req1: user has allow → true
  - req2: no policy → false
  - req3: deny rule → false
When: Calling svc.EnforceBatch(ctx, [req1, req2, req3])
Then:
  - Returns ([true, false, false], nil)
  - Slice length == 3 (same as input)
  - Order matches input order
```
- [ ] **Status:** Pending — `service_test.go:TestEnforceBatch_MixedResults`

### InvalidateCache Tests

#### Test Case: InvalidateCache reloads from DB
```
Test ID: AZ-SVC-060
Description: Verify InvalidateCache refreshes the in-memory model
Given: A policy added directly to casbin_rule via raw SQL (bypassing Service)
And:   The in-memory model does not yet have this policy
When: Calling svc.InvalidateCache(ctx)
Then:
  - The in-memory model is rebuilt from DB
  - Enforce now reflects the directly-inserted rule
  - Returns nil error
```
- [ ] **Status:** Pending — `service_test.go:TestInvalidateCache`

#### Test Case: InvalidateCache is blocking
```
Test ID: AZ-SVC-061
Description: Verify InvalidateCache completes before returning
Given: Large casbin_rule table (1,000+ rules)
When: Calling svc.InvalidateCache(ctx)
Then:
  - Returns only after full reload is complete
  - Immediately following call to Enforce uses fresh model
```
- [ ] **Status:** Pending — `service_test.go:TestInvalidateCache_Blocking`

---

## Role Management Tests

### AssignRole Tests

#### Test Case: AssignRole inserts to DB and Casbin
```
Test ID: AZ-ROLE-001
Description: Verify AssignRole writes to both role_assignments and casbin_rule
Given: No existing role assignments for subject
When: Calling svc.AssignRole(ctx, tenantID, "tenant:usr_001", "role:finance-manager", dom)
Then:
  - role_assignments row is inserted (is_active=TRUE)
  - casbin_rule g-row is inserted: ptype=g, v0="tenant:usr_001", v1="role:finance-manager", v2=dom
  - svc.HasRole(ctx, "tenant:usr_001", "role:finance-manager", dom) returns true
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_Success`

#### Test Case: AssignRole is idempotent (UPSERT)
```
Test ID: AZ-ROLE-002
Description: Verify repeated AssignRole calls do not duplicate rows
Given: AssignRole already called once with same params
When: Calling AssignRole with identical (subject, role, domain) a second time
Then:
  - No error returned
  - Still exactly 1 row in role_assignments
  - Still exactly 1 g-rule in casbin_rule
  - is_active remains TRUE
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_Idempotent`

#### Test Case: AssignRole with expiry stored correctly
```
Test ID: AZ-ROLE-003
Description: Verify WithExpiry stores expires_at in role_assignments
Given: expiry = time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
When: Calling AssignRole with authz.WithExpiry(expiry)
Then:
  - role_assignments.expires_at == expiry (within 1 second tolerance)
  - casbin_rule g-row still exists (not yet expired)
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_WithExpiry`

#### Test Case: AssignRole with assignedBy stored correctly
```
Test ID: AZ-ROLE-004
Description: Verify WithAssignedBy stores assigner in role_assignments
When: Calling AssignRole with authz.WithAssignedBy("tenant:usr_ceo_001")
Then:
  - role_assignments.assigned_by == "tenant:usr_ceo_001"
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_WithAssignedBy`

#### Test Case: AssignRole with delegatedBy stored correctly
```
Test ID: AZ-ROLE-005
Description: Verify WithDelegatedBy stores delegator in role_assignments
When: Calling AssignRole with authz.WithDelegatedBy("tenant:usr_cfo")
Then:
  - role_assignments.delegated_by == "tenant:usr_cfo"
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_WithDelegatedBy`

#### Test Case: AssignRole re-activates a previously revoked role
```
Test ID: AZ-ROLE-006
Description: Verify UPSERT re-activates an inactive role assignment
Given: role_assignments row with is_active=FALSE (previously revoked)
When: Calling AssignRole with the same (subject, role, domain)
Then:
  - role_assignments.is_active becomes TRUE
  - casbin_rule g-row is re-added
  - HasRole returns true
```
- [ ] **Status:** Pending — `roles_test.go:TestAssignRole_ReactivatesRevoked`

### RevokeRole Tests

#### Test Case: RevokeRole deactivates and removes g-rule
```
Test ID: AZ-ROLE-010
Description: Verify RevokeRole marks inactive and removes from Casbin
Given: Active role assignment for tenant:usr_james → role:sales-manager in dom
When: Calling svc.RevokeRole(ctx, "tenant:usr_james", "role:sales-manager", dom)
Then:
  - role_assignments row has is_active=FALSE
  - casbin_rule g-row for this triple is gone
  - HasRole returns false
  - role_assignments row still exists (audit trail preserved)
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeRole_Success`

#### Test Case: RevokeRole for non-existent assignment
```
Test ID: AZ-ROLE-011
Description: Verify RevokeRole does not error when assignment doesn't exist
Given: No role assignment exists for (subject, role, domain)
When: Calling svc.RevokeRole(ctx, subject, role, domain)
Then: Returns nil error (idempotent behavior)
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeRole_NotExist`

### GetRoles Tests

#### Test Case: GetRoles returns current in-memory roles
```
Test ID: AZ-ROLE-020
Description: Verify GetRoles reads from Casbin in-memory model
Given: tenant:usr_amina has been assigned role:finance-manager in dom
When: Calling svc.GetRoles(ctx, "tenant:usr_amina", dom)
Then:
  - Returns []string{"role:finance-manager"}
  - No database query is made (in-memory only)
```
- [ ] **Status:** Pending — `roles_test.go:TestGetRoles_HasRoles`

#### Test Case: GetRoles returns empty slice (not error) when no roles
```
Test ID: AZ-ROLE-021
Description: Verify GetRoles returns empty slice for subject with no roles
Given: tenant:usr_unknown has no role assignments in dom
When: Calling svc.GetRoles(ctx, "tenant:usr_unknown", dom)
Then:
  - Returns ([]string{}, nil) — empty slice, no error
  - NOT an error case
```
- [ ] **Status:** Pending — `roles_test.go:TestGetRoles_NoRoles`

#### Test Case: GetRoles returns multiple roles
```
Test ID: AZ-ROLE-022
Description: Verify GetRoles returns all roles for user
Given: usr_ceo has both role:tenant-admin and role:cfo in dom
When: Calling svc.GetRoles(ctx, "tenant:usr_ceo", dom)
Then:
  - Returns slice containing both "role:tenant-admin" and "role:cfo"
  - Order is not guaranteed (set semantics)
```
- [ ] **Status:** Pending — `roles_test.go:TestGetRoles_MultipleRoles`

### HasRole Tests

#### Test Case: HasRole returns true for active assignment
```
Test ID: AZ-ROLE-030
Description: Verify HasRole detects an active role membership
Given: g-rule exists for (tenant:usr_amina, role:finance-manager, dom)
When: Calling svc.HasRole(ctx, "tenant:usr_amina", "role:finance-manager", dom)
Then: Returns (true, nil)
```
- [ ] **Status:** Pending — `roles_test.go:TestHasRole_True`

#### Test Case: HasRole returns false after revocation
```
Test ID: AZ-ROLE-031
Description: Verify HasRole reflects revoked state immediately
Given: Role was just revoked via RevokeRole
When: Calling svc.HasRole immediately after
Then:
  - Returns (false, nil)
  - In-memory model is up to date (no cache lag)
```
- [ ] **Status:** Pending — `roles_test.go:TestHasRole_AfterRevoke`

#### Test Case: HasRole is domain-scoped
```
Test ID: AZ-ROLE-032
Description: Verify HasRole is scoped to the provided domain
Given: tenant:usr_amina has role:finance-manager in dom-1 but NOT in dom-2
When: Calling HasRole with dom-2
Then: Returns (false, nil)
```
- [ ] **Status:** Pending — `roles_test.go:TestHasRole_WrongDomain`

### GetAssignments Tests

#### Test Case: GetAssignments returns active and inactive rows
```
Test ID: AZ-ROLE-040
Description: Verify GetAssignments returns full audit history
Given: tenant:usr_james has:
  - role:sales-manager (is_active=TRUE)
  - role:sales-rep (is_active=FALSE — previously revoked)
When: Calling svc.GetAssignments(ctx, "tenant:usr_james", dom)
Then:
  - Returns both rows
  - is_active field correctly reflects each row's state
  - Ordered by created_at DESC (newest first)
```
- [ ] **Status:** Pending — `roles_test.go:TestGetAssignments_IncludesInactive`

#### Test Case: GetAssignments returns full metadata
```
Test ID: AZ-ROLE-041
Description: Verify all RoleAssignment fields are populated
Given: Role assigned with all options (expiry, assignedBy, delegatedBy)
When: Calling svc.GetAssignments(ctx, subject, dom)
Then:
  - ID is non-empty UUID
  - Subject, Role, Domain match input
  - TenantID matches
  - AssignedBy and DelegatedBy are set
  - ExpiresAt is non-nil and matches expiry
  - CreatedAt is set
```
- [ ] **Status:** Pending — `roles_test.go:TestGetAssignments_FullMetadata`

#### Test Case: GetAssignments returns empty slice for unknown subject
```
Test ID: AZ-ROLE-042
Description: Verify GetAssignments returns empty slice (not error) for unknown subject
Given: No assignments exist for the subject in any domain
When: Calling svc.GetAssignments(ctx, "tenant:usr_unknown", dom)
Then: Returns ([]RoleAssignment{}, nil)
```
- [ ] **Status:** Pending — `roles_test.go:TestGetAssignments_Unknown`

---

## Policy Management Tests

### AddPolicy Tests

#### Test Case: AddPolicy allow rule
```
Test ID: AZ-POL-001
Description: Verify AddPolicy inserts a valid allow p-rule
Given: Valid Policy{Subject:"role:cfo", Domain:"dom", Object:"invoice/*", Action:"*", Effect:"allow"}
When: Calling svc.AddPolicy(ctx, policy)
Then:
  - Returns nil error
  - casbin_rule contains the p-row
  - Enforce for a matching request returns true
```
- [ ] **Status:** Pending — `policies_test.go:TestAddPolicy_Allow`

#### Test Case: AddPolicy deny rule
```
Test ID: AZ-POL-002
Description: Verify AddPolicy inserts a valid deny p-rule
Given: Policy with Effect="deny"
When: Calling svc.AddPolicy(ctx, policy)
Then:
  - casbin_rule row has v4="deny"
  - Enforce for a matching request returns false (deny overrides any allow)
```
- [ ] **Status:** Pending — `policies_test.go:TestAddPolicy_Deny`

#### Test Case: AddPolicy rejects invalid effect
```
Test ID: AZ-POL-003
Description: Verify AddPolicy validates Effect value
Test Data:
  - Effect="" → error "effect must be..."
  - Effect="ALLOW" → error (uppercase not accepted)
  - Effect="permit" → error
  - Effect="block" → error
When: Calling svc.AddPolicy with invalid Effect
Then:
  - Returns non-nil error
  - casbin_rule is unchanged
```
- [ ] **Status:** Pending — `policies_test.go:TestAddPolicy_InvalidEffect`

#### Test Case: AddPolicy returns ErrPolicyConflict on duplicate
```
Test ID: AZ-POL-004
Description: Verify duplicate policy detection
Given: Policy already exists in casbin_rule
When: Calling svc.AddPolicy with identical (sub, dom, obj, act, eft)
Then:
  - Returns ErrPolicyConflict
  - HTTP status 409
  - No duplicate rows in casbin_rule
```
- [ ] **Status:** Pending — `policies_test.go:TestAddPolicy_Duplicate`

#### Test Case: AddPolicy rejects empty required fields
```
Test ID: AZ-POL-005
Description: Verify all policy fields are required
Test Data:
  - Subject="" → ErrInvalidRequest
  - Domain="" → ErrInvalidRequest
  - Object="" → ErrInvalidRequest
  - Action="" → ErrInvalidRequest
  - Effect="" → error (effect-specific message)
When: Calling AddPolicy with each field empty in turn
Then: Returns appropriate error for each case
```
- [ ] **Status:** Pending — `policies_test.go:TestAddPolicy_EmptyFields`

### RemovePolicy Tests

#### Test Case: RemovePolicy removes matching rule
```
Test ID: AZ-POL-010
Description: Verify RemovePolicy deletes exact match
Given: 2 policies in casbin_rule; target is the first
When: Calling svc.RemovePolicy(ctx, policy1)
Then:
  - policy1 is gone from casbin_rule
  - policy2 is untouched
  - Enforce for policy1's parameters now returns false
```
- [ ] **Status:** Pending — `policies_test.go:TestRemovePolicy_Success`

#### Test Case: RemovePolicy all 5 fields must match
```
Test ID: AZ-POL-011
Description: Verify RemovePolicy requires exact 5-field match
Given: Policy exists with Effect="allow"
When: Calling RemovePolicy with Effect="deny" (all other fields same)
Then:
  - The allow rule is NOT deleted (effect mismatch)
  - Returns nil error (no match is not an error in Casbin)
```
- [ ] **Status:** Pending — `policies_test.go:TestRemovePolicy_EffectMismatch`

### GetPolicies Tests

#### Test Case: GetPolicies returns all p-rules for domain
```
Test ID: AZ-POL-020
Description: Verify GetPolicies returns policies scoped to domain
Given: 3 p-rules for dom-1 and 2 p-rules for dom-2 in casbin_rule
When: Calling svc.GetPolicies(ctx, "dom-1")
Then:
  - Returns exactly 3 Policy structs
  - All have Domain="dom-1"
  - dom-2 policies are excluded
```
- [ ] **Status:** Pending — `policies_test.go:TestGetPolicies_DomainScoped`

#### Test Case: GetPolicies returns empty slice (not error) for unknown domain
```
Test ID: AZ-POL-021
Description: Verify GetPolicies handles domains with no policies
Given: No p-rules exist for "dom-unknown"
When: Calling svc.GetPolicies(ctx, "dom-unknown")
Then: Returns ([]Policy{}, nil) — not an error
```
- [ ] **Status:** Pending — `policies_test.go:TestGetPolicies_EmptyDomain`

#### Test Case: GetPolicies Policy struct fields correctly mapped
```
Test ID: AZ-POL-022
Description: Verify all Policy fields are populated from casbin_rule
Given: p-rule: ptype=p, v0=role:cfo, v1=dom, v2=invoice/*, v3=*, v4=allow
When: Calling svc.GetPolicies(ctx, "dom")
Then: Returned Policy has:
  - Subject = "role:cfo"
  - Domain  = "dom"
  - Object  = "invoice/*"
  - Action  = "*"
  - Effect  = "allow"
```
- [ ] **Status:** Pending — `policies_test.go:TestGetPolicies_FieldMapping`

---

## Temporal Role Tests

### Expiry & Lazy Revoke Tests

#### Test Case: revokeExpiredRoles fires on Enforce
```
Test ID: AZ-TEMP-001
Description: Verify lazy expiry is triggered on every Enforce call
Given: role_assignments row with expires_at = 2 hours ago AND is_active=TRUE
And:   casbin_rule g-row still exists (not yet cleaned up)
When: Calling svc.Enforce(ctx, req) for the expired subject
Then:
  - revokeExpiredRoles queries role_assignments WHERE expires_at < NOW()
  - role_assignments.is_active is set to FALSE
  - casbin_rule g-row is deleted
  - Enforce returns (false, nil)
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_LazyFire`

#### Test Case: Unexpired role is not revoked
```
Test ID: AZ-TEMP-002
Description: Verify roles with future expiry are not touched
Given: role_assignments row with expires_at = 1 hour FROM NOW
When: Calling svc.Enforce(ctx, req) for the subject
Then:
  - is_active remains TRUE
  - casbin_rule g-row is untouched
  - Enforce returns (true, nil) (assuming matching allow policy)
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_NotExpired`

#### Test Case: Permanent role (expires_at = NULL) is never expired
```
Test ID: AZ-TEMP-003
Description: Verify NULL expires_at means permanent role
Given: role_assignments row with expires_at=NULL and is_active=TRUE
When: Calling svc.Enforce(ctx, req) at any time
Then:
  - role_assignments.is_active remains TRUE
  - g-rule is preserved
  - Enforce returns (true, nil) (assuming matching policy)
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_Permanent`

#### Test Case: Multiple expired roles revoked in one Enforce call
```
Test ID: AZ-TEMP-004
Description: Verify all expired roles for a subject are revoked together
Given: Subject has 3 expired roles and 1 permanent role
When: Calling svc.Enforce(ctx, req) for the subject
Then:
  - All 3 expired roles have is_active=FALSE
  - All 3 g-rules removed from Casbin
  - Permanent role untouched
  - Enforce reflects only the permanent role's policies
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_Multiple`

#### Test Case: Expiry check uses partial index
```
Test ID: AZ-TEMP-005
Description: Verify expiry query uses idx_role_assignments_expires
Given: role_assignments with 10,000 rows; 100 have expires_at set
When: Calling EXPLAIN ANALYZE on the expiry check query
Then:
  - Index scan on idx_role_assignments_expires is used
  - Sequential scan on the full table is NOT used
  - Query time < 1ms even at 10k rows
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_IndexUsed`

#### Test Case: Expiry check failure is non-fatal
```
Test ID: AZ-TEMP-010
Description: Verify Enforce proceeds even if expiry check errors
Given: DB connection interrupted during revokeExpiredRoles query
When: Calling svc.Enforce(ctx, req)
Then:
  - Warning is logged (non-fatal error)
  - Enforce continues using current in-memory model
  - Returns enforcement result (true or false) based on cached state
  - Does NOT return DB error to caller
```
- [ ] **Status:** Pending — `roles_test.go:TestRevokeExpiredRoles_NonFatalError`

---

## Middleware Tests

### Fiber Middleware Tests

#### Test Case: Middleware returns 401 when Principal absent
```
Test ID: AZ-MID-001
Description: Verify 401 Unauthorized when no Principal in Locals
Given: Fiber request with no authz_principal in c.Locals
When: Handler with svc.Middleware("invoice", "read") is invoked
Then:
  - Response status is 401 Unauthorized
  - c.Next() is NOT called
  - Body contains "authentication required"
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_NoPrincipal`

#### Test Case: Middleware returns 401 for zero-value Principal
```
Test ID: AZ-MID-002
Description: Verify 401 when Principal has empty Subject
Given: c.Locals(authz.LocalsKeyPrincipal) = authz.Principal{Subject:"", Domain:"dom"}
When: Middleware is invoked
Then: Response status is 401 (Subject is empty → treated as unauthenticated)
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_EmptyPrincipal`

#### Test Case: Middleware returns 403 when Enforce returns false
```
Test ID: AZ-MID-010
Description: Verify 403 Forbidden when user lacks permission
Given: Valid Principal{Subject:"tenant:usr_001", Domain:"dom"}
And:   No allow policy for usr_001 on "invoice/read"
When: Middleware("invoice", "read") is invoked
Then:
  - Response status is 403 Forbidden
  - c.Next() is NOT called
  - Body contains "access denied"
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_Forbidden`

#### Test Case: Middleware calls c.Next() when Enforce returns true
```
Test ID: AZ-MID-011
Description: Verify next handler is called on successful authorization
Given: Valid Principal with matching allow policy
When: Middleware is invoked
Then:
  - Enforce returns (true, nil)
  - c.Next() is called (handler proceeds)
  - Response is from the inner handler
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_Allow`

#### Test Case: Middleware expands object with :id param
```
Test ID: AZ-MID-020
Description: Verify object becomes "object/id" when route param :id is present
Given: Route /invoices/:id with Middleware("invoice", "read")
And:   Request to /invoices/inv_123
And:   Policy allows "invoice/inv_123" read
When: Middleware is invoked
Then:
  - Enforce is called with Object = "invoice/inv_123"
  - Wildcard policy "invoice/*" also matches (keyMatch2)
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_ObjectExpansion`

#### Test Case: Middleware uses plain object when no :id param
```
Test ID: AZ-MID-021
Description: Verify object is used as-is when no :id route param
Given: Route /invoices (no :id) with Middleware("invoice", "read")
When: Middleware is invoked
Then:
  - Enforce is called with Object = "invoice" (no suffix)
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_NoIDParam`

#### Test Case: Middleware returns 500 on Enforce error
```
Test ID: AZ-MID-030
Description: Verify 500 Internal Server Error on unexpected Enforce error
Given: Casbin enforcer returns a non-nil error (e.g., model corruption)
When: Middleware is invoked
Then: Response status is 500 Internal Server Error
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_EnforceError`

#### Test Case: Middleware reads correct Locals key
```
Test ID: AZ-MID-040
Description: Verify Middleware reads from authz.LocalsKeyPrincipal constant
Then:
  - authz.LocalsKeyPrincipal == "authz_principal"
  - Middleware uses this key (not a hardcoded string)
  - Changing the constant propagates correctly
```
- [ ] **Status:** Pending — `middleware_test.go:TestMiddleware_LocalsKey`

---

## Domain Isolation Tests

#### Test Case: Policy in dom-1 does not match request for dom-2
```
Test ID: AZ-ISO-001
Description: Verify domain strict equality in matcher
Given: Policy: role:finance-manager | dom-1 | invoice/* | * | allow
And:   User: tenant:usr_amina → role:finance-manager in dom-1
When: Enforce({Subject:"tenant:usr_amina", Domain:"dom-2", Object:"invoice/123", Action:"read"})
Then:
  - Returns (false, nil)
  - dom-1 policy does NOT bleed into dom-2 evaluation
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_PolicyDoesNotCrossDomain`

#### Test Case: Role assignment in dom-1 does not grant access in dom-2
```
Test ID: AZ-ISO-002
Description: Verify g-rule is domain-scoped in Casbin matcher
Given: g-rule: tenant:usr_amina → role:finance-manager in dom-1 (NOT dom-2)
And:   Identical policy exists in BOTH dom-1 and dom-2
When: Enforce({Subject:"tenant:usr_amina", Domain:"dom-2", Object:"invoice/123", Action:"read"})
Then:
  - Returns (false, nil)
  - Role from dom-1 does NOT apply in dom-2
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_RoleDoesNotCrossDomain`

#### Test Case: Platform domain does not match tenant domain
```
Test ID: AZ-ISO-003
Description: Verify _platform_ policies are isolated from tenant policies
Given: Policy: role:platform-admin | _platform_ | * | * | allow
And:   user → role:platform-admin in _platform_
When: Enforce({Subject:"platform:usr_ops", Domain:"some-tenant-uuid", Object:"invoice/*", Action:"*"})
Then:
  - Returns (false, nil)
  - Platform policies cannot override tenant domain
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_PlatformVsTenant`

#### Test Case: Tenant domain does not match portal domain
```
Test ID: AZ-ISO-004
Description: Verify {tenantID} and {tenantID}:portal are distinct domains
Given: Role assignment: tenant:usr_admin → role:tenant-admin in {tenantID} (not portal)
And:   Policy: role:tenant-admin | {tenantID} | * | * | allow
When: Enforce({Subject:"tenant:usr_admin", Domain:"{tenantID}:portal", Object:"invoice/*", Action:"read"})
Then:
  - Returns (false, nil)
  - Tenant domain role does NOT apply in portal domain
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_TenantVsPortal`

#### Test Case: Tenant domain does not match API domain
```
Test ID: AZ-ISO-005
Description: Verify {tenantID} and {tenantID}:api are distinct domains
Given: Role/policy set up only in {tenantID} domain
When: Enforce with Domain = {tenantID}:api
Then: Returns (false, nil)
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_TenantVsAPI`

#### Test Case: Two tenants with identical role names are isolated
```
Test ID: AZ-ISO-010
Description: Verify role:tenant-admin in tenant-A does not grant access in tenant-B
Given:
  - role:tenant-admin assigned to usr_A in domain tenant-A-uuid
  - Same role:tenant-admin policy exists in tenant-B-uuid domain too
When: Enforce({Subject:"tenant:usr_A", Domain:"tenant-B-uuid", Object:"*", Action:"*"})
Then:
  - Returns (false, nil)
  - usr_A has no g-rule in tenant-B-uuid
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_CrossTenant`

#### Test Case: Wildcard subject "*" in deny rule stays domain-scoped
```
Test ID: AZ-ISO-020
Description: Verify wildcard subject deny rule only applies within its domain
Given: Policy: * | dom-1 | journal/closed/* | create | deny
When: Enforce({Subject:"tenant:usr_anyone", Domain:"dom-2", Object:"journal/closed/q1/je_1", Action:"create"})
Then:
  - Returns based on dom-2 policies (deny from dom-1 does NOT apply)
```
- [ ] **Status:** Pending — `isolation_test.go:TestDomainIsolation_WildcardSubjectDomainScoped`

---

## Integration Tests

### Full Authorization Workflow Tests

#### Test Case: Complete RBAC flow from setup to enforcement
```
Test ID: AZ-INT-001
Description: End-to-end: bootstrap policies → assign role → enforce
Given: Freshly initialized authz service (empty DB)
When: Following steps:
  1. AddPolicy: role:finance-manager | dom | invoice/* | * | allow
  2. AssignRole: tenant:usr_amina → role:finance-manager in dom
  3. Enforce: {Subject:tenant:usr_amina, Domain:dom, Object:invoice/123, Action:read}
Then:
  - Step 1: policy persisted to casbin_rule
  - Step 2: g-rule in casbin_rule + row in role_assignments
  - Step 3: Returns (true, nil)
```
- [ ] **Status:** Pending — `integration_test.go:TestFullRBACFlow`

#### Test Case: Employee termination — complete revocation flow
```
Test ID: AZ-INT-002
Description: End-to-end: terminate employee → all access denied
Given: Employee with role:sales-manager and allow policies
When: Following termination steps:
  1. RevokeRole(subject, "role:sales-manager", dom)
  2. AddPolicy(subject, dom, "*", "*", "deny")
Then:
  - Enforce for any object/action returns (false, nil)
  - Blanket deny prevents re-access even if role is re-assigned later
```
- [ ] **Status:** Pending — `integration_test.go:TestTerminationFlow`

#### Test Case: Auditor time-limited access flow
```
Test ID: AZ-INT-003
Description: End-to-end: assign expiring role → access during period → denied after expiry
Given: expiry = now + 10ms (very short for test)
When: Following steps:
  1. AssignRole with WithExpiry(expiry)
  2. Enforce immediately → (true, nil) — access granted
  3. Sleep until expiry passes (> 10ms)
  4. Enforce again → triggers lazy revoke → (false, nil)
Then:
  - Temporal role revocation works end-to-end
  - role_assignments.is_active=FALSE after step 4
```
- [ ] **Status:** Pending — `integration_test.go:TestTemporalRoleFlow`

#### Test Case: Multi-instance cache sync via InvalidateCache
```
Test ID: AZ-INT-010
Description: Verify instance B sees policy change made by instance A via InvalidateCache
Given: Two service instances sharing the same DB (simulated via two authz.New() calls on same pool)
When: Instance A calls AddPolicy (writes directly to DB via Casbin EnableAutoSave)
Then:
  - Instance B Enforce returns (false, nil) [still sees stale cache]
  - Instance B calls InvalidateCache
  - Instance B Enforce now returns (true, nil) [cache is fresh]
```
- [ ] **Status:** Pending — `integration_test.go:TestMultiInstanceSync`

#### Test Case: Role reconciliation query (rule 8)
```
Test ID: AZ-INT-020
Description: Verify reconciliation query detects inconsistency
Given: Active row in role_assignments with no matching g-rule in casbin_rule
When: Running reconciliation query from Business Rule 8
Then:
  - Query returns the orphaned row
  - Confirms inconsistency is detectable
  - Fix: re-add g-rule or mark inactive
```
- [ ] **Status:** Pending — `integration_test.go:TestReconciliation_DetectsOrphan`

#### Test Case: RLS on role_assignments is tenant-scoped
```
Test ID: AZ-INT-030
Description: Verify application_role can only read its own tenant's role_assignments
Given: role_assignments rows for tenant-A and tenant-B
And:   DB session with current_tenant_id() = tenant-A-uuid
When: Querying role_assignments
Then:
  - Only tenant-A rows are visible
  - tenant-B rows are invisible (RLS applied)
```
- [ ] **Status:** Pending — `integration_test.go:TestRLS_RoleAssignments`

#### Test Case: casbin_rule has no RLS tenant scoping (domain provides isolation)
```
Test ID: AZ-INT-031
Description: Verify casbin_rule is accessible by application_role for all domains
Given: casbin_rule has rows for multiple tenant domains
When: application_role queries casbin_rule
Then:
  - All rows are returned (no RLS filter)
  - Domain isolation is enforced at application level (r.dom == p.dom)
  - This is by design — Casbin needs the full ruleset in memory
```
- [ ] **Status:** Pending — `integration_test.go:TestRLS_CasbinRule_NoTenantFilter`

---

## Performance Tests

#### Test Case: Enforce latency (hot path) < 1ms
```
Test ID: AZ-PERF-001
Description: Verify p99 Enforce latency is sub-millisecond
Given: In-memory model loaded with 1,000 policies across 20 tenants
When: Calling svc.Enforce 10,000 times in a tight loop
Then:
  - p99 latency < 1ms per call
  - p50 latency < 200µs per call
  - No goroutine leaks
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkEnforce`

#### Test Case: EnforceBatch latency stays near-linear
```
Test ID: AZ-PERF-002
Description: Verify batch enforcement is efficient
Given: In-memory model with policies
When: Calling EnforceBatch with N=100 requests
Then:
  - Total latency < 5ms for 100 items
  - Scales near-linearly (not exponentially)
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkEnforceBatch`

#### Test Case: Policy load time at startup for 10,000 rules
```
Test ID: AZ-PERF-003
Description: Verify startup policy load finishes in < 200ms for 10k rules
Given: casbin_rule table with 10,000 rows
When: Calling authz.New(cfg) (which calls LoadPolicy)
Then:
  - Service is ready in < 200ms
  - All 10,000 rules are in memory
  - No timeout or partial load
```
- [ ] **Status:** Pending — `bench_test.go:TestLoadPolicy_10kRules`

#### Test Case: Policy load time for 100,000 rules
```
Test ID: AZ-PERF-004
Description: Verify large rule set loads in < 1s
Given: casbin_rule table with 100,000 rows (extreme scale)
When: Calling authz.New(cfg)
Then:
  - Service starts in < 1s
  - Memory usage < 100MB additional
```
- [ ] **Status:** Pending — `bench_test.go:TestLoadPolicy_100kRules`

#### Test Case: AssignRole latency
```
Test ID: AZ-PERF-010
Description: Verify AssignRole completes within 10ms p99
Given: Connected DB pool
When: Calling svc.AssignRole 1,000 times
Then:
  - p99 < 10ms per call (DB write + in-memory update)
  - No deadlocks or lock contention
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkAssignRole`

#### Test Case: AddPolicy latency
```
Test ID: AZ-PERF-011
Description: Verify AddPolicy completes within 10ms p99
Given: Connected DB pool
When: Calling svc.AddPolicy 1,000 times with distinct rules
Then: p99 < 10ms per call
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkAddPolicy`

#### Test Case: Memory footprint for 1,000 tenants
```
Test ID: AZ-PERF-020
Description: Verify memory usage is bounded for multi-tenant model
Given: 1,000 tenants each with 50 policies
When: Loading all 50,000 rules into the in-memory model
Then:
  - Total memory increase < 25MB (as documented)
  - No off-heap or unbounded allocation
```
- [ ] **Status:** Pending — `bench_test.go:TestMemoryFootprint_1kTenants`

#### Test Case: revokeExpiredRoles query is fast even at scale
```
Test ID: AZ-PERF-030
Description: Verify expiry check uses the partial index at scale
Given: role_assignments with 1,000,000 rows; 1,000 have expires_at set
When: Calling revokeExpiredRoles for a subject+domain pair
Then:
  - Uses idx_role_assignments_expires partial index
  - Query executes in < 1ms
  - Does not scan the full table
```
- [ ] **Status:** Pending — `bench_test.go:TestExpiry_IndexPerformance`

---

## Security Tests

#### Test Case: Privilege escalation — user cannot grant themselves a role
```
Test ID: AZ-SEC-001
Description: Verify authorization is checked before role assignment
Given: tenant:usr_regular_001 with role:finance-viewer (no admin powers)
When: The user's request handler calls AssignRole for themselves to role:tenant-admin
Then:
  - The CALLING HANDLER must check authz before calling AssignRole
  - authz.Enforce must return false for "role-assignments/write" action
  - AssignRole itself does NOT check the caller's permissions
  - This confirms authz must be applied at the handler/service layer above
```
- [ ] **Status:** Pending — `security_test.go:TestNoEscalation_RoleAssignment`

#### Test Case: SQL injection in subject field does not affect adapter
```
Test ID: AZ-SEC-002
Description: Verify parameterized queries prevent injection via subject
Given: Subject crafted as "tenant:'; DROP TABLE casbin_rule; --"
When: This subject is used in AssignRole or AddPolicy
Then:
  - DB query is parameterized (not string-concatenated)
  - casbin_rule table is NOT affected by injected SQL
  - The literal string is stored as the subject value (treated as data)
```
- [ ] **Status:** Pending — `security_test.go:TestSQLInjection_Subject`

#### Test Case: SQL injection in object field
```
Test ID: AZ-SEC-003
Description: Verify object field is also parameterized
Given: Object = "invoice/*; SELECT * FROM casbin_rule WHERE '1'='1"
When: AddPolicy is called with this object
Then:
  - No extra rows are returned or deleted
  - The literal string is stored verbatim
```
- [ ] **Status:** Pending — `security_test.go:TestSQLInjection_Object`

#### Test Case: Cross-tenant policy access is architecturally impossible
```
Test ID: AZ-SEC-010
Description: Verify Casbin matcher r.dom == p.dom prevents cross-tenant access
Given: 2 tenants, identical role names and policies, distinct UUIDs
When: Tenant-A user (with role in dom-A) requests resource in dom-B
Then:
  - Returns (false, nil) — enforcer sees r.dom != p.dom
  - No policy or role from dom-A applies in dom-B
  - This is a Casbin engine guarantee, not application code
```
- [ ] **Status:** Pending — `security_test.go:TestCrossTenant_Impossible`

#### Test Case: deny policy injection — attacker cannot add policies via Enforce
```
Test ID: AZ-SEC-020
Description: Verify Enforce call does not mutate policy state
Given: Malformed request with Object="*; DROP TABLE casbin_rule"
When: Calling svc.Enforce(ctx, req)
Then:
  - Enforce is read-only relative to the policy model
  - casbin_rule is unchanged
  - revokeExpiredRoles is the only write path in Enforce
```
- [ ] **Status:** Pending — `security_test.go:TestEnforce_ReadOnly`

#### Test Case: Blanket deny cannot be bypassed by wildcard role grant
```
Test ID: AZ-SEC-030
Description: Verify deny-override works even if user has all roles
Given: Blanket deny: tenant:usr_001 | dom | * | * | deny
And:   User also has role:tenant-admin with allow on everything
When: Calling Enforce for any resource/action
Then:
  - Returns (false, nil) — deny beats all allows
  - Even role:tenant-admin's wildcard allow is overridden
```
- [ ] **Status:** Pending — `security_test.go:TestDenyCannotBeBypassed`

#### Test Case: Expired role cannot be used by replaying old tokens
```
Test ID: AZ-SEC-040
Description: Verify lazy expiry catches old token replays
Given: Role expired 5 minutes ago; lazy revoke not yet triggered
When: User replays an old JWT (still valid in authn, but role expired in authz)
Then:
  - First Enforce call: revokeExpiredRoles fires → role revoked
  - Enforce returns (false, nil) → access denied
  - Subsequent calls: role is gone → always denied
```
- [ ] **Status:** Pending — `security_test.go:TestExpiredRole_OldTokenReplay`

#### Test Case: Platform domain cannot be injected via request
```
Test ID: AZ-SEC-050
Description: Verify a tenant user cannot claim _platform_ domain
Given: authn middleware enforces domain based on JWT actor type
And:   A tenant user crafts a request claiming Domain="_platform_"
When: Enforce is called with Domain="_platform_" for tenant subject
Then:
  - If no platform g-rule exists: returns (false, nil)
  - Platform policies DO NOT match if user has no platform role assignment
  - This requires authn middleware to enforce correct domain derivation
```
- [ ] **Status:** Pending — `security_test.go:TestPlatformDomain_NotInjectable`

---

## IAM Layer Tests (internal/platform)

### Flag Repository Tests

```go
func TestFlagRepo_ResolveForTenant_DefaultValue(t *testing.T) {
    repo := setupTestRepo(t) // testcontainers postgres

    tenantID := uuid.New()
    // No tenant_feature_flags row — should return default_value
    flags, err := repo.ResolveForTenant(ctx, tenantID)
    assert.NoError(t, err)
    assert.False(t, flags["finance"])
    assert.False(t, flags["finance.transactions.approval_workflow"])
}

func TestFlagRepo_ResolveForTenant_TenantOverride(t *testing.T) {
    repo := setupTestRepo(t)
    tenantID := uuid.New()

    repo.SetFlag(ctx, domain.SetFlagParams{TenantID: tenantID, FlagKey: "finance", Enabled: true})

    flags, _ := repo.ResolveForTenant(ctx, tenantID)
    assert.True(t, flags["finance"])
    assert.True(t, flags["finance.transactions"])
}
```

### Role Assignment Guard Tests

```go
func TestAssignRole_Guards(t *testing.T) {
    svc := setupTestIAM(t)

    t.Run("guard1_namespace", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: tenantUserID, RoleID: platformRoleID})
        assert.ErrorIs(t, err, domain.ErrForbidden)
        assert.Contains(t, err.Error(), "assignable to")
    })
    t.Run("guard2_cross_tenant", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: tenantAUserID, RoleID: tenantBRoleID})
        assert.ErrorIs(t, err, domain.ErrForbidden)
    })
    t.Run("guard3_delegation", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: newUserID, RoleID: controllerRoleID})
        // GrantedBy = accountantUserID (from context) who doesn't hold controller
        assert.ErrorIs(t, err, domain.ErrForbidden)
        assert.Contains(t, err.Error(), "you can only grant roles you hold")
    })
}
```

### Entity Scope Tests

```go
func TestEntityScope_SubtreeAccess(t *testing.T) {
    companyID   := createEntity(t, "company", nil)
    nairobiID   := createEntity(t, "region",  &companyID)
    westlandsID := createEntity(t, "branch",  &nairobiID)

    companyTxn   := createTransaction(t, companyID)
    nairobiTxn   := createTransaction(t, nairobiID)
    westlandsTxn := createTransaction(t, westlandsID)

    ctx := contextWithSession(t, createUserWithEntity(t, nairobiID))
    txns, _ := financeRepo.ListTransactions(ctx, domain.TransactionListParams{})
    ids := transactionIDs(txns)

    assert.Contains(t,    ids, nairobiTxn.ID,   "own entity — accessible")
    assert.Contains(t,    ids, westlandsTxn.ID, "child entity — accessible")
    assert.NotContains(t, ids, companyTxn.ID,   "parent entity — not accessible")
}
```

### Nav Generation Tests

```go
func TestBootService_BuildNav_FlagGating(t *testing.T) {
    svc := setupTestBoot(t)
    tenantID := createTenantWithFlags(t, map[string]bool{
        "finance":              true,
        "finance.transactions": true,
        "airline":              false,  // airline off
    })

    session := buildSessionForTenant(t, tenantID, map[string]bool{
        "finance.transactions.read": true,
        "airline.bookings.read":     true,  // has permission but flag is off
    })

    shell, _ := svc.BuildAppShell(ctx, session)
    pages := shell["pages"].([]any)
    labels := extractNavLabels(pages)

    assert.Contains(t,    labels, "Finance")  // flag on + permission = visible
    assert.NotContains(t, labels, "Airline")  // flag off = not visible regardless of permission
}
```

---

[Back to Index](README.md)
