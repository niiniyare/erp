package audit

import (
	"context"

	"github.com/niiniyare/erp/internal/shared/token"
)

type service struct {
	repo Repository
}

// NewService creates a new audit service.
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Record logs an audit event.
func (s *service) Record(ctx context.Context, event AuditEvent) error {
	authPayload := ctx.Value(token.AuthorizationPayloadKey).(*token.Payload)
	event.UserID = authPayload.UserID

	return s.repo.CreateAuditEvent(ctx, event)
}
