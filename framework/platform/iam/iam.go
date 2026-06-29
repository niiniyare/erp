// Package iam provides identity and access management for platform tenants.
package iam

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User status values.
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
)

// User is the domain model for an authenticated principal within a tenant.
type User struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	Email        string         `json:"email"`
	PasswordHash string         `json:"-"`
	Status       string         `json:"status"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Role is a named set of permissions assigned to users.
type Role struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store is the persistence interface for IAM entities.
type Store interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateRole(ctx context.Context, r *Role) error
	ListRoles(ctx context.Context) ([]*Role, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	RevokeRole(ctx context.Context, userID, roleID uuid.UUID) error
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
}
