package authz

import (
	"fmt"

	"github.com/niiniyare/erp/internal/shared/errors"
)

// ─── ERROR DEFINITIONS ─────────────────────────────────────────────────────

var (
	// Permission evaluation errors
	ErrPermissionEvaluationFailed = errors.NewBusinessError("PERMISSION_EVALUATION_FAILED", "Permission evaluation failed")
	ErrBulkEvaluationFailed       = errors.NewBusinessError("BULK_EVALUATION_FAILED", "Bulk permission evaluation failed")
	ErrInvalidEvaluationRequest   = errors.NewBusinessError("INVALID_EVALUATION_REQUEST", "Invalid permission evaluation request")

	// User permission errors
	ErrGetUserPermissionsFailed     = errors.NewBusinessError("GET_USER_PERMISSIONS_FAILED", "Failed to get user permissions")
	ErrCalculateRoleHierarchyFailed = errors.NewBusinessError("CALCULATE_ROLE_HIERARCHY_FAILED", "Failed to calculate role hierarchy")
	ErrUserNotFound                 = errors.NewBusinessError("USER_NOT_FOUND", "User not found")

	// Access request errors
	ErrCreateAccessRequestFailed  = errors.NewBusinessError("CREATE_ACCESS_REQUEST_FAILED", "Failed to create access request")
	ErrGetAccessRequestFailed     = errors.NewBusinessError("GET_ACCESS_REQUEST_FAILED", "Failed to get access request")
	ErrProcessAccessRequestFailed = errors.NewBusinessError("PROCESS_ACCESS_REQUEST_FAILED", "Failed to process access request")
	ErrListAccessRequestsFailed   = errors.NewBusinessError("LIST_ACCESS_REQUESTS_FAILED", "Failed to list access requests")
	ErrAccessRequestNotFound      = errors.NewBusinessError("ACCESS_REQUEST_NOT_FOUND", "Access request not found")
	ErrInvalidAccessRequestStatus = errors.NewBusinessError("INVALID_ACCESS_REQUEST_STATUS", "Invalid access request status")

	// Approval workflow errors
	ErrCreateWorkflowFailed         = errors.NewBusinessError("CREATE_WORKFLOW_FAILED", "Failed to create approval workflow")
	ErrGetWorkflowFailed            = errors.NewBusinessError("GET_WORKFLOW_FAILED", "Failed to get approval workflow")
	ErrUpdateWorkflowFailed         = errors.NewBusinessError("UPDATE_WORKFLOW_FAILED", "Failed to update approval workflow")
	ErrWorkflowNotFound             = errors.NewBusinessError("WORKFLOW_NOT_FOUND", "Approval workflow not found")
	ErrInvalidWorkflowConfiguration = errors.NewBusinessError("INVALID_WORKFLOW_CONFIG", "Invalid workflow configuration")

	// Conditional access errors
	ErrEvaluateConditionalAccessFailed = errors.NewBusinessError("EVALUATE_CONDITIONAL_ACCESS_FAILED", "Failed to evaluate conditional access")
	ErrCreateConditionalPolicyFailed   = errors.NewBusinessError("CREATE_CONDITIONAL_POLICY_FAILED", "Failed to create conditional access policy")
	ErrUpdateConditionalPolicyFailed   = errors.NewBusinessError("UPDATE_CONDITIONAL_POLICY_FAILED", "Failed to update conditional access policy")
	ErrConditionalPolicyNotFound       = errors.NewBusinessError("CONDITIONAL_POLICY_NOT_FOUND", "Conditional access policy not found")
	ErrInvalidConditionalPolicy        = errors.NewBusinessError("INVALID_CONDITIONAL_POLICY", "Invalid conditional access policy")

	// Permission management errors
	ErrGrantPermissionFailed      = errors.NewBusinessError("GRANT_PERMISSION_FAILED", "Failed to grant permission")
	ErrRevokePermissionFailed     = errors.NewBusinessError("REVOKE_PERMISSION_FAILED", "Failed to revoke permission")
	ErrListUserPermissionsFailed  = errors.NewBusinessError("LIST_USER_PERMISSIONS_FAILED", "Failed to list user permissions")
	ErrPermissionNotFound         = errors.NewBusinessError("PERMISSION_NOT_FOUND", "Permission not found")
	ErrInvalidPermissionOperation = errors.NewBusinessError("INVALID_PERMISSION_OPERATION", "Invalid permission operation")

	// Decision history and audit errors
	ErrGetDecisionHistoryFailed = errors.NewBusinessError("GET_DECISION_HISTORY_FAILED", "Failed to get decision history")
	ErrDecisionHistoryNotFound  = errors.NewBusinessError("DECISION_HISTORY_NOT_FOUND", "Decision history not found")

	// Cache management errors
	ErrInvalidateUserCacheFailed   = errors.NewBusinessError("INVALIDATE_USER_CACHE_FAILED", "Failed to invalidate user cache")
	ErrInvalidatePolicyCacheFailed = errors.NewBusinessError("INVALIDATE_POLICY_CACHE_FAILED", "Failed to invalidate policy cache")
	ErrGetCacheStatisticsFailed    = errors.NewBusinessError("GET_CACHE_STATISTICS_FAILED", "Failed to get cache statistics")
	ErrCacheOperationFailed        = errors.NewBusinessError("CACHE_OPERATION_FAILED", "Cache operation failed")

	// Service delegation errors
	ErrABACServiceUnavailable   = errors.NewBusinessError("ABAC_SERVICE_UNAVAILABLE", "ABAC service is unavailable")
	ErrAccessServiceUnavailable = errors.NewBusinessError("ACCESS_SERVICE_UNAVAILABLE", "Access service is unavailable")
	ErrServiceDelegationFailed  = errors.NewBusinessError("SERVICE_DELEGATION_FAILED", "Service delegation failed")

	// Type conversion errors
	ErrTypeConversionFailed = errors.NewBusinessError("TYPE_CONVERSION_FAILED", "Type conversion failed")
	ErrInvalidDataFormat    = errors.NewBusinessError("INVALID_DATA_FORMAT", "Invalid data format")
	ErrMissingRequiredField = errors.NewBusinessError("MISSING_REQUIRED_FIELD", "Missing required field")
)

