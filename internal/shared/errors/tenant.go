package errors

import (
	"fmt"
	"net/http"
)

// Tenant-related errors
// func ErrTenantNotFound(tenantID string) *BusinessError {
// 	return NewBusinessError("TENANT_NOT_FOUND", "Tenant not found").
// 		WithDetail("tenant_id", tenantID).
// 		WithHTTPStatus(http.StatusNotFound).
// 		WithSuggestion("Verify the tenant ID is correct")
// }

func ErrTenantSuspended(tenantID string) *BusinessError {
	return NewBusinessError("TENANT_SUSPENDED", "Tenant account is suspended").
		WithDetail("tenant_id", tenantID).
		WithHTTPStatus(http.StatusForbidden).
		WithSuggestion("Contact support to reactivate your account")
}

func ErrTenantLimitExceeded(tenantID, limitType string, current, max int64) *BusinessError {
	return NewBusinessError("TENANT_LIMIT_EXCEEDED", fmt.Sprintf("Tenant %s limit exceeded", limitType)).
		WithDetail("tenant_id", tenantID).
		WithDetail("limit_type", limitType).
		WithDetail("current", current).
		WithDetail("maximum", max).
		WithHTTPStatus(http.StatusPaymentRequired).
		WithSuggestion("Upgrade your plan to increase limits")
}
