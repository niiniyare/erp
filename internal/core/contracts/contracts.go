// Package contracts is the bounded-context facade for contract lifecycle management.
//
// External callers (handlers, wire providers) import only this package.
// Internal sub-packages (domain, repository, service) are implementation details.
package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
	"awo.so/internal/core/contracts/service"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
)

// ── Re-exported domain types ──────────────────────────────────────────────────

type (
	Contract              = domain.Contract
	ContractStatus        = domain.ContractStatus
	ContractType          = domain.ContractType
	CreateContractRequest = domain.CreateContractRequest
	UpdateContractRequest = domain.UpdateContractRequest
	ContractFilter        = domain.ContractFilter
	Option                = domain.Option
)

// ── Re-exported status constants ──────────────────────────────────────────────

const (
	StatusDraft           = domain.StatusDraft
	StatusPendingApproval = domain.StatusPendingApproval
	StatusActive          = domain.StatusActive
	StatusExpired         = domain.StatusExpired
	StatusTerminated      = domain.StatusTerminated
)

// ── Re-exported contract type constants ───────────────────────────────────────

const (
	TypeVendor   = domain.TypeVendor
	TypeCustomer = domain.TypeCustomer
	TypeEmployee = domain.TypeEmployee
	TypeService  = domain.TypeService
	TypeLease    = domain.TypeLease
	TypeOther    = domain.TypeOther
)

// ── Re-exported errors ────────────────────────────────────────────────────────

var (
	ErrContractNotFound       = domain.ErrContractNotFound
	ErrContractAlreadyExists  = domain.ErrContractAlreadyExists
	ErrInvalidTransition      = domain.ErrInvalidTransition
	ErrVersionConflict        = domain.ErrVersionConflict
	ErrCannotModifyTerminated = domain.ErrCannotModifyTerminated
	ErrCannotModifyExpired    = domain.ErrCannotModifyExpired
)

// ── Service interface ─────────────────────────────────────────────────────────

// Service is the public contract service interface consumed by handlers and
// other modules.
//
// All methods require an authenticated [iamcontract.SessionContext] in ctx,
// injected by the IAM middleware chain before any route handler runs.
type Service interface {
	// CRUD
	Create(ctx context.Context, req CreateContractRequest) (*Contract, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Contract, error)
	GetByNumber(ctx context.Context, number string) (*Contract, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateContractRequest) (*Contract, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ContractFilter) ([]*Contract, int64, error)

	// Lifecycle transitions
	Submit(ctx context.Context, id uuid.UUID, version int32) error
	Approve(ctx context.Context, id uuid.UUID, version int32) error
	Reject(ctx context.Context, id uuid.UUID, version int32) error
	Terminate(ctx context.Context, id uuid.UUID, reason string, version int32) error
}

// ── Functional options (re-exported for handler convenience) ─────────────────

var (
	WithContractType      = domain.WithContractType
	WithContractValue     = domain.WithContractValue
	WithEndDate           = domain.WithEndDate
	WithCounterpartyEmail = domain.WithCounterpartyEmail
	WithDescription       = domain.WithDescription
	WithTerms             = domain.WithTerms
	WithCreatedBy         = domain.WithCreatedBy
)

// ── Factory ───────────────────────────────────────────────────────────────────

// Dependencies holds everything needed to wire a ContractService.
type Dependencies struct {
	Store  db.Store
	Cache  cache.Service
	Tracer tracing.Service
	Logger logger.Logger
}

// NewService wires and returns a production-ready contract Service.
func NewService(deps Dependencies) Service {
	repo := repository.NewPostgres(deps.Store, deps.Tracer, deps.Logger)
	return service.NewContractService(repo, deps.Cache, deps.Tracer, deps.Logger)
}

// ── Request helpers (convenience constructors) ────────────────────────────────

// NewCreateRequest builds a CreateContractRequest with required fields.
// Optional fields can be added directly to the returned struct.
func NewCreateRequest(
	entityID uuid.UUID,
	title, counterparty, contractType, currencyCode string,
	startDate time.Time,
	value decimal.Decimal,
) CreateContractRequest {
	return CreateContractRequest{
		EntityID:         entityID,
		Title:            title,
		ContractType:     contractType,
		CounterpartyName: counterparty,
		StartDate:        startDate,
		Value:            value,
		CurrencyCode:     currencyCode,
	}
}
