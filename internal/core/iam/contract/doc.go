// Package contract defines the stable IAM consumer interface.
//
// All non-IAM modules (billing, inventory, HR, etc.) interact with IAM
// exclusively through this package. It exposes:
//
//   - [SessionContext]: read-only identity + runtime metadata carrier
//   - [AuthService]: narrow auth entrypoints (login/logout/validate)
//   - [FromContext] / [WithContext]: Go context helpers for service-layer code
//   - [InjectSessionContext]: Fiber middleware that bridges Fiber locals to Go context
//
// # Architectural contract
//
// IAM is the ONLY identity + authorization authority. Modules that import
// this package:
//
//   - MUST NOT call authzService.Enforce() directly
//   - MUST NOT import internal/core/iam/service or internal/core/iam/repository
//   - MUST NOT query IAM database tables
//   - MAY assume the request is already authorized by IAM middleware
//   - MAY read [SessionContext] freely — it carries no permission data
//
// # Usage example
//
//	// In a service method:
//	sc, ok := contract.FromContext(ctx)
//	if !ok {
//	    return nil, errors.New("unauthenticated")
//	}
//	tenantID := sc.TenantID()
//	if sc.FeatureEnabled("billing.autopay") {
//	    // feature-gated logic
//	}
package contract
