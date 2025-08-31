# Gin Handlers Inventory

## Executive Summary
This document inventories all current Gin framework handlers that need to be migrated to GOA framework.

**Total Gin Handlers Found**: 8 files  
**Total Endpoints**: 35+ routes  
**Handler Categories**: Health, CRUD APIs, Access Management, Analytics, Swagger UI

---

## 1. Health Check Handlers
**File**: `internal/api/handlers/health.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| GET | `/health` | `HealthHandler.Health` | ✅ Active |
| GET | `/ready` | `HealthHandler.Ready` | ✅ Active |

**Migration Priority**: High - Critical for monitoring
**Dependencies**: None

---

## 2. Entity Management Handlers  
**File**: `internal/api/handlers/entity.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| POST | `/api/v1/entities` | `EntityHandler.CreateEntity` | ✅ Active |
| GET | `/api/v1/entities` | `EntityHandler.ListEntities` | ✅ Active |
| GET | `/api/v1/entities/tree` | `EntityHandler.GetEntityTree` | ✅ Active |
| POST | `/api/v1/entities/sequence/next` | `EntityHandler.GetNextSequence` | ✅ Active |
| POST | `/api/v1/entities/sequence/reset` | `EntityHandler.ResetSequence` | ✅ Active |
| GET | `/api/v1/entities/:id` | `EntityHandler.GetEntity` | ✅ Active |
| PUT | `/api/v1/entities/:id` | `EntityHandler.UpdateEntity` | ✅ Active |
| DELETE | `/api/v1/entities/:id` | `EntityHandler.DeleteEntity` | ✅ Active |
| POST | `/api/v1/entities/:id/restore` | `EntityHandler.RestoreEntity` | ✅ Active |
| GET | `/api/v1/entities/:id/children` | `EntityHandler.GetEntityChildren` | ✅ Active |
| GET | `/api/v1/entities/:id/ancestors` | `EntityHandler.GetEntityAncestors` | ✅ Active |
| GET | `/api/v1/entities/:id/hierarchy` | `EntityHandler.GetEntityWithHierarchy` | ✅ Active |

**Migration Priority**: High - Core business functionality
**Dependencies**: Entity service, tracing, metrics

---

## 3. Tenant Management Handlers
**File**: `internal/api/handlers/tenant.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| POST | `/api/v1/tenants` | `TenantHandler.CreateTenant` | ✅ Active |
| GET | `/api/v1/tenants/:id` | `TenantHandler.GetTenant` | ✅ Active |
| PUT | `/api/v1/tenants/:id` | `TenantHandler.UpdateTenant` | ✅ Active |
| DELETE | `/api/v1/tenants/:id` | `TenantHandler.DeleteTenant` | ✅ Active |
| GET | `/api/v1/tenants` | `TenantHandler.ListTenants` | ✅ Active |
| GET | `/api/v1/tenants/subdomain/:subdomain` | `TenantHandler.GetTenantBySubdomain` | ✅ Active |

**Migration Priority**: High - Multi-tenancy core feature
**Dependencies**: Tenant service

---

## 4. User Management Handlers
**File**: `internal/api/handlers/user.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| POST | `/api/v1/users` | `UserHandler.CreateUser` | ✅ Active |
| GET | `/api/v1/users/:id` | `UserHandler.GetUser` | ✅ Active |
| PUT | `/api/v1/users/:id` | `UserHandler.UpdateUser` | ✅ Active |
| DELETE | `/api/v1/users/:id` | `UserHandler.DeleteUser` | ✅ Active |
| POST | `/api/v1/users/auth` | `UserHandler.AuthenticateUser` | ✅ Active |
| GET | `/api/v1/users/search` | `UserHandler.SearchUsers` | ✅ Active |
| PUT | `/api/v1/users/:id/password` | `UserHandler.UpdateUserPassword` | ✅ Active |
| GET | `/api/v1/users/:id/roles` | `UserHandler.GetUserRoles` | ✅ Active |

**Migration Priority**: High - Authentication and user management
**Dependencies**: Identity service, tracing, metrics

---

