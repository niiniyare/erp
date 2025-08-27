package policy

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/iam/model"
)

type Repository interface {
	GetPolicy(ctx context.Context, policyID uuid.UUID) (*model.Policy, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

