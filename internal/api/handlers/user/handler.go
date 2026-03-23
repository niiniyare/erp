package user

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/iam"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// UserHandler handles user-related HTTP requests with centralized error handling
type UserHandler struct {
	service   iam.UserService
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.Service
	validator *validator.Validate
}

// NewUserHandler creates a new user handler with dependencies
func NewUserHandler(
	userService iam.UserService,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *UserHandler {
	validator := validator.New()

	// Register custom validators
	registerCustomValidators(validator)

	return &UserHandler{
		service:   userService,
		logger:    logger,
		metrics:   metrics,
		tracer:    tracer,
		validator: validator,
	}
}

// ============================================================================
// CRUD OPERATIONS FOLLOWING 5-STEP PATTERN
// ============================================================================

// Create handles user creation (POST /api/v1/users)
// @Summary Create user
// @Description Creates a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body iam.CreateUserRequest true "User data"
// @Success 201 {object} User
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "user.Create")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Parse and validate request
	var req iam.CreateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Step 3: Delegate to service
	createdUser, err := h.service.RegisterNewUser(c.Context(), &req)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Convert to API response
	response := h.userToAPIResponse(createdUser)

	// Step 5: Return successful response
	span.SetAttributes(
		attribute.String("user.id", createdUser.ID.String()),
		attribute.String("user.email", createdUser.Email),
	)

	return h.Created(c, response)
}

// Get handles user retrieval by ID (GET /api/v1/users/:id)
// @Summary Get user by ID
// @Description Retrieves a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param view query string false "View type" Enums(default, detailed, security, public) default(default)
// @Success 200 {object} User
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) Get(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "user.Get")
	defer span.End()

	userID := c.Params("id")
	h.logger.InfoContext(ctx, "Retrieving user", logger.Fields{
		"user_id": userID,
		"method":  c.Method(),
		"path":    c.Path(),
	})

	// Step 2: Parse and Validate
	if userID == "" {
		return h.HandleError(c, errors.NewBusinessError("MISSING_USER_ID", "User ID is required").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Step 3: Extract Context
	view := c.Query("view", "default")

	// Step 4: Delegate to Service
	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_USER_ID", "Invalid user ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	userEntity, err := h.service.GetUserByID(ctx, userUUID)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Convert user to API response
	response := h.userToAPIResponse(userEntity)

	// Add view-specific fields
	if view == "detailed" {
		// For detailed view, include additional metadata
		response = h.enhanceDetailedView(response, userEntity)
	} else if view == "security" {
		// For security view, include security-specific fields
		response = h.enhanceSecurityView(response, userEntity)
	} else if view == "public" {
		// For public view, limit to safe fields only
		response = h.limitPublicView(response)
	}

	// Step 5: Return Response
	h.metrics.IncrementCounter("user_retrieved_total", metrics.Fields{
		"status": "success",
		"view":   view,
	})

	return h.Success(c, response)
}

// List handles user listing with pagination (GET /api/v1/users)
// @Summary List users
// @Description List users with pagination and filtering
// @Tags users
// @Accept json
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Param status query string false "Filter by account status" Enums(ACTIVE, INACTIVE, SUSPENDED, LOCKED, PENDING_VERIFICATION, ARCHIVED)
// @Param user_type query string false "Filter by user type" Enums(INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN, SYSTEM)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "user.List")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Extract pagination parameters and filters
	_, _ = h.ExtractPaginationParams(c)

	// Note: For now, we'll use basic listing. In a real implementation,
	// we would add filtering parameters to the service interface
	// status := c.Query("status")
	// userType := c.Query("user_type")

	// Step 3: Delegate to service
	// Note: The current service interface doesn't have a ListUsers method
	// This would need to be added to the iam.Service interface
	// For now, we'll return an error indicating this feature needs implementation
	return h.HandleError(c, errors.NewBusinessError("NOT_IMPLEMENTED", "User listing not yet implemented").
		WithHTTPStatus(501).
		WithCategory(errors.CategoryBusiness).
		WithSuggestion("Add ListUsers method to iam.Service interface"))
}

