package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	coreUser "github.com/niiniyare/erp/internal/core/user"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// UserHandler handles user-related HTTP requests following the data flow pattern
type UserHandler struct {
	service coreUser.Service
	tracing *tracing.TracingService
	metrics *metrics.MetricsService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service coreUser.Service, tracing *tracing.TracingService, metrics *metrics.MetricsService) *UserHandler {
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
	
	// Convert to service request
	serviceReq := &coreUser.CreateUserRequest{
		EntityID:               req.EntityID,
		Username:               &req.Username,
		Email:                  req.Email,
		Password:               req.Password,
		UserType:               req.UserType,
		AccountStatus:          req.AccountStatus,
		SessionTimeoutMinutes:  req.SessionTimeoutMinutes,
		MfaEnabled:             req.MfaEnabled,
		PasswordExpirationDays: req.PasswordExpirationDays,
		MaxFailedLogins:        req.MaxFailedLogins,
		PersonID:               req.PersonID,
		EmployeeID:             req.EmployeeID,
		CreatedBy:              req.CreatedBy,
	}
	
	// Call service layer
	user, err := h.service.CreateUser(ctx, serviceReq)
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
		case errors.Is(err, errors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case errors.Is(err, errors.ErrUsernameAlreadyExists):
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
		case errors.Is(err, errors.ErrUserNotFound):
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
	req := &coreUser.ListUsersRequest{
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
	serviceReq := &coreUser.UpdateUserRequest{
		Username:               req.Username,
		Email:                  req.Email,
		UserType:               req.UserType,
		AccountStatus:          req.AccountStatus,
		SessionTimeoutMinutes:  req.SessionTimeoutMinutes,
		MfaEnabled:             req.MfaEnabled,
		PasswordExpirationDays: req.PasswordExpirationDays,
		MaxFailedLogins:        req.MaxFailedLogins,
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
		case errors.Is(err, errors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, errors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case errors.Is(err, errors.ErrUsernameAlreadyExists):
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

// UpdateUserPassword handles password updates
func (h *UserHandler) UpdateUserPassword(ctx context.Context, p *user.UpdateUserPasswordPayload) (*user.UpdateUserPasswordResult, error) {
	h.logger.Info("Updating user password", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	// Get current user to verify current password
	currentUser, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, user.MakeNotFound(err)
	}

	// Verify current password
	if currentUser.PasswordHash == nil || !h.userService.VerifyPassword(*currentUser.PasswordHash, p.CurrentPassword) {
		return nil, user.MakeUnauthorized(fmt.Errorf("current password is incorrect"))
	}

	// Update password
	err = h.userService.UpdateUserPassword(ctx, userID, p.NewPassword)
	if err != nil {
		h.logger.Error("Failed to update user password", "user_id", p.UserID, "error", err)
		return nil, user.MakeInternalError(err)
	}

	return &user.UpdateUserPasswordResult{
		Success: true,
		Message: "Password updated successfully",
	}, nil
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
	err = h.service.DeleteUser(ctx, userID, false)
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
		case errors.Is(err, errors.ErrUserNotFound):
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
	user, err := h.service.AuthenticateUser(ctx, req.Identifier, req.Password)
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
		case errors.Is(err, errors.ErrUserNotFound):
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
		case errors.Is(err, errors.ErrUserNotFound):
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
			ID:          role.ID.String(),
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			RoleType:    role.RoleType,
			AssignedAt:  role.AssignedAt.Format(time.RFC3339),
		}
		
		if role.ExpiresAt != nil {
			roleResponses[i].ExpiresAt = role.ExpiresAt.Format(time.RFC3339)
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
	err = h.service.UpdatePassword(ctx, userID, req.NewPassword)
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
		case errors.Is(err, errors.ErrUserNotFound):
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
	EntityID               uuid.UUID `json:"entity_id" binding:"required"`
	Username               string    `json:"username" binding:"required"`
	Email                  string    `json:"email" binding:"required,email"`
	Password               string    `json:"password" binding:"required,min=8"`
	UserType               string    `json:"user_type" binding:"required"`
	AccountStatus          *string   `json:"account_status,omitempty"`
	SessionTimeoutMinutes  *int32    `json:"session_timeout_minutes,omitempty"`
	MfaEnabled             *bool     `json:"mfa_enabled,omitempty"`
	PasswordExpirationDays *int32    `json:"password_expiration_days,omitempty"`
	MaxFailedLogins        *int32    `json:"max_failed_logins,omitempty"`
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
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description string  `json:"description"`
	RoleType    string  `json:"role_type"`
	AssignedAt  string  `json:"assigned_at"`
	ExpiresAt   string  `json:"expires_at,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// SearchUsersResponse represents the response for searching users
type SearchUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Query  string        `json:"query"`
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
func (h *UserHandler) convertUserToResponse(u *coreUser.User) UserResponse {
	response := UserResponse{
		ID:            u.ID.String(),
		EntityID:      u.EntityID.String(),
		Email:         u.Email,
		UserType:      u.UserType,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}

	if u.PersonID != nil {
		personID := u.PersonID.String()
		response.PersonID = &personID
	}
	if u.EmployeeID != nil {
		employeeID := u.EmployeeID.String()
		response.EmployeeID = &employeeID
	}
	if u.Username != nil {
		response.Username = u.Username
	}
	if u.AccountStatus != nil {
		response.AccountStatus = u.AccountStatus
	}
	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format(time.RFC3339)
		response.LastLoginAt = &lastLogin
	}
	if u.MfaEnabled != nil {
		response.MfaEnabled = u.MfaEnabled
	}

	return response
}