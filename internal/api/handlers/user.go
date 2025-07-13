package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourproject/erp/internal/api/gen/user"
	"github.com/yourproject/erp/internal/core/user"
	"github.com/yourproject/erp/internal/shared/logger"
	"github.com/yourproject/erp/internal/shared/utils"
	"github.com/yourproject/erp/db/sqlc"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *user.UserService
	logger      logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *user.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(ctx context.Context, p *user.CreateUserPayload) (*user.CreateUserResult, error) {
	h.logger.Info("Creating user", "username", p.Username, "email", p.Email)

	// Parse entity ID
	entityID, err := uuid.Parse(p.EntityID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid entity_id: %w", err))
	}

	// Convert payload to user config
	config := user.UserConfig{
		EntityID:      entityID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Email:         p.Email,
		Username:      p.Username,
		Password:      p.Password,
		UserType:      p.UserType,
		SecurityLevel: p.SecurityLevel,
		IsActive:      p.IsActive != nil && *p.IsActive,
	}

	// Set optional fields
	if p.MiddleName != nil {
		config.MiddleName = p.MiddleName
	}
	if p.Phone != nil {
		config.Phone = p.Phone
	}
	if p.Department != nil {
		config.Department = p.Department
	}
	if p.Position != nil {
		config.Position = p.Position
	}
	if p.RoleNames != nil {
		config.RoleNames = p.RoleNames
	}

	// TODO: Get roles by name (this would typically come from a role service)
	rolesByName := make(map[string]uuid.UUID)
	for _, roleName := range config.RoleNames {
		// This is a placeholder - you would query the role service
		rolesByName[roleName] = uuid.New()
	}

	// Create the user
	result, err := h.userService.Step8_CreateUsers(ctx, rolesByName, entityID)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return nil, user.MakeInternalError(err)
	}

	// For now, return the first user created (in a real implementation, this would be different)
	var userID uuid.UUID
	for _, id := range result {
		userID = id
		break
	}

	// Get the created user details
	createdUser, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to retrieve created user", "error", err)
		return nil, user.MakeInternalError(err)
	}

	return &user.CreateUserResult{
		UserID:    createdUser.ID.String(),
		PersonID:  createdUser.PersonID.String(),
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetUser handles user retrieval by ID
func (h *UserHandler) GetUser(ctx context.Context, p *user.GetUserPayload) (*user.GetUserResult, error) {
	h.logger.Info("Getting user", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user", "user_id", p.UserID, "error", err)
		return nil, user.MakeNotFound(err)
	}

	return h.convertUserToResult(u), nil
}

// ListUsers handles user listing with pagination
func (h *UserHandler) ListUsers(ctx context.Context, p *user.ListUsersPayload) (*user.ListUsersResult, error) {
	h.logger.Info("Listing users", "limit", p.Limit, "offset", p.Offset)

	limit := int32(20)
	offset := int32(0)

	if p.Limit != nil {
		limit = *p.Limit
	}
	if p.Offset != nil {
		offset = *p.Offset
	}

	users, err := h.userService.ListUsers(ctx, p.UserType, p.AccountStatus, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		return nil, user.MakeInternalError(err)
	}

	userResults := make([]*user.UserResult, len(users))
	for i, u := range users {
		userResults[i] = h.convertUserToUserResult(u)
	}

	return &user.ListUsersResult{
		Users:  userResults,
		Total:  int32(len(users)), // In a real implementation, get actual count
		Limit:  limit,
		Offset: offset,
	}, nil
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(ctx context.Context, p *user.UpdateUserPayload) (*user.UserResult, error) {
	h.logger.Info("Updating user", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	// Build update parameters
	params := sqlc.UpdateUserParams{
		ID: userID,
	}

	if p.Username != nil {
		params.Username = p.Username
	}
	if p.Email != nil {
		params.Email = *p.Email
	}
	if p.UserType != nil {
		params.UserType = *p.UserType
	}
	if p.AccountStatus != nil {
		params.AccountStatus = p.AccountStatus
	}
	if p.MfaEnabled != nil {
		params.MfaEnabled = p.MfaEnabled
	}
	if p.SessionTimeoutMinutes != nil {
		params.SessionTimeoutMinutes = p.SessionTimeoutMinutes
	}

	updatedUser, err := h.userService.UpdateUser(ctx, userID, params)
	if err != nil {
		h.logger.Error("Failed to update user", "user_id", p.UserID, "error", err)
		return nil, user.MakeInternalError(err)
	}

	return h.convertUserToUserResult(updatedUser), nil
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
func (h *UserHandler) DeleteUser(ctx context.Context, p *user.DeleteUserPayload) (*user.DeleteUserResult, error) {
	h.logger.Info("Deleting user", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	err = h.userService.SoftDeleteUser(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to delete user", "user_id", p.UserID, "error", err)
		return nil, user.MakeInternalError(err)
	}

	return &user.DeleteUserResult{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// AuthenticateUser handles user authentication
func (h *UserHandler) AuthenticateUser(ctx context.Context, p *user.AuthenticateUserPayload) (*user.AuthenticateUserResult, error) {
	h.logger.Info("Authenticating user", "identifier", p.Identifier)

	authenticatedUser, err := h.userService.AuthenticateUser(ctx, p.Identifier, p.Password)
	if err != nil {
		h.logger.Error("Authentication failed", "identifier", p.Identifier, "error", err)
		if strings.Contains(err.Error(), "locked") {
			return nil, user.MakeAccountLocked(err)
		}
		return nil, user.MakeUnauthorized(err)
	}

	// TODO: Generate JWT token (this would be handled by an auth service)
	token := "jwt-token-placeholder"
	expiresAt := time.Now().Add(8 * time.Hour)

	return &user.AuthenticateUserResult{
		User:      h.convertUserToUserResult(authenticatedUser),
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

// GetUserRoles handles retrieving user roles
func (h *UserHandler) GetUserRoles(ctx context.Context, p *user.GetUserRolesPayload) (*user.GetUserRolesResult, error) {
	h.logger.Info("Getting user roles", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	roles, err := h.userService.GetUserRoles(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user roles", "user_id", p.UserID, "error", err)
		return nil, user.MakeInternalError(err)
	}

	roleResults := make([]*user.UserRoleResult, len(roles))
	for i, role := range roles {
		roleResults[i] = &user.UserRoleResult{
			ID:          role.ID.String(),
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			RoleType:    role.RoleType,
			AssignedAt:  role.AssignedAt.Format(time.RFC3339),
		}
		
		if role.ExpiresAt != nil {
			roleResults[i].ExpiresAt = role.ExpiresAt.Format(time.RFC3339)
		}
	}

	return &user.GetUserRolesResult{
		Roles: roleResults,
	}, nil
}

// GetUserPermissions handles retrieving user permissions
func (h *UserHandler) GetUserPermissions(ctx context.Context, p *user.GetUserPermissionsPayload) (*user.GetUserPermissionsResult, error) {
	h.logger.Info("Getting user permissions", "user_id", p.UserID)

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, user.MakeBadRequest(fmt.Errorf("invalid user_id: %w", err))
	}

	permissions, err := h.userService.GetUserEffectivePermissions(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions", "user_id", p.UserID, "error", err)
		return nil, user.MakeInternalError(err)
	}

	permissionResults := make([]*user.UserPermissionResult, len(permissions))
	for i, perm := range permissions {
		permissionResults[i] = &user.UserPermissionResult{
			ID:               perm.ID.String(),
			Name:             perm.Name,
			DisplayName:      perm.DisplayName,
			ResourceName:     perm.ResourceName,
			ActionName:       perm.ActionName,
			PermissionSource: perm.PermissionSource,
		}
		
		if perm.SourceRole != nil {
			permissionResults[i].SourceRole = *perm.SourceRole
		}
	}

	return &user.GetUserPermissionsResult{
		Permissions: permissionResults,
	}, nil
}

// SearchUsers handles user search
func (h *UserHandler) SearchUsers(ctx context.Context, p *user.SearchUsersPayload) (*user.SearchUsersResult, error) {
	h.logger.Info("Searching users", "query", p.Query)

	limit := int32(20)
	offset := int32(0)

	if p.Limit != nil {
		limit = *p.Limit
	}
	if p.Offset != nil {
		offset = *p.Offset
	}

	// TODO: Implement proper search in the service
	// For now, this is a placeholder
	users := []*user.UserSearchResult{}

	return &user.SearchUsersResult{
		Users:  users,
		Total:  int32(len(users)),
		Limit:  limit,
		Offset: offset,
	}, nil
}

// Helper functions

// convertUserToResult converts a sqlc.User to user.GetUserResult
func (h *UserHandler) convertUserToResult(u *sqlc.User) *user.GetUserResult {
	result := &user.GetUserResult{
		ID:        u.ID.String(),
		EntityID:  u.EntityID.String(),
		Username:  u.Username,
		Email:     u.Email,
		UserType:  u.UserType,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}

	if u.PersonID != nil {
		result.PersonID = u.PersonID.String()
	}
	if u.EmployeeID != nil {
		result.EmployeeID = u.EmployeeID.String()
	}
	if u.AccountStatus != nil {
		result.AccountStatus = *u.AccountStatus
	}
	if u.LastLoginAt != nil {
		result.LastLoginAt = u.LastLoginAt.Format(time.RFC3339)
	}
	if u.MfaEnabled != nil {
		result.MfaEnabled = *u.MfaEnabled
	}

	return result
}

// convertUserToUserResult converts a sqlc.User to user.UserResult
func (h *UserHandler) convertUserToUserResult(u *sqlc.User) *user.UserResult {
	result := &user.UserResult{
		ID:        u.ID.String(),
		EntityID:  u.EntityID.String(),
		Username:  u.Username,
		Email:     u.Email,
		UserType:  u.UserType,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}

	if u.PersonID != nil {
		result.PersonID = u.PersonID.String()
	}
	if u.EmployeeID != nil {
		result.EmployeeID = u.EmployeeID.String()
	}
	if u.AccountStatus != nil {
		result.AccountStatus = *u.AccountStatus
	}
	if u.LastLoginAt != nil {
		result.LastLoginAt = u.LastLoginAt.Format(time.RFC3339)
	}
	if u.MfaEnabled != nil {
		result.MfaEnabled = *u.MfaEnabled
	}

	return result
}