package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/gen/user"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// UserGoaHandler implements the GOA user service following the data flow pattern
type UserGoaHandler struct {
	userService              identity.Service
	accessRequestService     request.AccessRequestService
	conditionalAccessService conditional.ConditionalAccessService
	analyticsService         analytics.UserAnalyticsService
	iamService               interface{} // TODO: Replace with proper IAM service when available
	tracing                  tracing.TracingService
	metrics                  *metrics.MetricsService
}

// NewUserGoaHandler creates a new GOA user handler following Clean Architecture pattern
func NewUserGoaHandler(userSvc identity.Service, accessSvc request.AccessRequestService, conditionalSvc conditional.ConditionalAccessService, analyticsSvc analytics.UserAnalyticsService, tracing tracing.TracingService, metrics *metrics.MetricsService) user.Service {
	// Create temporary IAM adapter for user service integration
	iamAdapter := &identityToIAMAdapter{identityService: userSvc}

	return &UserGoaHandler{
		userService:              userSvc,
		accessRequestService:     accessSvc,
		conditionalAccessService: conditionalSvc,
		analyticsService:         analyticsSvc,
		iamService:               iamAdapter,
		tracing:                  tracing,
		metrics:                  metrics,
	}
}

// Create creates a new user following the data flow pattern
func (h *UserGoaHandler) Create(ctx context.Context, p *user.CreateUserPayload) (*user.User, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.create",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.email", p.Email),
			attribute.String("user.type", string(p.UserType)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_create_duration", metrics.Fields{
		"operation": "create",
	})
	defer timer.Stop()

	// Integrate with existing user creation logic using h.userService.CreateUser
	createReq := &identity.CreateUserRequest{
		EntityID: uuid.New(),                          // TODO: Get from context
		Username: getStringValue(p.Username, p.Email), // Use email as username fallback
		Email:    p.Email,
		Password: getStringValue(p.Password, "TempPassword123!"), // Default temp password
		UserType: string(p.UserType),
	}

	domainUser, err := h.userService.RegisterNewUser(ctx, createReq)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("user_create_failed_total", metrics.Fields{
			"reason": "service_error",
		})

		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, "", user.MakeBadRequest(err)
	}

	// Convert domain user to GOA response
	userResult := &user.User{
		ID:        domainUser.ID.String(),
		Username:  &domainUser.Username,
		Email:     domainUser.Email,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		UserType:  p.UserType,
		Status:    user.AccountStatus(domainUser.AccountStatus),
		CreatedAt: domainUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: domainUser.UpdatedAt.Format(time.RFC3339),
	}

	h.metrics.IncrementCounter("user_create_total", metrics.Fields{
		"user_type": p.UserType,
	})
	span.SetAttributes(attribute.String("result.user_id", userResult.ID))

	return userResult, "default", nil
}

func (h *UserGoaHandler) AuthorizeAction(context.Context, *user.AuthorizeActionPayload) (res *user.AuthorizationResult, err error) {
	return nil, errors.New("NOT implemented")
}

// Get retrieves a user by ID following the data flow pattern
func (h *UserGoaHandler) Get(ctx context.Context, p *user.GetPayload) (*user.User, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.get",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_get_duration", metrics.Fields{
		"operation": "get",
	})
	defer timer.Stop()

	// Integrate with existing user retrieval logic using h.userService.GetUserByID
	userID, err := uuid.Parse(p.ID)
	if err != nil {
		h.metrics.IncrementCounter("user_get_failed_total", metrics.Fields{
			"reason": "invalid_id",
		})
		return nil, "", user.MakeBadRequest(fmt.Errorf("invalid user ID format"))
	}

	domainUser, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("user_get_failed_total", metrics.Fields{
			"reason": "not_found",
		})

		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, "", user.MakeNotFound(err)
	}

	userResult := &user.User{
		ID:        domainUser.ID.String(),
		Username:  &domainUser.Username,
		Email:     domainUser.Email,
		FirstName: "", // TODO: Get from Person entity if available
		LastName:  "", // TODO: Get from Person entity if available
		UserType:  user.UserType(domainUser.UserType),
		Status:    user.AccountStatus(domainUser.AccountStatus),
		CreatedAt: domainUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: domainUser.UpdatedAt.Format(time.RFC3339),
	}

	h.metrics.IncrementCounter("user_get_total", metrics.Fields{})
	return userResult, "default", nil
}

