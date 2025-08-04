package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/gen/user"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/identity"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// UserHandler handles user-related HTTP requests following the data flow pattern
type UserHandler struct {
	service identity.Service
	tracing tracing.TracingService
	metrics *metrics.MetricsService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service identity.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) *UserHandler {
	return &UserHandler{
		service: service,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateUser handles user creation following the data flow pattern
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.create_user",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users",
	})
	defer timer.Stop()

	// Log request
	logger.InfoContext(ctx, "Processing create user request",
		logger.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
		})

	// Parse and validate request
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Convert to service request with proper type handling
	serviceReq := &identity.CreateUserRequest{
		EntityID:   req.EntityID,
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		UserType:   req.UserType,
		PersonID:   req.PersonID,
		EmployeeID: req.EmployeeID,
	}

	// Handle optional pointer fields with defaults
	if req.AccountStatus != nil {
		serviceReq.AccountStatus = *req.AccountStatus
	} else {
		serviceReq.AccountStatus = "ACTIVE"
	}

	if req.SessionTimeoutMinutes != nil {
		serviceReq.SessionTimeoutMinutes = *req.SessionTimeoutMinutes
	} else {
		serviceReq.SessionTimeoutMinutes = 30
	}

	if req.MfaEnabled != nil {
		serviceReq.MfaEnabled = *req.MfaEnabled
	} else {
		serviceReq.MfaEnabled = false
	}

	// Call service layer
	user, err := h.service.RegisterNewUser(ctx, serviceReq)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to create user",
			logger.Fields{"error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case errors.Is(err, sharedErrors.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusCreated, h.convertUserToResponse(user))
}

// GetUser handles user retrieval by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.get_user",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Add user ID to span
	span.SetAttributes(attribute.String("user.id", userID.String()))

	// Call service layer
	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to get user",
			logger.Fields{"user_id": userID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, h.convertUserToResponse(user))
}

// ListUsers handles user listing with pagination
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.list_users",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users",
	})
	defer timer.Stop()

	// Parse query parameters
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	userType := c.Query("user_type")
	accountStatus := c.Query("account_status")

	// Add query parameters to span
	span.SetAttributes(
		attribute.Int("query.limit", limit),
		attribute.Int("query.offset", offset),
		attribute.String("query.user_type", userType),
		attribute.String("query.account_status", accountStatus),
	)

	// Build service request
	req := &identity.ListUsersRequest{
		Limit:  limit,
		Offset: offset,
	}

	if userType != "" {
		req.UserType = &userType
	}
	if accountStatus != "" {
		req.AccountStatus = &accountStatus
	}

	// Call service layer
	users, err := h.service.ListUsers(ctx, req)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to list users",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.convertUserToResponse(user)
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, ListUsersResponse{
		Users:  userResponses,
		Total:  len(users),
		Limit:  limit,
		Offset: offset,
	})
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.update_user",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse and validate request
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Convert to service request
	serviceReq := &identity.UpdateUserRequest{
		Username:              req.Username,
		Email:                 req.Email,
		UserType:              req.UserType,
		AccountStatus:         req.AccountStatus,
		SessionTimeoutMinutes: req.SessionTimeoutMinutes,
		MfaEnabled:            req.MfaEnabled,
	}

	// Call service layer
	user, err := h.service.UpdateUser(ctx, userID, serviceReq)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to update user",
			logger.Fields{"user_id": userID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, sharedErrors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case errors.Is(err, sharedErrors.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, h.convertUserToResponse(user))
}

