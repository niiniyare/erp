# IAM Restructuring Mapping Table

## Progress Tracking

| Test Case / Spec                | Implementation Location                           | Verified | Notes |
|---------------------------------|---------------------------------------------------|----------|-------|
| ABAC policy evaluation          | @internal/core/abac/policy_evaluation_engine.go  | ✅       | 1900+ lines mature engine |
| Attribute resolver              | @internal/core/abac/attribute_service.go         | ✅       |  attribute collection |
| Identity user lifecycle         | @internal/core/identity/service.go               | ✅       | 317+ lines mature service |
| Session management              | @internal/core/identity/service.go               | ✅       | Integrated in identity service |
| Role + permission assignment    | @internal/core/access/request/access_request_service.go | ✅ | 806+ lines workflow logic |
| Access decision enforcement     | @internal/core/abac/hybrid_evaluator.go          | ✅       | RBAC-ABAC hybrid evaluation |
| Audit logging                   | @internal/core/analytics/user_analytics_service.go | ✅    | 1300+ lines behavioral analysis |
| Access analytics reports        | @internal/core/analytics/user_analytics_service.go | ✅    | Rich analytics & reporting |

## Restructuring Status

### Authentication Domain (@internal/core/iam/authn/)
- [✅] User Management Service (wraps @internal/core/identity/)
- [✅] Authentication Service 
- [⚠️] Session Management (stubbed - not implemented in legacy service)
- [⚠️] Password & MFA Management (partial - MFA stubbed)

### Authorization Domain (@internal/core/iam/authz/)
- [ ] Permission Evaluation Service (wraps @internal/core/abac/)
- [ ] Access Request Workflows (wraps @internal/core/access/)
- [ ] RBAC + ABAC Hybrid Service

### Policy Domain (@internal/core/iam/policy/)
- [ ] Policy Management (wraps @internal/core/abac/activities/)
- [ ] Policy Evaluation Engine
- [ ] Policy Caching & Lifecycle

### Repository Layer (@internal/core/iam/repo/)
- [ ] User Repository
- [ ] Role Repository  
- [ ] Policy Repository
- [ ] Permission Repository

## Test Conversion Status

### Unit Tests Converted to testify.suite + table-driven
- [✅] @internal/core/iam/authn/user_management_test.go
- [✅] @internal/core/iam/authn/authentication_test.go  
- [✅] @internal/core/iam/authz/permission_test.go
- [✅] @internal/core/iam/authz/compatibility_test.go
- [✅] @internal/core/iam/policy/management_test.go
- [✅] @internal/core/iam/repo/user_test.go

### Legacy Tests Requiring Migration
- [ ] @internal/core/abac/*_test.go files
- [ ] @internal/core/identity/*_test.go files  
- [ ] @internal/core/access/*_test.go files
- [ ] @internal/core/analytics/*_test.go files

## Implementation Verification

### Completed
- [✅] Test design principles documented
- [✅] Fail-first test structure verified
- [✅] Table-driven testify.suite pattern established

### In Progress
- [ ] Legacy codebase audit
- [ ] Adapter service implementations
- [ ] Integration point verification

### Pending
- [ ] Performance benchmarking
- [ ] Contract testing (legacy vs new)
- [ ] End-to-end workflow testing