// List retrieves users with pagination following the data flow pattern
func (h *UserGoaHandler) List(ctx context.Context, p *user.ListPayload) (*user.ListResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.list",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_list_duration", metrics.Fields{
		"operation": "list",
	})
	defer timer.Stop()

	// Integrate with existing user listing logic using h.userService.ListUsers
	// Convert pagination parameters
	page := int(p.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(p.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Create list request for identity service
	listReq := &identity.ListUsersRequest{
		Limit:  pageSize,
		Offset: offset,
		// Note: ListUsersRequest might not have these fields, using basic pagination
	}

	// Get users from identity service
	domainUsers, err := h.userService.ListUsers(ctx, listReq)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("user_list_failed_total", metrics.Fields{
			"reason": "service_error",
		})

		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, user.MakeBadRequest(err)
	}

	// Convert users to GOA format
	var goaUsers []*user.User
	for _, u := range domainUsers {
		goaUsers = append(goaUsers, &user.User{
			ID:        u.ID.String(),
			Username:  &u.Username,
			Email:     u.Email,
			FirstName: "", // TODO: Get from Person entity if available
			LastName:  "", // TODO: Get from Person entity if available
			UserType:  user.UserType(u.UserType),
			Status:    user.AccountStatus(u.AccountStatus),
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	// Create pagination metadata
	totalUsers := uint(len(goaUsers)) // TODO: Get actual count from service
	hasMore := len(goaUsers) == pageSize
	hasNext := hasMore
	hasPrev := page > 1
	totalPages := uint((int(totalUsers) + pageSize - 1) / pageSize)

	paginationMeta := &user.PaginationMeta{
		CurrentPage: uint(page),
		PageSize:    uint(pageSize),
		TotalItems:  totalUsers,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}

	result := &user.ListResult{
		Data:       goaUsers,
		Pagination: paginationMeta,
	}

	h.metrics.IncrementCounter("user_list_total", metrics.Fields{})
	return result, nil
}

// Update updates an existing user following the data flow pattern
func (h *UserGoaHandler) Update(ctx context.Context, p *user.UpdateUserPayload) (*user.User, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.update",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_update_duration", metrics.Fields{
		"operation": "update",
	})
	defer timer.Stop()

	// TODO: Integrate with existing user update logic using h.userService.UpdateUser
	userResult := &user.User{
		ID:        p.ID,
		Username:  p.Username,
		Email:     *p.Email,
		FirstName: *p.FirstName,
		LastName:  *p.LastName,
		UserType:  *p.UserType,
		Status:    "ACTIVE",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("user_update_total", metrics.Fields{})
	return userResult, "default", nil
}

// Deactivate deactivates a user following the data flow pattern
func (h *UserGoaHandler) Deactivate(ctx context.Context, p *user.DeactivatePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.deactivate",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_deactivate_duration", metrics.Fields{
		"operation": "deactivate",
	})
	defer timer.Stop()

	logger.Info("User deactivate called", logger.Fields{
		"id": p.ID,
	})

	// TODO: Integrate with existing user deactivation logic using h.userService.DeleteUser
	h.metrics.IncrementCounter("user_deactivate_total", metrics.Fields{})
	return nil
}

// Permissions retrieves user permissions following the data flow pattern
func (h *UserGoaHandler) Permissions(ctx context.Context, p *user.PermissionsPayload) (*user.UserPermissions, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.permissions",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_permissions_duration", metrics.Fields{
		"operation": "permissions",
	})
	defer timer.Stop()

	// TODO: Integrate with existing user permissions logic using h.userService.GetUserRoles
	result := &user.UserPermissions{
		UserID:      p.ID,
		Roles:       []*user.RoleInfo{},
		Permissions: []*user.PermissionInfo{},
		ComputedAt:  "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("user_permissions_total", metrics.Fields{})
	return result, nil
}

// AssignRole assigns a role to a user following the data flow pattern
func (h *UserGoaHandler) AssignRole(ctx context.Context, p *user.AssignRolePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.assign_role",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.UserID),
			attribute.String("role.id", p.RoleID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_assign_role_duration", metrics.Fields{
		"operation": "assign_role",
	})
	defer timer.Stop()

	logger.Info("User assign role called", logger.Fields{
		"user_id": p.UserID,
		"role_id": p.RoleID,
	})

	// TODO: Integrate with existing role assignment logic using h.userService.AssignUserRole
	h.metrics.IncrementCounter("user_assign_role_total", metrics.Fields{})
	return nil
}

// RemoveRole removes a role from a user following the data flow pattern
func (h *UserGoaHandler) RemoveRole(ctx context.Context, p *user.RemoveRolePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "user.remove_role",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.UserID),
			attribute.String("role.id", p.RoleID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("user_remove_role_duration", metrics.Fields{
		"operation": "remove_role",
	})
	defer timer.Stop()

	logger.Info("User remove role called", logger.Fields{
		"user_id": p.UserID,
		"role_id": p.RoleID,
	})

	// TODO: Integrate with existing role removal logic using h.userService.RevokeUserRole
	h.metrics.IncrementCounter("user_remove_role_total", metrics.Fields{})
	return nil
}

// ABAC Methods Implementation

// GetAttributes gets user attributes for ABAC evaluation
func (h *UserGoaHandler) GetAttributes(ctx context.Context, p *user.GetAttributesPayload) (*user.UserAttributesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.get_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_get_attributes_duration", metrics.Fields{
		"operation": "get_attributes",
	})
	defer timer.Stop()

	// TODO: Implement ABAC attribute retrieval
	result := &user.UserAttributesResult{
		UserID:      p.ID,
		UserType:    "standard",
		RetrievedAt: "2024-01-01T00:00:00Z",
		Attributes: map[string]any{
			"department": "engineering",
			"role":       "developer",
			"level":      "senior",
		},
		DerivedAttributes: map[string]any{
			"access_level": "high",
		},
	}

	h.metrics.IncrementCounter("user_get_attributes_total", metrics.Fields{})
	return result, nil
}