// DeleteUser handles user soft deletion
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.delete_user",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Add user ID to span
	span.SetAttributes(attribute.String("user.id", userID.String()))

	// Call service layer (soft delete)
	err = h.service.DeleteUser(ctx, userID)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to delete user",
			logger.Fields{"user_id": userID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// AuthenticateUser handles user authentication
func (h *UserHandler) AuthenticateUser(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.authenticate_user",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/auth",
	})
	defer timer.Stop()

	// Parse and validate request
	var req AuthenticateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/auth",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Add identifier to span (but not password for security)
	span.SetAttributes(attribute.String("auth.identifier", req.Identifier))

	// Call service layer
	user, err := h.service.Authenticate(ctx, req.Identifier, req.Password)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/auth",
			"error_type": "authentication_error",
		})

		// Log error (without password)
		logger.ErrorContext(ctx, "Authentication failed",
			logger.Fields{"identifier": req.Identifier, "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case strings.Contains(err.Error(), "locked"):
			c.JSON(http.StatusLocked, gin.H{"error": "Account is locked"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
		return
	}

	// TODO: Generate JWT token (this would be handled by an auth service)
	token := "jwt-token-placeholder"
	expiresAt := time.Now().Add(8 * time.Hour)

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/auth",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, AuthenticateUserResponse{
		User:      h.convertUserToResponse(user),
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// GetUserRoles handles retrieving user roles
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.get_user_roles",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/roles",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/roles",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Add user ID to span
	span.SetAttributes(attribute.String("user.id", userID.String()))

	// Call service layer
	roles, err := h.service.GetUserRoles(ctx, userID)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/roles",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to get user roles",
			logger.Fields{"user_id": userID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Convert to response format
	roleResponses := make([]UserRoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = UserRoleResponse{
			ID:             role.ID.String(),
			UserID:         role.UserID.String(),
			RoleID:         role.RoleID.String(),
			EntityID:       role.EntityID.String(),
			AssignmentType: role.AssignmentType,
			AssignedAt:     role.AssignedAt.Format(time.RFC3339),
			IsActive:       role.IsActive,
		}

		if role.ExpiresAt != nil {
			expiry := role.ExpiresAt.Format(time.RFC3339)
			roleResponses[i].ExpiresAt = &expiry
		}
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/roles",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, GetUserRolesResponse{
		Roles: roleResponses,
	})
}

// UpdateUserPassword handles password updates
func (h *UserHandler) UpdateUserPassword(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.update_user_password",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/password",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/password",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse and validate request
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/password",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Add user ID to span
	span.SetAttributes(attribute.String("user.id", userID.String()))

	// Call service layer
	err = h.service.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/password",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to update user password",
			logger.Fields{"user_id": userID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/password",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// SearchUsers handles user search
func (h *UserHandler) SearchUsers(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.search_users",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/search",
	})
	defer timer.Stop()

	// Parse query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Add query parameters to span
	span.SetAttributes(
		attribute.String("search.query", query),
		attribute.Int("search.limit", limit),
		attribute.Int("search.offset", offset),
	)

	// Call service layer
	users, err := h.service.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/search",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to search users",
			logger.Fields{"query": query, "error": err.Error()})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.convertUserToResponse(user)
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/search",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, SearchUsersResponse{
		Users:  userResponses,
		Total:  len(users),
		Limit:  limit,
		Offset: offset,
		Query:  query,
	})
}

