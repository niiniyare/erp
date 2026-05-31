# Multi-Tenant Module - Test Cases

[<-- Back to Index](README.md)

## Table of Contents
- [Domain Layer Tests](#domain-layer-tests)
- [Repository Layer Tests](#repository-layer-tests)
- [Service Layer Tests](#service-layer-tests)
- [Workflow Tests](#workflow-tests)
- [Activity Tests](#activity-tests)
- [Row-Level Security Tests](#row-level-security-tests)
- [Integration Tests](#integration-tests)
- [Performance Tests](#performance-tests)
- [Security Tests](#security-tests)

---

## Domain Layer Tests

### Tenant Entity Tests

#### Test Case: Valid Tenant Creation
```
Test ID: TN-DOM-001
Description: Verify tenant creation with valid data via NewTenant()
Given: Valid name "Acme Corporation" and email "admin@acme.com"
When: Calling domain.NewTenant(name, email)
Then:
  - Tenant is created successfully (no error)
  - ID is generated as UUID
  - Slug is auto-generated as "acme-corporation"
  - Status defaults to PENDING
  - Timezone defaults to "UTC"
  - CurrencyCode defaults to "USD"
  - Metadata and Settings maps are initialized (non-nil)
  - CreatedAt and UpdatedAt are set to current time
  - Email is normalized to lowercase
  - Name is trimmed of whitespace
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestNewTenant_Valid` (4 table cases)

#### Test Case: Tenant Creation with Missing Name
```
Test ID: TN-DOM-002
Description: Verify tenant creation fails without name
Given: Empty name "" and valid email "admin@acme.com"
When: Calling domain.NewTenant("", email)
Then:
  - Returns nil tenant and ErrTenantNameRequired
  - Whitespace-only names are also rejected
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestNewTenant_ValidationErrors` (7 table cases)

#### Test Case: Tenant Creation with Missing/Invalid Email
```
Test ID: TN-DOM-003
Description: Verify tenant creation validates email
Test Data:
  - Missing: "" → ErrTenantEmailRequired
  - Invalid: "not-an-email" → ErrInvalidEmail
  - Invalid: "user@" → ErrInvalidEmail
  - Valid: "admin@company.com" → success
  - Valid: "test.user+tag@domain.co.uk" → success
When: Calling domain.NewTenant(name, email)
Then:
  - Invalid emails return appropriate error
  - Valid emails are accepted and normalized to lowercase
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestNewTenant_ValidationErrors` (covered in same suite)

#### Test Case: Tenant Slug Generation
```
Test ID: TN-DOM-004
Description: Verify automatic slug generation
Test Data:
  - "ACME Corporation & Co." → "acme-corporation-and-co"
  - "Smith & Sons Ltd." → "smith-and-sons-ltd"
  - "ABC-123 Company!!!" → "abc-123-company"
When: Creating tenant without explicit slug
Then:
  - Slug is generated from name
  - Special characters are removed/replaced
  - Slug conforms to URL standards (lowercase, hyphens)
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestNewTenant_Valid` (slug verified per case)

#### Test Case: Functional Options
```
Test ID: TN-DOM-005
Description: Verify functional options for NewTenant
Test Data:
  - WithSubdomain("acme-corp") → sets subdomain
  - WithSubdomain("admin") → returns ErrInvalidSubdomain (reserved)
  - WithSubdomain("-invalid-") → returns ErrInvalidSubdomain (format)
  - WithLimits(industry, invalidSize) → returns ErrInvalidCompanySize
  - WithLimits(industry, "SMALL") → sets company size
  - WithTimezone("Africa/Nairobi") → sets timezone
  - WithCurrency("KES") → sets currency code
  - WithStatus(StatusActive) → overrides default status
When: Calling domain.NewTenant(name, email, opts...)
Then:
  - Valid options are applied to the tenant
  - Invalid options return errors and prevent creation
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestWithSubdomain` (13 cases), `TestWithLimits_CompanySizeValidation` (8 cases), `TestWithTimezoneAndCurrency`, `TestWithStatus`

### Status State Machine Tests

#### Test Case: Valid Status Transitions
```
Test ID: TN-DOM-010
Description: Test all valid tenant status transitions
Test Data:
  - PENDING → ACTIVE (activation)
  - ACTIVE → SUSPENDED (suspension)
  - ACTIVE → ARCHIVED (archival)
  - SUSPENDED → ACTIVE (reactivation)
  - SUSPENDED → ARCHIVED (archival from suspension)
When: Calling CanTransitionTo() for each pair
Then:
  - All valid transitions return true
  - Transition methods (Activate, Suspend, Archive) succeed
  - UpdatedAt is updated on each transition
```
- [x] **Status:** Implemented — `domain/status_test.go:TestCanTransitionTo_Valid` (5 valid transitions)

#### Test Case: Invalid Status Transitions
```
Test ID: TN-DOM-011
Description: Test blocked status transitions per business rules
Test Data:
  - PENDING → SUSPENDED: blocked (cannot suspend before activation)
  - PENDING → ARCHIVED: blocked (must activate first)
  - ARCHIVED → ACTIVE: blocked (terminal state)
  - ARCHIVED → SUSPENDED: blocked (terminal state)
  - SUSPENDED → PENDING: blocked (cannot revert to initial)
When: Calling CanTransitionTo() for each pair
Then:
  - All invalid transitions return false
  - Transition methods return appropriate errors
```
- [x] **Status:** Implemented — `domain/status_test.go:TestCanTransitionTo_Blocked` (12 blocked transitions)

#### Test Case: Specific Transition Errors
```
Test ID: TN-DOM-012
Description: Verify specific error types for business method calls
Test Data:
  - Activate() on ACTIVE tenant → ErrAlreadyActive
  - Activate() on ARCHIVED tenant → ErrCannotActivateArchivedTenant
  - Suspend() on SUSPENDED tenant → ErrAlreadySuspended
  - Suspend() on ARCHIVED tenant → ErrCannotSuspendArchivedTenant
  - Archive() on ARCHIVED tenant → ErrAlreadyArchived
When: Calling business methods on tenants in various states
Then:
  - Each scenario returns the correct sentinel error
  - Tenant state is not modified on error
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestActivate` (4 cases), `TestSuspend` (4 cases), `TestArchive` (4 cases)

#### Test Case: Status Parsing
```
Test ID: TN-DOM-013
Description: Verify ParseTenantStatus validation
Test Data:
  - "ACTIVE" → StatusActive, nil
  - "PENDING" → StatusPending, nil
  - "SUSPENDED" → StatusSuspended, nil
  - "ARCHIVED" → StatusArchived, nil
  - "INVALID" → "", error
  - "active" → "", error (case-sensitive)
  - "" → "", error
When: Calling ParseTenantStatus(s)
Then:
  - Valid uppercase status strings are parsed correctly
  - Invalid strings return descriptive errors
```
- [x] **Status:** Implemented — `domain/status_test.go:TestParseTenantStatus` (9 cases incl. DEACTIVATED rejected)

### Suspend Metadata Tests

#### Test Case: Suspension Stores Reason
```
Test ID: TN-DOM-014
Description: Verify suspension reason is recorded in metadata
Given: Active tenant
When: Calling tenant.Suspend("payment_failure")
Then:
  - Status changes to SUSPENDED
  - Metadata["suspension_reason"] == "payment_failure"
  - Metadata["suspended_at"] is set to current time
  - Metadata map is initialized if nil
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestSuspend` + `TestSuspend_InitializesMetadata`

### Archive Tests

#### Test Case: Archive Sets DeletedAt
```
Test ID: TN-DOM-015
Description: Verify archive sets soft-delete timestamp
Given: Active tenant
When: Calling tenant.Archive()
Then:
  - Status changes to ARCHIVED
  - DeletedAt is set to current time (non-nil)
  - UpdatedAt is updated
  - IsSoftDeleted() returns true
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestArchive` (verifies DeletedAt + IsSoftDeleted)

### Predicate Tests

#### Test Case: Tenant Predicates
```
Test ID: TN-DOM-016
Description: Verify all tenant predicate methods
Test Data:
  - ACTIVE tenant: IsActive()=true, IsSuspended()=false, IsArchived()=false
  - SUSPENDED tenant: IsActive()=false, IsSuspended()=true, IsArchived()=false
  - ARCHIVED tenant: IsActive()=false, IsSuspended()=false, IsArchived()=true
  - Tenant with DeletedAt: IsSoftDeleted()=true
  - Tenant without DeletedAt: IsSoftDeleted()=false
When: Calling predicate methods
Then: Each predicate returns the expected boolean
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestPredicates` (5 table cases)

### Enum Validation Tests

#### Test Case: PlanType Values
```
Test ID: TN-DOM-020
Description: Verify PlanType enum values are uppercase
Then:
  - PlanBasic == "BASIC"
  - PlanProfessional == "PROFESSIONAL"
  - PlanEnterprise == "ENTERPRISE"
```
- [x] **Status:** Implemented — `domain/types_test.go:TestPlanTypeConstants` (3 cases)

#### Test Case: CompanySize Validation
```
Test ID: TN-DOM-021
Description: Verify CompanySize enum and validation
Test Data:
  - Valid: "STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE"
  - Invalid: "startup", "Small", "huge", ""
When: Calling ValidCompanySize(s)
Then:
  - Valid uppercase values return true
  - Invalid/lowercase values return false
```
- [x] **Status:** Implemented — `domain/types_test.go:TestCompanySizeConstants` (5 cases) + `TestValidCompanySize` (10 cases)

#### Test Case: AccountingMethod Validation
```
Test ID: TN-DOM-022
Description: Verify AccountingMethod enum and validation
Test Data:
  - Valid: "ACCRUAL", "CASH"
  - Invalid: "accrual", "cash", "HYBRID", ""
When: Calling ValidAccountingMethod(s)
Then:
  - Valid uppercase values return true
  - Invalid/lowercase values return false
```
- [x] **Status:** Implemented — `domain/types_test.go:TestAccountingMethodConstants` + `TestValidAccountingMethod` (6 cases)

### Configuration Tests

#### Test Case: Default Configuration Creation
```
Test ID: TN-DOM-030
Description: Verify NewDefaultConfig creates proper defaults
Given: Valid tenant UUID
When: Calling domain.NewDefaultConfig(tenantID)
Then:
  - TenantID matches input
  - MaxUsers == 100
  - MaxEntities == 1000
  - MaxTransactionsPerMonth == 10000
  - StorageQuota == 10GB (10 * 1024 * 1024 * 1024)
  - AccountingMethod == "ACCRUAL"
  - FiscalYearStartMonth == 1
  - DefaultCurrency == "USD"
  - DateFormat == "YYYY-MM-DD"
  - NumberFormat == "1,234.56"
  - LanguageCode == "en"
```
- [x] **Status:** Implemented — `domain/config_test.go:TestNewDefaultConfig` (all 11 fields) + `TestNewDefaultConfig_DifferentIDs`

### Usage Tests

#### Test Case: Usage Within Limits
```
Test ID: TN-DOM-040
Description: Verify IsWithinLimits check
Given: TenantUsage with ActiveUsers=50, TotalEntities=500
And: TenantConfiguration with MaxUsers=100, MaxEntities=1000
When: Calling usage.IsWithinLimits(config)
Then: Returns true

Given: TenantUsage with ActiveUsers=150
And: TenantConfiguration with MaxUsers=100
When: Calling usage.IsWithinLimits(config)
Then: Returns false
```
- [x] **Status:** Implemented — `domain/usage_test.go:TestIsWithinLimits` (8 table cases)

#### Test Case: Usage Percentage Calculation
```
Test ID: TN-DOM-041
Description: Verify UsagePercentage returns highest ratio
Given: Usage at 50% users, 80% entities, 30% transactions
When: Calling usage.UsagePercentage(config)
Then: Returns 0.8 (highest resource utilization)
```
- [x] **Status:** Implemented — `domain/usage_test.go:TestUsagePercentage` (6 table cases)

### Subdomain Validation Tests

#### Test Case: Reserved Subdomain Rejection
```
Test ID: TN-DOM-050
Description: Verify reserved subdomains are rejected
Test Data:
  - "admin", "api", "www", "app", "cdn", "static", "mail"
  - "dev", "staging", "prod", "test", "localhost"
  - "dashboard", "portal", "auth", "login", "signup"
  - "billing", "payment", "support", "help", "docs", "status"
  - "blog", "news", "about", "contact", "legal", "privacy", "terms"
When: Using WithSubdomain(reserved) option
Then: Returns ErrInvalidSubdomain for each reserved name
```
- [x] **Status:** Implemented — `domain/types_test.go:TestReservedSubdomains` (36 reserved + 2 non-reserved) + `domain/tenant_test.go:TestWithSubdomain` (reserved cases)

#### Test Case: Subdomain Format Validation
```
Test ID: TN-DOM-051
Description: Verify subdomain format rules (RFC 1035)
Test Data:
  - Valid: "acme", "acme-corp", "acme123", "a1b2c3"
  - Invalid: "-acme" (leading hyphen), "acme-" (trailing hyphen)
  - Invalid: "ACME" → normalized to "acme"
  - Invalid: "acme corp" (space), "acme.corp" (dot)
  - Invalid: "" (empty), string > 63 chars (too long)
When: Using WithSubdomain(s) or isValidSubdomain(s)
Then:
  - Valid subdomains pass
  - Invalid subdomains return ErrInvalidSubdomain
  - Uppercase is auto-lowered
```
- [x] **Status:** Implemented — `domain/tenant_test.go:TestWithSubdomain` (13 table cases incl. format + reserved)

---

## Repository Layer Tests

### CRUD Tests

#### Test Case: Create Tenant in Database
```
Test ID: TN-REPO-001
Description: Verify tenant creation via repository
Given: Valid domain.Tenant entity
When: Calling repo.Create(ctx, tenant)
Then:
  - Tenant is persisted to database
  - All fields are mapped correctly via SQLC
  - Unique constraints (slug) are enforced
  - Returns nil error on success
```
- [ ] **Status:** Pending

#### Test Case: Get Tenant by ID
```
Test ID: TN-REPO-002
Description: Verify tenant retrieval by UUID
Given: Existing tenant in database
When: Calling repo.GetByID(ctx, id)
Then:
  - Returns correct tenant with all fields mapped
  - Returns ErrTenantNotFound for non-existent ID
  - Tracing span is created and closed
```
- [ ] **Status:** Pending

#### Test Case: Get Tenant by Subdomain
```
Test ID: TN-REPO-003
Description: Verify tenant lookup by subdomain
Given: Tenant with subdomain "acme-corp"
When: Calling repo.GetBySubdomain(ctx, "acme-corp")
Then:
  - Returns matching tenant
  - Returns ErrTenantNotFound if no match
```
- [ ] **Status:** Pending

#### Test Case: List Tenants with Filters
```
Test ID: TN-REPO-004
Description: Verify SQL-level pagination and filtering
Given: 50 tenants with various statuses and industries
When: Calling repo.List(ctx, filter) with:
  - StatusFilter="ACTIVE", Limit=10, Offset=0
Then:
  - Returns at most 10 active tenants
  - Returns accurate total count
  - Uses FilterTenants SQLC query (not in-memory filtering)
  - Respects sort order
```
- [ ] **Status:** Pending

#### Test Case: Subdomain Uniqueness Check
```
Test ID: TN-REPO-005
Description: Verify subdomain existence check
Given: Tenant with subdomain "taken-domain"
When: Calling repo.SubdomainExists(ctx, "taken-domain")
Then: Returns true
When: Calling repo.SubdomainExists(ctx, "available-domain")
Then: Returns false
```
- [ ] **Status:** Pending

#### Test Case: Soft Delete
```
Test ID: TN-REPO-006
Description: Verify soft delete sets deleted_at
Given: Active tenant
When: Calling repo.SoftDelete(ctx, id)
Then:
  - DeletedAt is set to current time
  - Tenant is excluded from normal list queries
  - Tenant data remains intact for recovery
```
- [ ] **Status:** Pending

### Bulk Operation Tests

#### Test Case: Bulk Status Update
```
Test ID: TN-REPO-010
Description: Verify bulk status update via repository
Given: 5 active tenants
When: Calling repo.BulkUpdateStatus(ctx, ids, StatusSuspended)
Then:
  - All 5 tenants are updated to SUSPENDED
  - Operation is atomic
```
- [ ] **Status:** Pending

#### Test Case: Bulk Soft Delete
```
Test ID: TN-REPO-011
Description: Verify bulk soft delete
Given: 3 tenants
When: Calling repo.BulkSoftDelete(ctx, ids)
Then:
  - All 3 tenants have DeletedAt set
  - Other tenants are unaffected
```
- [ ] **Status:** Pending

### Configuration Repository Tests

#### Test Case: Create and Get Configuration
```
Test ID: TN-REPO-020
Description: Verify configuration CRUD via repository
Given: Newly created tenant
When: Calling repo.CreateDefaultConfig(ctx, tenantID) then repo.GetConfig(ctx, tenantID)
Then:
  - Configuration is created with default values
  - Retrieved configuration matches what was created
  - Returns ErrConfigurationNotFound for missing config
```
- [ ] **Status:** Pending

### Analytics Repository Tests

#### Test Case: Growth Statistics Query
```
Test ID: TN-REPO-030
Description: Verify growth stats use SQLC view query
Given: Tenants created across multiple months
When: Calling repo.GetGrowthStats(ctx, 6)
Then:
  - Returns monthly growth data for last 6 months
  - NewTenants and ActiveNewTenants counts are accurate
```
- [ ] **Status:** Pending

#### Test Case: Status Distribution Query
```
Test ID: TN-REPO-031
Description: Verify status distribution query
Given: Mix of ACTIVE, SUSPENDED, PENDING tenants
When: Calling repo.GetStatusDistribution(ctx)
Then:
  - Counts per status are accurate
  - Percentages sum to ~100%
```
- [ ] **Status:** Pending

### Mapper Tests

#### Test Case: SQLC to Domain Mapping
```
Test ID: TN-REPO-040
Description: Verify bidirectional SQLC ↔ domain converters
Given: SQLC-generated row with all fields populated
When: Calling toDomain(row)
Then:
  - All fields are mapped correctly
  - Nullable fields (pgtype) are handled properly
  - Status string is converted to TenantStatus type
  - JSON metadata/settings are deserialized
```
- [ ] **Status:** Pending

---

## Service Layer Tests

### TenantService Tests

#### Test Case: Create Tenant via Service
```
Test ID: TN-SVC-001
Description: Verify full tenant creation flow
Given: Valid CreateTenantRequest
When: Calling service.Create(ctx, req)
Then:
  - Input validation passes (go-playground/validator)
  - Subdomain uniqueness is checked
  - Reserved subdomains are rejected
  - CompanySize is validated if provided
  - domain.NewTenant() is called with correct params
  - Repo.Create() is called
  - Result is cached
  - Tracing span is created
  - Returns created tenant
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestCreate_Success` + `TestCreate_ValidCompanySize` (5 sizes)

#### Test Case: Create Tenant Validation Failures
```
Test ID: TN-SVC-002
Description: Verify service-level validation
Test Data:
  - Missing name → ErrInvalidRequest
  - Missing email → ErrInvalidRequest
  - Invalid email format → ErrInvalidRequest
  - Subdomain > 63 chars → ErrInvalidRequest
  - Reserved subdomain "admin" → ErrInvalidSubdomain
  - Invalid CompanySize "huge" → ErrInvalidCompanySize
  - Taken subdomain → ErrSubdomainTaken
When: Calling service.Create(ctx, req) with invalid data
Then: Returns appropriate error without database changes
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestCreate_ValidationErrors` (4 cases) + `TestCreate_SubdomainTaken`

#### Test Case: Get Tenant by ID with Cache
```
Test ID: TN-SVC-010
Description: Verify cache-first retrieval pattern
Given: Tenant exists in both cache and database
When: Calling service.GetByID(ctx, id) first time (cache miss)
Then:
  - Cache is checked first
  - On miss, repo.GetByID() is called
  - Result is cached with TTL (30 min)
  - Tracing records cache.hit=false

When: Calling service.GetByID(ctx, id) second time (cache hit)
Then:
  - Cached result is returned
  - Repo is NOT called
  - Tracing records cache.hit=true
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestGetByID_Found` + `TestGetByID_NotFound`

#### Test Case: Update Tenant
```
Test ID: TN-SVC-011
Description: Verify tenant update with cache invalidation
Given: Existing tenant with cached data
When: Calling service.Update(ctx, id, req)
Then:
  - Validation passes
  - Subdomain uniqueness checked if changing
  - Cache is invalidated BEFORE update
  - Repo.Update() is called
  - Fresh data is fetched and re-cached
  - Returns updated tenant
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestUpdate_Success` + `TestUpdate_ValidationErrors` (3 cases) + `TestUpdate_SubdomainTaken`

#### Test Case: Activate Tenant
```
Test ID: TN-SVC-020
Description: Verify activation flow
Given: Tenant in PENDING status
When: Calling service.Activate(ctx, id)
Then:
  - Tenant is loaded from repo
  - Domain Activate() validates transition
  - Repo.Update() persists status change
  - Cache is invalidated
  - Returns nil error
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestActivate` (4 table cases)

#### Test Case: Suspend Tenant
```
Test ID: TN-SVC-021
Description: Verify suspension flow
Given: Active tenant
When: Calling service.Suspend(ctx, id, "payment_failure")
Then:
  - Tenant is loaded from repo
  - Domain Suspend() validates transition and stores reason
  - Repo.Update() persists status change
  - Cache is invalidated
  - Returns nil error
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestSuspend` (4 table cases)

#### Test Case: Archive Tenant
```
Test ID: TN-SVC-022
Description: Verify archive flow (terminal state)
Given: Active or suspended tenant
When: Calling service.Archive(ctx, id)
Then:
  - Tenant is loaded from repo
  - Domain Archive() validates transition, sets DeletedAt
  - Repo.SoftDelete() is called
  - Cache is invalidated
  - Returns nil error
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestArchive` (4 table cases)

#### Test Case: List Tenants with Pagination
```
Test ID: TN-SVC-030
Description: Verify list with pagination defaults
Given: 50 tenants in database
When: Calling service.List(ctx, filter) with Limit=0
Then:
  - Limit defaults to 20 (not unlimited)
  - Offset defaults to 0 if negative
  - Limit capped at 100
  - Returns tenants and total count
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestList_PaginationDefaults` (4 table cases)

#### Test Case: Resolve Tenant ID
```
Test ID: TN-SVC-040
Description: Verify subdomain → UUID resolution
Given: Tenant with subdomain "acme-corp"
When: Calling service.ResolveTenantID(ctx, "acme-corp")
Then: Returns tenant's UUID
When: Calling service.ResolveTenantID(ctx, "nonexistent")
Then: Returns uuid.Nil and error
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestResolveTenantID` + `TestResolveTenantID_NotFound`

#### Test Case: Validate Tenant Access
```
Test ID: TN-SVC-041
Description: Verify access validation checks active status
Given: Active tenant
When: Calling service.ValidateTenantAccess(ctx, id)
Then: Returns nil

Given: Suspended tenant
When: Calling service.ValidateTenantAccess(ctx, id)
Then: Returns ErrTenantSuspended
```
- [x] **Status:** Implemented — `service/tenant_test.go:TestValidateTenantAccess` (4 table cases) + `TestExistsTenant`

### ProvisioningService Tests

#### Test Case: Provision Tenant via Temporal
```
Test ID: TN-SVC-050
Description: Verify provisioning delegates to Temporal workflow
Given: Valid ProvisioningInput
When: Calling provisioning.Provision(ctx, input)
Then:
  - Temporal workflow is started
  - Returns ProvisioningResult with tenant_id and status
  - Tracing span covers the operation
```
- [ ] **Status:** Pending

### AnalyticsService Tests

#### Test Case: Growth Statistics
```
Test ID: TN-SVC-060
Description: Verify analytics service delegates and caches
Given: Growth data in database
When: Calling analytics.GetGrowthStats(ctx, 6)
Then:
  - Repository query is executed
  - Results are returned with monthly breakdown
  - Tracing span is created
```
- [ ] **Status:** Pending

#### Test Case: Status Distribution
```
Test ID: TN-SVC-061
Description: Verify status distribution analytics
When: Calling analytics.GetStatusDistribution(ctx)
Then:
  - Returns counts for each status
  - Percentages are calculated correctly
```
- [ ] **Status:** Pending

---

## Workflow Tests

### Provisioning Workflow Tests

#### Test Case: Successful Provisioning Workflow
```
Test ID: TN-WF-001
Description: Test complete provisioning workflow (saga pattern)
Given: Valid ProvisioningInput
When: Executing ProvisioningWorkflow
Then:
  - Step 1: ProvisionTenantActivity creates tenant (PENDING)
  - Step 2: CreateDefaultConfigActivity creates config
  - Step 3: InitUsageActivity initializes usage stats
  - Step 4: ActivateTenantActivity activates tenant
  - Step 5: SendWelcomeNotificationActivity sends notification
  - Workflow returns ProvisioningResult
  - All steps execute in order
```
- [ ] **Status:** Pending

#### Test Case: Provisioning Rollback on Failure
```
Test ID: TN-WF-002
Description: Test saga compensation on workflow failure
Given: ProvisioningInput that will fail at step 3 (InitUsage)
When: Workflow encounters error
Then:
  - CleanupTenantActivity is called (compensation)
  - Tenant is soft-deleted
  - No orphaned config/usage records remain
  - Workflow returns error with details
```
- [ ] **Status:** Pending

#### Test Case: Provisioning Activity Retry
```
Test ID: TN-WF-003
Description: Verify Temporal retry policies on activities
Given: Transient failure in CreateDefaultConfigActivity
When: Activity fails on first attempt
Then:
  - Temporal retries the activity per retry policy
  - MaxAttempts respected
  - Backoff interval applied
  - Workflow succeeds if retry succeeds
```
- [ ] **Status:** Pending

### Bulk Operations Workflow Tests

#### Test Case: Bulk Status Update Workflow
```
Test ID: TN-WF-010
Description: Test bulk status update via Temporal
Given: List of 10 tenant IDs and target status SUSPENDED
When: Executing BulkOperationWorkflow
Then:
  - BulkUpdateStatusActivity is called with all IDs
  - All tenants are updated
  - Workflow returns success count
```
- [ ] **Status:** Pending

#### Test Case: Bulk Soft Delete Workflow
```
Test ID: TN-WF-011
Description: Test bulk soft delete via Temporal
Given: List of 5 tenant IDs
When: Executing BulkOperationWorkflow with operation="soft_delete"
Then:
  - BulkSoftDeleteActivity is called
  - All tenants are soft-deleted
  - Workflow returns success count
```
- [ ] **Status:** Pending

---

## Activity Tests

#### Test Case: ProvisionTenantActivity
```
Test ID: TN-ACT-001
Description: Verify provision activity wraps service call
Given: Valid ProvisioningInput
When: Calling activities.ProvisionTenantActivity(ctx, input)
Then:
  - Delegates to ProvisioningService.Provision()
  - Creates tracing span "activity.ProvisionTenant"
  - Returns ProvisioningResult on success
  - Returns wrapped error on failure
```
- [ ] **Status:** Pending

#### Test Case: CreateDefaultConfigActivity
```
Test ID: TN-ACT-002
Description: Verify config creation activity
Given: Valid tenant UUID
When: Calling activities.CreateDefaultConfigActivity(ctx, tenantID)
Then:
  - Delegates to repo.CreateDefaultConfig()
  - Creates tracing span
  - Returns nil on success
```
- [ ] **Status:** Pending

#### Test Case: ActivateTenantActivity
```
Test ID: TN-ACT-003
Description: Verify activation activity
Given: PENDING tenant UUID
When: Calling activities.ActivateTenantActivity(ctx, tenantID)
Then:
  - Delegates to TenantService.Activate()
  - Creates tracing span
  - Returns nil on success
```
- [ ] **Status:** Pending

#### Test Case: CleanupTenantActivity (Compensation)
```
Test ID: TN-ACT-004
Description: Verify cleanup activity for saga rollback
Given: Partially provisioned tenant
When: Calling activities.CleanupTenantActivity(ctx, tenantID)
Then:
  - Delegates to TenantService.Delete() (soft delete)
  - Creates tracing span
  - Returns nil on success
```
- [ ] **Status:** Pending

#### Test Case: BulkUpdateStatusActivity
```
Test ID: TN-ACT-005
Description: Verify bulk status update activity
Given: List of tenant IDs and target status
When: Calling activities.BulkUpdateStatusActivity(ctx, ids, status)
Then:
  - Delegates to repo.BulkUpdateStatus()
  - Creates tracing span
  - Returns nil on success
```
- [ ] **Status:** Pending

---

## Row-Level Security Tests

#### Test Case: Tenant Context Setting
```
Test ID: TN-RLS-001
Description: Verify tenant context management via RLS
Given: Valid tenant UUID
When: Calling store.SetTenantContext(ctx, tenantID)
Then:
  - Session variable app.current_tenant_id is set
  - Subsequent queries are scoped to this tenant
  - Context persists for the transaction/session
```
- [ ] **Status:** Pending

#### Test Case: RLS Data Isolation
```
Test ID: TN-RLS-002
Description: Test row-level security prevents cross-tenant access
Given: Tenant A and Tenant B with data in shared tables
When: Querying with Tenant A context
Then:
  - Only Tenant A data is returned
  - INSERT auto-populates tenant_id for Tenant A
  - UPDATE only affects Tenant A rows
  - DELETE only removes Tenant A rows
  - Tenant B data is completely invisible
```
- [ ] **Status:** Pending

#### Test Case: RLS Context Reset
```
Test ID: TN-RLS-003
Description: Verify context reset clears tenant scope
Given: Active tenant context
When: Calling store.ResetTenantContext(ctx)
Then:
  - Session variable is cleared
  - Subsequent queries have no tenant scope
```
- [ ] **Status:** Pending

#### Test Case: WithTenantContext Transaction
```
Test ID: TN-RLS-004
Description: Verify scoped transaction via WithTenantContext
Given: Valid tenant UUID
When: Calling store.WithTenant(ctx, id, func)
Then:
  - Tenant context is set before function executes
  - All operations in function are tenant-scoped
  - Context is properly cleaned up after function returns
  - Errors in function cause proper rollback
```
- [ ] **Status:** Pending

---

## Integration Tests

### Backward Compatibility Tests

#### Test Case: Service Adapter Interface
```
Test ID: TN-INT-001
Description: Verify tenantServiceAdapter implements Service interface
Given: tenant.NewService(deps) returns Service
When: Calling all Service interface methods
Then:
  - CreateTenant → delegates to TenantService.Create
  - GetTenantByID → delegates to TenantService.GetByID
  - GetTenantBySubdomain → delegates to TenantService.GetBySubdomain
  - UpdateTenant → delegates to TenantService.Update
  - DeleteTenant → delegates to TenantService.Delete
  - ActivateTenant → delegates to TenantService.Activate
  - DeactivateTenant → delegates to TenantService.Suspend with "deactivated"
  - ListTenants → delegates to TenantService.List
  - ProvisionTenant → delegates to ProvisioningService.Provision
  - All cache, RLS, and context methods work
```
- [ ] **Status:** Pending

#### Test Case: Type Alias Compatibility
```
Test ID: TN-INT-002
Description: Verify type aliases in top-level package
Then:
  - tenant.Tenant == domain.Tenant
  - tenant.TenantStatus == domain.TenantStatus
  - tenant.StatusActive == "ACTIVE"
  - tenant.StatusPending == "PENDING"
  - tenant.StatusSuspended == "SUSPENDED"
  - tenant.StatusArchived == "ARCHIVED"
  - tenant.PlanBasic == "BASIC"
  - tenant.CompanySizeSmall == "SMALL"
  - All error re-exports match domain errors
```
- [ ] **Status:** Pending

### Wire Integration Tests

#### Test Case: Dependency Injection Wiring
```
Test ID: TN-INT-010
Description: Verify Wire provider wiring compiles
Given: All Wire provider functions in wire/services.go
When: NewTenantService, NewTenantTemporalIntegration are called
Then:
  - Dependencies resolve correctly
  - tenant.Service is created
  - TemporalIntegration registers workflows and activities
```
- [ ] **Status:** Pending

### Temporal Integration Tests

#### Test Case: Temporal Registration
```
Test ID: TN-INT-020
Description: Verify workflow/activity registration with Temporal
Given: TemporalIntegration instance
When: Calling RegisterWithPlatform(platform)
Then:
  - ProvisioningWorkflow is registered
  - BulkOperationWorkflow is registered
  - All 8 activities are registered
  - Logger records registration success
```
- [ ] **Status:** Pending

### Complete Tenant Lifecycle Test

#### Test Case: End-to-End Lifecycle
```
Test ID: TN-INT-030
Description: Test complete tenant lifecycle
Given: New tenant creation request
When: Following complete lifecycle:
  1. Create tenant (PENDING by default)
  2. Activate tenant (PENDING → ACTIVE)
  3. Perform operations (cache, RLS context)
  4. Suspend tenant (ACTIVE → SUSPENDED)
  5. Reactivate tenant (SUSPENDED → ACTIVE)
  6. Archive tenant (ACTIVE → ARCHIVED, terminal)
  7. Verify archived tenant cannot be reactivated
Then:
  - Each step succeeds with correct state
  - Cache is invalidated at each transition
  - Tracing spans cover all operations
  - Final state: ARCHIVED with DeletedAt set
```
- [ ] **Status:** Pending

---

## Performance Tests

#### Test Case: Multi-Tenant Query Performance
```
Test ID: TN-PERF-001
Description: Test query performance with many tenants
Given: Database with 1000+ tenants and representative data
When: Executing common queries across tenants
Then:
  - Simple queries (GetByID) < 50ms
  - List with filters < 200ms
  - RLS overhead < 10% vs non-RLS
  - Concurrent queries don't degrade significantly
```
- [ ] **Status:** Pending

#### Test Case: Cache Performance
```
Test ID: TN-PERF-002
Description: Test cache hit rate and performance improvement
Given: Warm cache with tenant data
When: Repeated GetByID calls
Then:
  - Cache hit rate > 90% for repeated lookups
  - Cached response < 5ms vs ~50ms uncached
  - Cache invalidation on update is immediate
```
- [ ] **Status:** Pending

#### Test Case: Provisioning Performance
```
Test ID: TN-PERF-003
Description: Test provisioning throughput
Given: Temporal workflow environment
When: Processing 10 concurrent provisioning requests
Then:
  - Each provisioning completes within 30 seconds
  - No conflicts between concurrent operations
  - Database connections managed efficiently
```
- [ ] **Status:** Pending

---

## Security Tests

#### Test Case: Cross-Tenant Data Access Prevention
```
Test ID: TN-SEC-001
Description: Verify RLS prevents cross-tenant data leaks
Given: Tenants A and B with separate data
When: Tenant A attempts to access Tenant B data via:
  - Direct UUID guessing
  - Modified API requests
  - SQL injection attempts
Then:
  - All cross-tenant access is blocked
  - Audit logs capture attempts
  - No data leakage occurs
```
- [ ] **Status:** Pending

#### Test Case: Tenant Context Hijacking Prevention
```
Test ID: TN-SEC-002
Description: Verify tenant context cannot be manipulated
Given: Authenticated user for Tenant A
When: Attempting to set context to Tenant B
Then:
  - ValidateTenantAccess() blocks unauthorized access
  - SetTenant() only succeeds for active, authorized tenants
  - Session isolation is maintained
```
- [ ] **Status:** Pending

#### Test Case: Input Sanitization
```
Test ID: TN-SEC-003
Description: Verify all inputs are sanitized
Test Data:
  - Name with HTML: "<script>alert('xss')</script>" → stored as-is (output encoding)
  - Subdomain with injection: "admin'; DROP TABLE--" → rejected (format validation)
  - Email with overflow: "a"×300 + "@domain.com" → rejected (length validation)
When: Creating/updating tenant with malicious inputs
Then:
  - SQL injection is prevented (SQLC parameterized queries)
  - Format validations reject malformed input
  - No server errors or stack traces leaked
```
- [ ] **Status:** Pending