// SetAttributes sets/updates user attributes for ABAC
func (h *UserGoaHandler) SetAttributes(ctx context.Context, p *user.SetUserAttributesPayload) (*user.SetUserAttributesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.set_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_set_attributes_duration", metrics.Fields{
		"operation": "set_attributes",
	})
	defer timer.Stop()

	// TODO: Implement ABAC attribute setting
	result := &user.SetUserAttributesResult{
		UserID:      p.ID,
		Operation:   "set_attributes",
		CompletedAt: "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("user_set_attributes_total", metrics.Fields{})
	return result, nil
}

// BulkUpdateAttributes bulk updates user attributes for ABAC
func (h *UserGoaHandler) BulkUpdateAttributes(ctx context.Context, p *user.BulkUpdateUserAttributesPayload) (*user.BulkUpdateUserAttributesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.bulk_update_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.Int("users.count", len(p.Updates)),
		))
	defer span.End()

	timer := h.metrics.Timer("user_bulk_update_attributes_duration", metrics.Fields{
		"operation": "bulk_update_attributes",
	})
	defer timer.Stop()

	// TODO: Implement bulk ABAC attribute updates
	result := &user.BulkUpdateUserAttributesResult{
		BulkUpdateID: "bulk-123",
		StartedAt:    "2024-01-01T00:00:00Z",
		CompletedAt:  "2024-01-01T00:00:01Z",
	}

	h.metrics.IncrementCounter("user_bulk_update_attributes_total", metrics.Fields{
		"count": len(p.Updates),
	})
	return result, nil
}

// CheckPermission checks if user has permission using ABAC context
func (h *UserGoaHandler) CheckPermission(ctx context.Context, p *user.CheckPermissionPayload) (*user.PermissionCheckResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.check_permission",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.UserID),
			attribute.String("action", p.Action),
		))
	defer span.End()

	timer := h.metrics.Timer("user_check_permission_duration", metrics.Fields{
		"operation": "check_permission",
	})
	defer timer.Stop()

	// TODO: Implement ABAC permission checking
	result := &user.PermissionCheckResult{
		Allowed:          true,
		Decision:         "permit",
		EvaluationTimeMs: 5,
	}

	h.metrics.IncrementCounter("user_check_permission_total", metrics.Fields{
		"action": p.Action,
	})
	return result, nil
}

