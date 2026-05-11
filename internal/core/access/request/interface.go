//go:build ignore

package request

import (
	"context"

	"github.com/google/uuid"
)

// UserService defines the interface for user-related operations needed by the access request service.
type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
}