// Update handles user updates (PUT /api/v1/users/:id)
// @Summary Update user
// @Description Updates an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param user body iam.UpdateUserRequest true "User update data"
// @Success 200 {object} User
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "user.Update")
	defer span.End()

	userID := c.Params("id")
	h.logger.InfoContext(ctx, "Updating user", logger.Fields{
		"user_id": userID,
		"method":  c.Method(),
		"path":    c.Path(),
	})

	// Step 2: Parse and Validate
	var req iam.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_JSON", "Invalid JSON payload").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_USER_ID", "Invalid user ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		return h.HandleError(c, err)
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	updatedUser, err := h.service.UpdateUser(ctx, userUUID, &req)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Convert user to API response
	response := h.userToAPIResponse(updatedUser)

	// Step 5: Return Response
	return h.Success(c, response)
}

// Delete handles user deletion (DELETE /api/v1/users/:id)
// @Summary Delete user
// @Description Deletes a user (soft delete, sets status to ARCHIVED)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Success 204
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "user.Delete")
	defer span.End()

	userID := c.Params("id")
	h.logger.InfoContext(ctx, "Deleting user", logger.Fields{
		"user_id": userID,
		"method":  c.Method(),
		"path":    c.Path(),
	})

	// Step 2: Parse and Validate
	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_USER_ID", "Invalid user ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	if err := h.service.DeleteUser(ctx, userUUID); err != nil {
		return h.HandleError(c, err)
	}

	// Step 5: Return Response (204 No Content)
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// AUTHENTICATION ENDPOINTS
// ============================================================================

// Authenticate handles user authentication (POST /api/v1/users/authenticate)
// @Summary Authenticate user
// @Description Authenticates a user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body iam.AuthenticateRequest true "Authentication credentials"
// @Success 200 {object} User
// @Router /api/v1/users/authenticate [post]
func (h *UserHandler) Authenticate(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "user.Authenticate")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Parse and validate request
	var req iam.AuthenticateRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Step 3: Delegate to service
	userEntity, err := h.service.Authenticate(c.Context(), req.Identifier, req.Password)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Set security attributes
	span.SetAttributes(
		attribute.String("user.id", userEntity.ID.String()),
	)

	// Step 5: Return successful response
	return h.Success(c, h.userToAPIResponse(userEntity))
}

// ChangePassword handles password changes (POST /api/v1/users/:id/change-password)
// @Summary Change user password
// @Description Changes a user's password
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param request body iam.ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id}/change-password [post]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "user.ChangePassword")
	defer span.End()
	c.SetUserContext(ctx)

	userID := c.Params("id")

	// Step 2: Parse and validate request
	var req iam.ChangePasswordRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_USER_ID", "Invalid user ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Delegate to service
	if err := h.service.ChangePassword(c.Context(), userUUID, req.CurrentPassword, req.NewPassword); err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Set attributes
	span.SetAttributes(
		attribute.String("user.id", userUUID.String()),
	)

	// Step 5: Return successful response
	return h.Success(c, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// ============================================================================
// BASE HANDLER METHODS (from BaseHandler pattern)
// ============================================================================

// HandleError provides centralized error handling
func (h *UserHandler) HandleError(c *fiber.Ctx, err error) error {
	requestID := h.getRequestID(c)
	tenantID := h.getTenantID(c)

	// Convert to HTTP error using existing system
	httpErr := errors.ToHTTPError(err)
	httpErr.RequestID = requestID

	// Log the error with context
	h.logger.Error("Request error", logger.Fields{
		"error":      err.Error(),
		"request_id": requestID,
		"tenant_id":  tenantID,
		"method":     c.Method(),
		"path":       c.Path(),
		"ip":         c.IP(),
		"status":     httpErr.Status,
		"code":       httpErr.Code,
	})

	// Record metrics
	h.recordErrorMetrics(c, httpErr)

	return c.Status(httpErr.Status).JSON(httpErr)
}

// ValidateRequest validates request data
func (h *UserHandler) ValidateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return errors.NewBusinessError("INVALID_JSON", "Invalid JSON format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Ensure request body contains valid JSON")
	}

	if err := h.validator.Struct(req); err != nil {
		return err
	}

	return nil
}