// ─── ERROR WRAPPING FUNCTIONS ──────────────────────────────────────────────

// WrapPermissionEvaluationError wraps permission evaluation errors with context
func WrapPermissionEvaluationError(err error, userID, resourceType, action string) error {
	if err == nil {
		return nil
	}

	result := ErrPermissionEvaluationFailed.
		WithDetail("user_id", userID).
		WithDetail("resource_type", resourceType).
		WithDetail("action", action)
	result.Err = err
	return result
}

// WrapBulkEvaluationError wraps bulk evaluation errors with context
func WrapBulkEvaluationError(err error, requestCount int) error {
	if err == nil {
		return nil
	}

	result := ErrBulkEvaluationFailed.
		WithDetail("request_count", fmt.Sprintf("%d", requestCount))
	result.Err = err
	return result
}

// WrapAccessRequestError wraps access request errors with context
func WrapAccessRequestError(err error, operation, requestID string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch operation {
	case "create":
		baseErr = ErrCreateAccessRequestFailed
	case "get":
		baseErr = ErrGetAccessRequestFailed
	case "process":
		baseErr = ErrProcessAccessRequestFailed
	case "list":
		baseErr = ErrListAccessRequestsFailed
	default:
		baseErr = errors.NewBusinessError("ACCESS_REQUEST_ERROR", fmt.Sprintf("Access request %s failed", operation))
	}

	result := baseErr.WithDetail("operation", operation)
	if requestID != "" {
		result = result.WithDetail("request_id", requestID)
	}
	result.Err = err
	return result
}