## 5. Access Request Handlers
**File**: `internal/api/handlers/access_request_handler.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| POST | `/api/v1/access-requests` | `AccessRequestHandler.CreateAccessRequest` | ✅ Active |
| GET | `/api/v1/access-requests/:id` | `AccessRequestHandler.GetAccessRequest` | ✅ Active |
| POST | `/api/v1/access-requests/:id/process` | `AccessRequestHandler.ProcessAccessRequest` | ✅ Active |
| DELETE | `/api/v1/access-requests/:id` | `AccessRequestHandler.RevokeAccessRequest` | ✅ Active |
| GET | `/api/v1/access-requests` | `AccessRequestHandler.ListAccessRequests` | ✅ Active |
| GET | `/api/v1/access-requests/stats` | `AccessRequestHandler.GetAccessRequestStats` | ✅ Active |
| POST | `/api/v1/conditional-access/evaluate` | `AccessRequestHandler.EvaluateConditionalAccess` | ✅ Active |
| POST | `/api/v1/conditional-access/rules` | `AccessRequestHandler.CreateConditionalAccessRule` | ✅ Active |

**Migration Priority**: High - ABAC and access control
**Dependencies**: Access request service, conditional access service, analytics service

---

## 6. User Analytics Handlers
**File**: `internal/api/handlers/access_request_handler.go` (continued)

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| GET | `/api/v1/analytics/users/:user_id/behavior` | `AccessRequestHandler.GetUserBehaviorAnalytics` | ✅ Active |
| GET | `/api/v1/analytics/users/:user_id/risk` | `AccessRequestHandler.GetUserRiskAssessment` | ✅ Active |
| GET | `/api/v1/analytics/users/:user_id/insights` | `AccessRequestHandler.GetUserPersonalizedInsights` | ✅ Active |
| POST | `/api/v1/analytics/users/:user_id/detect-anomalies` | `AccessRequestHandler.DetectUserAnomalies` | ✅ Active |

**Migration Priority**: Medium - Analytics features
**Dependencies**: Analytics service

---

## 7. Swagger UI Handler
**File**: `internal/api/handlers/swagger.go`

| Method | Route | Handler Function | Status |
|--------|-------|------------------|--------|
| GET | `/swagger-ui/*filepath` | `SwaggerHandler.ServeSwaggerUI` | ✅ Active |

**Migration Priority**: Medium - Documentation interface
**Dependencies**: Embedded static files

---

## 8. Router Infrastructure  
**File**: `internal/api/handlers/router.go`

**Middleware Used**:
- `gin.Recovery()` - Panic recovery
- `middleware.RequestLogger()` - HTTP request logging
- `middleware.TracingMiddleware(tracing)` - Distributed tracing
- `middleware.MetricsMiddleware(metrics)` - Prometheus metrics
- Custom CORS middleware - Cross-origin requests

**Route Groups**:
- `/api/v1/tenants/*` - Tenant management
- `/api/v1/entities/*` - Entity management  
- `/api/v1/users/*` - User management
- `/api/v1/access-requests/*` - Access control
- `/api/v1/conditional-access/*` - Conditional access
- `/api/v1/analytics/*` - User analytics

---

## Migration Complexity Analysis

### High Complexity (5-8 hours each)
- **Entity Handlers** - 12 endpoints, complex business logic
- **User Handlers** - 8 endpoints, authentication flows
- **Access Request Handlers** - 8+ endpoints, ABAC integration

### Medium Complexity (2-4 hours each)
- **Tenant Handlers** - 6 endpoints, straightforward CRUD
- **Analytics Handlers** - 4 endpoints, data aggregation

### Low Complexity (1-2 hours each)
- **Health Handlers** - 2 endpoints, simple responses
- **Swagger Handler** - 1 endpoint, static file serving

---

## Current GOA Coverage Check

**Existing GOA Services** (already implemented):
- ✅ ABAC Service - `/abac/*` routes
- ✅ Auth Service - `/auth/*` routes  
- ✅ Organization Service - `/api/v1/organizations/*` routes
- ✅ Tenant Service - `/api/v1/tenants/*` routes (GOA version exists!)
- ✅ User Service - `/api/v1/users/*` routes (GOA version exists!)
- ✅ OpenAPI Service - `/openapi.json` route

**Overlap Analysis**:
- **Tenant endpoints**: Both Gin and GOA versions exist - GOA should be used
- **User endpoints**: Both Gin and GOA versions exist - GOA should be used
- **Entities**: Only Gin version exists - needs GOA design
- **Access Requests**: Only Gin version exists - needs GOA design
- **Health/Swagger**: Only Gin versions exist - need GOA design

---

## Risk Assessment

### High Risk Areas
1. **Authentication flows** - Password updates, role management
2. **File uploads** - Entity attachments, user avatars
3. **Complex queries** - Search, analytics, hierarchy traversal
4. **CORS handling** - Cross-origin API access

### Low Risk Areas  
1. **Basic CRUD** - Standard create/read/update/delete operations
2. **Health checks** - Simple status endpoints
3. **Static file serving** - Swagger UI assets

---

## Dependencies Map

```
Gin Handlers Dependencies:
├── Services
│   ├── tenant.Service
│   ├── entity.Service  
│   ├── identity.Service
│   ├── request.AccessRequestService
│   ├── conditional.ConditionalAccessService
│   └── analytics.UserAnalyticsService
├── Infrastructure  
│   ├── tracing.TracingService
│   ├── metrics.MetricsService
│   └── middleware (RequestLogger, TracingMiddleware, MetricsMiddleware)
└── Static Assets
    └── internal/api/swagger/* (embedded files)
```

---

## Next Steps

### Phase 1.2: Design GOA Equivalents
1. **Entities** - Create `internal/api/design/entities.go`
2. **Access Requests** - Create `internal/api/design/access_requests.go`  
3. **Health** - Create `internal/api/design/health.go`
4. **Analytics** - Create `internal/api/design/analytics.go`

### Phase 1.3: Validation
- [ ] Test all current Gin endpoints are responding
- [ ] Document request/response formats for each endpoint
- [ ] Verify middleware behavior (logging, tracing, metrics)

---

**Document Version**: 1.0  
**Created**: 2025-08-04  
**Last Updated**: 2025-08-04  
**Total Endpoints to Migrate**: 35+