// GetSessionAttributes gets user session attributes for ABAC
func (h *UserGoaHandler) GetSessionAttributes(ctx context.Context, p *user.GetSessionAttributesPayload) (*user.SessionAttributesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.get_session_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.UserID),
			attribute.String("session.id", p.SessionID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_get_session_attributes_duration", metrics.Fields{
		"operation": "get_session_attributes",
	})
	defer timer.Stop()

	// TODO: Implement session attribute retrieval
	result := &user.SessionAttributesResult{
		SessionID: p.SessionID,
		UserID:    p.UserID,
		Status:    "active",
		CreatedAt: "2024-01-01T00:00:00Z",
		Attributes: map[string]any{
			"ip_address":  "192.168.1.1",
			"location":    "office",
			"device_type": "laptop",
		},
	}

	h.metrics.IncrementCounter("user_get_session_attributes_total", metrics.Fields{})
	return result, nil
}

// SetSessionContext sets session context for user
func (h *UserGoaHandler) SetSessionContext(ctx context.Context, p *user.SetSessionContextPayload) (*user.SetSessionContextResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.set_session_context",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.UserID),
			attribute.String("session.id", p.SessionID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_set_session_context_duration", metrics.Fields{
		"operation": "set_session_context",
	})
	defer timer.Stop()

	// TODO: Implement session context setting
	result := &user.SetSessionContextResult{
		SessionID:      p.SessionID,
		Operation:      "set_session_context",
		CompletedAt:    "2024-01-01T00:00:00Z",
		ContextUpdated: true,
	}

	h.metrics.IncrementCounter("user_set_session_context_total", metrics.Fields{})
	return result, nil
}

// GetUserContext gets user context for ABAC
func (h *UserGoaHandler) GetUserContext(ctx context.Context, p *user.GetUserContextPayload) (*user.UserContextResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.get_user_context",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_get_user_context_duration", metrics.Fields{
		"operation": "get_user_context",
	})
	defer timer.Stop()

	// TODO: Implement user context retrieval
	result := &user.UserContextResult{
		UserID:      p.ID,
		RetrievedAt: "2024-01-01T00:00:00Z",
		UserAttributes: map[string]any{
			"department": "engineering",
			"role":       "developer",
		},
		DerivedAttributes: map[string]any{
			"access_level": "high",
		},
	}

	h.metrics.IncrementCounter("user_get_user_context_total", metrics.Fields{})
	return result, nil
}

// ValidateAttributes validates user attributes for ABAC compliance
func (h *UserGoaHandler) ValidateAttributes(ctx context.Context, p *user.ValidateAttributesPayload) (*user.AttributeValidationResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.validate_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_validate_attributes_duration", metrics.Fields{
		"operation": "validate_attributes",
	})
	defer timer.Stop()

	// TODO: Implement attribute validation
	result := &user.AttributeValidationResult{
		UserID:        p.ID,
		ValidationID:  "valid-123",
		ValidatedAt:   "2024-01-01T00:00:00Z",
		OverallStatus: "valid",
	}

	h.metrics.IncrementCounter("user_validate_attributes_total", metrics.Fields{})
	return result, nil
}

// RefreshAttributes refreshes user attributes from authoritative sources
func (h *UserGoaHandler) RefreshAttributes(ctx context.Context, p *user.RefreshAttributesPayload) (*user.RefreshAttributesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "user.refresh_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.id", p.ID),
		))
	defer span.End()

	timer := h.metrics.Timer("user_refresh_attributes_duration", metrics.Fields{
		"operation": "refresh_attributes",
	})
	defer timer.Stop()

	// TODO: Implement attribute refresh from external sources
	result := &user.RefreshAttributesResult{
		UserID:      p.ID,
		RefreshID:   "refresh-123",
		StartedAt:   "2024-01-01T00:00:00Z",
		CompletedAt: "2024-01-01T00:00:01Z",
	}

	h.metrics.IncrementCounter("user_refresh_attributes_total", metrics.Fields{})
	return result, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Helper functions
// ──────────────────────────────────────────────────────────────────────────────

// getStringValue returns the value of a string pointer or a default value
func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// getIntValue returns the value of an int pointer or a default value
func getIntValue(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
