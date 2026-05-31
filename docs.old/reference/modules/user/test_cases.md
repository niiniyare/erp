# IAM Module -  Test Cases

## Table of Contents
- [Core Domain Model Tests](#core-domain-model-tests)
- [Repository Layer Tests](#repository-layer-tests)
- [Service Layer Tests](#service-layer-tests)
- [API Integration Tests](#api-integration-tests)
- [Hybrid Access Control (RBAC + ABAC) Tests](#hybrid-access-control-rbac-abac-tests)
- [Multi-tenant Isolation Tests](#multi-tenant-isolation-tests)
- [Security Tests](#security-tests)
- [Performance Tests](#performance-tests)
- [End-to-End (E2E) Workflow Tests](#end-to-end-e2e-workflow-tests)

## Core Domain Model Tests

### Person Model
```
Test ID: IAM-CORE-001
Description: Verify Person model creation with valid data.
Given: Valid parameters for a new person (first_name, last_name, person_type, etc.).
When: Creating a new Person instance.
Then:
  - Person object is created successfully.
  - All fields are set correctly.
  - `id` and `created_at`/`updated_at` timestamps are generated automatically.
  - `tenant_id` is correctly assigned from the context.
```
```
Test ID: IAM-CORE-002
Description: Verify Person model validation for invalid data.
Given: Invalid parameters (e.g., null first_name, invalid person_type enum).
When: Attempting to create a Person with invalid data.
Then:
  - Validation error is returned.
  - Error message clearly indicates the validation failure.
  - No Person record is created.
```

### Employee Model
```
Test ID: IAM-CORE-003
Description: Verify Employee model creation and linking to a Person.
Given: A valid, existing Person and valid employee parameters (employee_number, hire_date).
When: Creating a new Employee instance linked to the Person.
Then:
  - Employee object is created successfully.
  - `person_id` correctly links to the Person record.
  - Self-referential `manager_id` can be set and retrieved correctly.
  - `deleted_at` is null, indicating it's not soft-deleted.
```

### User Model
```
Test ID: IAM-CORE-004
Description: Verify User model creation and linking to Person/Employee.
Given: A valid, existing Person and Employee.
When: Creating a new User instance linked to them.
Then:
  - User object is created successfully.
  - `person_id` and `employee_id` are linked correctly.
  - `password_hash` is stored securely and is not the raw password.
  - `account_status` defaults to 'ACTIVE'.
  - `failed_login_attempts` defaults to 0.
```
```
Test ID: IAM-CORE-005
Description: Verify User model security feature defaults.
Given: A newly created User.
When: Inspecting the User object.
Then:
  - `mfa_enabled` is false by default.
  - `lockout_until` is null by default.
  - `password_changed_at` is set to the current time.
```

### Role & Permission Models
```
Test ID: IAM-CORE-006
Description: Verify Role creation with hierarchy.
Given: An existing Role to act as a parent.
When: Creating a new Role with a `parent_role_id`.
Then:
  - The child role is created successfully.
  - The parent-child relationship is correctly established.
  - The system prevents circular role dependencies (e.g., a role cannot be its own parent).
```
```
Test ID: IAM-CORE-007
Description: Verify Permission creation with risk levels and categories.
Given: Valid permission parameters (name, resource_type, action, risk_level).
When: Creating a new Permission.
Then:
  - The permission is created successfully.
  - `risk_level` and `action_category` are validated against their respective enums.
  - `requires_approval` flag is set correctly.
```

## Repository Layer Tests

### CRUD Operations
```
Test ID: IAM-REPO-001
Description: Test basic CRUD operations for User, Person, and Employee.
Given: A valid tenant context.
When: Performing Create, Read, Update, and Soft-Delete operations.
Then:
  - Records are created successfully in the database.
  - Records can be retrieved accurately by ID.
  - Updates modify the correct fields and update the `updated_at` timestamp.
  - Soft-delete sets the `deleted_at` timestamp and the record is no longer retrieved by standard queries.
```

### Tenant Isolation (RLS)
```
Test ID: IAM-REPO-002
Description: Verify tenant isolation for all IAM entities.
Given: Two tenants (Tenant A and Tenant B) with their own sets of users, roles, and policies.
When: A user from Tenant A attempts to query for resources belonging to Tenant B.
Then:
  - The query returns an empty result set or a "not found" error.
  - No data from Tenant B is ever visible to Tenant A.
  - This must be tested for all tables: `persons`, `employees`, `users`, `roles`, `permissions`, `policies`, etc.
```
```
Test ID: IAM-REPO-003
Description: Verify that creating a resource without a tenant context fails.
Given: A request to create a User or Role.
When: The database query is executed without `app.current_tenant_id` being set.
Then:
  - The database throws a foreign key violation or a specific RLS-related error.
  - No record is created.
```

### Complex Queries
```
Test ID: IAM-REPO-004
Description: Test retrieval of a user's effective permissions, including inherited roles.
Given: A user with a directly assigned role, which inherits from a parent role.
When: Querying for the user's effective permissions.
Then:
  - The result set includes permissions from both the direct and the inherited roles.
  - The query correctly navigates the role hierarchy.
```

## Service Layer Tests

### Authentication (authn) Service
```
Test ID: IAM-SVC-001
Description: Test successful user authentication.
Given: An active user with a known password.
When: Calling `authn.Service.Authenticate` with correct credentials.
Then:
  - An authentication token (e.g., JWT) is returned.
  - `last_login_at` timestamp is updated for the user.
  - `failed_login_attempts` is reset to 0.
  - A new session is created in `user_sessions`.
```
```
Test ID: IAM-SVC-002
Description: Test failed user authentication and lockout mechanism.
Given: An active user with a known password.
When: Calling `authn.Service.Authenticate` with an incorrect password multiple times.
Then:
  - `failed_login_attempts` is incremented on each failure.
  - After the configured number of attempts, `account_status` is set to 'LOCKED'.
  - `lockout_until` is set to a future timestamp.
  - Subsequent login attempts fail immediately with a "locked" error, even with the correct password.
```
```
Test ID: IAM-SVC-003
Description: Test MFA enablement and verification.
Given: A user who has not enabled MFA.
When: The user calls `authn.Service.EnableMFA`.
Then:
  - A secret key and setup information (e.g., QR code) are generated.
  - The user's `mfa_enabled` flag is set to true after verification.
  - Subsequent logins require a valid MFA token.
```

### Authorization (authz) Service
```
Test ID: IAM-SVC-004
Description: Test basic ABAC permission evaluation (ALLOW).
Given: A policy that allows access when specific attributes match.
When: `authz.Service.EvaluatePermission` is called with a context that satisfies the policy.
Then:
  - The decision returned is `ALLOW`.
  - The response includes the ID of the policy that granted access.
```
```
Test ID: IAM-SVC-005
Description: Test ABAC permission evaluation (DENY).
Given: A policy that denies access.
When: `authz.Service.EvaluatePermission` is called with a context that matches the DENY policy's target.
Then:
  - The decision returned is `DENY`.
  - The `Deny-Overrides` combining algorithm ensures this result even if an ALLOW policy also matches.
```
```
Test ID: IAM-SVC-006
Description: Test ABAC evaluation with dynamic attributes (user vs. resource).
Given: A policy allowing managers to view records of employees in their own department (`user.department_id == resource.department_id`).
When: A manager requests a record from their department.
Then:
  - The decision is `ALLOW`.
When: The same manager requests a record from another department.
Then:
  - The decision is `DENY`.
```
```
Test ID: IAM-SVC-007
Description: Test ABAC evaluation with environmental attributes.
Given: A policy allowing access only during business hours (`environment.time_of_day` between 09:00-17:00).
When: A user requests access at 14:00.
Then:
  - The decision is `ALLOW`.
When: The same user requests access at 20:00.
Then:
  - The decision is `DENY`.
```

### Policy (policy) Service
```
Test ID: IAM-SVC-008
Description: Test policy creation and validation.
Given: A request to create a new policy.
When: `policy.Service.CreatePolicy` is called with an invalid rule structure (e.g., malformed JSON).
Then:
  - The service returns a validation error.
  - No policy is created in the database.
```
```
Test ID: IAM-SVC-009
Description: Test policy cache invalidation.
Given: A policy is cached by the ABAC engine.
When: `policy.Service.UpdatePolicy` is called for that policy.
Then:
  - The service invalidates the relevant cache entries.
  - The next evaluation for that policy fetches the updated version from the database.
```

## API Integration Tests

```
Test ID: IAM-API-001
Description: Test that API endpoints require a valid JWT token.
Given: Any protected IAM API endpoint (e.g., GET /api/v1/users/{user_id}).
When: A request is made without a valid `Authorization` header.
Then:
  - The API returns a `401 Unauthorized` status code.
```
```
Test ID: IAM-API-002
Description: Test that API endpoints enforce tenant isolation via the `X-Tenant-ID` header.
Given: A user with a valid JWT for Tenant A.
When: The user makes a request to GET /api/v1/users/{user_id} for a user in Tenant B, while passing `X-Tenant-ID` for Tenant A.
Then:
  - The API returns a `404 Not Found` or `403 Forbidden` status code, not the user data.
```
```
Test ID: IAM-API-003
Description: Test the user creation endpoint with valid data.
Given: A valid request body for creating a new user.
When: A POST request is made to `/api/v1/users`.
Then:
  - The API returns a `201 Created` status code.
  - The response body contains the newly created user object.
  - The `Location` header points to the new user's resource URL.
```

## Hybrid Access Control (RBAC + ABAC) Tests

```
Test ID: IAM-HYBRID-001
Description: Test that a user's role grants them a base set of permissions.
Given: A user with the "Accountant" role, which has permission to "read" financial reports.
When: An ABAC evaluation is performed for the user to "read" a financial report with no other restrictive attributes.
Then:
  - The decision is `ALLOW` based on the user's role.
```
```
Test ID: IAM-HYBRID-002
Description: Test that an ABAC policy can override an RBAC permission.
Given: A user with the "Accountant" role (ALLOWS read), but a DENY policy is in place for accessing reports outside of business hours.
When: The user attempts to read a financial report at 21:00.
Then:
  - The decision is `DENY` because the more specific, attribute-based policy takes precedence.
```
```
Test ID: IAM-HYBRID-003
Description: Test dynamic role activation.
Given: A user has a "Project Manager" role that is only active when `user.is_on_project_x` is true.
When: The user's attributes are updated to set `is_on_project_x` to true.
Then:
  - A subsequent permission evaluation grants permissions associated with the "Project Manager" role.
```
```
Test ID: IAM-HYBRID-004
Description: Test temporary access via role delegation.
Given: User A delegates their "Auditor" role to User B for 48 hours.
When: User B attempts to access an audit resource within the 48-hour window.
Then:
  - The decision is `ALLOW`.
When: User B attempts the same access after 48 hours.
Then:
  - The decision is `DENY`.
```

## Multi-tenant Isolation Tests

```
Test ID: IAM-MULTI-TENANT-001
Description: Ensure creating a user in Tenant A does not affect Tenant B.
Given: Admin for Tenant A and Admin for Tenant B.
When: Admin A creates a new user "testuser@a.com".
Then:
  - Admin B cannot see, modify, or authenticate as "testuser@a.com".
  - A query for that user from Tenant B's context returns nothing.
```
```
Test ID: IAM-MULTI-TENANT-002
Description: Ensure roles and policies are strictly separated by tenant.
Given: Tenant A has a role "SuperAdmin" and Tenant B has a role with the same name "SuperAdmin".
When: A permission is added to Tenant A's "SuperAdmin" role.
Then:
  - The permissions for Tenant B's "SuperAdmin" role remain unchanged.
```

## Security Tests

```
Test ID: IAM-SEC-001
Description: Test for privilege escalation.
Given: A standard user without administrative privileges.
When: The user attempts to call an admin-only API endpoint (e.g., POST /api/v1/roles) by manipulating the request.
Then:
  - The API returns a `403 Forbidden` error.
  - The action is logged as a security event in the audit trail.
```
```
Test ID: IAM-SEC-002
Description: Test password policy enforcement.
Given: A password policy requiring 12 characters, uppercase, lowercase, number, and symbol.
When: A user attempts to change their password to "password123".
Then:
  - The service rejects the password change with a clear error message about the policy requirements.
```
```
Test ID: IAM-SEC-003
Description: Test session hijacking prevention.
Given: An active user session.
When: An attacker attempts to use the same session token from a different IP address or device.
Then:
  - The system detects the change in context (if IP/device binding is enabled).
  - The session is invalidated, and the user is forced to re-authenticate.
```

## Performance Tests

### Golden Path Testing

| ID | Group | Feature/Method | Preconditions (Given) | Action (When) | Expected (Then) | Edge/Negative | Trace (File:Func) |
|----|-------|----------------|------------------------|---------------|-----------------|---------------|-------------------|
| E2E-008 | Golden Path | HappyPathUserJourney | New organization setup | Complete user setup → role assignment → policy evaluation → resource access | Entire happy path completes without errors | Any step failure breaks the chain | internal/core/iam/e2e/golden_path_test.go::TestHappyPathUserJourney |
| E2E-009 | Golden Path | DailyOperations | Established system | Authentication → permission checks → resource access → logout | Typical daily operations work smoothly | Performance degradation, intermittent failures | internal/core/iam/e2e/golden_path_test.go::TestDailyOperations |
| E2E-010 | Golden Path | AdministrativeOperations | Admin user with full permissions | User management → role management → policy updates → audit review | Administrative operations complete successfully | Permission issues, data consistency problems | internal/core/iam/e2e/golden_path_test.go::TestAdministrativeOperations |