// Request and Response Types

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	EntityID               uuid.UUID  `json:"entity_id" binding:"required"`
	Username               string     `json:"username" binding:"required"`
	Email                  string     `json:"email" binding:"required,email"`
	Password               string     `json:"password" binding:"required,min=8"`
	UserType               string     `json:"user_type" binding:"required"`
	AccountStatus          *string    `json:"account_status,omitempty"`
	SessionTimeoutMinutes  *int32     `json:"session_timeout_minutes,omitempty"`
	MfaEnabled             *bool      `json:"mfa_enabled,omitempty"`
	PasswordExpirationDays *int32     `json:"password_expiration_days,omitempty"`
	MaxFailedLogins        *int32     `json:"max_failed_logins,omitempty"`
	PersonID               *uuid.UUID `json:"person_id,omitempty"`
	EmployeeID             *uuid.UUID `json:"employee_id,omitempty"`
	CreatedBy              *uuid.UUID `json:"created_by,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username               *string `json:"username,omitempty"`
	Email                  *string `json:"email,omitempty"`
	UserType               *string `json:"user_type,omitempty"`
	AccountStatus          *string `json:"account_status,omitempty"`
	SessionTimeoutMinutes  *int32  `json:"session_timeout_minutes,omitempty"`
	MfaEnabled             *bool   `json:"mfa_enabled,omitempty"`
	PasswordExpirationDays *int32  `json:"password_expiration_days,omitempty"`
	MaxFailedLogins        *int32  `json:"max_failed_logins,omitempty"`
}

// AuthenticateUserRequest represents the request to authenticate a user
type AuthenticateUserRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// UpdatePasswordRequest represents the request to update a user's password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID            string    `json:"id"`
	EntityID      string    `json:"entity_id"`
	PersonID      *string   `json:"person_id,omitempty"`
	EmployeeID    *string   `json:"employee_id,omitempty"`
	Username      *string   `json:"username,omitempty"`
	Email         string    `json:"email"`
	UserType      string    `json:"user_type"`
	AccountStatus *string   `json:"account_status,omitempty"`
	IsActive      bool      `json:"is_active"`
	LastLoginAt   *string   `json:"last_login_at,omitempty"`
	MfaEnabled    *bool     `json:"mfa_enabled,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserRoleResponse represents a user role in API responses
type UserRoleResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	RoleID         string  `json:"role_id"`
	EntityID       string  `json:"entity_id"`
	AssignmentType string  `json:"assignment_type"`
	AssignedAt     string  `json:"assigned_at"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	IsActive       bool    `json:"is_active"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// SearchUsersResponse represents the response for searching users
type SearchUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Query  string         `json:"query"`
}