// WrapWorkflowError wraps workflow errors with context
func WrapWorkflowError(err error, operation, workflowID string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch operation {
	case "create":
		baseErr = ErrCreateWorkflowFailed
	case "get":
		baseErr = ErrGetWorkflowFailed
	case "update":
		baseErr = ErrUpdateWorkflowFailed
	default:
		baseErr = errors.NewBusinessError("WORKFLOW_ERROR", fmt.Sprintf("Workflow %s failed", operation))
	}

	result := baseErr.WithDetail("operation", operation)
	if workflowID != "" {
		result = result.WithDetail("workflow_id", workflowID)
	}
	result.Err = err
	return result
}

// WrapConditionalAccessError wraps conditional access errors with context
func WrapConditionalAccessError(err error, operation, policyID string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch operation {
	case "evaluate":
		baseErr = ErrEvaluateConditionalAccessFailed
	case "create":
		baseErr = ErrCreateConditionalPolicyFailed
	case "update":
		baseErr = ErrUpdateConditionalPolicyFailed
	default:
		baseErr = errors.NewBusinessError("CONDITIONAL_ACCESS_ERROR", fmt.Sprintf("Conditional access %s failed", operation))
	}

	result := baseErr.WithDetail("operation", operation)
	if policyID != "" {
		result = result.WithDetail("policy_id", policyID)
	}
	result.Err = err
	return result
}

// WrapPermissionManagementError wraps permission management errors with context
func WrapPermissionManagementError(err error, operation, userID, resourceType, action string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch operation {
	case "grant":
		baseErr = ErrGrantPermissionFailed
	case "revoke":
		baseErr = ErrRevokePermissionFailed
	case "list":
		baseErr = ErrListUserPermissionsFailed
	default:
		baseErr = errors.NewBusinessError("PERMISSION_MANAGEMENT_ERROR", fmt.Sprintf("Permission %s failed", operation))
	}

	result := baseErr.
		WithDetail("operation", operation).
		WithDetail("user_id", userID).
		WithDetail("resource_type", resourceType).
		WithDetail("action", action)
	result.Err = err
	return result
}

// WrapServiceDelegationError wraps service delegation errors with context
func WrapServiceDelegationError(err error, service, operation string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch service {
	case "abac":
		baseErr = ErrABACServiceUnavailable
	case "access":
		baseErr = ErrAccessServiceUnavailable
	default:
		baseErr = ErrServiceDelegationFailed
	}

	result := baseErr.
		WithDetail("service", service).
		WithDetail("operation", operation)
	result.Err = err
	return result
}

// WrapCacheError wraps cache operation errors with context
func WrapCacheError(err error, operation string, details map[string]string) error {
	if err == nil {
		return nil
	}

	var baseErr *errors.BusinessError
	switch operation {
	case "invalidate_user":
		baseErr = ErrInvalidateUserCacheFailed
	case "invalidate_policy":
		baseErr = ErrInvalidatePolicyCacheFailed
	case "get_statistics":
		baseErr = ErrGetCacheStatisticsFailed
	default:
		baseErr = ErrCacheOperationFailed
	}

	result := baseErr.WithDetail("operation", operation)
	for key, value := range details {
		result = result.WithDetail(key, value)
	}
	result.Err = err
	return result
}

// WrapTypeConversionError wraps type conversion errors with context
func WrapTypeConversionError(err error, fromType, toType, field string) error {
	if err == nil {
		return nil
	}

	result := ErrTypeConversionFailed.
		WithDetail("from_type", fromType).
		WithDetail("to_type", toType).
		WithDetail("field", field)
	result.Err = err
	return result
}

// ─── ERROR VALIDATION FUNCTIONS ────────────────────────────────────────────

// IsPermissionEvaluationError checks if error is a permission evaluation error
func IsPermissionEvaluationError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		return businessErr.Code == ErrPermissionEvaluationFailed.Code ||
			businessErr.Code == ErrBulkEvaluationFailed.Code ||
			businessErr.Code == ErrInvalidEvaluationRequest.Code
	}
	return false
}