// Success returns a standardized success response
func (h *UserHandler) Success(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// SuccessWithMeta returns success response with metadata
func (h *UserHandler) SuccessWithMeta(c *fiber.Ctx, data interface{}, meta map[string]interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta":       meta,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// Created returns a 201 response
func (h *UserHandler) Created(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.Status(fiber.StatusCreated).JSON(response)
}

// ExtractPaginationParams extracts pagination parameters
func (h *UserHandler) ExtractPaginationParams(c *fiber.Ctx) (offset, limit int) {
	offset = c.QueryInt("offset", 0)
	limit = c.QueryInt("limit", 20)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return offset, limit
}

// Helper methods
func (h *UserHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("requestid"); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

func (h *UserHandler) getTenantID(c *fiber.Ctx) string {
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		return tenantID.(string)
	}
	return ""
}

func (h *UserHandler) recordErrorMetrics(c *fiber.Ctx, httpErr *errors.HTTPError) {
	h.metrics.IncrementCounter("api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   string(rune(httpErr.Status)),
	})
}

func (h *UserHandler) recordSuccessMetrics(c *fiber.Ctx) {
	h.metrics.IncrementCounter("api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

// Custom validator registration
func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		return len(value) == 36 && value[8] == '-' && value[13] == '-' && value[18] == '-' && value[23] == '-'
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// User represents the API response structure for a user
type User struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id,omitempty"`
	EntityID      string                 `json:"entity_id,omitempty"`
	PersonID      string                 `json:"person_id,omitempty"`
	EmployeeID    string                 `json:"employee_id,omitempty"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username,omitempty"`
	UserType      string                 `json:"user_type"`
	Status        string                 `json:"account_status"`
	IsActive      bool                   `json:"is_active"`
	EmailVerified bool                   `json:"email_verified,omitempty"`
	Phone         string                 `json:"phone,omitempty"`
	PhoneVerified bool                   `json:"phone_verified,omitempty"`
	LastLoginAt   *string                `json:"last_login_at,omitempty"`
	LastLoginIP   string                 `json:"last_login_ip,omitempty"`
	MFAEnabled    bool                   `json:"mfa_enabled,omitempty"`
	MFAMethod     string                 `json:"mfa_method,omitempty"`
	Timezone      string                 `json:"timezone,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Roles         []string               `json:"roles,omitempty"`
	Permissions   []string               `json:"permissions,omitempty"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     *string                `json:"updated_at,omitempty"`
}

// userToAPIResponse converts a user entity to API user response
func (h *UserHandler) userToAPIResponse(userEntity *iam.User) *User {
	if userEntity == nil {
		return nil
	}

	apiUser := &User{
		ID:        userEntity.ID.String(),
		Email:     userEntity.Email,
		UserType:  "user", // Default user type
		Status:    string(userEntity.AccountStatus),
		IsActive:  userEntity.IsActive,
		CreatedAt: userEntity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Optional fields
	apiUser.TenantID = userEntity.TenantID.String()
	apiUser.EntityID = userEntity.EntityID.String()
	if userEntity.PersonID != nil {
		apiUser.PersonID = userEntity.PersonID.String()
	}
	if userEntity.EmployeeID != nil {
		apiUser.EmployeeID = userEntity.EmployeeID.String()
	}
	if userEntity.LastLoginAt != nil {
		lastLoginStr := userEntity.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		apiUser.LastLoginAt = &lastLoginStr
	}
	if userEntity.UserAttributes != nil {
		apiUser.Metadata = userEntity.UserAttributes
	}

	updatedAtStr := userEntity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	apiUser.UpdatedAt = &updatedAtStr

	return apiUser
}

// enhanceDetailedView adds additional fields for detailed view
func (h *UserHandler) enhanceDetailedView(user *User, userEntity *iam.User) *User {
	// Add MFA information
	user.MFAEnabled = userEntity.MfaEnabled

	// TODO: Add roles and permissions when available in the model
	// user.Roles = userEntity.Roles
	// user.Permissions = userEntity.Permissions

	return user
}

// enhanceSecurityView adds security-specific fields
func (h *UserHandler) enhanceSecurityView(user *User, userEntity *iam.User) *User {
	// Include security-relevant fields
	user.MFAEnabled = userEntity.MfaEnabled

	// Remove sensitive preferences and metadata for security view
	user.Preferences = nil
	user.Metadata = nil

	return user
}

// limitPublicView removes sensitive fields for public view
func (h *UserHandler) limitPublicView(user *User) *User {
	// Create a minimal public user object
	publicUser := &User{
		ID:       user.ID,
		Username: user.Username,
		UserType: user.UserType,
	}

	return publicUser
}