// AuthenticateUserResponse represents the response for user authentication
type AuthenticateUserResponse struct {
	User      UserResponse `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
}

// GetUserRolesResponse represents the response for getting user roles
type GetUserRolesResponse struct {
	Roles []UserRoleResponse `json:"roles"`
}

// Helper functions

// convertUserToResponse converts a domain User to UserResponse
func (h *UserHandler) convertUserToResponse(u *identity.User) UserResponse {
	response := UserResponse{
		ID:        u.ID.String(),
		EntityID:  u.EntityID.String(),
		Email:     u.Email,
		UserType:  u.UserType,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.PersonID != nil {
		personID := u.PersonID.String()
		response.PersonID = &personID
	}
	if u.EmployeeID != nil {
		employeeID := u.EmployeeID.String()
		response.EmployeeID = &employeeID
	}
	if u.Username != "" {
		response.Username = &u.Username
	}
	status := string(u.AccountStatus)
	response.AccountStatus = &status
	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format(time.RFC3339)
		response.LastLoginAt = &lastLogin
	}
	response.MfaEnabled = &u.MfaEnabled

	return response
}

// ===== PERMISSION EVALUATION HANDLERS =====

// EvaluatePermission handles single permission evaluation
func (h *UserHandler) EvaluatePermission(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.evaluate_permission",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/permissions/evaluate",
	})
	defer timer.Stop()

	// Parse and validate request
	var req EvaluatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/permissions/evaluate",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ABAC functionality has been moved to the dedicated ABAC service
	// Users should use the /api/v1/abac/evaluate endpoint instead

	// Increment error counter
	h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
		"method":     c.Request.Method,
		"endpoint":   "/api/v1/users/permissions/evaluate",
		"error_type": "service_moved",
	})

	// Return error response indicating service has moved
	c.JSON(http.StatusGone, gin.H{
		"error":       "ABAC functionality has been moved to the dedicated ABAC service",
		"message":     "Please use /api/v1/abac/evaluate endpoint instead",
		"redirect_to": "/api/v1/abac/evaluate",
	})
	return
}

// BulkEvaluatePermissions handles bulk permission evaluation
func (h *UserHandler) BulkEvaluatePermissions(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.bulk_evaluate_permissions",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/permissions/bulk-evaluate",
	})
	defer timer.Stop()

	// Parse and validate request
	var req BulkEvaluatePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/permissions/bulk-evaluate",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ABAC functionality has been moved to the dedicated ABAC service
	// Users should use the /api/v1/abac/evaluate-bulk endpoint instead

	// Increment error counter
	h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
		"method":     c.Request.Method,
		"endpoint":   "/api/v1/users/permissions/bulk-evaluate",
		"error_type": "service_moved",
	})

	// Return error response indicating service has moved
	c.JSON(http.StatusGone, gin.H{
		"error":       "ABAC functionality has been moved to the dedicated ABAC service",
		"message":     "Please use /api/v1/abac/evaluate-bulk endpoint instead",
		"redirect_to": "/api/v1/abac/evaluate-bulk",
	})
	return
}

// GetUserEffectivePermissions handles getting user effective permissions
func (h *UserHandler) GetUserEffectivePermissions(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.get_user_effective_permissions",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/effective-permissions",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/effective-permissions",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Note: entity filtering not implemented in current service interface

	// Add user ID to span
	span.SetAttributes(attribute.String("user.id", userID.String()))

	// ABAC functionality has been moved to the dedicated ABAC service
	// Users should use the appropriate ABAC endpoints instead

	// Increment error counter
	h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
		"method":     c.Request.Method,
		"endpoint":   "/api/v1/users/{id}/effective-permissions",
		"error_type": "service_moved",
	})

	// Return error response indicating service has moved
	c.JSON(http.StatusGone, gin.H{
		"error":   "ABAC functionality has been moved to the dedicated ABAC service",
		"message": "Please use the appropriate ABAC endpoints instead",
		"user_id": userID.String(),
	})
	return
}

// ===== ROLE MANAGEMENT HANDLERS =====

// AssignUserRole handles role assignment to users
func (h *UserHandler) AssignUserRole(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.assign_user_role",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/roles",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/roles",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse and validate request
	var req AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/roles",
			"error_type": "validation_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Call service layer
	err = h.service.AssignUserRole(ctx, userID, req.RoleID, req.EntityID)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{id}/roles",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to assign user role",
			logger.Fields{"user_id": userID.String(), "role_id": req.RoleID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, sharedErrors.ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		case strings.Contains(err.Error(), "already assigned"):
			c.JSON(http.StatusConflict, gin.H{"error": "Role already assigned"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{id}/roles",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusCreated, AssignUserRoleResponse{
		Success:      true,
		AssignmentID: uuid.New().String(), // TODO: Get actual assignment ID from service
		Message:      "Role assigned successfully",
	})
}

// RevokeUserRole handles role revocation from users
func (h *UserHandler) RevokeUserRole(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.revoke_user_role",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{user_id}/roles/{role_id}",
	})
	defer timer.Stop()

	// Get user ID from path
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{user_id}/roles/{role_id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid user ID",
			logger.Fields{"user_id": userIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get role ID from path
	roleIDStr := c.Param("role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{user_id}/roles/{role_id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid role ID",
			logger.Fields{"role_id": roleIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Get entity ID from query parameter
	entityIDStr := c.Query("entity_id")
	if entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id query parameter is required"})
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{user_id}/roles/{role_id}",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid entity ID",
			logger.Fields{"entity_id": entityIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	// Call service layer
	err = h.service.RevokeUserRole(ctx, userID, roleID, entityID)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/{user_id}/roles/{role_id}",
			"error_type": "business_error",
		})

		// Log error
		logger.ErrorContext(ctx, "Failed to revoke user role",
			logger.Fields{"user_id": userID.String(), "role_id": roleID.String(), "error": err.Error()})

		// Handle specific errors
		switch {
		case errors.Is(err, sharedErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, sharedErrors.ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		case strings.Contains(err.Error(), "assignment not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": "Role assignment not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/{user_id}/roles/{role_id}",
		"status":   "success",
	})

	// Success response
	c.JSON(http.StatusOK, RevokeUserRoleResponse{
		Success: true,
		Message: "Role revoked successfully",
	})
}

// GetRoleHierarchy handles getting role hierarchy
func (h *UserHandler) GetRoleHierarchy(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.get_role_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/roles/{role_id}/hierarchy",
	})
	defer timer.Stop()

	// Get role ID from path
	roleIDStr := c.Param("role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/roles/{role_id}/hierarchy",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid role ID",
			logger.Fields{"role_id": roleIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// ABAC functionality has been moved to the dedicated ABAC service
	// This functionality is no longer available through the user service

	// Increment error counter
	h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
		"method":     c.Request.Method,
		"endpoint":   "/api/v1/users/roles/{role_id}/hierarchy",
		"error_type": "service_moved",
	})

	// Return error response indicating service has moved
	c.JSON(http.StatusGone, gin.H{
		"error":   "ABAC functionality has been moved to the dedicated ABAC service",
		"message": "This functionality is no longer available through the user service",
		"role_id": roleID.String(),
	})
	return
}

// TestPolicy handles ABAC policy testing
func (h *UserHandler) TestPolicy(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.test_policy",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/users/policies/{policy_id}/test",
	})
	defer timer.Stop()

	// Get policy ID from path
	policyIDStr := c.Param("policy_id")
	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/policies/{policy_id}/test",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid policy ID",
			logger.Fields{"policy_id": policyIDStr, "error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	// Parse and validate request
	var req TestPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/users/policies/{policy_id}/test",
			"error_type": "validation_error",
		})

		logger.ErrorContext(ctx, "Invalid request payload",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ABAC functionality has been moved to the dedicated ABAC service
	// Users should use the appropriate ABAC endpoints for policy testing

	// Increment error counter
	h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
		"method":     c.Request.Method,
		"endpoint":   "/api/v1/users/policies/{policy_id}/test",
		"error_type": "service_moved",
	})

	// Return error response indicating service has moved
	c.JSON(http.StatusGone, gin.H{
		"error":     "ABAC functionality has been moved to the dedicated ABAC service",
		"message":   "Please use the appropriate ABAC endpoints for policy testing",
		"policy_id": policyID.String(),
	})
	return
}

// ===== ADDITIONAL REQUEST/RESPONSE TYPES =====

// Permission Evaluation Types
type EvaluatePermissionRequest struct {
	UserID       uuid.UUID      `json:"user_id" binding:"required"`
	ResourceName string         `json:"resource_name" binding:"required"`
	ActionName   string         `json:"action_name" binding:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

type PermissionEvaluationResponse struct {
	Allowed          bool     `json:"allowed"`
	PolicyDecisions  []string `json:"policy_decisions"`
	EffectiveRoles   []string `json:"effective_roles"`
	EvaluationTimeMS int32    `json:"evaluation_time_ms"`
	CacheHit         bool     `json:"cache_hit"`
}

type BulkEvaluatePermissionsRequest struct {
	Requests []EvaluatePermissionRequest `json:"requests" binding:"required"`
}

type BulkEvaluatePermissionsResponse struct {
	Results               []PermissionEvaluationResponse `json:"results"`
	TotalEvaluationTimeMS int32                          `json:"total_evaluation_time_ms"`
}

type EffectivePermissionResponse struct {
	PermissionID   string `json:"permission_id"`
	PermissionName string `json:"permission_name"`
	ResourceName   string `json:"resource_name"`
	ActionName     string `json:"action_name"`
	Effect         string `json:"effect"`
	GrantedByRole  string `json:"granted_by_role"`
	AssignmentType string `json:"assignment_type"`
	EntityID       string `json:"entity_id"`
}

type GetUserEffectivePermissionsResponse struct {
	Permissions []EffectivePermissionResponse `json:"permissions"`
	TotalCount  int                           `json:"total_count"`
}

// Role Management Types
type AssignUserRoleRequest struct {
	RoleID         uuid.UUID  `json:"role_id" binding:"required"`
	EntityID       uuid.UUID  `json:"entity_id" binding:"required"`
	AssignmentType string     `json:"assignment_type" binding:"required"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Justification  string     `json:"justification,omitempty"`
}

type AssignUserRoleResponse struct {
	Success      bool   `json:"success"`
	AssignmentID string `json:"assignment_id"`
	Message      string `json:"message"`
}

type RevokeUserRoleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RoleHierarchyResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	DisplayName      string  `json:"display_name"`
	Level            int32   `json:"level"`
	ParentRoleID     *string `json:"parent_role_id,omitempty"`
	ChildrenCount    int32   `json:"children_count"`
	PermissionsCount int32   `json:"permissions_count"`
}

type GetRoleHierarchyResponse struct {
	Roles      []RoleHierarchyResponse `json:"roles"`
	TotalDepth int32                   `json:"total_depth"`
}

// Policy Testing Types
type TestPolicyRequest struct {
	UserID       uuid.UUID      `json:"user_id" binding:"required"`
	ResourceName string         `json:"resource_name" binding:"required"`
	ActionName   string         `json:"action_name" binding:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

type TestPolicyResponse struct {
	PolicyID      string         `json:"policy_id"`
	PolicyName    string         `json:"policy_name"`
	Effect        string         `json:"effect"`
	TargetMatches bool           `json:"target_matches"`
	RuleResult    bool           `json:"rule_result"`
	Details       map[string]any `json:"details"`
}

// UserGoaHandler implements the GOA user service following the data flow pattern
type UserGoaHandler struct {
	userService              identity.Service
	accessRequestService     request.AccessRequestService
	conditionalAccessService conditional.ConditionalAccessService
	analyticsService         analytics.UserAnalyticsService
	tracing                  tracing.TracingService
	metrics                  *metrics.MetricsService
}

// NewUserGoaHandler creates a new GOA user handler following Clean Architecture pattern
func NewUserGoaHandler(userSvc identity.Service, accessSvc request.AccessRequestService, conditionalSvc conditional.ConditionalAccessService, analyticsSvc analytics.UserAnalyticsService, tracing tracing.TracingService, metrics *metrics.MetricsService) user.Service {
	return &UserGoaHandler{
		userService:              userSvc,
		accessRequestService:     accessSvc,
		conditionalAccessService: conditionalSvc,
		analyticsService:         analyticsSvc,
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

	// TODO: Integrate with existing user creation logic using h.userService.CreateUser
	userResult := &user.User{
		ID:        "mock-user-id",
		Username:  p.Username,
		Email:     p.Email,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		UserType:  p.UserType,
		Status:    "ACTIVE",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("user_create_total", metrics.Fields{
		"user_type": p.UserType,
	})
	span.SetAttributes(attribute.String("result.user_id", userResult.ID))

	return userResult, "default", nil
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

	// TODO: Integrate with existing user retrieval logic using h.userService.GetUserByID
	mockUsername := "mockuser"
	userResult := &user.User{
		ID:        p.ID,
		Username:  &mockUsername,
		Email:     "mock@example.com",
		FirstName: "Mock",
		LastName:  "User",
		UserType:  "INTERNAL",
		Status:    "ACTIVE",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
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

	// TODO: Integrate with existing user listing logic using h.userService.ListUsers
	result := &user.ListResult{
		Data:       []*user.User{},
		Pagination: &user.PaginationMeta{CurrentPage: 1, PageSize: 20, TotalItems: 0, TotalPages: 0, HasNext: false, HasPrev: false},
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