// IsAccessRequestError checks if error is an access request error
func IsAccessRequestError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrCreateAccessRequestFailed.Code,
			ErrGetAccessRequestFailed.Code,
			ErrProcessAccessRequestFailed.Code,
			ErrListAccessRequestsFailed.Code,
			ErrAccessRequestNotFound.Code,
			ErrInvalidAccessRequestStatus.Code:
			return true
		}
	}
	return false
}

// IsWorkflowError checks if error is a workflow error
func IsWorkflowError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrCreateWorkflowFailed.Code,
			ErrGetWorkflowFailed.Code,
			ErrUpdateWorkflowFailed.Code,
			ErrWorkflowNotFound.Code,
			ErrInvalidWorkflowConfiguration.Code:
			return true
		}
	}
	return false
}

// IsConditionalAccessError checks if error is a conditional access error
func IsConditionalAccessError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrEvaluateConditionalAccessFailed.Code,
			ErrCreateConditionalPolicyFailed.Code,
			ErrUpdateConditionalPolicyFailed.Code,
			ErrConditionalPolicyNotFound.Code,
			ErrInvalidConditionalPolicy.Code:
			return true
		}
	}
	return false
}

// IsServiceDelegationError checks if error is a service delegation error
func IsServiceDelegationError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrABACServiceUnavailable.Code,
			ErrAccessServiceUnavailable.Code,
			ErrServiceDelegationFailed.Code:
			return true
		}
	}
	return false
}

// IsCacheError checks if error is a cache operation error
func IsCacheError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrInvalidateUserCacheFailed.Code,
			ErrInvalidatePolicyCacheFailed.Code,
			ErrGetCacheStatisticsFailed.Code,
			ErrCacheOperationFailed.Code:
			return true
		}
	}
	return false
}

// IsRetryableError checks if error is retryable
func IsRetryableError(err error) bool {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrABACServiceUnavailable.Code,
			ErrAccessServiceUnavailable.Code,
			ErrServiceDelegationFailed.Code,
			ErrCacheOperationFailed.Code:
			return true
		}
	}
	return false
}

// GetErrorCategory returns the category of the error for monitoring/alerting
func GetErrorCategory(err error) string {
	if businessErr, ok := err.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case ErrPermissionEvaluationFailed.Code,
			ErrBulkEvaluationFailed.Code,
			ErrInvalidEvaluationRequest.Code:
			return "permission_evaluation"
		case ErrCreateAccessRequestFailed.Code,
			ErrGetAccessRequestFailed.Code,
			ErrProcessAccessRequestFailed.Code,
			ErrListAccessRequestsFailed.Code,
			ErrAccessRequestNotFound.Code,
			ErrInvalidAccessRequestStatus.Code:
			return "access_request"
		case ErrCreateWorkflowFailed.Code,
			ErrGetWorkflowFailed.Code,
			ErrUpdateWorkflowFailed.Code,
			ErrWorkflowNotFound.Code,
			ErrInvalidWorkflowConfiguration.Code:
			return "approval_workflow"
		case ErrEvaluateConditionalAccessFailed.Code,
			ErrCreateConditionalPolicyFailed.Code,
			ErrUpdateConditionalPolicyFailed.Code,
			ErrConditionalPolicyNotFound.Code,
			ErrInvalidConditionalPolicy.Code:
			return "conditional_access"
		case ErrGrantPermissionFailed.Code,
			ErrRevokePermissionFailed.Code,
			ErrListUserPermissionsFailed.Code,
			ErrPermissionNotFound.Code,
			ErrInvalidPermissionOperation.Code:
			return "permission_management"
		case ErrABACServiceUnavailable.Code,
			ErrAccessServiceUnavailable.Code,
			ErrServiceDelegationFailed.Code:
			return "service_delegation"
		case ErrInvalidateUserCacheFailed.Code,
			ErrInvalidatePolicyCacheFailed.Code,
			ErrGetCacheStatisticsFailed.Code,
			ErrCacheOperationFailed.Code:
			return "cache_management"
		case ErrTypeConversionFailed.Code,
			ErrInvalidDataFormat.Code,
			ErrMissingRequiredField.Code:
			return "data_conversion"
		}
	}
	return "unknown"
}
