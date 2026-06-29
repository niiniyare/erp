package repository

import (
	"context"

	"awo.so/internal/core/contracts/domain"
	"github.com/google/uuid"
)

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

// Repository defines all data operations for the contracts module.
type Repository interface {
	Create(ctx context.Context, c *domain.Contract) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Contract, error)
	GetByNumber(ctx context.Context, number string) (*domain.Contract, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateContractRequest) (*domain.Contract, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ContractFilter) ([]*domain.Contract, int64, error)
	NumberExists(ctx context.Context, number string) (bool, error)
}
